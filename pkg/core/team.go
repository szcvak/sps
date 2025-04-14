package core

type Team struct {
	DbId         int64
	Code         string
	Members      []*Player
	IsPractice   bool
	TotalMembers int32
	Creator      *Player
}

func NewTeam(code string) *Team {
	return &Team{
		Code: code,
	}
}
