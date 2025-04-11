package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
)

var (
	cardsCsv     map[int][]string = nil
	locationsCsv map[int][]string = nil
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

func LocationIds() []int {
	if locationsCsv == nil {
		slog.Error("locations.csv has not been loaded yet!")
		return make([]int, 0)
	}

	temp := make([]int, 0)

	for id, _ := range locationsCsv {
		temp = append(temp, id)
	}

	return temp
}
