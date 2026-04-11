package postgres

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Postgres struct {
	Conn *gorm.DB
}

func New() (*Postgres, error) {
	db, err := gorm.Open(sqlite.Open("practice7.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Postgres{Conn: db}, nil
}
