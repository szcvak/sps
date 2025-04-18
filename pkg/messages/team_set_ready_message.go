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
	if wrapper.Player.TeamId == nil {
		return
	}
	
	slog.Info("changed team ready status", "playerId", wrapper.Player.DbId, "isReady", t.ready)

	tm := core.GetTeamManager()
	tm.UpdateReady(wrapper.Player, t.ready)

	for _, member := range tm.Teams[*wrapper.Player.TeamId].Members {
		if member.Wrapper != nil {
			msg := NewTeamMessage(member.Wrapper)
			payload := msg.Marshal()
		
			member.Wrapper.Send(msg.PacketId(), msg.PacketVersion(), payload)
		}
	}
}
