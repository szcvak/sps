package messages

import (
	"context"
	"log/slog"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type AskProfileMessage struct {
	highId int32
	lowId int32
}

func NewAskProfileMessage() *AskProfileMessage {
	return &AskProfileMessage{}
}

func (a *AskProfileMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)
	defer stream.Close()

	a.highId, _ = stream.ReadInt()
	a.lowId, _ = stream.ReadInt()
}

func (a *AskProfileMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.State() != core.StateLoggedIn {
		return
	}

	player, err := dbm.LoadPlayerByIds(context.Background(), a.highId, a.lowId)
	
	if err != nil {
		slog.Error("failed to find player by ids!", "err", err)
		return
	}
	
	msg := NewProfileMessage(player)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
