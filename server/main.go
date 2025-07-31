package main

import (
	"database/sql"
	"fmt"
	"log"
	"system_api/router"
	"system_api/store"

	_ "modernc.org/sqlite"
)

const sqliteDB = "questionSystem.db"

func main() {
	db, err := sql.Open("sqlite", sqliteDB)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := store.InitDb(db); err != nil {
		fmt.Println(err)
		return
	}

	r := router.SetupRouter(db)

	r.Run(":8080")
}
