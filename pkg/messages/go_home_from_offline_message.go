package messages

import (
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
	"log/slog"
)

type GoHomeFromOfflineMessage struct{}

func NewGoHomeFromOfflineMessage() *GoHomeFromOfflineMessage {
	return &GoHomeFromOfflineMessage{}
}

func (g *GoHomeFromOfflineMessage) Unmarshal(_ []byte) {}

func (g *GoHomeFromOfflineMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.State() != core.StateLoggedIn {
		return
	}

	if wrapper.Player.TutorialState != 2 {
		wrapper.Player.TutorialState++

		if err := dbm.Exec("update players set tutorial_state = $1 where id = $2", wrapper.Player.TutorialState, wrapper.Player.DbId); err != nil {
			slog.Error("failed to update tutorial_state!", "playerId", wrapper.Player.DbId, "err", err)
		}
	}

	msg := NewOwnHomeDataMessage(wrapper, dbm)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
