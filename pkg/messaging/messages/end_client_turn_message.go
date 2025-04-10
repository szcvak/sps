package messages

import (
	"fmt"
	"os"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
	"github.com/szcvak/sps/pkg/messaging"
)

type EndClientTurnMessage struct {
	amount core.VInt
	id core.VInt
}

func NewEndClientTurnMessage() *EndClientTurnMessage {
	return &EndClientTurnMessage{}
}

func (e *EndClientTurnMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)
	defer stream.Close()

	_, _ = stream.ReadBool()

	_, _ = stream.ReadVInt()
	_, _ = stream.ReadVInt()

	e.amount, _ = stream.ReadVInt()
	e.id, _ = stream.ReadVInt()
}

func (e *EndClientTurnMessage) Process(wrapper *core.ClientWrapper, _ *database.Manager) {
	_, exists := messaging.ClientCommands[int(e.id)]

	if !exists {
		_, _ = fmt.Fprintf(os.Stderr, "client command %d does not exist\n", e.id)
		return
	}

	fmt.Printf("received client command: %d (%d amount)\n", e.id, e.amount)

	//msg := factory()
	//msg.Unmarshal(
}
