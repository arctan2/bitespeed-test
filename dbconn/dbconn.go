package dbconn

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
)

var TLS_NAME = "my_custom_tls"

func registerTLSConfig() error {
	rootCertPool := x509.NewCertPool()

	certPath, present := os.LookupEnv("CERT_PATH")
	if !present {
		log.Fatal("CERT_PATH not set")
	}

	fmt.Println("registering: ", certPath)
	
	pem, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read CA file: %w", err)
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		return fmt.Errorf("failed to append CA cert")
	}

	tlsConfig := &tls.Config{
		RootCAs: rootCertPool,
	}

	err = mysql.RegisterTLSConfig(TLS_NAME, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to register TLS config: %w", err)
	}
	fmt.Println("registered tls!")
	return nil
}

func DbConn() *sql.DB {
	var db *sql.DB

	user, present := os.LookupEnv("DB_USER")
	if !present {
		log.Fatal("DB_USER not set")
	}

	pwd, present := os.LookupEnv("DB_PASSWORD")
	if !present {
		log.Fatal("DB_PASSWORD not set")
	}

	addr, present := os.LookupEnv("DB_ADDR")
	if !present {
		log.Fatal("DB_ADDR not set")
	}

	dbname, present := os.LookupEnv("DB_NAME")
	if !present {
		log.Fatal("DB_NAME not set")
	}

    cfg := mysql.NewConfig()
    cfg.User = user
    cfg.Passwd = pwd
    cfg.Net = "tcp"
    cfg.Addr = addr
    cfg.DBName = dbname
	cfg.ParseTime = true
	cfg.TLSConfig = TLS_NAME

	err := registerTLSConfig()
	if err != nil {
		log.Fatal("TLS setup failed:", err)
	}

    db, err = sql.Open("mysql", cfg.FormatDSN())
    if err != nil {
		fmt.Println("Unable to make DB connection")
		return nil
    }

    pingErr := db.Ping()
    if pingErr != nil {
		fmt.Println("ping error:", pingErr.Error())
		return nil
    }

	fmt.Println("connected to", cfg.DBName)

	return db
}

