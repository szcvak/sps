package messages

import (
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
	"log/slog"
)

type TeamSetReadyMessage struct {
	ready bool
}

func NewTeamSetReadyMessage() *TeamSetReadyMessage {
	return &TeamSetReadyMessage{}
}

func (t *TeamSetReadyMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)
	defer stream.Close()

	t.ready, _ = stream.ReadBool()
	_, _ = stream.ReadVInt()
}

func (t *TeamSetReadyMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	slog.Info("changed ready status", "playerId", wrapper.Player.DbId, "isReady", t.ready)

	wrapper.Player.TeamIsReady = t.ready

	err := dbm.Exec("update team_members set is_ready = $1 where id = $2", t.ready, wrapper.Player.DbId)

	if err != nil {
		slog.Error("failed to update player's ready status!", "err", err)
		return
	}

	msg := NewTeamMessage(wrapper, dbm, 2)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
