package main

import (
	"fmt"

	"github.com/niccolot/BlogAggregator/internal/config"
)

type state struct {
	cfg *config.Config
}

type command struct {
	cmdName string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.cmdName]
	if !ok {
		return fmt.Errorf("error: command %s not found", cmd.cmdName)
	}

	errCmd := handler(s,cmd)
	if errCmd != nil {
		return fmt.Errorf("error while executing '%s' command: %v", cmd.cmdName, errCmd)
	}
	
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("usage: login <username>")
	}

	errSet := s.cfg.SetUser(cmd.args[0])
	if errSet != nil {
		return fmt.Errorf("error setting username: %v", errSet)
	}

	fmt.Println("username has been correctly set")

	return nil
}