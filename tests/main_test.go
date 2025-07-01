package test

import (
	"fmt"
	"log"
	"testing"
	"server/dbconn"
)

func ClearDataFromDatabase() (error) {
	db := dbconn.DbConn()

	query := `truncate table Contact`

	_, err := db.Exec(query)

	if err != nil {
		return err
	}

	fmt.Println("Truncated all tables successfully.")
	return nil
}

func TestMain(m *testing.M) {
	if err := ClearDataFromDatabase(); err != nil {
		log.Fatal(err.Error())
	}
	m.Run()
}
