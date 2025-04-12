package core

import "time"

type PlayerBrawler struct {
	BrawlerId int32 `db:"brawler_id"`

	Trophies        int32 `db:"trophies"`
	HighestTrophies int32 `db:"highest_trophies"`

	PowerLevel  int32 `db:"power_level"`
	PowerPoints int32 `db:"power_points"`

	SelectedGadget    *int32 `db:"selected_gadget"`
	SelectedStarPower *int32 `db:"selected_star_power"`
	SelectedGear1     *int32 `db:"selected_gear1"`
	SelectedGear2     *int32 `db:"selected_gear2"`

	UnlockedSkinIds []int32          `db:"unlocked_skins"`
	Cards           map[string]int32 `db:"cards"`

	SelectedSkinId int32 `db:"selected_skin"`
}

type PlayerCurrency struct {
	CurrencyId int32 `db:"currency_id"`
	Balance    int64 `db:"balance"`
}

type Player struct {
	Name   string `db:"name"`
	Region string `db:"region"`

	HighId int32 `db:"high_id"`
	LowId  int32 `db:"low_id"`

	Token string `db:"token"`

	DbId int64 `db:"id"`

	ProfileIcon int32 `db:"profile_icon"`

	Trophies        int32 `db:"trophies"`
	HighestTrophies int32 `db:"highest_trophies"`

	SoloVictories int32 `db:"solo_victories"`
	DuoVictories  int32 `db:"duo_victories"`
	TrioVictories int32 `db:"trio_victories"`

	BattleHints   bool  `db:"battle_hints"`
	ControlMode   int32 `db:"control_mode"`
	TutorialState int32 `db:"tutorial_state"`

	CoinBooster int32 `db:"coin_booster"`
	CoinDoubler int32 `db:"coin_doubler"`
	CoinsReward int32 `db:"coins_reward"`
	Experience  int32 `db:"experience"`

	CreatedAt time.Time `db:"created_at"`
	LastLogin time.Time `db:"last_login"`

	Brawlers map[int32]*PlayerBrawler
	Wallet   map[int32]*PlayerCurrency
	
	AllianceId   *int64 `db:"alliance_id"`
	AllianceRole int16  `db:"alliance_role"`

	state PlayerState
}

func NewPlayer() *Player {
	return &Player{
		Brawlers: make(map[int32]*PlayerBrawler),
		Wallet:   make(map[int32]*PlayerCurrency),

		state: StateSession,
	}
}

func (p *Player) SetState(state PlayerState) {
	p.state = state
}

func (p *Player) State() PlayerState {
	return p.state
}
