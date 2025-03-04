package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"fmt"

	_ "github.com/go-sql-driver/mysql"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type RatingCards struct {
	ID       string `json:"id" gorm:"primaryKey"`
	Question string `json:"question"`
	Category string `json:"category"`
	OrderId  int32  `json:"orderId"`
}

var db *gorm.DB

func GetDb() *gorm.DB {
	return db
}

func init() {
	// hostname := "mysql.foxnhound.mysql.database.azure.com"
	hostname := "foxnhound-mysql-servera899faa2.mysql.database.azure.com"
	port := "3306"
	username := "sqladmin_jH5JKsj_54KJH"
	password := "jHGJ7JKsd(sjd)jkh%"
	dbname := "foxnhound-db"
	caCertPath := "certs/DigiCertGlobalRootCA.crt.pem"
	// dsn := "devuser:devpassword@tcp(127.0.0.1:3306)/fox_and_hound?charset=utf8mb4&parseTime=True&loc=Local"

	// Load the CA certificate
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Fatalf("Failed to read CA certificate: %v", err)
		return
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Register a custom TLS config
	err = mysqlDriver.RegisterTLSConfig("custom", &tls.Config{
		RootCAs: caCertPool,
	})
	if err != nil {
		log.Fatalf("Failed to register custom TLS config: %v", err)
		return
	}

	// DNS resolution debug code
	addrs, err := net.LookupHost(hostname)
	if err != nil {
		log.Fatalf("Failed to resolve hostname %s: %v", hostname, err)
	}
	log.Printf("Resolved hostname %s to addresses: %v", hostname, addrs)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=custom", username, password, hostname, port, dbname)
	log.Printf("Connecting to database with DSN: %s", dsn)
	db, err = gorm.Open(mysql.New(mysql.Config{
		DriverName: "mysql",
		DSN:        dsn,
	}))
		
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := db.AutoMigrate(&RatingCards{}); err != nil {
		log.Fatal("Failed to migrate database schema:", err)
	}
}

func getRatingCardDtoObject(w http.ResponseWriter, r *http.Request) {
	responses := getRatingCards()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

func getRatingCards() []RatingCards {
	var ratingCards []RatingCards
	result := db.Find(&ratingCards)
	if result.Error != nil {
		log.Println("Error fetching rating cards:", result.Error)
		return []RatingCards{}
	}
	log.Printf("Found %d rating cards in the database", len(ratingCards))
	return ratingCards
}
