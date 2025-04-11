package messages

import (
	"fmt"
	"github.com/szcvak/sps/pkg/messaging"
	"os"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type ChangeAvatarNameMessage struct {
	name string
}

func NewChangeAvatarNameMessage() *ChangeAvatarNameMessage {
	return &ChangeAvatarNameMessage{}
}

func (c *ChangeAvatarNameMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)
	defer stream.Close()

	c.name, _ = stream.ReadString()
}

func (c *ChangeAvatarNameMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.State() != core.StateLoggedIn {
		return
	}

	if wrapper.Player.DbId <= 0 {
		return
	}

	err := dbm.Exec("update players set Name = $1 where id = $2", c.name, wrapper.Player.DbId)

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to change player's Name: %w\n", err)
		return
	}

	fmt.Printf("changed player %d's Name from %s to %s\n", wrapper.Player.DbId, wrapper.Player.Name, c.name)

	wrapper.Player.Name = c.name

	msg := messaging.NewAvailableServerCommandMessage(201, c.name)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
