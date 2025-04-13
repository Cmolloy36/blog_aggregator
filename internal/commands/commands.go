package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Cmolloy36/blog_aggregator/internal/config"
	"github.com/Cmolloy36/blog_aggregator/internal/database"
	"github.com/google/uuid"
)

type State struct {
	Db           *database.Queries
	ConfigStruct *config.Config
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	FunctionMap map[string]func(*State, Command) error
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.FunctionMap[name] = f
}

func (c *Commands) Run(s *State, cmd Command) error {
	fcn, ok := c.FunctionMap[cmd.Name]
	if !ok {
		return fmt.Errorf("error: \"%s\" is not registered", cmd.Name)
	}

	err := fcn(s, cmd)
	if err != nil {
		return err
	}

	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	if len(cmd.Args) != 0 {
		log.Fatalf("error: \"list\" does not expect an additional argument")
	}

	usersList, err := s.Db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if len(usersList) == 0 {
		return fmt.Errorf("there are no users in the database")
	}

	for _, user := range usersList {
		append := ""
		if user == s.ConfigStruct.Current_user_name {
			append = " (current)"
		}
		fmt.Printf("* %s\n", user+append)
	}

	return nil
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		log.Fatalf("error: \"login\" expects a username argument")
	}

	name := cmd.Args[0]

	_, err := s.Db.GetUser(context.Background(), name)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Fatalf("unexpected error occurred: %v", err)
		} else {
			log.Fatalf("%s does not exist!", name)
		}
	}

	s.ConfigStruct.Current_user_name = name
	// fmt.Printf("%v", s.ConfigStruct.Current_user_name)
	fmt.Printf("The user has been set: %s\n", s.ConfigStruct.Current_user_name)
	s.ConfigStruct.SetUser(s.ConfigStruct.Current_user_name)
	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("error: \"register\" expects a username argument")
	}

	name := cmd.Args[0]

	emptyUser := database.User{}

	user, err := s.Db.GetUser(context.Background(), name)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Fatalf("unexpected error occurred: %v", err)
		}
	} else if user != emptyUser {
		log.Fatalf("%s already exists!", name)
	}

	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	}

	s.Db.CreateUser(context.Background(), userParams)

	s.ConfigStruct.Current_user_name = cmd.Args[0]
	// fmt.Printf("%v", s.ConfigStruct.Current_user_name)
	fmt.Printf("The user has been registered: %s\n", s.ConfigStruct.Current_user_name)
	s.ConfigStruct.SetUser(s.ConfigStruct.Current_user_name)
	return nil
}

func HandlerReset(s *State, cmd Command) error {
	err := s.Db.ResetUsers(context.Background())
	if err != nil {
		log.Fatalf("unexpected error occurred: %v", err)
	}

	fmt.Println("The database has been reset.")

	return nil
}
