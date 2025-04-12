package messages

import (
	"context"
	"log/slog"

	"github.com/szcvak/sps/pkg/hub"
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type AllianceChatMessage struct {
	content string
}

func NewAllianceChatMessage() *AllianceChatMessage {
	return &AllianceChatMessage{}
}

func (a *AllianceChatMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)
	defer stream.Close()

	a.content, _ = stream.ReadString()
}

func (a *AllianceChatMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.State() != core.StateLoggedIn {
		return
	}
	
	if wrapper.Player.AllianceId == nil {
		return
	}
	
	message, err := dbm.AddAllianceMessage(context.Background(), *wrapper.Player.AllianceId, wrapper.Player, 2, a.content, nil)
	
	if err != nil {
		slog.Error("failed to add alliance message!", "err", err)
		return
	}
	
	msg := NewAllianceChatServerMessage(*message, *wrapper.Player.AllianceId)

	messageHub := hub.GetHub()
	messageHub.BroadcastToAlliance(*wrapper.Player.AllianceId, msg)
}
