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
)

func main() {
	errEnv := godotenv.Load()
	if errEnv != nil {
		log.Fatalf(fmt.Sprintf("error loading environment variables: %v", errEnv))
	}

	dbURL := os.Getenv("DB_URL")
	db, errDB := sql.Open("postgres", dbURL)
	if errDB != nil {
		log.Fatalf(fmt.Sprintf("error creating APIconfig: %v", errDB))
	}
	
	defer db.Close()

	dbQueries := database.New(db)

	cfg := config.Read()

	s := state.State{
		Db: dbQueries,
		Cfg: cfg,
	}

	cmds := commands.Commands{}
	cmds.Init()

	cliArgs := os.Args
	if len(cliArgs) < 2 {
		log.Fatal("enter a valid command")
	}

	cmd := commands.Command{
		CmdName: cliArgs[1],
		Args: cliArgs[2:],
	}

	errCmd := cmds.Run(&s, cmd)
	if errCmd != nil {
		log.Fatal(errCmd)
	}
}