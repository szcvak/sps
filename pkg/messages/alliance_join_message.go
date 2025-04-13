package messages

import (
	"context"
	"log/slog"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
	"github.com/szcvak/sps/pkg/hub"
)

type AllianceJoinMessage struct {
	highId     int32
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

	err := dbm.AddAllianceMember(context.Background(), wrapper.Player, int64(a.allianceId))

	if err != nil {
		slog.Error("failed to add alliance member!", "err", err)
		return
	}

	h := hub.GetHub()
	h.UpdateAllianceMembership(wrapper, nil, wrapper.Player.AllianceId)

	message, err := dbm.AddAllianceMessage(context.Background(), *wrapper.Player.AllianceId, wrapper.Player, 43, "", wrapper.Player)

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

	h.BroadcastToAlliance(*wrapper.Player.AllianceId, serverMsg)
}
