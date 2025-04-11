package messages

import (
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type GoHomeFromOfflineMessage struct{}

func NewGoHomeFromOfflineMessage() *GoHomeFromOfflineMessage {
	return &GoHomeFromOfflineMessage{}
}

func (g *GoHomeFromOfflineMessage) Unmarshal(_ []byte) {}

func (g *GoHomeFromOfflineMessage) Process(wrapper *core.ClientWrapper, _ *database.Manager) {
	msg := NewOwnHomeDataMessage(wrapper)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
