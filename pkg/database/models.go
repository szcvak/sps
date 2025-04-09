package database

import "errors"

// --- Database tables --- //

const (
	players = `CREATE TABLE IF NOT EXISTS players (
	id bigserial primary key,
    
    name varchar(15) not null,
    
    high_id int not null,
    low_id int not null,
    
    profile_icon int not null default 0,
    
    token text unique not null,
    region text not null,
    
    created_at timestamptz not null default current_timestamp,
    last_login timestamptz not null default current_timestamp
);`

	playerProgression = `CREATE TABLE IF NOT EXISTS player_progression (
	player_id bigint primary key references players (id) on delete cascade,
	
	solo_victories int not null default 0,
	duo_victories int not null default 0,
	trio_victories int not null default 0,
	
	trophies int not null default 0,
	highest_trophies int not null default 0
);`

	playerBrawlers = `CREATE TABLE IF NOT EXISTS player_brawlers (
	player_id bigint references players (id) on delete cascade,
	brawler_id int not null,
	
	trophies int not null default 0,
	highest_trophies int not null default 0,
	
	power_level int not null default 1 check (power_level between 1 and 11),
	power_points int not null default 0,
	
	selected_gadget int default null,
	selected_star_power int default null,
	selected_gear1 int default null,
	selected_gear2 int default null,
	
	unlocked_at timestamptz not null default current_timestamp,
	
	primary key (player_id, brawler_id)
);`

	playerUnlockedStarPowers = `CREATE TABLE IF NOT EXISTS player_unlocked_star_powers (
	player_id bigint references players (id) on delete cascade,
	brawler_id int not null,
	star_power_id int not null,
	
	unlocked_at timestamptz not null default current_timestamp,
	
	primary key (player_id, brawler_id, star_power_id),
	foreign key (player_id, brawler_id) references player_brawlers (player_id, brawler_id) on delete cascade
);`

	playerUnlockedGadgets = `CREATE TABLE IF NOT EXISTS player_unlocked_gadgets (
	player_id bigint references players (id) on delete cascade,
	brawler_id int not null,
	gadget_id int not null,
	
	unlocked_at timestamptz not null default current_timestamp,
	
	primary key (player_id, brawler_id, gadget_id),
	foreign key (player_id, brawler_id) references player_brawlers (player_id, brawler_id) on delete cascade
);`

	playerUnlockedGears = `CREATE TABLE IF NOT EXISTS player_unlocked_gears (
	player_id bigint references players (id) on delete cascade,
	brawler_id int not null,
	gear_id int not null,
	
	unlocked_at timestamptz not null default current_timestamp,
	
	primary key (player_id, brawler_id, gear_id),
	foreign key (player_id, brawler_id) references player_brawlers (player_id, brawler_id) on delete cascade
);`

	playerWallet = `CREATE TABLE IF NOT EXISTS player_wallet (
	player_id bigint references players (id) on delete cascade,
	currency_id int not null,
	
	balance int not null check ( balance >= 0 ),
	
	primary key (player_id, currency_id)
);`
)

// --- Identifiers --- //

const (
	CurrencyCoins  int32 = 1
	CurrencyGems         = 2
	CurrencyBling        = 3
	CurrencyChips        = 5
	CurrencyElixir       = 6
)

// --- Errors --- //

var (
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrPlayerNotFound       = errors.New("player not found")
)

// --- Other --- //

var (
	defaultCurrencies      = []int32{CurrencyCoins, CurrencyGems, CurrencyChips}
	defaultCurrencyBalance = make(map[int32]int32)
)

func init() {
	defaultCurrencyBalance[CurrencyCoins] = 1000
	defaultCurrencyBalance[CurrencyGems] = 50
	defaultCurrencyBalance[CurrencyBling] = 0
	defaultCurrencyBalance[CurrencyChips] = 0
	defaultCurrencyBalance[CurrencyElixir] = 0
}
