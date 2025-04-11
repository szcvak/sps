package messages

import (
	"github.com/szcvak/sps/pkg/messaging"
	"log/slog"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type EndClientTurnMessage struct {
	amount core.VInt
	id     core.VInt

	stream *core.ByteStream

	unmarshalled bool
}

func NewEndClientTurnMessage() *EndClientTurnMessage {
	return &EndClientTurnMessage{
		unmarshalled: false,
	}
}

func (e *EndClientTurnMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)

	_, _ = stream.ReadBool()

	for i := 0; i < 3; i++ {
		_, _ = stream.ReadVInt()
	}

	e.amount, _ = stream.ReadVInt()
	e.id, _ = stream.ReadVInt()

	e.unmarshalled = true
	e.stream = stream
}

func (e *EndClientTurnMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if !e.unmarshalled {
		return
	}

	if wrapper.Player.DbId <= 0 || wrapper.Player.State() != core.StateLoggedIn {
		return
	}

	factory, exists := messaging.ClientCommands[int(e.id)]

	if !exists {
		slog.Warn("got unknown client command", "id", e.id, "amt", e.amount)
		return
	}

	slog.Info("got client command", "id", e.id, "amt", e.amount)

	msg := factory()
	msg.UnmarshalStream(e.stream)
	msg.Process(wrapper, dbm)
}
