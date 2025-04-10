package messages

import (
	"fmt"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type ClientCapabilitiesMessage struct{}

func NewClientCapabilitiesMessage() *ClientCapabilitiesMessage {
	return &ClientCapabilitiesMessage{}
}

func (c *ClientCapabilitiesMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)
	defer stream.Close()

	ping, _ := stream.ReadVInt()

	fmt.Println("client's ping:", ping)
}

func (c *ClientCapabilitiesMessage) Process(_ *core.ClientWrapper, _ *database.Manager) {}
