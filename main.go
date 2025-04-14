package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/Cmolloy36/blog_aggregator/commands"
	"github.com/Cmolloy36/blog_aggregator/internal/config"
	"github.com/Cmolloy36/blog_aggregator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(fmt.Errorf("error: %w", err))
	}

	db, err := sql.Open("postgres", cfg.Db_url)
	if err != nil {
		fmt.Println(fmt.Errorf("error: %w", err))
	}

	dbQueries := database.New(db)

	st := &commands.State{
		Db:           dbQueries,
		ConfigStruct: &cfg,
	}

	commandsStruct := commands.Commands{
		FunctionMap: make(map[string]func(*commands.State, commands.Command) error),
	}

	// Handler Commands

	commandsStruct.Register("addfeed", commands.HandlerAddFeed)

	commandsStruct.Register("agg", commands.HandlerAggregator)

	commandsStruct.Register("feeds", commands.HandlerFeeds)

	commandsStruct.Register("login", commands.HandlerLogin)

	commandsStruct.Register("register", commands.HandlerRegister)

	commandsStruct.Register("reset", commands.HandlerReset)

	commandsStruct.Register("users", commands.HandlerUsers)

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
