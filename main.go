package main

import (
	"fmt"
	"os"

	"github.com/Cmolloy36/blog_aggregator/internal/commands"
	"github.com/Cmolloy36/blog_aggregator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(fmt.Errorf("error: %w", err))
	}

	st := &commands.State{
		ConfigStruct: &cfg,
	}

	commandsStruct := commands.Commands{
		FunctionMap: make(map[string]func(*commands.State, commands.Command) error),
	}

	commandsStruct.Register("login", commands.HandlerLogin)

	args := os.Args
	if len(args) < 2 {
		fmt.Println(fmt.Errorf("error: provide at least 2 arguments"))
		os.Exit(1)
	}

	commandName := args[1]

	commandArgs := []string{}

	if len(args) > 2 {
		commandArgs = args[2:]
	}

	commandStruct := commands.Command{
		Name: commandName,
		Args: commandArgs,
	}

	err = commandsStruct.Run(st, commandStruct)
	if err != nil {
		fmt.Println(fmt.Errorf("%w", err))
		os.Exit(1)
	}
}
