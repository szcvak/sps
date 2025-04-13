package csv

import "fmt"

func LoadAll() error {
	if err := loadCards(); err != nil {
		return fmt.Errorf("failed to load cards: %w\n", err)
	}

	if err := loadLocations(); err != nil {
		return fmt.Errorf("failed to load locations: %w\n", err)
	}

	if err := loadCharacters(); err != nil {
		return fmt.Errorf("failed to load characters: %w\n", err)
	}

	if err := loadThumbnails(); err != nil {
		return fmt.Errorf("failed to load thumbnails: %w\n", err)
	}

	return nil
}
