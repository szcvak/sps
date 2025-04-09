package main

import (
	"fmt"

	"github.com/szcvak/sps/pkg/network"
)

func main() {
	server := network.NewServer("0.0.0.0:9339")
	err := server.Serve()

	if err != nil {
		fmt.Println(err)
	}
}
