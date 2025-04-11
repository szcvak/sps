package messaging

import (
	"log/slog"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type ServerCommand interface {
	Marshal(stream *core.ByteStream)
}

type ClientCommand interface {
	UnmarshalStream(stream *core.ByteStream)
	Process(wrapper *core.ClientWrapper, dbm *database.Manager)
}

var (
	ServerCommands = make(map[int]func(payload interface{}) ServerCommand)
	ClientCommands = make(map[int]func() ClientCommand)
)

func init() {
	// --- Server --- //
	ServerCommands[201] = func(payload interface{}) ServerCommand {
		switch v := payload.(type) {
		case string:
			return NewChangeAvatarNameCommand(v)
		default:
			return nil
		}
	}

	ServerCommands[203] = func(payload interface{}) ServerCommand {
		switch v := payload.(type) {
		case *DeliveryLogic:
			return v
		}

		return nil
	}

	// --- Client --- //
	ClientCommands[509] = func() ClientCommand { return NewClientSelectControlModeCommand() }
	ClientCommands[500] = func() ClientCommand { return NewClientGatchaCommand() }
}

// --- Server commands --- //

type ChangeAvatarNameCommand struct {
	name string
}

func NewChangeAvatarNameCommand(name string) *ChangeAvatarNameCommand {
	return &ChangeAvatarNameCommand{
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

// --- Client commands --- //

type ClientSelectControlModeCommand struct {
	controlMode core.VInt
}

type ClientGatchaCommand struct {
	boxType core.VInt
}

func NewClientSelectControlModeCommand() *ClientSelectControlModeCommand {
	return &ClientSelectControlModeCommand{}
}

func NewClientGatchaCommand() *ClientGatchaCommand {
	return &ClientGatchaCommand{}
}

// Control mode //

func (c *ClientSelectControlModeCommand) UnmarshalStream(stream *core.ByteStream) {
	for i := 0; i < 4; i++ {
		stream.ReadVInt()
	}

	c.controlMode, _ = stream.ReadVInt()
}

func (c *ClientSelectControlModeCommand) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	wrapper.Player.ControlMode = int32(c.controlMode)

	if err := dbm.Exec("update players set control_mode = $1 where id = $2", c.controlMode, wrapper.Player.DbId); err != nil {
		slog.Error("failed to update control mode!", "err", err, "playerId", wrapper.Player.DbId)
		return
	}

	slog.Info("changed control mode", "playerId", wrapper.Player.DbId, "mode", c.controlMode)
}

// Gatcha //

func (c *ClientGatchaCommand) UnmarshalStream(stream *core.ByteStream) {
	for i := 0; i < 4; i++ {
		stream.ReadVInt()
	}

	c.boxType, _ = stream.ReadVInt()
}

func (c *ClientGatchaCommand) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	slog.Info("giving box", "playerId", wrapper.Player.DbId, "boxType", c.boxType)

	logic := NewDeliveryLogic(wrapper, dbm)
	logic.GenerateRewards(int32(c.boxType))

	msg := NewAvailableServerCommandMessage(203, logic)

	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
