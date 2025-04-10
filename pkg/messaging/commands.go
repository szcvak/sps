package messaging

import (
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type ServerCommand interface {
	Marshal(stream *core.ByteStream)
}

type ClientCommand interface {
	Unmarshal(data []byte)
	Process(wrapper *core.ClientWrapper, dbm *database.Manager)
}

var (
	ServerCommands = make(map[int]func(payload interface{}) ServerCommand)
	ClientCommands = make(map[int]func() ClientCommand)
)

func registerServerCommand(id int, factory func(payload interface{}) ServerCommand) {
	ServerCommands[id] = factory
}

func registerClientCommand(id int, factory func() ClientCommand) {
	ClientCommands[id] = factory
}

func init() {
	// --- Server --- //
	ServerCommands[201] = func(payload interface{}) ServerCommand { switch v := payload.(type) { case string: return NewChangeAvatarNameCommand(v); default: return nil }}

	// --- Client --- //
}

// --- Commands --- //

type ChangeAvatarNameCommand struct {
	name string
}

func NewChangeAvatarNameCommand(name string) *ChangeAvatarNameCommand {
	return &ChangeAvatarNameCommand {
		name: name,
	}
}

func (c *ChangeAvatarNameCommand) Marshal(stream *core.ByteStream) {
	stream.Write(core.VInt(201))
	stream.Write(c.name)

	stream.Write(byte(0))

	stream.Write(core.VInt(1))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(1))
}
