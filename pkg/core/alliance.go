package core

import "time"

type AllianceMessage struct {
	Id         int64  `db:"id"`
	AllianceId int64  `db:"alliance_id"`
	PlayerId   *int64 `db:"player_id"`

	PlayerHighId int32  `db:"player_high_id"`
	PlayerLowId  int32  `db:"player_low_id"`
	PlayerName   string `db:"player_name"`
	PlayerRole   int16  `db:"player_role"`
	PlayerExp    int32  `db:"player_exp"`
	PlayerIcon   int32  `db:"player_icon"`

	Type    int16  `db:"message_type"`
	Content string `db:"message_content"`

	TargetId   *int64 `db:"target_id"`
	TargetName string `db:"target_name"`

	Timestamp time.Time `db:"created_at"`
}

type AllianceMember struct {
	PlayerId int64
	Role     int16

	Name string

	Experience int32
	Trophies   int32

	ProfileIcon int32
	LowId       int32
	HighId      int32
}

type Alliance struct {
	Id int64

	Name        string
	Description string

	BadgeId int32
	Type    int16

	RequiredTrophies int32
	TotalTrophies    int32

	CreatorId *int64

	Members      []AllianceMember
	TotalMembers int32

	Region string
}

func NewAlliance(id int64) *Alliance {
	return &Alliance{
		Id: id,
	}
}
