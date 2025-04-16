package messages

import (
	"context"
	"log/slog"

	"github.com/szcvak/sps/pkg/hub"
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type AllianceKickMessage struct {
	highId int32
	lowId int32
	reason string
}

func NewAllianceKickMessage() *AllianceKickMessage {
	return &AllianceKickMessage{}
}

func (a *AllianceKickMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)
	defer stream.Close()

	a.highId, _ = stream.ReadInt()
	a.lowId, _ = stream.ReadInt()
	a.reason, _ = stream.ReadString()
}

func (a *AllianceKickMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.AllianceId == nil {
		return
	}
	
	if wrapper.Player.AllianceRole == 1 {
		return
	}
	
	var wr *core.ClientWrapper
	var player *core.Player
	var err error
	
	playerId := -1
	
	h := hub.GetHub()
	i := -1
	
	for w, _ := range h.ClientsByAID[*wrapper.Player.AllianceId] {
		i++
		
		if w.Player == nil {
			continue
		}
		
		if w.Player.HighId == int32(a.highId) && w.Player.LowId == int32(a.lowId) {
			if w.Player.AllianceId == nil {
				return
			}
			
			playerId = i
			wr = w
			player = wr.Player
			
			break
		}
	}
	
	if playerId == -1 {
		player, err = dbm.LoadPlayerByIds(context.Background(), int32(a.highId), int32(a.lowId))
	
		if err != nil {
			slog.Error("failed to fetch player!", "err", err)
			return
		}
		
		playerId = int(player.DbId)
	}
	
	if player == nil {
		return
	}
	
	if player.AllianceRole == 2 {
		return
	}
	
	if player.AllianceRole > wrapper.Player.AllianceRole {
		return
	}
	
	err = dbm.Exec("delete from alliance_members where player_id=$1", playerId)
	
	if err != nil {
		slog.Error("failed to delete player!", "err", err)
		return
	}
	
	if wr != nil {
		h := hub.GetHub()
		h.UpdateAllianceMembership(wr, nil, wrapper.Player.AllianceId)
		
		msg := NewAllianceResponseMessage(100)
		wr.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
	}
	
	player.AllianceId = nil
	player.AllianceRole = 0
	
	message, err := dbm.AddAllianceMessage(context.Background(), *wrapper.Player.AllianceId, wrapper.Player, 41, "", player)

	if err != nil {
		slog.Error("failed to send alliance message!", "err", err)
		return
	}

	serverMsg := NewAllianceChatServerMessage(*message, *wrapper.Player.AllianceId)
	h.BroadcastToAlliance(*wrapper.Player.AllianceId, serverMsg)
	
	serverMsg2 := NewMyAllianceMessage(wrapper, dbm)
	h.BroadcastToAlliance(*wrapper.Player.AllianceId, serverMsg2)
	
	msg := NewAllianceResponseMessage(70)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
