package messages

import (
	"context"
	"log/slog"
	
	"github.com/szcvak/sps/pkg/hub"
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type AllianceJoinMessage struct {
	highId int32
	allianceId int32
}

func NewAllianceJoinMessage() *AllianceJoinMessage {
	return &AllianceJoinMessage{}
}

func (a *AllianceJoinMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)
	
	a.highId, _ = stream.ReadInt()
	a.allianceId, _ = stream.ReadInt()
}

func (a *AllianceJoinMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.AllianceId != nil {
		return
	}

	err := dbm.Exec(
		"update alliances set total_trophies = total_trophies + $1 where id = $2",
		wrapper.Player.Trophies, a.allianceId + 1,
	)
	
	if err != nil {
		slog.Error("failed to update alliances!", "err", err)
		return
	}
	
	err = dbm.Exec(
		"insert into alliance_members (alliance_id, player_id) values ($1, $2)",
		a.allianceId + 1, wrapper.Player.DbId,
	)
	
	if err != nil {
		slog.Error("failed to update alliance_members!", "err", err)
		return
	}
	
	wrapper.Player.AllianceRole = 1
	wrapper.Player.AllianceId = new(int64)
	*wrapper.Player.AllianceId = int64(a.allianceId + 1)
	
	message, err := dbm.AddAllianceMessage(context.Background(), *wrapper.Player.AllianceId, wrapper.Player, 4, "", wrapper.Player)
	
	if err != nil {
		slog.Error("failed to send alliance message!", "err", err)
		return
	}
	
	msg := NewAllianceEventMessage(40)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
	
	msg2 := NewMyAllianceMessage(wrapper, dbm)
	wrapper.Send(msg2.PacketId(), msg2.PacketVersion(), msg2.Marshal())
	
	msg3 := NewClanStreamMessage(wrapper, dbm)
	wrapper.Send(msg3.PacketId(), msg3.PacketVersion(), msg3.Marshal())
	
	serverMsg := NewAllianceChatServerMessage(*message, *wrapper.Player.AllianceId)
	
	messageHub := hub.GetHub()
	messageHub.BroadcastToAlliance(*wrapper.Player.AllianceId, serverMsg)
}
