package messaging

import (
	"github.com/szcvak/sps/pkg/config"
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/csv"
	"github.com/szcvak/sps/pkg/database"
	
	"encoding/json"
	"log/slog"
	"time"
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
	ClientCommands[507] = func() ClientCommand { return NewClientSelectSkinCommand() }
	ClientCommands[508] = func() ClientCommand { return NewClientUnlockSkinCommand() }
	ClientCommands[509] = func() ClientCommand { return NewClientSelectControlModeCommand() }
	ClientCommands[510] = func() ClientCommand { return NewClientBuyCoinDoubler() }
	ClientCommands[511] = func() ClientCommand { return NewClientBuyCoinBooster() }
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

type ClientBuyCoinDoubler struct{}
type ClientBuyCoinBooster struct{}

type ClientUnlockSkinCommand struct {
	skin core.DataRef
}

type ClientSelectSkinCommand struct{
	skin core.DataRef
}

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

func NewClientBuyCoinDoubler() *ClientBuyCoinDoubler {
	return &ClientBuyCoinDoubler{}
}

func NewClientBuyCoinBooster() *ClientBuyCoinBooster {
	return &ClientBuyCoinBooster{}
}

func NewClientUnlockSkinCommand() *ClientUnlockSkinCommand {
	return &ClientUnlockSkinCommand{}
}

func NewClientSelectSkinCommand() *ClientSelectSkinCommand {
	return &ClientSelectSkinCommand{}
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
	event := em.GetCurrentEventPtr(int32(c.eventSlot) - 1)

	flag := false
	
	for _, v := range event.SeenBy {
		if v == wrapper.Player.DbId {
			flag = true
			break
		}
	}
	
	if flag {
		return
	}
	
	amount := int64(event.Config.CoinsToWin + (event.Config.CoinsToClaim / 2))

	slog.Info("giving event coins", "playerId", wrapper.Player.DbId, "amount", amount)

	if err := dbm.Exec("update player_wallet set balance = $1 where player_id = $2 and currency_id = $3", wrapper.Player.Wallet[config.CurrencyCoins].Balance + amount, wrapper.Player.DbId, config.CurrencyCoins); err != nil {
		slog.Error("failed to update player wallet!", "err", err)
		return
	}
	
	wrapper.Player.Wallet[config.CurrencyCoins].Balance += amount
	event.SeenBy = append(event.SeenBy, wrapper.Player.DbId)
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

// --- Buy coin doubler --- //

func (c *ClientBuyCoinDoubler) UnmarshalStream(stream *core.ByteStream) {}

func (c *ClientBuyCoinDoubler) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	newBalance := wrapper.Player.Wallet[config.CurrencyGems].Balance - config.CoinDoublerPrice
	
	if newBalance < 0 {
		return
	}
	
	newCoinDoubler := wrapper.Player.CoinDoubler + config.CoinDoublerReward
	
	if err := dbm.Exec("update players set coin_doubler = $1 where id = $2", newCoinDoubler, wrapper.Player.DbId); err != nil {
		slog.Error("failed to update player's coin doubler!", "err", err)
		return
	}
	
	if err := dbm.Exec("update player_wallet set balance = $1 where player_id = $2 and currency_id = $3", newBalance, wrapper.Player.DbId, config.CurrencyGems); err != nil {
		slog.Error("failed to update player's balance!", "err", err)
		return
	}

	wrapper.Player.CoinDoubler = newCoinDoubler
}

// --- Buy coin booster --- //

func (c *ClientBuyCoinBooster) UnmarshalStream(stream *core.ByteStream) {}

func (c *ClientBuyCoinBooster) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	newBalance := wrapper.Player.Wallet[config.CurrencyGems].Balance - config.CoinBoosterPrice
	
	if newBalance < 0 {
		return
	}
	
	now := time.Now().Unix()
	newCoinBooster := wrapper.Player.CoinBooster + config.CoinBoosterReward
	
	if wrapper.Player.CoinBooster < int32(now) {
		newCoinBooster = int32(now) + config.CoinBoosterReward
	}
	
	if err := dbm.Exec("update players set coin_booster = $1 where id = $2", newCoinBooster, wrapper.Player.DbId); err != nil {
		slog.Error("failed to update player's coin booster!", "err", err)
		return
	}
	
	if err := dbm.Exec("update player_wallet set balance = $1 where player_id = $2 and currency_id = $3", newBalance, wrapper.Player.DbId, config.CurrencyGems); err != nil {
		slog.Error("failed to update player's balance!", "err", err)
		return
	}

	wrapper.Player.CoinBooster = newCoinBooster
}

// --- Unlock skin --- //

func (c *ClientUnlockSkinCommand) UnmarshalStream(stream *core.ByteStream) {
	for i := 0; i < 4; i++ {
		_, _ = stream.ReadVInt()
	}

	c.skin, _ = stream.ReadDataRef()
}

func (c *ClientUnlockSkinCommand) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.State() != core.StateLoggedIn {
		return
	}

	contains := false

	for _, value := range csv.Skins() {
		if value == c.skin.S {
			contains = true
			break
		}
	}

	if !contains {
		return
	}

	if csv.IsSkinDefault(c.skin.S) {
		return
	}
	
	brawler := csv.GetBrawlerForSkin(c.skin.S)
	_, exists := wrapper.Player.Brawlers[brawler]
	
	if !exists {
		return
	}
	
	price := csv.GetSkinPrice(c.skin.S)
	newBalance := wrapper.Player.Wallet[config.CurrencyGems].Balance - int64(price)
	
	if newBalance < 0 {
		return
	}
	
	if err := dbm.Exec("update player_wallet set balance = $1 where player_id = $2 and currency_id = $3", newBalance, wrapper.Player.DbId, config.CurrencyGems); err != nil {
		slog.Error("failed to update player wallet!")
		return
	}
	
	wrapper.Player.Wallet[config.CurrencyGems].Balance = newBalance
	wrapper.Player.Brawlers[brawler].UnlockedSkinIds = append(wrapper.Player.Brawlers[brawler].UnlockedSkinIds, c.skin.S)
	wrapper.Player.Brawlers[brawler].SelectedSkinId = c.skin.S
	
	data, err := json.Marshal(wrapper.Player.Brawlers[brawler].UnlockedSkinIds)
	
	if err != nil {
		slog.Error("failed to marshal unlocked skin ids!", "err", err)
		return
	}
	
	if err := dbm.Exec("update player_brawlers set unlocked_skins = $1, selected_skin = $2 where player_id = $3 and brawler_id = $4", data, wrapper.Player.Brawlers[brawler].SelectedSkinId, wrapper.Player.DbId, brawler); err != nil {
		slog.Error("failed to update player's brawlers!")
		return
	}
}

// --- Select skin --- //

func (c *ClientSelectSkinCommand) UnmarshalStream(stream *core.ByteStream) {
	for i := 0; i < 4; i++ {
		_, _ = stream.ReadVInt()
	}

	c.skin, _ = stream.ReadDataRef()
}

func (c *ClientSelectSkinCommand) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	brawler := csv.GetBrawlerForSkin(c.skin.S)
	data, exists := wrapper.Player.Brawlers[brawler]
	
	if !exists {
		return
	}
	
	f := false
	
	for _, skin := range data.UnlockedSkinIds {
		if skin == c.skin.S {
			f = true
			break
		}
	}
	
	if !f {
		return
	}
	
	if err := dbm.Exec("update player_brawlers set selected_skin = $1 where player_id = $2 and brawler_id = $3", c.skin.S, wrapper.Player.DbId, brawler); err != nil {
		slog.Error("failed to update player's selected skin!")
		return
	}
	
	wrapper.Player.Brawlers[brawler].SelectedSkinId = c.skin.S
}
