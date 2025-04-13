package config

import (
	"github.com/mroth/weightedrand/v2"
)

const (
	// --- Crypto configuration --- //

	Rc4Key      = "fhsd6f86f67rt8fw78fw789we78r9789wer6re"
	Rc4KeyNonce = "nonce"

	// --- General configuration --- //

	MaximumRank         = 20
	MaximumUpgradeLevel = 5

	// --- New brawler configuration --- //

	NewBrawlerPowerLevel       int32 = 1
	NewBrawlerTrophies         int32 = 0
	NewBrawlerPowerPoints      int32 = 0
	NewPlayerStartingBrawlerId int32 = 0

	// --- New player configuration --- //

	NewPlayerTrophies int32 = 0
	
	// --- Offer configuration --- //
	
	CoinBoosterPrice  int64 = 20
	CoinBoosterReward int32 = 7 * 24 * 60 * 60

	CoinDoublerPrice  int64 = 50
	CoinDoublerReward int32 = 1000
)

const (
	CurrencyCoins  int32 = 1
	CurrencyGems         = 2
	CurrencyBling        = 3
	CurrencyChips        = 5
	CurrencyElixir       = 6
)

var (
	DefaultCurrencies = []int32{
		CurrencyCoins, CurrencyGems, CurrencyBling, CurrencyChips, CurrencyElixir,
	}

	DefaultCurrencyBalance = map[int32]int32{
		CurrencyCoins:  1000,
		CurrencyGems:   1000,
		CurrencyBling:  5000,
		CurrencyChips:  0,
		CurrencyElixir: 0,
	}

	BoxRewardRarityChooser, _ = weightedrand.NewChooser(
		weightedrand.NewChoice(0, 25),
		weightedrand.NewChoice(1, 20),
		weightedrand.NewChoice(2, 15),
		weightedrand.NewChoice(3, 12),
		weightedrand.NewChoice(4, 8),
		weightedrand.NewChoice(5, 5),
	)
)
