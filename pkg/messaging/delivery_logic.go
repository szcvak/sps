package messaging

import (
	//"log/slog"
	//"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type DeliveryLogic struct{}

func NewDeliveryLogic() *DeliveryLogic {
	return &DeliveryLogic{}
}

func (d *DeliveryLogic) GenerateRewards(box int32) {
	id, _, _ := getBoxData(box)
	rewardsAmt := 1
	
	if id == 11 { rewardsAmt = 10 }
	
	rewards := make(map[int]map[string]int32)
	
	for i := 0; i < rewardsAmt; i++ {
		rewards[i] = make(map[string]int32)
		
		rewards[i]["rarity"] = 1
		
		switch rewards[i]["rarity"] {
		case 1:
			
		}
	}
}

// --- Helpers --- //

func getBoxData(id int32) (int32, int32, int32) {
	if id == 1 {
		return 10, database.CurrencyCoins, 100
	} else if id == 2 {
		return 10, database.CurrencyGems, 10
	} else if id == 3 {
		return 11, database.CurrencyGems, 80
	}
	
	return id, database.CurrencyCoins, 0
}

func getRarityFromId(rarity int32) string {
	switch rarity {
	case 0:
		return "common"
	case 1:
		return "rare"
	case 2:
		return "super_rare"
	case 3:
		return "epic"
	case 4:
		return "mega_epic"
	}
	
	return "legendary"
}
