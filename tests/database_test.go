package test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http/httptest"
	"testing"

	"server/dbconn"
	"server/mux"
)

type RequestData struct {
	Email string `json:"email"`
	PhoneNumber int64 `json:"phoneNumber"`
}

func TestDistinct(t *testing.T) {
	if err := ClearDataFromDatabase(); err != nil {
		log.Fatal(err.Error())
	}

	db := dbconn.DbConn()

	m := mux.NewMuxWithDb(db)

	l := []RequestData{
		RequestData{"email1", 123456},
		RequestData{"email2", 123457},
		RequestData{"email3", 123458},
	}

	for _, v := range l {
		j, err := json.Marshal(&v)

		if err != nil {
			t.Fatalf("json marshal failed: %s", err.Error())
		}

		req := httptest.NewRequest("POST", "/identify", bytes.NewReader(j))
		rr := httptest.NewRecorder()

		m.ServeHTTP(rr, req)

		var body mux.IdentifyResponse

		err = json.NewDecoder(rr.Result().Body).Decode(&body)

		if err != nil {
			t.Fatalf(err.Error())
		}

		if rr.Result().StatusCode != 200 {
			t.Fatalf("invalid status code")
		}
	}
}

func TestSecondary(t *testing.T) {
	if err := ClearDataFromDatabase(); err != nil {
		log.Fatal(err.Error())
	}

	db := dbconn.DbConn()

	m := mux.NewMuxWithDb(db)

	l := []RequestData{
		RequestData{"email1", 123456},
		RequestData{"email1", 123457},
		RequestData{"email2", 123458},
		RequestData{"email3", 123458},
	}

	for _, v := range l {
		j, err := json.Marshal(&v)

		if err != nil {
			t.Fatalf("json marshal failed: %s", err.Error())
		}

		req := httptest.NewRequest("POST", "/identify", bytes.NewReader(j))
		rr := httptest.NewRecorder()

		m.ServeHTTP(rr, req)

		var body mux.IdentifyResponse

		err = json.NewDecoder(rr.Result().Body).Decode(&body)

		if err != nil {
			t.Fatalf(err.Error())
		}

		if rr.Result().StatusCode != 200 {
			t.Fatalf("invalid status code")
		}

	}

	r := db.QueryRow(`select count(*) from Contact`)

	if r == nil {
		t.Fatalf("error while QueryRow")
	}

	c := -1
	r.Scan(&c)

	if c != len(l) {
		t.Fatalf("records length are unequal")
	}
}

func TestPrimaryKeyMatchesOne(t *testing.T) {
	if err := ClearDataFromDatabase(); err != nil {
		log.Fatal(err.Error())
	}

	db := dbconn.DbConn()

	m := mux.NewMuxWithDb(db)

	l := []RequestData{
		RequestData{"george@hillvalley.edu", 919191},
		RequestData{"biffsucks@hillvalley.edu", 717171},
		RequestData{"george@hillvalley.edu", 717171},
	}

	for _, v := range l {
		j, err := json.Marshal(&v)

		if err != nil {
			t.Fatalf("json marshal failed: %s", err.Error())
		}

		req := httptest.NewRequest("POST", "/identify", bytes.NewReader(j))
		rr := httptest.NewRecorder()

		m.ServeHTTP(rr, req)

		var body mux.IdentifyResponse

		err = json.NewDecoder(rr.Result().Body).Decode(&body)

		if err != nil {
			t.Fatalf(err.Error())
		}

		if rr.Result().StatusCode != 200 {
			t.Fatalf("invalid status code")
		}

	}

	r := db.QueryRow(`select count(*) from Contact`)

	if r == nil {
		t.Fatalf("error while QueryRow")
	}

	c := -1
	r.Scan(&c)

	if c != 2 {
		t.Fatalf("edge case failed, count: %d, expected: %d", c, 2)
	}
}

func TestPrimaryKeyMatchesTwo(t *testing.T) {
	if err := ClearDataFromDatabase(); err != nil {
		log.Fatal(err.Error())
	}

	db := dbconn.DbConn()

	m := mux.NewMuxWithDb(db)

	l := []RequestData{
		RequestData{"george@hillvalley.edu", 919191},
		RequestData{"george@hillvalley.edu", 229888},
		RequestData{"biffsucks@hillvalley.edu", 717171},
		RequestData{"biffsucks@hillvalley.edu", 333333},
		RequestData{"george@hillvalley.edu", 717171},
	}

	for _, v := range l {
		j, err := json.Marshal(&v)

		if err != nil {
			t.Fatalf("json marshal failed: %s", err.Error())
		}

		req := httptest.NewRequest("POST", "/identify", bytes.NewReader(j))
		rr := httptest.NewRecorder()

		m.ServeHTTP(rr, req)

		var body mux.IdentifyResponse

		err = json.NewDecoder(rr.Result().Body).Decode(&body)

		if err != nil {
			t.Fatalf(err.Error())
		}

		if rr.Result().StatusCode != 200 {
			t.Fatalf("invalid status code")
		}

	}

	r := db.QueryRow(`select count(*) from Contact`)

	if r == nil {
		t.Fatalf("error while QueryRow")
	}

	c := -1
	r.Scan(&c)

	if c != len(l) - 1 {
		t.Fatalf("edge case failed, count: %d, expected: %d", c, len(l) - 1)
	}
}
