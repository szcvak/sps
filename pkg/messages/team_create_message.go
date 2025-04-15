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
	defer stream.Close()

	t.teamType, _ = stream.ReadVInt()
	t.event, _ = stream.ReadVInt()
}

func (t *TeamCreateMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if t.event < 1 || t.event > 4 {
		slog.Error("invalid event!", "event", t.event)
		return
	}

	if wrapper.Player.TeamId != nil {
		slog.Error("can't create team, player is already in a team", "playerId", wrapper.Player.DbId, "teamId", *wrapper.Player.TeamId)
		return
	}

	tm := core.GetTeamManager()
	tm.CreateTeam(wrapper, int32(t.event))
	
	if wrapper.Player.TeamId == nil {
		slog.Error("failed to create team!")
		return
	}
	
	slog.Info("created team", "type", t.teamType, "event", t.event, "teamId", *wrapper.Player.TeamId)

	msg := NewTeamMessage(wrapper)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
