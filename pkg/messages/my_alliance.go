package messages

import (
	"context"
	"log/slog"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type MyAllianceMessage struct {
	wrapper *core.ClientWrapper
	dbm *database.Manager
}

func NewMyAllianceMessage(wrapper *core.ClientWrapper, dbm *database.Manager) *MyAllianceMessage {
	return &MyAllianceMessage {
		wrapper: wrapper,
		dbm: dbm,
	}
}

func (m *MyAllianceMessage) PacketId() uint16 {
	return 24399
}

func (m *MyAllianceMessage) PacketVersion() uint16 {
	return 1
}

func (m *MyAllianceMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(4)

	if m.wrapper.Player.AllianceId == nil {
		stream.Write(core.VInt(0))
		stream.Write(false)

		return stream.Buffer()
	}
	
	a, err := m.dbm.LoadAlliance(context.Background(), *m.wrapper.Player.AllianceId)
	
	if err != nil {
		slog.Error("failed to load alliance!", "err", err)
		return stream.Buffer()
	}
	
	stream.Write(core.VInt(1)) // online members
	stream.Write(true)
	stream.Write(core.DataRef{25, int32(m.wrapper.Player.AllianceRole)})
	stream.Write(0)
	stream.Write(int32(a.Id))
	stream.Write(a.Name)
	stream.Write(core.DataRef{8, a.BadgeId})
	stream.Write(core.VInt(a.Type))
	stream.Write(core.VInt(1)) // online members
	stream.Write(core.VInt(a.TotalTrophies))
	stream.Write(core.DataRef{0, 1})
	stream.Write(core.VInt(a.TotalMembers))
	
	return stream.Buffer()
}
