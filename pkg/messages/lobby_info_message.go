package messages

import "github.com/szcvak/sps/pkg/core"

type LobbyInfoMessage struct {
	online int32
}

func NewLobbyInfoMessage(online int32) *LobbyInfoMessage {
	return &LobbyInfoMessage {
		online: online,
	}
}

func (l *LobbyInfoMessage) PacketId() uint16 {
	return 23457
}

func (l *LobbyInfoMessage) PacketVersion() uint16 {
	return 1
}

func (l *LobbyInfoMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(32)
	
	stream.Write(core.VInt(l.online))
	
	stream.Write(core.VInt(0))
	
	return stream.Buffer()
}
