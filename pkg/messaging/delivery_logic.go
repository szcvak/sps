package messaging

import (
	"github.com/mroth/weightedrand/v2"
	"github.com/szcvak/sps/pkg/config"
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/csv"
	"github.com/szcvak/sps/pkg/database"
	"log/slog"
	"math/rand"
	"strconv"
)

type DeliveryLogic struct {
	wrapper *core.ClientWrapper
	dbm     *database.Manager
	boxId   int32

	rewards map[int]map[string]int32
}

func NewDeliveryLogic(wrapper *core.ClientWrapper, dbm *database.Manager) *DeliveryLogic {
	return &DeliveryLogic{
		wrapper: wrapper,
		dbm:     dbm,
		boxId:   -1,

		rewards: make(map[int]map[string]int32),
	}
}

func (d *DeliveryLogic) GenerateRewards(box int32) {
	id, currency, price := getBoxData(box)

	if d.wrapper.Player.Wallet[currency].Balance-int64(price) < 0 {
		slog.Error("failed to generate rewards!", "err", "player balance would go negative")
		return
	}

	d.wrapper.Player.Wallet[currency].Balance -= int64(price)
	err := d.dbm.Exec("update player_wallet set balance = $1 where player_id = $2 and currency_id = $3", d.wrapper.Player.Wallet[currency].Balance, d.wrapper.Player.DbId, currency)

	if err != nil {
		slog.Error("failed to update players balance!", "playerId", d.wrapper.Player.DbId, "err", err)
		return
	}

	d.boxId = id

	rewardsAmt := 1

	if id == 11 {
		rewardsAmt = 10
	}

	for i := 0; i < rewardsAmt; i++ {
		d.rewards[i] = make(map[string]int32)

		d.rewards[i]["rarity"] = int32(config.BoxRewardRarityChooser.Pick())

		switch d.rewards[i]["rarity"] {
		case 0:
			// elixir or brawler
			chooser, _ := weightedrand.NewChooser(
				weightedrand.NewChoice(0, 25),
				weightedrand.NewChoice(1, 75),
			)

			switch chooser.Pick() {
			case 0:
				d.rewards[i]["amount"] = 2
				d.rewards[i]["dataref_high"] = 23
				d.rewards[i]["dataref_low"] = 0
				d.rewards[i]["reward_id"] = 3

				balance := d.wrapper.Player.Wallet[config.CurrencyElixir].Balance + 2
				d.wrapper.Player.Wallet[config.CurrencyElixir].Balance = balance

				err = d.dbm.Exec("update player_wallet set balance = $1 where player_id = $2 and currency_id = $3", balance, d.wrapper.Player.DbId, config.CurrencyElixir)

				if err != nil {
					slog.Error("failed to update elixir", "playerId", d.wrapper.Player.DbId, "err", err)
				}

				break
			case 1:
				brawlers := csv.GetBrawlersWithRarity("common")
				brawlerId := brawlers[rand.Intn(len(brawlers))]

				d.rewards[i]["amount"] = 1
				d.rewards[i]["dataref_high"] = 23
				d.rewards[i]["dataref_low"] = brawlerId

				_, exists := d.wrapper.Player.Brawlers[brawlerId]

				if exists {
					d.rewards[i]["reward_id"] = 2
					d.wrapper.Player.Wallet[config.CurrencyChips].Balance += 1

					err = d.dbm.Exec("update player_wallet set balance = $1 where player_id = $2 and currency_id = $3", d.wrapper.Player.Wallet[config.CurrencyChips].Balance, d.wrapper.Player.DbId, config.CurrencyChips)

					if err != nil {
						slog.Error("failed to update chips", "playerId", d.wrapper.Player.DbId, "err", err)
					}

					break
				}

				d.rewards[i]["reward_id"] = 1

				cardsKey := strconv.Itoa(int(brawlerId))

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
					Cards:             map[string]int32{cardsKey: 1},
					SelectedSkinId:    0,
				}

				d.wrapper.Player.Brawlers[brawlerId] = brawler

				stmt := `
				insert into player_brawlers (
					player_id, brawler_id, trophies, highest_trophies,
					power_level, power_points,
					unlocked_skins, selected_skin, cards
				)
				values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

				err = d.dbm.Exec(stmt, d.wrapper.Player.DbId, brawlerId, brawler.Trophies, brawler.HighestTrophies, brawler.PowerLevel, brawler.PowerPoints, brawler.UnlockedSkinIds, brawler.SelectedSkinId, brawler.Cards)

				if err != nil {
					slog.Error("failed to insert new brawler!", "playerId", d.wrapper.Player.DbId, "err", err)
				}
			}
		case 1:
			// elixir or brawler
			d.rewards[i]["amount"] = 2
			d.rewards[i]["dataref_high"] = 23
			d.rewards[i]["dataref_low"] = 0
			d.rewards[i]["reward_id"] = 3

			balance := d.wrapper.Player.Wallet[config.CurrencyElixir].Balance + 2
			d.wrapper.Player.Wallet[config.CurrencyElixir].Balance = balance

			err = d.dbm.Exec("update player_wallet set balance = $1 where player_id = $2 and currency_id = $3", balance, d.wrapper.Player.DbId, config.CurrencyElixir)

			if err != nil {
				slog.Error("failed to update elixir", "playerId", d.wrapper.Player.DbId, "err", err)
			}

			break
		}
	}
}

func (d *DeliveryLogic) Marshal(stream *core.ByteStream) {
	if d.boxId == -1 {
		return
	}

	stream.Write(core.VInt(d.boxId))
	stream.Write(core.VInt(len(d.rewards)))

	for _, data := range d.rewards {
		stream.Write(core.VInt(data["rarity"]))
		stream.Write(core.VInt(data["amount"]))
		stream.Write(core.ScId{data["dataref_high"], data["dataref_low"]})
		stream.Write(core.VInt(data["reward_id"]))
	}
}

// --- Helpers --- //

func getBoxData(id int32) (int32, int32, int32) {
	if id == 1 {
		return 10, config.CurrencyCoins, 100
	} else if id == 2 {
		return 10, config.CurrencyGems, 10
	} else if id == 3 {
		return 11, config.CurrencyGems, 80
	}

	return id, config.CurrencyCoins, 0
}

func getRarityFromId(rarity int32) string {
	switch rarity {
	case 0:
		return "common"
	case 1:
		return "rare"
	case 2:
		return "super_rare"
	case 3:
		return "epic"
	case 4:
		return "mega_epic"
	}

	return "legendary"
}
