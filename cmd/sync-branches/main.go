package main

import (
	"autoMerger/internal/application/sync"
	"autoMerger/internal/interfaces/cli"
	"autoMerger/internal/interfaces/controller"
	"fmt"
	"os"
)

func main() {
	
	args := cli.ParseArgs(os.Args[1:])

	if args.ShowHelp {
		cli.PrintHelp()
		return
	}
	
    service := sync.NewSyncService()
    controller := controller.NewSyncController(service)
    if err := controller.Run(args); err != nil {
		fmt.Printf("error %s", err)
	}
}