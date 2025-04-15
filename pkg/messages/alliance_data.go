package messages

import (
	"context"
	"log/slog"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type AllianceDataMessage struct {
	wrapper *core.ClientWrapper
	dbm     *database.Manager
	id      int64
}

func NewAllianceDataMessage(wrapper *core.ClientWrapper, dbm *database.Manager, id int64) *AllianceDataMessage {
	return &AllianceDataMessage{
		wrapper: wrapper,
		id:      id,
		dbm:     dbm,
	}
}

func (a *AllianceDataMessage) PacketId() uint16 {
	return 24301
}

func (a *AllianceDataMessage) PacketVersion() uint16 {
	return 1
}

func (a *AllianceDataMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(32)

	alliance, err := a.dbm.LoadAlliance(context.Background(), a.id)

	if err != nil {
		slog.Error("failed to load alliance data!", "err", err)
		return stream.Buffer()
	}

	stream.Write(0)
	stream.Write(int32(alliance.Id))

	stream.Write(alliance.Name)
	stream.Write(core.DataRef{8, alliance.BadgeId})

	stream.Write(core.VInt(alliance.Type))
	stream.Write(core.VInt(alliance.TotalMembers))
	stream.Write(core.VInt(alliance.TotalTrophies))
	stream.Write(core.VInt(alliance.RequiredTrophies))

	stream.Write(core.DataRef{0, 1})

	stream.Write(alliance.Description)
	
	stream.Write(core.VInt(alliance.TotalMembers))

	for _, member := range alliance.Members {
		stream.Write(member.HighId)
		stream.Write(member.LowId)

		stream.Write(member.Name)
		stream.Write(core.VInt(member.Role))

		level := -1

		for i := 0; i < len(core.RequiredExp)-1; i++ {
			if core.RequiredExp[i] <= 0 && 0 < core.RequiredExp[i+1] {
				level = i + 1
				break
			}
		}

		if level == -1 {
			level = len(core.RequiredExp)
		}

		stream.Write(core.VInt(level))
		stream.Write(core.VInt(member.Trophies))
		stream.Write(core.DataRef{28, member.ProfileIcon})
	}

	return stream.Buffer()
}
