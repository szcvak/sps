package messages

import (
	"context"
	"log/slog"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type JoinableAlliancesMessage struct {
	dbm *database.Manager
}

func NewJoinableAlliancesMessage(dbm *database.Manager) *JoinableAlliancesMessage {
	return &JoinableAlliancesMessage {
		dbm: dbm,
	}
}

func (j *JoinableAlliancesMessage) PacketId() uint16 {
	return 24304
}

func (j *JoinableAlliancesMessage) PacketVersion() uint16 {
	return 1
}

func (j *JoinableAlliancesMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(16)
	
	alliances, err := j.dbm.GetAlliances(context.Background())
	
	if err != nil {
		slog.Error("failed to get alliances!", "err", err)
		return stream.Buffer()
	}

	stream.Write(core.VInt(len(alliances)))
	

	for _, a := range alliances {
		stream.Write(0)
		stream.Write(int32(a.Id))
		
		stream.Write(a.Name)
		
		stream.Write(core.VInt(8))
		stream.Write(core.VInt(a.BadgeId))
		
		stream.Write(core.VInt(a.Type))
		stream.Write(core.VInt(a.TotalMembers))
		stream.Write(core.VInt(a.TotalTrophies))
		stream.Write(core.VInt(a.RequiredTrophies))
		
		stream.Write(core.VInt(0))
	}
	
	return stream.Buffer()
}
