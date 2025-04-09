package network

import (
	"github.com/szcvak/sps/pkg/messaging"
	"github.com/szcvak/sps/pkg/messaging/messages"
)

var ClientRegistry = make(map[uint16]func() messaging.ClientMessage)

func registerClientMessage(id uint16, factory func() messaging.ClientMessage) {
	ClientRegistry[id] = factory
}

func init() {
	registerClientMessage(10101, func() messaging.ClientMessage { return messages.NewLoginMessage() })
	registerClientMessage(10108, func() messaging.ClientMessage { return messages.NewKeepAliveMessage() })
}
