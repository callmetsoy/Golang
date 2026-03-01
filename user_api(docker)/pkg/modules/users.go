package modules

import (
	"database/sql"
	"time"
)

type User struct {
	ID        int          `db:"id" json:"id"`
	Name      string       `db:"name" json:"name"`
	Email     string       `db:"email" json:"email"`
	Password  string       `json:"password" db:"password"`
	Age       int          `db:"age" json:"age"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	DeletedAt sql.NullTime `db:"deleted_at" json:"-"`
}
