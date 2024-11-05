package main

import (
	"encoding/json"
	"fmt"

	"github.com/niccolot/BlogAggregator/internal/config"
)

func main() {
	cfg := config.Read()
	cfg.SetUser("nico")
	cfg = config.Read()

	dat, _ := json.Marshal(cfg)
	fmt.Println(string(dat))
}