package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/csv"
	"github.com/szcvak/sps/pkg/database"
	"github.com/szcvak/sps/pkg/hub"
	"github.com/szcvak/sps/pkg/network"
)

func main() {
	if err := csv.LoadAll(); err != nil {
		slog.Error("failed to load cards!", "err", err)
		return
	}

	core.InitEventManager(core.DefaultSchedules())
	core.InitTeamManager()
	
	hub.InitHub()

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
	errChan := make(chan error, 1)

	go func() {
		errChan <- server.Serve()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case _ = <-stop:
		server.Close()
	case err := <-errChan:
		slog.Error("faled to serve!", "err", err)
	}

	if em := core.GetEventManager(); em != nil {
		em.Close()
	}
}
