package main

import (
	"fmt"
	"os"

	"github.com/szcvak/sps/pkg/database"
	"github.com/szcvak/sps/pkg/network"
)

func main() {
	dbm, err := database.NewManager()

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unable to connect to database: %v\n", err)
		return
	}

	defer dbm.Close()

	err = dbm.CreateDefault()

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create default tables: %v\n", err)
		return
	}

	server := network.NewServer("0.0.0.0:9339", dbm)
	err = server.Serve()

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unable to start server: %v\n", err)
	}
}
