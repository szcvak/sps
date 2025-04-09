package messaging

import "github.com/szcvak/sps/pkg/core"

type ClientMessage interface {
	Unmarshal(payload []byte)
	Process(player *core.ClientWrapper)
}

type ServerMessage interface {
	Marshal() []byte
	PacketId() uint16
	PacketVersion() uint16
}
