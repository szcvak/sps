package messages

import (
	"time"
	
	"github.com/szcvak/sps/pkg/core"
)

type TeamStreamMessage struct {
	wrapper *core.ClientWrapper
	initial bool
}

func NewTeamStreamMessage(wrapper *core.ClientWrapper, initial bool) *TeamStreamMessage {
	return &TeamStreamMessage {
		wrapper: wrapper,
		initial: initial,
	}
}

func (t *TeamStreamMessage) PacketId() uint16 {
	return 24131
}

func (t *TeamStreamMessage) PacketVersion() uint16 {
	return 1
}

func (t *TeamStreamMessage) Marshal() []byte {
	if t.wrapper.Player.TeamId == nil {
		return []byte{}
	}
	
	id := *t.wrapper.Player.TeamId
	
	tm := core.GetTeamManager()
	team := tm.Teams[id]
	
	if team == nil {
		return []byte{}
	}
	
	if len(team.Messages) == 0 {
		return []byte{}
	}
	
	now := time.Now().Unix()
	
	stream := core.NewByteStreamWithCapacity(16)
	
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(id))
	
	// messages
	
	if t.initial {
		stream.Write(core.VInt(len(team.Messages)))
	
		for i, message := range team.Messages {
			role := 0
	
			if team.Creator == message.PlayerId {
				role = 2
			}
			
			stream.Write(core.VInt(message.Type))
			
			stream.Write(core.VInt(0))
			stream.Write(core.VInt(i))
			
			stream.Write(core.VInt(message.PlayerHighId))
			stream.Write(core.VInt(message.PlayerLowId))
			
			stream.Write(message.PlayerName)
			
			stream.Write(core.VInt(role))
			stream.Write(core.VInt(now - message.Timestamp))
			stream.Write(core.VInt(0))
			
			if message.Type != 4 {
				stream.Write(message.Content)
				continue
			}
			
			stream.Write(core.VInt(message.Event))
			
			stream.Write(core.VInt(1))
			stream.Write(core.VInt(0))
			
			stream.Write(core.VInt(message.TargetId))
			stream.Write(message.TargetName)
		}
		
		return stream.Buffer()
	}
	
	message := team.Messages[len(team.Messages)-1]
	
	role := 0
	
	if team.Creator == message.PlayerId {
		role = 2
	}
	
	stream.Write(core.VInt(1))
	
	stream.Write(core.VInt(message.Type))
	
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(len(team.Messages)-1))
		
	stream.Write(core.VInt(message.PlayerHighId))
	stream.Write(core.VInt(message.PlayerLowId))
		
	stream.Write(message.PlayerName)
		
	stream.Write(core.VInt(role))
	stream.Write(core.VInt(now - message.Timestamp))
	stream.Write(core.VInt(0))
	
	if message.Type != 4 {
		stream.Write(message.Content)
		return stream.Buffer()
	}
	
	stream.Write(core.VInt(message.Event))
	
	stream.Write(core.VInt(1))
	stream.Write(core.VInt(0))
	
	stream.Write(core.VInt(message.TargetId))
	stream.Write(message.TargetName)
	
	return stream.Buffer()
}
