package models

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

func GetDB() (*gorm.DB, error) {
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbDomain := os.Getenv("POSTGRES_DOMAIN")
	dbPort := os.Getenv("POSTGRES_PORT")

	connectionString := fmt.Sprintf("host=%s user=%s dbname=collabtest sslmode=disable password=%s port=%s", dbDomain, user, password, dbPort)
	db, err := gorm.Open("postgres", connectionString)
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err, "connection string": connectionString}).Error("Could not connect to database")
		return nil, err
	}
	return db, nil
}

type Project struct {
	gorm.Model
	User User
	Hash string `gorm:"primary_key"`
	Name string
	Runs []Run
}

type Run struct {
	gorm.Model
	Project Project
	Stdout  string
	Stderr  string
}

type User struct {
	gorm.Model
	GithubId uint
}
