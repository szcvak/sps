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
	registerClientMessage(10212, func() messaging.ClientMessage { return messages.NewChangeAvatarNameMessage() })
	registerClientMessage(10107, func() messaging.ClientMessage { return messages.NewClientCapabilitiesMessage() })
	registerClientMessage(14102, func() messaging.ClientMessage { return messages.NewEndClientTurnMessage() })
}
