package messages

import (
	"context"
	"log/slog"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
	"github.com/szcvak/sps/pkg/hub"
)

type AllianceCreateMessage struct {
	name             string
	description      string
	badge            core.DataRef
	allianceType     core.VInt
	requiredTrophies core.VInt
}

func NewAllianceCreateMessage() *AllianceCreateMessage {
	return &AllianceCreateMessage{}
}

func (a *AllianceCreateMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)

	a.name, _ = stream.ReadString()
	a.description, _ = stream.ReadString()
	a.badge, _ = stream.ReadDataRef()
	a.allianceType, _ = stream.ReadVInt()
	a.requiredTrophies, _ = stream.ReadVInt()
}

func (a *AllianceCreateMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.AllianceId != nil {
		return
	}

	err := dbm.CreateAlliance(context.Background(), a.name, a.description, a.badge.S, int32(a.allianceType), int32(a.requiredTrophies), wrapper.Player)

	if err != nil {
		slog.Error("failed to create alliance!", "err", err)
		return
	}

	h := hub.GetHub()
	h.UpdateAllianceMembership(wrapper, nil, wrapper.Player.AllianceId)

	msg := NewAllianceResponseMessage(20)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())

	msg2 := NewMyAllianceMessage(wrapper, dbm)
	wrapper.Send(msg2.PacketId(), msg2.PacketVersion(), msg2.Marshal())

	msg3 := NewClanStreamMessage(wrapper, dbm)
	wrapper.Send(msg3.PacketId(), msg3.PacketVersion(), msg3.Marshal())
}
