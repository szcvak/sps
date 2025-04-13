package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
)

var (
	cardsCsv      map[int][]string = nil
	skinsCsv      map[int][]string = nil
	locationsCsv  map[int][]string = nil
	charactersCsv map[int][]string = nil
	thumbnailsCsv map[int][]string = nil
)

// --- Private methods --- //

func loadCards() error {
	if cardsCsv != nil {
		return nil
	}

	slog.Info("loading cards.csv")

	file, err := os.Open("assets/csv_logic/cards.csv")

	if err != nil {
		return err
	}

	defer file.Close()

	cardsCsv = make(map[int][]string)

	reader := csv.NewReader(file)
	line := 0

	for {
		record, err_ := reader.Read()

		if err_ != nil {
			if err_ == io.EOF {
				break
			} else {
				return fmt.Errorf("error during reading: %w\n", err_)
			}
		}

		if line < 2 {
			line++
			continue
		}

		cardsCsv[line-2] = record

		line++
	}

	return nil
}

func loadLocations() error {
	if locationsCsv != nil {
		return nil
	}

	slog.Info("loading locations.csv")

	file, err := os.Open("assets/csv_logic/locations.csv")

	if err != nil {
		return err
	}

	defer file.Close()

	locationsCsv = make(map[int][]string)

	reader := csv.NewReader(file)
	line := 0

	for {
		record, err_ := reader.Read()

		if err_ != nil {
			if err_ == io.EOF {
				break
			} else {
				return fmt.Errorf("error during reading: %w\n", err_)
			}
		}

		if line < 2 {
			line++
			continue
		}

		locationsCsv[line-2] = record

		line++
	}

	return nil
}

func loadCharacters() error {
	if charactersCsv != nil {
		return nil
	}

	slog.Info("loading characters.csv")

	file, err := os.Open("assets/csv_logic/characters.csv")

	if err != nil {
		return err
	}

	defer file.Close()

	charactersCsv = make(map[int][]string)

	reader := csv.NewReader(file)
	line := 0

	for {
		record, err_ := reader.Read()

		if err_ != nil {
			if err_ == io.EOF {
				break
			} else {
				return fmt.Errorf("error during reading: %w\n", err_)
			}
		}

		if line < 2 {
			line++
			continue
		}

		charactersCsv[line-2] = record

		line++
	}

	return nil
}

func loadThumbnails() error {
	if thumbnailsCsv != nil {
		return nil
	}

	slog.Info("loading player_thumbnails.csv")

	file, err := os.Open("assets/csv_logic/player_thumbnails.csv")

	if err != nil {
		return err
	}

	defer file.Close()

	thumbnailsCsv = make(map[int][]string)

	reader := csv.NewReader(file)
	line := 0

	for {
		record, err_ := reader.Read()

		if err_ != nil {
			if err_ == io.EOF {
				break
			} else {
				return fmt.Errorf("error during reading: %w\n", err_)
			}
		}

		if line < 2 {
			line++
			continue
		}

		thumbnailsCsv[line-2] = record

		line++
	}

	return nil
}

func loadSkins() error {
	if skinsCsv != nil {
		return nil
	}

	slog.Info("loading skins.csv")

	file, err := os.Open("assets/csv_logic/skins.csv")

	if err != nil {
		return err
	}

	defer file.Close()

	skinsCsv = make(map[int][]string)

	reader := csv.NewReader(file)
	line := 0

	for {
		record, err_ := reader.Read()

		if err_ != nil {
			if err_ == io.EOF {
				break
			} else {
				return fmt.Errorf("error during reading: %w\n", err_)
			}
		}

		if line < 2 {
			line++
			continue
		}

		skinsCsv[line-2] = record

		line++
	}

	return nil
}

// --- Public methods --- //

func CardIds() []int {
	if cardsCsv == nil {
		slog.Error("cards.csv has not been loaded yet!")
		return make([]int, 0)
	}

	temp := make([]int, 0)

	for id, _ := range cardsCsv {
		temp = append(temp, id)
	}

	return temp
}

func IsCardUnlocked(card int) bool {
	if cardsCsv == nil {
		slog.Error("cards.csv has not been loaded yet!")
		return false
	}

	for line, row := range cardsCsv {
		if line != card {
			continue
		}

		data := row[5]

		return data == "unlock"
	}

	slog.Error("failed to find card!", "cardId", card)
	return false
}

func GetBrawlersWithRarity(rarity string) []int32 {
	if cardsCsv == nil {
		slog.Error("cards.csv has not been loaded yet!")
		return []int32{}
	}

	brawlers := make([]int32, 0)

	for line, row := range cardsCsv {
		if row[5] == "unlock" && row[10] == rarity {
			brawlers = append(brawlers, int32(line))
		}
	}

	return brawlers
}

func GetBrawlerId(card int32) int32 {
	if cardsCsv == nil {
		slog.Error("cards.csv has not been loaded yet!")
		return -1
	}

	brawler := ""

	for line, row := range cardsCsv {
		if int32(line) == card {
			brawler = row[3]
			break
		}
	}

	return GetCharacterIdByName(brawler)
}

func LocationIds() []int32 {
	if locationsCsv == nil {
		slog.Error("locations.csv has not been loaded yet!")
		return make([]int32, 0)
	}

	temp := make([]int32, 0)

	for id, _ := range locationsCsv {
		temp = append(temp, int32(id))
	}

	return temp
}

func GetLocationsByGamemode(gamemode string) []int32 {
	if locationsCsv == nil {
		slog.Error("locations.csv has not been loaded yet!")
		return []int32{}
	}

	ids := make([]int32, 0)

	for line, row := range locationsCsv {
		if row[11] == gamemode {
			ids = append(ids, int32(line))
		}
	}

	return ids
}

func GetCharacterIdByName(name string) int32 {
	if charactersCsv == nil {
		slog.Error("characters.csv has not been loaded yet!")
		return 0
	}

	for line, row := range charactersCsv {
		if row[0] == name {
			return int32(line)
		}
	}

	slog.Error("failed to find character!", "name", name, "err", "not found in file")
	return 0
}

func GetCardForCharacter(charId int32) (int32, bool) {
	if charactersCsv == nil || cardsCsv == nil {
		slog.Error("characters.csv or cards.csv not loaded yet!")
		return -1, false
	}

	charName := ""
	charRow, charFound := charactersCsv[int(charId)]

	if !charFound || len(charRow) == 0 {
		slog.Warn("character id not found in characters.csv", "charId", charId)
		return -1, false
	}

	charName = charRow[0]

	if charName == "" {
		slog.Warn("character name is empty in characters.csv", "charId", charId)
		return -1, false
	}

	for cardIdInt, cardRow := range cardsCsv {
		if len(cardRow) > 5 && cardRow[5] == "unlock" && len(cardRow) > 3 && cardRow[3] == charName {
			return int32(cardIdInt), true
		}
	}

	slog.Warn("no card found for character", "charId", charId, "charName", charName)

	return -1, false
}

func GetTrophiesForThumbnail(id int32) int32 {
	if thumbnailsCsv == nil {
		slog.Error("player_thumbnails.csv has not been loaded yet!")
		return 0
	}

	for line, row := range thumbnailsCsv {
		if int32(line) == id {
			amount, err := strconv.Atoi(row[2])

			if err != nil {
				slog.Error("failed to convert trophies thumbnail to int!", "err", err)
				return 0
			}

			return int32(amount)
		}
	}

	slog.Error("failed to find trophies for thumbnail!", "id", id, "err", "not found in file")
	return 0
}

func GetBrawlerForThumbnail(id int32) string {
	if thumbnailsCsv == nil {
		slog.Error("player_thumbnails.csv has not been loaded yet!")
		return ""
	}

	for line, row := range thumbnailsCsv {
		if int32(line) == id {
			return row[4]
		}
	}

	slog.Error("failed to find brawler for thumbnail!", "id", id, "err", "not found in file")
	return ""
}

func GetExperienceLevelForThumbnail(id int32) int32 {
	if thumbnailsCsv == nil {
		slog.Error("player_thumbnails.csv has not been loaded yet!")
		return 0
	}
	
	if int(id) > len(thumbnailsCsv) {
		slog.Error("thumbnail out of range")
		return 0
	}
	
	row, exists := thumbnailsCsv[int(id)]
	
	if !exists {
		slog.Error("thumbnail does not exist")
		return 0
	}

	amount, err := strconv.Atoi(row[1])

	if err != nil {
		slog.Error("failed to convert thumbnail experience to int!", "err", err)
		return 0
	}

	return int32(amount)

	slog.Error("failed to find experience for thumbnail!", "id", id, "err", "not found in file")
	return 0
}

func Thumbnails() []int32 {
	if thumbnailsCsv == nil {
		slog.Error("player_thumbnail.csv has not been loaded yet!")
		return make([]int32, 0)
	}

	temp := make([]int32, 0)

	for id, _ := range thumbnailsCsv {
		temp = append(temp, int32(id))
	}

	return temp
}

func Skins() []int32 {
	if skinsCsv == nil {
		slog.Error("skins.csv has not been loaded yet!")
		return make([]int32, 0)
	}

	temp := make([]int32, 0)

	for id, _ := range skinsCsv {
		temp = append(temp, int32(id))
	}

	return temp
}

func IsSkinDefault(id int32) bool {
	if charactersCsv == nil {
		slog.Error("characters.csv has not been loaded yet!")
		return false
	}
	
	if skinsCsv == nil {
		slog.Error("skins.csv has not been loaded yet!")
		return false
	}
	
	defaultSkins := make(map[string]bool)
	
	for _, row := range charactersCsv {
		defaultSkins[row[20]] = true
	}
	
	if int(id) > len(skinsCsv) {
		slog.Error("skin out of range")
		return false
	}
	
	name := skinsCsv[int(id)][0]
	_, e :=defaultSkins[name]
	
	return e
}

func GetSkinPrice(id int32) int32 {
	if skinsCsv == nil {
		slog.Error("skins.csv has not been loaded yet!")
		return 0
	}
	
	if int(id) > len(skinsCsv) {
		return 0
	}

	row, exists := skinsCsv[int(id)]
	
	if !exists {
		slog.Error("row does not exist!")
		return 0
	}
	
	price, err := strconv.Atoi(row[3])
	
	if err != nil {
		slog.Error("failed to convert to int!", "err", err)
		return 0
	}

	return int32(price)
}

func GetBrawlerForSkin(id int32) int32 {
	if charactersCsv == nil {
		slog.Error("characters.csv has not been loaded yet!")
		return 0
	}
	
	if skinsCsv == nil {
		slog.Error("skins.csv has not been loaded yet!")
		return 0
	}
	
	if int(id) > len(skinsCsv) {
		slog.Error("skin out of range")
		return 0
	}
	
	row, exists := skinsCsv[int(id)]
	
	if !exists {
		slog.Error("skin does not exist")
		return 0
	}
	
	character := row[1]
	
	for i, c := range charactersCsv {
		if c[0] == character {
			return int32(i)
		}
	}
	
	return 0
}
