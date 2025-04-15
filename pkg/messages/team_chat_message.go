package messages

import (
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type TeamChatMessage struct {
	message string
}

func NewTeamChatMessage() *TeamChatMessage {
	return &TeamChatMessage{}
}

func (t *TeamChatMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)
	defer stream.Close()

	t.message, _ = stream.ReadString()
}

func (t *TeamChatMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.TeamId == nil {
		return
	}
	
	tm := core.GetTeamManager()
	tm.AddMessage(wrapper.Player, t.message)

	for _, member := range tm.Teams[*wrapper.Player.TeamId].Members {
		if member.Wrapper != nil {
			msg := NewTeamStreamMessage(member.Wrapper, false)
			payload := msg.Marshal()
			
			member.Wrapper.Send(msg.PacketId(), msg.PacketVersion(), payload)
		}
	}
}
