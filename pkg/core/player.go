package core

type Player struct {
	Name     string
	LoggedIn bool

	HighId int32
	LowId  int32

	Token string
}

func NewPlayer(name string) *Player {
	return &Player{
		Name:     name,
		LoggedIn: false,

		HighId: 0,
		LowId:  0,

		Token: "",
	}
}
