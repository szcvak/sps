package core

type PlayerState int32

const (
	StateSession PlayerState = iota
	StateLogin
	StateLoggedIn
)
