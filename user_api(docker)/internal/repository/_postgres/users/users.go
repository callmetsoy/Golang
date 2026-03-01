package users

import (
	"fmt"
	"user_api/internal/repository/_postgres"
	"user_api/pkg/modules"
)

type Repository struct {
	db *_postgres.PGXDialect
}

func NewUserRepository(db *_postgres.PGXDialect) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(u modules.User) (int, error) {
	tx, err := r.db.DB.Begin()
	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	var id int
	query := `INSERT INTO users (name, email, password, age) VALUES ($1, $2, $3, $4) RETURNING id`
	err = tx.QueryRow(query, u.Name, u.Email, u.Password, u.Age).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	auditQuery := `INSERT INTO audit_logs (user_id, action) VALUES ($1, $2)`
	_, err = tx.Exec(auditQuery, id, "USER_CREATED")
	if err != nil {
		return 0, fmt.Errorf("failed to insert audit log: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) GetByID(id int) (*modules.User, error) {
	var user modules.User

	err := r.db.DB.Get(&user, "SELECT * FROM users WHERE id=$1 AND deleted_at IS NULL", id)
	if err != nil {
		return nil, fmt.Errorf("user %d not found", id)
	}
	return &user, nil
}

func (r *Repository) GetAll() ([]modules.User, error) {
	var users []modules.User

	err := r.db.DB.Select(&users, "SELECT * FROM users WHERE deleted_at IS NULL")
	return users, err
}

func (r *Repository) Update(u modules.User) error {

	tx, err := r.db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		"UPDATE users SET name=$1, email=$2, age=$3 WHERE id=$4 AND deleted_at IS NULL",
		u.Name, u.Email, u.Age, u.ID,
	)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("update failed: user not found")
	}

	auditQuery := `INSERT INTO audit_logs (user_id, action) VALUES ($1, $2)`
	_, err = tx.Exec(auditQuery, u.ID, "USER_UPDATED")
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}

	return tx.Commit()
}
func (r *Repository) Delete(id int) (int64, error) {

	tx, err := r.db.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec("UPDATE users SET deleted_at = NOW() WHERE id=$1 AND deleted_at IS NULL", id)
	if err != nil {
		return 0, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, fmt.Errorf("user not found")
	}

	auditQuery := `INSERT INTO audit_logs (user_id, action) VALUES ($1, $2)`
	_, err = tx.Exec(auditQuery, id, "USER_DELETED")
	if err != nil {
		return 0, fmt.Errorf("failed to insert audit log: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return count, nil
}
