package models

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

func GetDB() (*gorm.DB, error) {
	hostname := os.Getenv("HOSTNAME")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")

	connectionString := fmt.Sprintf("host=%s user=%s dbname=collabtest sslmode=disable password=%s", hostname, user, password)
	db, err := gorm.Open("postgres", connectionString)
	if err != nil {
		logrus.Errorf("%s", err)
		return nil, err
	}
	return db, nil
}
