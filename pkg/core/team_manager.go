package core

import (
	"math/rand/v2"
	"log/slog"
	"sync"
	"time"
)

type TeamManager struct {
	Teams map[int32]*Team
}

var (
	teamManagerInstance *TeamManager
	teamManagerOnce     sync.Once
)

func InitTeamManager() {
	teamManagerOnce.Do(func() {
		tm := &TeamManager{
			Teams: make(map[int32]*Team),
		}

		teamManagerInstance = tm

		slog.Info("team manager initialized")
	})
}

func GetTeamManager() *TeamManager {
	if teamManagerInstance == nil {
		panic("team manager not initialized")
	}

	return teamManagerInstance
}

func (tm *TeamManager) NextId() int32 {
	id := int32(0)
	
	for {
		id = int32(999 + rand.IntN(999999))
		
		if tm.Teams[id] != nil {
			continue
		}
		
		break
	}
	
	return id
}

func (tm *TeamManager) CreateTeam(wrapper *ClientWrapper, event int32) {
	creator := wrapper.Player
	
	if creator.TeamId != nil {
		return
	}
	
	teamId := tm.NextId()
	
	teamPlayer := TeamPlayer{
		PlayerId: creator.DbId,
		Name: creator.Name,
		HighId: creator.HighId,
		LowId: creator.LowId,
		SelectedBrawler: ScId{creator.SelectedCardHigh, creator.SelectedCardLow},
		SelectedSkin: ScId{29, creator.Brawlers[creator.SelectedCardLow].SelectedSkinId},
		IsReady: false,
		IsCreator: true,
		Status: 3,
		Wrapper: wrapper,
	}
	
	team := &Team{
		Id: teamId,
		Members: []TeamPlayer{teamPlayer},
		Creator: creator.DbId,
		Event: event,
		Messages: []TeamMessage{},
		PostAd: false,
	}
	
	tm.Teams[teamId] = team
	
	creator.TeamId = new(int32)
	*creator.TeamId = teamId
	
	slog.Info("created team", "creatorId", creator.DbId, "teamId", teamId, "event", event)
}

func (tm *TeamManager) JoinTeam(wrapper *ClientWrapper, id int32) {
	player := wrapper.Player
	
	if player.TeamId != nil {
		return
	}
	
	data := tm.Teams[id]
	
	if data == nil {
		return
	}
	
	teamPlayer := TeamPlayer{
		PlayerId: player.DbId,
		Name: player.Name,
		HighId: player.HighId,
		LowId: player.LowId,
		SelectedBrawler: ScId{player.SelectedCardHigh, player.SelectedCardLow},
		SelectedSkin: ScId{29, player.Brawlers[player.SelectedCardLow].SelectedSkinId},
		IsCreator: false,
		IsReady: false,
		Status: 3,
		Wrapper: wrapper,
	}
	
	tm.Teams[id].Members = append(tm.Teams[id].Members, teamPlayer)
	
	player.TeamId = new(int32)
	*player.TeamId = id
}

func (tm *TeamManager) LeaveTeam(player *Player) {
	if player.TeamId == nil {
		return
	}
	
	if tm.Teams[*player.TeamId] == nil {
		player.TeamId = nil
		return
	}
	
	id := *player.TeamId
	index := -1
	
	for i, data := range tm.Teams[id].Members {
		if data.PlayerId == player.DbId {
			index = i
			break
		}
	}
	
	if index == -1 {
		return
	}
	
	if tm.Teams[id].Creator == player.DbId {
		if len(tm.Teams[id].Members)-1 > 0 {
			for _, member := range tm.Teams[id].Members {
				if member.IsCreator {
					member.IsCreator = false
					continue
				}
				
				member.IsCreator = true
				tm.Teams[id].Creator = member.PlayerId
				
				break
			}
		}
	}
	
	if len(tm.Teams[id].Members)-1 <= 0 {
		tm.Teams[id] = nil
	} else {
		tm.Teams[id].Members = remove(tm.Teams[id].Members, index)
	}
	
	player.TeamId = nil
}

func (tm *TeamManager) GetTeamIdForPlayer(player *Player) (int32, bool) {
	for id, team := range tm.Teams {
		if team == nil {
			continue
		}
		
		found := false
		
		for _, plr := range team.Members {
			if plr.PlayerId == player.DbId {
				found = true
				break
			}
		}
		
		if found {
			return int32(id), true
		}
	}
	
	return 0, false
}

func (tm *TeamManager) UpdateBrawler(player *Player) {
	if player.TeamId == nil {
		return
	}
	
	team := tm.Teams[*player.TeamId]
	
	if team == nil {
		return
	}
	
	memberIdx := -1
	
	for i, member := range team.Members {
		if member.PlayerId == player.DbId {
			memberIdx = i
			break
		}
	}
	
	if memberIdx == -1 {
		return
	}
	
	tm.Teams[*player.TeamId].Members[memberIdx].SelectedBrawler = ScId{player.SelectedCardHigh, player.SelectedCardLow}
}

func (tm *TeamManager) UpdateReady(player *Player, value bool) {
	if player.TeamId == nil {
		return
	}
	
	team := tm.Teams[*player.TeamId]
	
	if team == nil {
		return
	}
	
	memberIdx := -1
	
	for i, member := range team.Members {
		if member.PlayerId == player.DbId {
			memberIdx = i
			break
		}
	}
	
	if memberIdx == -1 {
		return
	}
	
	tm.Teams[*player.TeamId].Members[memberIdx].IsReady = value
}

func (tm *TeamManager) SetStatus(player *Player, value int16) {
	if player.TeamId == nil {
		return
	}
	
	team := tm.Teams[*player.TeamId]
	
	if team == nil {
		return
	}
	
	memberIdx := -1
	
	for i, member := range team.Members {
		if member.PlayerId == player.DbId {
			memberIdx = i
			break
		}
	}
	
	if memberIdx == -1 {
		return
	}
	
	tm.Teams[*player.TeamId].Members[memberIdx].Status = value
}

func (tm *TeamManager) AssignWrapper(wrapper *ClientWrapper) {
	player := wrapper.Player
	
	if player.TeamId == nil {
		return
	}
	
	team := tm.Teams[*player.TeamId]
	
	if team == nil {
		return
	}
	
	memberIdx := -1
	
	for i, member := range team.Members {
		if member.PlayerId == player.DbId {
			memberIdx = i
			break
		}
	}
	
	if memberIdx == -1 {
		return
	}
	
	tm.Teams[*player.TeamId].Members[memberIdx].Wrapper = wrapper
}

func (tm *TeamManager) TogglePractice(player *Player) {
	if player.TeamId == nil {
		return
	}
	
	data := tm.Teams[*player.TeamId]
	
	if data == nil {
		player.TeamId = nil
		return
	}
	
	if tm.Teams[*player.TeamId].Creator != player.DbId {
		return
	}
	
	tm.Teams[*player.TeamId].IsPractice = !tm.Teams[*player.TeamId].IsPractice
}

func (tm *TeamManager) TogglePostAd(player *Player) {
	if player.TeamId == nil {
		return
	}
	
	data := tm.Teams[*player.TeamId]
	
	if data == nil {
		player.TeamId = nil
		return
	}
	
	if tm.Teams[*player.TeamId].Creator != player.DbId {
		return
	}
	
	tm.Teams[*player.TeamId].PostAd = !tm.Teams[*player.TeamId].PostAd
}

func (tm *TeamManager) Kick(player *Player, highId int32, lowId int32) *ClientWrapper {
	if player.TeamId == nil {
		return nil
	}
	
	if tm.Teams[*player.TeamId] == nil {
		player.TeamId = nil
		return nil
	}
	
	if tm.Teams[*player.TeamId].Creator != player.DbId {
		return nil
	}
	
	id := *player.TeamId
	index := -1
	
	var wrapper *ClientWrapper
	
	for i, data := range tm.Teams[id].Members {
		if data.HighId == highId && data.LowId == lowId {
			index = i
			wrapper = data.Wrapper
			
			break
		}
	}
	
	if index == -1 {
		return nil
	}
	
	if wrapper != nil {
		wrapper.Player.TeamId = nil
	}
	
	tm.Teams[id].Members = remove(tm.Teams[id].Members, index)
	
	return wrapper
}

func (tm *TeamManager) AddMessage(player *Player, message string) {
	if player.TeamId == nil {
		return
	}
	
	id := *player.TeamId
	data := tm.Teams[id]
	
	if data == nil {
		return
	}
	
	entry := TeamMessage {
		PlayerId: player.DbId,
		PlayerName: player.Name,
		PlayerHighId: player.HighId,
		PlayerLowId: player.LowId,
		Content: message,
		Timestamp: time.Now().Unix(),
		Type: 2,
	}
	
	tm.Teams[id].Messages = append(tm.Teams[id].Messages, entry)
}

func (tm *TeamManager) AddMessageExtra(player *Player, type_ int32, event int32, targetName string, targetId int32) {
	if player.TeamId == nil {
		return
	}
	
	id := *player.TeamId
	data := tm.Teams[id]
	
	if data == nil {
		return
	}
	
	entry := TeamMessage {
		PlayerId: player.DbId,
		PlayerName: player.Name,
		PlayerHighId: player.HighId,
		PlayerLowId: player.LowId,
		Timestamp: time.Now().Unix(),
		Type: type_,
		Event: event,
		TargetName: targetName,
		TargetId: targetId,
	}
	
	tm.Teams[id].Messages = append(tm.Teams[id].Messages, entry)
}

// --- Helper functions --- //

func remove[T any](s []T, i int) []T {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
