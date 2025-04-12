package messaging 

import (
	"fmt"
	"log/slog"
	"math/rand"
	"strconv"
	"time"
	"encoding/json"

	"github.com/mroth/weightedrand/v2"
	"github.com/szcvak/sps/pkg/config"
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/csv"
	"github.com/szcvak/sps/pkg/database"
)

const (
	RewardIdBrawler int32 = 1
	RewardIdChips int32 = 2
	RewardIdElixir int32 = 3
	RewardIdCoinDoubler int32 = 4
	RewardIdCoinBooster int32 = 5
	
	DataRefClassCard int32 = 23

	DataRefInstanceElixir int32 = 0
)

type RewardItem struct {
	Rarity      int32
	Amount      int32
	RewardId    int32
	DataRef     core.DataRef
}

type BoxConfig struct {
	Id           int32
	Price        int32
	CurrencyId   int32
	RewardsCount int
}

var boxConfigs = map[int32]BoxConfig{
	1: {Id: 10, Price: 100, CurrencyId: config.CurrencyCoins, RewardsCount: 1},
	2: {Id: 10, Price: 10, CurrencyId: config.CurrencyGems, RewardsCount: 1},
	3: {Id: 11, Price: 80, CurrencyId: config.CurrencyGems, RewardsCount: 10},
}

type RarityConfig struct {
	Name         string
	ElixirAmount int32
	ChipAmount   int32

	WeightElixir  uint
	WeightBrawler uint
	WeightBooster uint
	WeightDoubler uint
}

var rarityConfigs = map[int32]RarityConfig{
	0: {Name: "common", ElixirAmount: 2, ChipAmount: 1, WeightElixir: 80, WeightBrawler: 20, WeightBooster: 0, WeightDoubler: 0},
	1: {Name: "rare", ElixirAmount: 2, ChipAmount: 2, WeightElixir: 80, WeightBrawler: 20, WeightBooster: 0, WeightDoubler: 0},
	2: {Name: "super_rare", ElixirAmount: 3, ChipAmount: 4, WeightElixir: 30, WeightBrawler: 10, WeightBooster: 30, WeightDoubler: 30},
	3: {Name: "epic", ElixirAmount: 5, ChipAmount: 10, WeightElixir: 80, WeightBrawler: 20, WeightBooster: 0, WeightDoubler: 0},
	4: {Name: "mega_epic", ElixirAmount: 7, ChipAmount: 25, WeightElixir: 80, WeightBrawler: 20, WeightBooster: 0, WeightDoubler: 0},
	5: {Name: "legendary", ElixirAmount: 10, ChipAmount: 60, WeightElixir: 80, WeightBrawler: 20, WeightBooster: 0, WeightDoubler: 0},
}

type DeliveryLogic struct {
	wrapper *core.ClientWrapper
	dbm     *database.Manager
	boxId   int32
	rewards []RewardItem
}

func NewDeliveryLogic(wrapper *core.ClientWrapper, dbm *database.Manager) *DeliveryLogic {
	return &DeliveryLogic{
		wrapper: wrapper,
		dbm:     dbm,
		boxId:   -1,
		rewards: make([]RewardItem, 0),
	}
}

func (d *DeliveryLogic) GenerateRewards(boxTypeIdentifier int32) error {
	boxConf, ok := boxConfigs[boxTypeIdentifier]
	
	if !ok {
		err := fmt.Errorf("unknown box type identifier: %d", boxTypeIdentifier)
		slog.Error("failed to generate rewards!", "err", err)
		
		return err
	}

	player := d.wrapper.Player
	wallet, exists := player.Wallet[boxConf.CurrencyId]
	
	if !exists || wallet.Balance < int64(boxConf.Price) {
		err := fmt.Errorf("insufficient funds for box %d (Type %d): need %d of currency %d, have %d",
			boxConf.Id, boxTypeIdentifier, boxConf.Price, boxConf.CurrencyId, wallet.Balance)

		slog.Warn("will not giwe out rewards", "playerId", player.DbId, "err", err)
		
		return err
	}

	newBalance := wallet.Balance - int64(boxConf.Price)
	
	err := d.dbm.Exec("update player_wallet set balance = $1 where player_id = $2 and currency_id = $3",
		newBalance, player.DbId, boxConf.CurrencyId)

	if err != nil {
		slog.Error("failed to update player!", "playerId", player.DbId, "currency", boxConf.CurrencyId, "err", err)
		return fmt.Errorf("failed to deduct box cost: %w", err)
	}

	wallet.Balance = newBalance
	
	d.boxId = boxConf.Id
	d.rewards = make([]RewardItem, 0, boxConf.RewardsCount)

	for i := 0; i < boxConf.RewardsCount; i++ {
		reward, err := d.generateSingleReward()
		
		if err != nil {
			slog.Error("failed to generate reward!", "boxId", d.boxId, "itemIndex", i, "err", err)
			continue
		}
		
		if reward != nil {
			d.rewards = append(d.rewards, *reward)
		}
	}

	slog.Info("generated rewards", "playerId", player.DbId, "boxId", d.boxId, "count", len(d.rewards))
	
	return nil
}

func (d *DeliveryLogic) generateSingleReward() (*RewardItem, error) {
	rarityId := int32(config.BoxRewardRarityChooser.Pick())
	rarityConf, ok := rarityConfigs[rarityId]
	
	if !ok {
		return nil, fmt.Errorf("invalid rarity picked: %d", rarityId)
	}

	choices := []weightedrand.Choice[int, uint]{}
	
	if rarityConf.WeightElixir > 0 {
		choices = append(choices, weightedrand.NewChoice(0, rarityConf.WeightElixir))
	}
	
	if rarityConf.WeightBrawler > 0 {
		choices = append(choices, weightedrand.NewChoice(1, rarityConf.WeightBrawler))
	}
	
	if rarityConf.WeightBooster > 0 {
		choices = append(choices, weightedrand.NewChoice(2, rarityConf.WeightBooster))
	}
	
	if rarityConf.WeightDoubler > 0 {
		choices = append(choices, weightedrand.NewChoice(3, rarityConf.WeightDoubler))
	}

	chooser, err := weightedrand.NewChooser(choices...)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create reward type chooser for rarity %d: %w", rarityId, err)
	}

	rewardType := chooser.Pick()

	switch rewardType {
	case 0: // elixir
		return d.grantElixir(rarityConf.ElixirAmount, rarityId)
	case 1: // Brawler
		characters := csv.GetBrawlersWithRarity(rarityConf.Name)
		
		if len(characters) == 0 {
			slog.Warn("no brawlers found for rarity, will give elixir", "rarity", rarityConf.Name)
			return d.grantElixir(rarityConf.ElixirAmount, rarityId)
		}
		
		selected := characters[rand.Intn(len(characters))]
		return d.grantBrawler(selected, rarityConf.ChipAmount, rarityId)
	case 2: // booster
		boosterDuration := int32(259200) // 3 days
		return d.grantCoinBooster(boosterDuration, rarityId)
	case 3: // doubler
		amount := int32(200)
		return d.grantCoinDoubler(amount, rarityId)
	default:
		return nil, fmt.Errorf("unknown reward type picked: %d", rewardType)
	}
}

func (d *DeliveryLogic) grantElixir(amount int32, rarity int32) (*RewardItem, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("invalid elixir amount: %d", amount)
	}
	
	player := d.wrapper.Player
	currencyId := config.CurrencyElixir

	newBalance := player.Wallet[int32(currencyId)].Balance + int64(amount)
	
	err := d.dbm.Exec("update player_wallet set balance = $1 where player_id = $2 and currency_id = $3",
		newBalance, player.DbId, currencyId)
		
	if err != nil {
		slog.Error("failed to update elixir", "playerId", player.DbId, "err", err)
		return nil, fmt.Errorf("error granting elixir: %w", err)
	}

	player.Wallet[int32(currencyId)].Balance = newBalance

	return &RewardItem{
		Rarity:      rarity,
		Amount:      amount,
		RewardId:    RewardIdElixir,
		DataRef:     core.DataRef{DataRefClassCard, DataRefInstanceElixir},
	}, nil
}

func (d *DeliveryLogic) grantChips(amount int32, id int32, rarity int32) (*RewardItem, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("invalid chip amount: %d", amount)
	}
	
	player := d.wrapper.Player
	currencyId := config.CurrencyChips

	newBalance := player.Wallet[int32(currencyId)].Balance + int64(amount)
	
	err := d.dbm.Exec("update player_wallet set balance = $1 where player_id = $2 and currency_id = $3",
		newBalance, player.DbId, currencyId)
		
	if err != nil {
		slog.Error("failed to update chips", "playerId", player.DbId, "err", err)
		return nil, fmt.Errorf("error granting chips: %w", err)
	}

	player.Wallet[int32(currencyId)].Balance = newBalance

	return &RewardItem{
		Rarity:      rarity,
		Amount:      amount,
		RewardId:    RewardIdChips,
		DataRef:     core.DataRef{DataRefClassCard, id},
	}, nil
}

func (d *DeliveryLogic) grantCoinBooster(duration int32, rarity int32) (*RewardItem, error) {
	if duration <= 0 {
		return nil, fmt.Errorf("invalid booster duration: %d", duration)
	}
	
	player := d.wrapper.Player
	
	now := time.Now().Unix()
	newBoosterEndTime := player.CoinBooster

	if now >= int64(newBoosterEndTime) {
		newBoosterEndTime = int32(now) + duration
	} else {
		newBoosterEndTime += duration
	}

	err := d.dbm.Exec("update players set coin_booster = $1 where id = $2", newBoosterEndTime, player.DbId)
	
	if err != nil {
		slog.Error("failed to update coin_booster!", "playerId", player.DbId, "err", err)
		return nil, fmt.Errorf("error granting booster: %w", err)
	}

	player.CoinBooster = newBoosterEndTime

	return &RewardItem{
		Rarity:      rarity,
		Amount:      duration,
		RewardId:    RewardIdCoinBooster,
		DataRef:     core.DataRef{DataRefClassCard, 0},
	}, nil
}

func (d *DeliveryLogic) grantCoinDoubler(amount int32, rarity int32) (*RewardItem, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("invalid doubler amount: %d", amount)
	}
	
	player := d.wrapper.Player
	newAmount := amount + player.CoinDoubler
	
	err := d.dbm.Exec("update players set coin_doubler = $1 where id = $2", newAmount, player.DbId)
	
	if err != nil {
		slog.Error("failed to update coin_doubler!", "playerId", player.DbId, "err", err)
		return nil, fmt.Errorf("error granting doubler: %w", err)
	}

	player.CoinDoubler = newAmount

	return &RewardItem{
		Rarity:      rarity,
		Amount:      amount,
		RewardId:    RewardIdCoinDoubler,
		DataRef:     core.DataRef{DataRefClassCard, 0},
	}, nil
}

func (d *DeliveryLogic) grantBrawler(cardId int32, chipAmount int32, rarity int32) (*RewardItem, error) {
	player := d.wrapper.Player

	brawlerId := csv.GetBrawlerId(cardId)

	_, exists := player.Brawlers[brawlerId]

	if exists {
		return d.grantChips(chipAmount, cardId, rarity)
	} else {
		cardsKey := strconv.Itoa(int(cardId))
		newBrawlerCards := map[string]int32{cardsKey: 1}

		cardsJson, err := json.Marshal(newBrawlerCards)
		
		if err != nil {
			return nil, fmt.Errorf("failed to marshal new brawler cards to json: %w", err)
		}
		
		skinsJson, err := json.Marshal([]int32{0})
		
		if err != nil {
			return nil, fmt.Errorf("failed to marshal new brawler skins to json: %w", err)
		}


		brawler := &core.PlayerBrawler{
			BrawlerId:         brawlerId,
			Trophies:          config.NewBrawlerTrophies,
			HighestTrophies:   config.NewBrawlerTrophies,
			PowerLevel:        config.NewBrawlerPowerLevel,
			PowerPoints:       config.NewBrawlerPowerPoints,
			SelectedGadget:    nil,
			SelectedStarPower: nil,
			SelectedGear1:     nil,
			SelectedGear2:     nil,
			UnlockedSkinIds:   []int32{0},
			Cards:             newBrawlerCards,
			SelectedSkinId:    0,
		}

		stmt := `
			INSERT INTO player_brawlers (
				player_id, brawler_id, trophies, highest_trophies,
				power_level, power_points,
				unlocked_skins, selected_skin, cards
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

		err = d.dbm.Exec(stmt,
			player.DbId,
			brawler.BrawlerId,
			brawler.Trophies, brawler.HighestTrophies,
			brawler.PowerLevel, brawler.PowerPoints,
			skinsJson,
			brawler.SelectedSkinId,
			cardsJson,
		)

		if err != nil {
			slog.Error("failed to insert new brawler!", "playerId", player.DbId, "brawler", brawlerId, "err", err)

			return nil, fmt.Errorf("error granting new brawler: %w", err)
		}

		player.Brawlers[brawlerId] = brawler

		return &RewardItem{
			Rarity:      rarity,
			Amount:      1,
			RewardId:    RewardIdBrawler,
			DataRef:     core.DataRef{DataRefClassCard, cardId},
		}, nil
	}
}

func (d *DeliveryLogic) Marshal(stream *core.ByteStream) {
	if d.boxId == -1 || len(d.rewards) == 0 {
		return
	}

	stream.Write(core.VInt(d.boxId))
	stream.Write(core.VInt(len(d.rewards)))

	for _, item := range d.rewards {
		stream.Write(core.VInt(item.Rarity))
		stream.Write(core.VInt(item.Amount))
		stream.Write(core.ScId{item.DataRef.F, item.DataRef.S})
		stream.Write(core.VInt(item.RewardId))
	}
}
