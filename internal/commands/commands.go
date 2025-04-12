package commands

import (
	"fmt"

	"github.com/Cmolloy36/blog_aggregator/internal/config"
)

type State struct {
	ConfigStruct *config.Config
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	FunctionMap map[string]func(*State, Command) error
}

func HandlerLogin(s *State, cmd Command) error {

	if len(cmd.Args) == 0 {
		return fmt.Errorf("error: \"login\" expects a username argument")
	}

	s.ConfigStruct.Current_user_name = cmd.Args[0]
	// fmt.Printf("%v", s.ConfigStruct.Current_user_name)
	fmt.Printf("The user has been set: %s\n", s.ConfigStruct.Current_user_name)
	s.ConfigStruct.SetUser(s.ConfigStruct.Current_user_name)
	return nil
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
