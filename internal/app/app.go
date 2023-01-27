package app

import (
	"vpn-wg/internal/store/jsondb"
)

func Run() {
	db, err := jsondb.New("./db")
	if err != nil {
		panic(err)
	}
	if err := db.Init(); err != nil {
		panic(err)
	}

}
