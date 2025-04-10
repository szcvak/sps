package csv

import (
	"fmt"
	"os"
	"io"
	"encoding/csv"
)

var (
	cardsCsv map[int][]string = nil
	locationsCsv map[int][]string = nil
)

// --- Private methods --- //

func initCards() error {
	if cardsCsv != nil {
		return nil
	}

	fmt.Println("initializing cards.csv")

	file, err := os.Open("assets/csv_logic/cards.csv")

	if err != nil {
		return fmt.Errorf("failed to load cards.csv: %w\n", err)
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

func initLocations() error {
	if locationsCsv != nil {
		return nil
	}

	fmt.Println("initializing locations.csv")

	file, err := os.Open("assets/csv_logic/locations.csv")

	if err != nil {
		return fmt.Errorf("failed to load locations.csv: %w\n", err)
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
		_, _ = fmt.Fprintf(os.Stderr, "can't call CardIds() because cards.csv has not been initialized yet\n")
		return make([]int, 0)
	}

	temp := make([]int, 0)

	for id, _ := range cardsCsv {
		temp = append(temp, id)
	}

	return temp
}

func LocationIds() []int {
	if locationsCsv == nil {
		_, _ = fmt.Fprintf(os.Stderr, "can't call LocationIds() because locations.csv has not been initialized yet\n")
		return make([]int, 0)
	}

	temp := make([]int, 0)

	for id, _ := range locationsCsv {
		temp = append(temp, id)
	}

	return temp
}
