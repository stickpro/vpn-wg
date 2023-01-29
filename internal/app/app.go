package app

import (
	"vpn-wg/internal/config"
	"vpn-wg/internal/store/jsondb"
)

func Run() {
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	db, err := jsondb.New("./db", cfg.Server, cfg.Global)
	if err != nil {
		panic(err)
	}

	if err := db.Init(); err != nil {
		panic(err)
	}

}
