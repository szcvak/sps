package messages

import (
	"log/slog"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type AllianceEditMessage struct {
	description      string
	badge            core.DataRef
	allianceType     core.VInt
	requiredTrophies core.VInt
}

func NewAllianceEditMessage() *AllianceEditMessage {
	return &AllianceEditMessage{}
}

func (a *AllianceEditMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)

	a.description, _ = stream.ReadString()
	a.badge, _ = stream.ReadDataRef()
	a.allianceType, _ = stream.ReadVInt()
	a.requiredTrophies, _ = stream.ReadVInt()
}

func (a *AllianceEditMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.AllianceId == nil {
		return
	}

	// TODO: change?
	if wrapper.Player.AllianceRole < 2 {
		return
	}

	err := dbm.Exec(
		"update alliances set description = $1, badge_id = $2, type = $3, required_trophies = $4 where id = $5",
		a.description, a.badge.S, a.allianceType, a.requiredTrophies, *wrapper.Player.AllianceId,
	)

	if err != nil {
		slog.Error("failed to update alliance!", "err", err)
		return
	}

	msg := NewAllianceEventMessage(10)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())

	msg2 := NewMyAllianceMessage(wrapper, dbm)
	wrapper.Send(msg2.PacketId(), msg2.PacketVersion(), msg2.Marshal())
}
