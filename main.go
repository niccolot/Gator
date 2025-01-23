package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/niccolot/BlogAggregator/internal/commands"
	"github.com/niccolot/BlogAggregator/internal/config"
	"github.com/niccolot/BlogAggregator/internal/database"
	"github.com/niccolot/BlogAggregator/internal/state"
	"github.com/peterh/liner"
)

func main() {
	errEnv := godotenv.Load()
	if errEnv != nil {
		log.Fatalf(fmt.Sprintf("error loading environment variables: %v", errEnv))
	}

	dbURL := os.Getenv("DB_URL")
	db, errDB := sql.Open("postgres", dbURL)
	if errDB != nil {
		log.Fatalf(fmt.Sprintf("error opening database: %v", errDB))
	}
	
	defer db.Close()

	dbQueries := database.New(db)
	cfg := config.Read()

	s := state.State{
		Db: dbQueries,
		Cfg: cfg,
		Aggregating: false,
		StopAggregation: make(chan bool),
	}

	cmds := commands.Commands{}
	cmds.Init()

	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	// cli logger
	var warning = log.New(os.Stdout, "\u001b[33mWARNING: \u001B[0m", 0)

	for {
		fmt.Println()
		input, err := line.Prompt("Gator > ")
		if err != nil {
			if err == liner.ErrPromptAborted {
				break
			}
			fmt.Println("Error reading line:", err)
			continue
		}

		line.AppendHistory(input)
		s.Cfg.CmdHistory = append(s.Cfg.CmdHistory, input)

		cmdName, args := commands.ParseInput(input)
		if cmdName == "exit" || cmdName == "q" || cmdName == "quit" {
			break
		}

		cmd := commands.Command{
			CmdName: cmdName,
			Args: args,
		}

		commands.Run(cmd, &cmds, &s, warning)
		
		fmt.Println()
	}
}