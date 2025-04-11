package messages

import (
	"github.com/szcvak/sps/pkg/database"
	"log/slog"
	"math/rand"
	"strconv"
	"time"

	"github.com/szcvak/sps/pkg/config"
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/csv"
)

type OwnHomeDataMessage struct {
	wrapper *core.ClientWrapper
	dbm     *database.Manager
}

func NewOwnHomeDataMessage(wrapper *core.ClientWrapper, dbm *database.Manager) *OwnHomeDataMessage {
	return &OwnHomeDataMessage{
		wrapper: wrapper,
		dbm:     dbm,
	}
}

func (o *OwnHomeDataMessage) PacketId() uint16 {
	return 24101
}

func (o *OwnHomeDataMessage) PacketVersion() uint16 {
	return 1
}

func (o *OwnHomeDataMessage) Marshal() []byte {
	player := o.wrapper.Player

	player.SetState(core.StateLoggedIn)

	stream := core.NewByteStreamWithCapacity(128)

	brawlerTrophies := 0
	locations := csv.LocationIds()

	if config.MaximumRank <= 34 {
		brawlerTrophies = progressStart[config.MaximumRank-1]
	} else {
		brawlerTrophies = progressStart[33] + 50*(config.MaximumRank-1)
	}

	stream.Write(core.VInt(2017189)) // timestamp
	stream.Write(core.VInt(10))      // new band timer

	stream.Write(core.VInt(player.Trophies))
	stream.Write(core.VInt(player.HighestTrophies))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(player.Experience))

	stream.Write(core.ScId{28, player.ProfileIcon})

	// played game modes count
	stream.Write(core.VInt(7))

	for i := 0; i < 7; i++ {
		stream.Write(core.VInt(i))
	}

	// Brawlers with selected skins
	nonZeroSkins := make([]int32, 0)

	for _, data := range player.Brawlers {
		if data.SelectedSkinId != 0 {
			nonZeroSkins = append(nonZeroSkins, data.SelectedSkinId)
		}
	}

	stream.Write(core.VInt(len(nonZeroSkins)))

	for skin := range nonZeroSkins {
		stream.Write(core.DataRef{29, int32(skin)})
	}

	// overall owned skins for Brawlers

	nonZeroSkins = nil

	for _, data := range player.Brawlers {
		for skin := range data.UnlockedSkinIds {
			if skin != 0 {
				nonZeroSkins = append(nonZeroSkins, int32(skin))
			}
		}
	}

	stream.Write(core.VInt(len(nonZeroSkins)))

	for skin := range nonZeroSkins {
		stream.Write(core.DataRef{29, int32(skin)})
	}

	// end

	stream.Write(true)
	stream.Write(core.VInt(0))

	stream.Write(core.VInt(player.CoinsReward))

	stream.Write(false)

	stream.Write(core.VInt(player.ControlMode))
	stream.Write(player.BattleHints)
	stream.Write(core.VInt(0)) // coin doubler

	// coin booster

	now := time.Now().Unix()

	if int64(player.CoinBooster)-now > 0 {
		stream.Write(core.VInt(int64(player.CoinBooster) - now))
	} else {
		stream.Write(core.VInt(0))
		player.CoinBooster = int32(now)

		if err := o.dbm.Exec("update players set coin_booster = $1 where id = $2", player.CoinBooster, player.DbId); err != nil {
			slog.Error("failed to update coin_booster!", "playerId", player.DbId, "err", err)
		}
	}

	// end

	stream.Write(core.VInt(0))
	stream.Write(false)

	stream.Write(core.LogicLong{0, 1})
	stream.Write(core.LogicLong{0, 1})
	stream.Write(core.LogicLong{0, 1})
	stream.Write(core.LogicLong{0, 1})

	stream.Write(core.DataRef{0, 1})

	stream.Write(core.VInt(0))

	stream.Write(true)
	stream.Write(true)

	stream.Write(core.VInt(2017189))

	stream.Write(core.VInt(100))
	stream.Write(core.VInt(10))
	stream.Write(core.VInt(80))
	stream.Write(core.VInt(10))
	stream.Write(core.VInt(20))
	stream.Write(core.VInt(50))
	stream.Write(core.VInt(50))
	stream.Write(core.VInt(1000))
	stream.Write(core.VInt(7 * 24))

	stream.Write(core.VInt(brawlerTrophies))

	stream.Write(core.VInt(50))
	stream.Write(core.VInt(9999))

	stream.Write([]core.VInt{1, 2, 5, 10, 20, 60})
	stream.Write([]core.VInt{3, 10, 20, 60, 200, 500})
	stream.Write([]core.VInt{0, 30, 80, 170, 0, 0})

	// events

	stream.Write(core.VInt(4))

	for i := 0; i < 4; i++ {
		stream.Write(core.VInt(i + 1))
		stream.Write(core.VInt(requiredBrawlers[i]))
	}

	// disponible events

	stream.Write(core.VInt(4))

	for i := 0; i < 4; i++ {
		stream.Write(core.VInt(i + 1))
		stream.Write(core.VInt(i + 1))

		stream.Write(core.VInt(1))
		stream.Write(core.VInt(39120))
		stream.Write(core.VInt(8))
		stream.Write(core.VInt(8))
		stream.Write(core.VInt(999))

		stream.Write(false)
		stream.Write(i == 4)

		stream.Write(core.ScId{15, int32(locations[rand.Intn(len(locations))])})

		stream.Write(core.VInt(0))
		stream.Write(core.VInt(2))

		stream.Write("szcvak's servers")
		stream.Write(false)
	}

	// coming soon events

	stream.Write(core.VInt(4))

	for i := 0; i < 4; i++ {
		stream.Write(core.VInt(i + 1))
		stream.Write(core.VInt(i + 1))

		stream.Write(core.VInt(1337))
		stream.Write(core.VInt(39120))
		stream.Write(core.VInt(8))
		stream.Write(core.VInt(8))
		stream.Write(core.VInt(999))

		stream.Write(false)
		stream.Write(i == 4)

		stream.Write(core.ScId{15, int32(locations[rand.Intn(len(locations))])})

		stream.Write(core.VInt(0))
		stream.Write(core.VInt(2))

		stream.Write("szcvak's servers")
		stream.Write(false)
	}

	// end

	stream.Write(core.VInt(config.MaximumUpgradeLevel))

	for i := 0; i < config.MaximumUpgradeLevel; i++ {
		stream.Write(core.VInt(i + 1))
	}

	// milestones

	core.EmbedMilestones(stream)

	// end

	stream.Write(player.HighId)
	stream.Write(player.LowId)

	stream.Write(core.VInt(0))

	for i := 0; i < 3; i++ {
		stream.Write(core.LogicLong{player.HighId, player.LowId})
	}

	stream.Write(player.Name)
	stream.Write(player.Name != "Brawler")

	stream.Write(1)

	// motorised arrays

	stream.Write(core.VInt(5))

	cards := make(map[int32]int32)

	for _, data := range player.Brawlers {
		for card, amt := range data.Cards {
			id, e := strconv.Atoi(card)

			if e != nil {
				slog.Error("failed to convert card to int!", "card", card)
				continue
			}

			cards[int32(id)] = amt
		}
	}

	stream.Write(core.VInt(len(cards) + len(config.DefaultCurrencies)))

	for card, amt := range cards {
		stream.Write(core.ScId{23, card})
		stream.Write(core.VInt(amt))
	}

	// resources

	for i := 0; i < len(config.DefaultCurrencies); i++ {
		stream.Write(core.ScId{5, config.DefaultCurrencies[i]})
		stream.Write(core.VInt(player.Wallet[config.DefaultCurrencies[i]].Balance))
	}

	// end

	stream.Write(core.VInt(len(player.Brawlers)))

	for id, data := range player.Brawlers {
		stream.WriteDataRef(core.DataRef{16, id})
		stream.Write(core.VInt(data.Trophies))
	}

	stream.Write(core.VInt(len(player.Brawlers)))

	for id, data := range player.Brawlers {
		stream.WriteDataRef(core.DataRef{16, id})
		stream.Write(core.VInt(data.HighestTrophies))
	}

	stream.Write(core.VInt(0))
	stream.Write(core.VInt(len(player.Brawlers)))

	for id, _ := range player.Brawlers {
		stream.WriteDataRef(core.DataRef{16, id})
		stream.Write(core.VInt(2))
	}

	stream.Write(core.VInt(player.Wallet[config.CurrencyGems].Balance))
	stream.Write(core.VInt(0))

	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))

	stream.Write(core.VInt(player.TutorialState))

	stream.Write(core.VInt(2017189))

	player.CoinsReward = 0

	if err := o.dbm.Exec("update players set coins_reward = $1 where id = $2", 0, player.DbId); err != nil {
		slog.Error("failed to update coins_reward!", "playerId", player.DbId, "err", err)
	}

	return stream.Buffer()
}

var (
	requiredBrawlers = []int{0, 0, 0, 0}
	progressStart    = []int{0, 10, 20, 30, 40, 60, 80, 100, 120, 140, 160, 180, 220, 260, 300, 340, 380, 420, 460, 500, 550, 600, 650, 700, 750, 800, 850, 900, 950, 1000, 1050, 1100, 1150, 1200}
)
