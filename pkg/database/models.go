package database

import "errors"

// --- Database tables --- //

const (
	players = `create table if not exists players (
	id bigserial primary key,
    
    name varchar(15) not null,
    
    high_id int not null,
    low_id int not null unique,
    
    profile_icon int not null default 0,
    
    token text unique not null,
    region text not null,

	battle_hints boolean not null default true,
	control_mode smallint not null default 0,
	tutorial_state int not null default 0,

	coin_booster int not null default 0,
	coin_doubler int not null default 0,
	coins_reward int not null default 0,
	
	selected_card_high int not null default 16,
	selected_card_low int not null default 0,

    created_at timestamptz not null default current_timestamp,
    last_login timestamptz not null default current_timestamp
);`

	playerProgression = `create table if not exists player_progression (
	player_id bigint primary key references players (id) on delete cascade,
	
	solo_victories int not null default 0,
	duo_victories int not null default 0,
	trio_victories int not null default 0,
	
	trophies int not null default 0,
	highest_trophies int not null default 0,
	
	experience int not null default 0
);`

	playerBrawlers = `create table if not exists player_brawlers (
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
	
	unlocked_skins jsonb not null default '[]'::jsonb,
	cards jsonb not null default '{}'::jsonb,

	selected_skin int not null default 0,
	
	unlocked_at timestamptz not null default current_timestamp,
	
	primary key (player_id, brawler_id)
);`

	playerUnlockedStarPowers = `create table if not exists player_unlocked_star_powers (
	player_id bigint references players (id) on delete cascade,
	brawler_id int not null,
	star_power_id int not null,
	
	unlocked_at timestamptz not null default current_timestamp,
	
	primary key (player_id, brawler_id, star_power_id),
	foreign key (player_id, brawler_id) references player_brawlers (player_id, brawler_id) on delete cascade
);`

	playerUnlockedGadgets = `create table if not exists player_unlocked_gadgets (
	player_id bigint references players (id) on delete cascade,
	brawler_id int not null,
	gadget_id int not null,
	
	unlocked_at timestamptz not null default current_timestamp,
	
	primary key (player_id, brawler_id, gadget_id),
	foreign key (player_id, brawler_id) references player_brawlers (player_id, brawler_id) on delete cascade
);`

	playerUnlockedGears = `create table if not exists player_unlocked_gears (
	player_id bigint references players (id) on delete cascade,
	brawler_id int not null,
	gear_id int not null,
	
	unlocked_at timestamptz not null default current_timestamp,
	
	primary key (player_id, brawler_id, gear_id),
	foreign key (player_id, brawler_id) references player_brawlers (player_id, brawler_id) on delete cascade
);`

	playerWallet = `create table if not exists player_wallet (
	player_id bigint references players (id) on delete cascade,
	currency_id int not null,
	
	balance int not null check ( balance >= 0 ),
	
	primary key (player_id, currency_id)
);`

	alliances = `create table if not exists alliances (
    id bigserial primary key,

    name varchar(30) not null unique,
    description text default null,
	
    badge_id int not null default 0,
    type smallint not null default 1,
	
    required_trophies int not null default 0,
    total_trophies int not null default 0,
	
	region text not null,

    created_at timestamptz not null default current_timestamp,
    creator_id bigint references players (id) on delete set null
);`

	allianceMembers = `create table if not exists alliance_members (
    alliance_id bigint references alliances (id) on delete cascade,
    player_id bigint references players (id) on delete cascade,

    role smallint not null default 1,
    joined_at timestamptz not null default current_timestamp,

    primary key (alliance_id, player_id)
);`

	allianceMessages = `create table if not exists alliance_messages (
    id bigserial primary key,
    alliance_id bigint references alliances (id) on delete cascade not null,
    player_id bigint references players (id) on delete set null,

    player_high_id int not null,
    player_low_id int not null,
    player_name varchar(15) not null,
    player_role smallint not null,
    player_icon int not null,

    message_type smallint not null,
    message_content text,

    target_id bigint default null,
    target_name varchar(15) default null,

    created_at timestamptz not null default current_timestamp
);`
)

// --- Errors --- //

var (
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrPlayerNotFound       = errors.New("player not found")
)

// --- Other --- //

var (
	defaultUnlockedSkinsJson = `[0]`
	defaultBrawlerCards      = `{"0": 1}`
	defaultSkinId            = 0
)
