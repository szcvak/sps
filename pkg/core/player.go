package core

type Player struct {
	name     string
	loggedIn bool
}

func NewPlayer(name string) *Player {
	return &Player{
		name:     name,
		loggedIn: false,
	}
}
