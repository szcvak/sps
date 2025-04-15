package messages

import (
	"log/slog"
	
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type TeamToggleMemberSideMessage struct {
	oldSlot core.VInt
	newSlot core.VInt
}

func NewTeamToggleMemberSideMessage() *TeamToggleMemberSideMessage {
	return &TeamToggleMemberSideMessage{}
}

func (t *TeamToggleMemberSideMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)
	defer stream.Close()
	
	t.oldSlot, _ = stream.ReadVInt()
	t.newSlot, _ = stream.ReadVInt()
}

func (t *TeamToggleMemberSideMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.TeamId == nil {
		return
	}

	slog.Info("HEREEE", "old", t.oldSlot, "new", t.newSlot)
}
