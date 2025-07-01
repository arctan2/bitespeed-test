package mux

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type MuxWithDb struct {
	db *sql.DB
	mux *http.ServeMux
}

type ContactField struct {
	PrimaryContatctId int64 `json:"primaryContatctId"`
	Emails []string `json:"emails"`
	PhoneNumbers []string `json:"phoneNumbers"`
	SecondaryContactIds []int64 `json:"secondaryContactIds"`
}

type IdentifyResponse struct {
	Contact ContactField `json:"contact"`
}

func NewMuxWithDb(db *sql.DB) *http.ServeMux {
	m := &MuxWithDb{db, http.NewServeMux()}
	m.setupRoutes()
	return m.mux
}

func (m *MuxWithDb) setupRoutes() {
	m.mux.HandleFunc("POST /identify", m.identifyRoute)
}

type Contact struct {
	Id int64 `json:"id"`
	PhoneNumber string `json:"phone_number"`
	Email string `json:"email"`
	LinkedId sql.NullInt64 `json:"linked_id"`
	LinkedPrecendence string `json:"linked_precendence"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at"`
}

func (m *MuxWithDb) insertNewContact(ph, email, preced string, linkedId int64) int64 {
	q := `INSERT INTO Contact (phone_number, email, linked_precendence, linked_id) VALUES (?, ?, ?, ?)`

	var l *int64 = &linkedId

	if linkedId == -1 {
		l = nil
	}

	result, err := m.db.Exec(q, ph, email, preced, l)

	if err != nil {
		return -1
	}

	id, err := result.LastInsertId()

	if err != nil {
		return -1
	}

	return id
}

func (m *MuxWithDb) identifyRoute(w http.ResponseWriter, r *http.Request) {
	var (
		primaryContatctId int64
		emails []string = []string{}
		phoneNumbers []string = []string{}
		secondaryContactIds []int64 = []int64{}
	)

	var body struct {
		Email string `json:"email"`
		Ph int64 `json:"phoneNumber"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		fmt.Println("error decoding body:", err.Error())
		return
	}

	if body.Email == "" || body.Ph == 0 {
		fmt.Println("email or phoneNumber is empty")
		return
	}

	q := `select
		id,
		phone_number,
		email,
		linked_id,
		linked_precendence,
		created_at,
		updated_at,
		deleted_at
	from Contact where email = ? or phone_number = ?`

	rows, err := m.db.Query(q, body.Email, body.Ph)

	if err != nil {
		fmt.Println("error getting contacts: ", err.Error())
		return
	}

	defer rows.Close()

	var identified []Contact

	for rows.Next() {
		c := Contact{}

		if err := rows.Scan(
			&c.Id,
			&c.PhoneNumber,
			&c.Email,
			&c.LinkedId,
			&c.LinkedPrecendence,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.DeletedAt,
		); err != nil {
			fmt.Println("error scanning row: ", err.Error())
			return
		}

		identified = append(identified, c)
	}

	ph := strconv.FormatInt(body.Ph, 10)

	if len(identified) == 0 {
		lastId := m.insertNewContact(ph, body.Email, "primary", -1)
		if lastId == -1 {
			fmt.Println("lastId is -1 after insert.")
			return
		}

		primaryContatctId = lastId
		emails = append(emails, body.Email)
		phoneNumbers = append(phoneNumbers, ph)
	} else {
		primaries := []int{}

		for idx, it := range identified {
			if it.LinkedPrecendence == "primary" {
				primaries = append(primaries, idx)
			}
		}

		if len(primaries) == 1 {
			primary := identified[0]

			if primary.Email == body.Email && primary.PhoneNumber == ph {
				fmt.Println("exact duplicate record")
				return
			}

			lastId := m.insertNewContact(ph, body.Email, "secondary", primary.Id)
			if lastId == -1 {
				fmt.Println("lastId is -1 after insert.")
				return
			}

			q := `select
				id,
				phone_number,
				email,
				linked_id,
				linked_precendence,
				created_at,
				updated_at,
				deleted_at
			from Contact where id = ?`

			r := m.db.QueryRow(q, lastId)

			if r == nil {
				fmt.Println("row is nil")
				return
			}

			var n Contact

			if err := r.Scan(
				&n.Id,
				&n.PhoneNumber,
				&n.Email,
				&n.LinkedId,
				&n.LinkedPrecendence,
				&n.CreatedAt,
				&n.UpdatedAt,
				&n.DeletedAt,
			); err != nil {
				fmt.Println(err.Error())
				return
			}

			identified = append(identified, n)
			primaryContatctId = primary.Id
		} else {
			c1 := &identified[primaries[0]]
			c2 := &identified[primaries[1]]

			if c1.Email == body.Email && c1.PhoneNumber == ph {
				fmt.Println("exact match in 2 primary situation")
				return
			}

			if c2.Email == body.Email && c2.PhoneNumber == ph {
				fmt.Println("exact match in 2 primary situation")
				return
			}

			q := `update Contact set linked_precendence = 'secondary', linked_id = ?, updated_at = NOW() where id = ?`

			if c1.CreatedAt.After(c2.CreatedAt) {
				c1, c2 = c2, c1
			}

			primaryContatctId = c1.Id
			secondaryId := c2.Id
			c2.LinkedPrecendence = "secondary"

			_, err := m.db.Exec(q, primaryContatctId, secondaryId)

			if err != nil {
				fmt.Println(err.Error())
				return
			}

			_, err = m.db.Exec(`update Contact set linked_id = ?, updated_at = NOW() where linked_id = ?`, primaryContatctId, c2.Id)

			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}

		emailSet := StringSet{}
		phoneNumbersSet := StringSet{}

		for _, v := range identified {
			if v.LinkedPrecendence == "primary" {
				continue
			}

			emailSet.Add(v.Email)
			phoneNumbersSet.Add(v.PhoneNumber)
			secondaryContactIds = append(secondaryContactIds, v.Id)
		}
	}

	res := IdentifyResponse {
		Contact: ContactField{
			primaryContatctId,
			emails,
			phoneNumbers,
			secondaryContactIds,
		},
	}

	json.NewEncoder(w).Encode(&res)
}
