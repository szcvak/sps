package messaging

import (
	"github.com/szcvak/sps/pkg/config"
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/csv"
	"github.com/szcvak/sps/pkg/database"
	"log/slog"
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
	ClientCommands[500] = func() ClientCommand { return NewClientGatchaCommand() }
	ClientCommands[504] = func() ClientCommand { return NewClientEventActionCommand() }
	ClientCommands[506] = func() ClientCommand { return NewClientProfileIconCommand() }
	ClientCommands[509] = func() ClientCommand { return NewClientSelectControlModeCommand() }
	ClientCommands[513] = func() ClientCommand { return NewClientSelectBattleHintsCommand() }
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

type ClientProfileIconCommand struct {
	profileIcon core.DataRef
}

type ClientEventActionCommand struct {
	eventSlot core.VInt
	action    core.VInt
}

type ClientSelectBattleHintsCommand struct{}

func NewClientSelectControlModeCommand() *ClientSelectControlModeCommand {
	return &ClientSelectControlModeCommand{}
}

func NewClientSelectBattleHintsCommand() *ClientSelectBattleHintsCommand {
	return &ClientSelectBattleHintsCommand{}
}

func NewClientGatchaCommand() *ClientGatchaCommand {
	return &ClientGatchaCommand{}
}

func NewClientProfileIconCommand() *ClientProfileIconCommand {
	return &ClientProfileIconCommand{}
}

func NewClientEventActionCommand() *ClientEventActionCommand {
	return &ClientEventActionCommand{}
}

// --- Control mode --- //

func (c *ClientSelectControlModeCommand) UnmarshalStream(stream *core.ByteStream) {
	for i := 0; i < 4; i++ {
		_, _ = stream.ReadVInt()
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

// --- Gatcha --- //

func (c *ClientGatchaCommand) UnmarshalStream(stream *core.ByteStream) {
	for i := 0; i < 4; i++ {
		_, _ = stream.ReadVInt()
	}

	c.boxType, _ = stream.ReadVInt()
}

func (c *ClientGatchaCommand) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	slog.Info("giving box", "playerId", wrapper.Player.DbId, "boxType", c.boxType)

	logic := NewDeliveryLogic(wrapper, dbm)
	_ = logic.GenerateRewards(int32(c.boxType))

	msg := NewAvailableServerCommandMessage(203, logic)

	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}

// --- Profile icon --- //

func (c *ClientProfileIconCommand) UnmarshalStream(stream *core.ByteStream) {
	for i := 0; i < 4; i++ {
		_, _ = stream.ReadVInt()
	}

	c.profileIcon, _ = stream.ReadDataRef()
	slog.Info("READ PROFILE", "profileIcon", c.profileIcon.S)
}

func (c *ClientProfileIconCommand) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.State() != core.StateLoggedIn {
		return
	}

	contains := false

	for _, value := range csv.Thumbnails() {
		if value == c.profileIcon.S {
			contains = true
			break
		}
	}

	if !contains {
		return
	}

	if csv.GetTrophiesForThumbnail(c.profileIcon.S) > wrapper.Player.Trophies {
		return
	}

	_, exists := wrapper.Player.Brawlers[csv.GetCharacterIdByName(csv.GetBrawlerForThumbnail(c.profileIcon.S))]

	if !exists {
		return
	}

	requiredExperience := csv.GetExperienceLevelForThumbnail(c.profileIcon.S)
	experience := core.RequiredExp[requiredExperience]

	if experience > wrapper.Player.Experience {
		return
	}

	wrapper.Player.ProfileIcon = c.profileIcon.S

	err := dbm.Exec("update players set profile_icon = $1 where id = $2", c.profileIcon.S, wrapper.Player.DbId)

	if err != nil {
		slog.Error("failed to update profile icon!", "err", err)
		return
	}
}

// --- Event action command --- //

func (c *ClientEventActionCommand) UnmarshalStream(stream *core.ByteStream) {
	for i := 0; i < 4; i++ {
		_, _ = stream.ReadVInt()
	}

	c.eventSlot, _ = stream.ReadVInt()
	c.action, _ = stream.ReadVInt()
}

func (c *ClientEventActionCommand) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if c.action != 2 {
		slog.Error("invalid action!", "action", c.action)
		return
	}

	em := core.GetEventManager()
	slot := em.GetSlotConfig(int32(c.eventSlot) - 1)

	amount := int64(slot.CoinsToWin + (slot.CoinsToClaim / 2))

	slog.Info("giving event coins", "playerId", wrapper.Player.DbId, "amount", amount)

	wrapper.Player.Wallet[config.CurrencyCoins].Balance += amount

	if err := dbm.Exec("update player_wallet set balance = $1 where player_id = $2 and currency_id = $3", wrapper.Player.Wallet[config.CurrencyCoins].Balance, wrapper.Player.DbId, config.CurrencyCoins); err != nil {
		slog.Error("failed to update player wallet!", "err", err)
		return
	}
}

// --- Select battle hints --- //

func (c *ClientSelectBattleHintsCommand) UnmarshalStream(stream *core.ByteStream) {}

func (c *ClientSelectBattleHintsCommand) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if err := dbm.Exec("update players set battle_hints = not battle_hints where id = $1", wrapper.Player.DbId); err != nil {
		slog.Error("failed to update player battle hints!", "err", err)
		return
	}

	wrapper.Player.BattleHints = !wrapper.Player.BattleHints
}
