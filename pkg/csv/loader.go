package csv

import "fmt"

func LoadAll() error {
	if err := initCards(); err != nil {
		return fmt.Errorf("failed to load: %w\n", err)
	}

	if err := initLocations(); err != nil {
		return fmt.Errorf("failed to load: %w\n", err)
	}

	return nil
}
