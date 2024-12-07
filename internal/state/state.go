package state

import (
	"github.com/niccolot/BlogAggregator/internal/config"
	"github.com/niccolot/BlogAggregator/internal/database"
)

type State struct {
	Db *database.Queries
	Cfg *config.Config
}