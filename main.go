package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"server/dbconn"
	"server/mux"
)

var (
	headers = map[string]string{
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers":     "Content-Type, Authorization",
		"Access-Control-Allow-Credentials": "true",
		"Content-Type":                     "application/json",
	}

	responseHeaders = HeaderMiddleware(headers)
)

func HeaderMiddleware(headers map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for key, value := range headers {
				w.Header().Set(key, value)
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func printIpv4(ADDR string) {
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)

	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			fmt.Printf("\thttp://%s%s\n", ipv4, ADDR)
		}
	}
	fmt.Printf("\thttp://localhost%s\n", ADDR)
}


func main() {
	var db *sql.DB = dbconn.DbConn()

	if db == nil {
		log.Fatal("unable to make db")
		return
	}

	m := responseHeaders(mux.NewMuxWithDb(db))

	port, present := os.LookupEnv("PORT")
	if !present {
		log.Fatal("PORT not set")
	}

	port = ":" + port

	server := http.Server{
		Addr: port,
		Handler: m,
	}

	fmt.Println("listening on: ")
	printIpv4(port)
	log.Fatal(server.ListenAndServe())
}

