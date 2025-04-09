package messages

import (
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type KeepAliveMessage struct{}

func NewKeepAliveMessage() *KeepAliveMessage {
	return &KeepAliveMessage{}
}

func (l *KeepAliveMessage) Unmarshal(_ []byte) {}

func (l *KeepAliveMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	msg := NewKeepAliveOkMessage()
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
