package core

type TeamPlayer struct {
	PlayerId int64
	Name string
	IsReady bool
	IsCreator bool
	Status int16
	HighId  int32
	LowId int32
	SelectedBrawler ScId
	SelectedSkin ScId
	Wrapper *ClientWrapper
}

type TeamMessage struct {
	PlayerId int64
	PlayerName string
	PlayerHighId int32
	PlayerLowId int32
	Content string
	Timestamp int64
	Type int32
	Event int32
	TargetName string
	TargetId int32
}

type Team struct {
	Members      []TeamPlayer
	Messages     []TeamMessage
	IsPractice   bool
	Creator      int64
	Event        int32
	Id           int32
	PostAd       bool
}
