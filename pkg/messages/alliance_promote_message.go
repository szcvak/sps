package messages

import (
	"github.com/szcvak/sps/pkg/hub"
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
	
	"log/slog"
	"context"
)

type AlliancePromoteMessage struct {
	highId int32
	lowId int32
	role core.VInt
}

func NewAlliancePromoteMessage() *AlliancePromoteMessage {
	return &AlliancePromoteMessage{}
}

func (a *AlliancePromoteMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)
	defer stream.Close()
	
	a.highId, _ = stream.ReadInt()
	a.lowId, _ = stream.ReadInt()
	a.role, _ = stream.ReadVInt()
}

//10=sentings saved,20=band created,22=name not accepted,40=band joined,42=band is full,43=ban is not open,45=not enough trophies,46=banned,47=band join limit,48=cant join in another region,50=request sent,51=request not sent,no longer open for requests,52=pending request,53=request rejected too low trophies,54=request rejected, banned,55=request rejected because of region limit,70=kick success,71=elders have kick cooldown,80=you left the band,81=promote success,82=demote success,90=accept success,91=reject success,92=not accepted because player is already in a band,93=cannot find request,94=request already handled,95=handle failed,no rights,96=accept failed, target has been banned,100=you have been kicked,101=you have been promoted,102=you have been demoted,

func (a *AlliancePromoteMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.AllianceId == nil {
		return
	}
	
	if wrapper.Player.AllianceRole != 2 && wrapper.Player.AllianceRole != 4 {
		return
	}
	
	if a.role > 4 || a.role < 1 {
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
	
	isDemoted := int32(0)
	
	if (int16(a.role) < player.AllianceRole && int32(a.role) != 2) || (player.AllianceRole == 2 && a.role != 2) {
		isDemoted = int32(1)
	}
	
	err = dbm.Exec("update alliance_members set role=$1 where player_id=$2 and alliance_id=$3", int16(a.role), playerId, *wrapper.Player.AllianceId)
	
	if err != nil {
		slog.Error("failed to update alliance member!", "err", err)
		return
	}
	
	if wr != nil {
		wr.Player.AllianceRole = int16(a.role)
		
		msg := NewAllianceResponseMessage(101 + isDemoted)
		wr.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
	}
	
	if a.role == 2 {
		err = dbm.Exec("update alliance_members set role=$1 where player_id=$2 and alliance_id=$3", int16(4), wrapper.Player.DbId, *wrapper.Player.AllianceId)
	
		if err != nil {
			slog.Error("failed to update alliance member!", "err", err)
			return
		}
		
		wrapper.Player.AllianceRole = 4
	}
	
	msg2 := NewAllianceResponseMessage(81 + isDemoted)
	wrapper.Send(msg2.PacketId(), msg2.PacketVersion(), msg2.Marshal())
	
	message, err := dbm.AddAllianceMessage(context.Background(), *wrapper.Player.AllianceId, wrapper.Player, 45 + int16(isDemoted), "", player)

	if err != nil {
		slog.Error("failed to send alliance message!", "err", err)
		return
	}
	
	serverMsg := NewAllianceChatServerMessage(*message, *wrapper.Player.AllianceId)
	h.BroadcastToAlliance(*wrapper.Player.AllianceId, serverMsg)
	
	if a.role == 2 {
		message, err = dbm.AddAllianceMessage(context.Background(), *wrapper.Player.AllianceId, wrapper.Player, 46, "", wrapper.Player)

		if err != nil {
			slog.Error("failed to send alliance message!", "err", err)
			return
		}
	
		serverMsg = NewAllianceChatServerMessage(*message, *wrapper.Player.AllianceId)
		h.BroadcastToAlliance(*wrapper.Player.AllianceId, serverMsg)
	}
	
	msg3 := NewAllianceDataMessage(wrapper, dbm, *wrapper.Player.AllianceId)
	wrapper.Send(msg3.PacketId(), msg3.PacketVersion(), msg3.Marshal())
}
