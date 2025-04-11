package main

import (
	"log/slog"

	"github.com/szcvak/sps/pkg/database"
	"github.com/szcvak/sps/pkg/network"
	"github.com/szcvak/sps/pkg/csv"
)

func main() {
	if err := csv.LoadAll(); err != nil {
		slog.Error("failed to load cards!", "err", err)
	}

	dbm, err := database.NewManager()

	if err != nil {
		slog.Error("failed to connect to psql!", "err", err)
		return
	}

	defer dbm.Close()

	err = dbm.CreateDefault()

	if err != nil {
		slog.Error("failed to create default db tables!", "err", err)
		return
	}

	server := network.NewServer("0.0.0.0:9339", dbm)
	err = server.Serve()

	if err != nil {
		slog.Error("failed to start serving!", "err", err)
	}
}
