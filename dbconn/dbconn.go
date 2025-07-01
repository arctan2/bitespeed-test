package dbconn

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
)

func DbConn() *sql.DB {
	var db *sql.DB

	user, present := os.LookupEnv("USER")
	if !present {
		log.Fatal("USER not set")
	}

	pwd, present := os.LookupEnv("PWD")
	if !present {
		log.Fatal("PWD not set")
	}

	addr, present := os.LookupEnv("ADDR")
	if !present {
		log.Fatal("ADDR not set")
	}

	dbname, present := os.LookupEnv("DB_NAME")
	if !present {
		log.Fatal("DB_NAME not set")
	}

	// cert, present := os.LookupEnv("CERT")
	// if !present {
	// 	log.Fatal("DB_NAME not set")
	// }

    cfg := mysql.NewConfig()
    cfg.User = user
    cfg.Passwd = pwd
    cfg.Net = "tcp"
    cfg.Addr = addr
    cfg.DBName = dbname
	cfg.ParseTime = true

    // Get a database handle.
    var err error
    db, err = sql.Open("mysql", cfg.FormatDSN())
    if err != nil {
		fmt.Println("Unable to make DB connection")
		return nil
    }

    pingErr := db.Ping()
    if pingErr != nil {
		fmt.Println(pingErr.Error())
		return nil
    }

	fmt.Println("connected to", cfg.DBName)

	return db
}

