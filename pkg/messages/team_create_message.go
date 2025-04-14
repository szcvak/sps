package messages

import (
	"context"
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
	defer stream.Close()

	t.teamType, _ = stream.ReadVInt()
	t.event, _ = stream.ReadVInt()
}

func (t *TeamCreateMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if t.event < 1 || t.event > 4 {
		return
	}

	if wrapper.Player.TeamCode != nil {
		slog.Info("player is already in team", "playerId", wrapper.Player.DbId, "teamCode", *wrapper.Player.TeamCode)
		return
	}

	em := core.GetEventManager()
	event := em.GetCurrentEvent(int32(t.event - 1))

	if event.Config.Gamemode == core.GameModeShowdown {
		return
	}

	code := core.GenerateTeamCode()
	err := dbm.CreateTeam(context.Background(), code, wrapper.Player)

	if err != nil {
		slog.Error("failed to create team!", "err", err)
		return
	}

	slog.Info("created team", "type", t.teamType, "event", t.event, "code", code)

	msg := NewTeamMessage(wrapper, dbm, int32(t.event))
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
