package messages

import (
	"log/slog"
	
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type TeamCreateMessage struct {
	teamType core.VInt
	event    core.VInt
}

func NewTeamCreateMessage() *TeamCreateMessage {
	return &TeamCreateMessage{}
}

func (t *TeamCreateMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)
	
	t.teamType, _ = stream.ReadVInt()
	t.event, _ = stream.ReadVInt()
}

func (t *TeamCreateMessage) Process(wrapper *core.ClientWrapper, _ *database.Manager) {
	if t.event < 1 || t.event > 4 {
		return
	}
	
	em := core.GetEventManager()
	event := em.GetCurrentEvent(int32(t.event))
	
	if event.Config.Gamemode == core.GameModeShowdown {
		return
	}
	
	slog.Info("created team", "type", t.teamType, "event", t.event)
}
