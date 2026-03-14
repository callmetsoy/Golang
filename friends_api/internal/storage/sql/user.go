package sqlStorage

import (
	"fmt"
	"github.com/callmetsoy/user_api/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"strings"
	"sync"
)

type UserRepo struct {
	db *sqlx.DB
}

var (
	userRepoInstance *UserRepo
	once             sync.Once
)

func NewUserRepo(db *sqlx.DB) {
	once.Do(func() {
		userRepoInstance = &UserRepo{
			db: db,
		}
	})
}

func GetUserRepo() *UserRepo {
	return userRepoInstance
}

func (r *UserRepo) GetPaginatedUsers(
	page int, pageSize int, filters map[string]interface{}, orderBy string, orderDir string,
) (domain.PaginatedUsers, error) {
	var users []domain.User
	offset := (page - 1) * pageSize

	baseQuery := `SELECT id, name, email, gender, birth_date FROM users`
	countQuery := `SELECT COUNT(*) FROM users`

	// Filtering
	var whereClauses []string
	var args []interface{}
	argIndex := 1
	for field, value := range filters {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	if len(whereClauses) > 0 {
		whereSQL := " WHERE " + strings.Join(whereClauses, " AND ")
		baseQuery += whereSQL
		countQuery += whereSQL
	}

	// Ordering
	if orderBy != "" {
		orderDir = strings.ToUpper(orderDir)
		if orderDir != "ASC" && orderDir != "DESC" {
			orderDir = "ASC"
		}
		baseQuery += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)
	} else {
		baseQuery += " ORDER BY id ASC"
	}

	// Pagination
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	// Count total
	var totalCount int
	if err := r.db.QueryRow(countQuery, args[:len(args)-2]...).Scan(&totalCount); err != nil {
		return domain.PaginatedUsers{}, err
	}

	// Execute query
	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return domain.PaginatedUsers{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Gender, &u.BirthDate); err != nil {
			return domain.PaginatedUsers{}, err
		}
		users = append(users, u)
	}

	return domain.PaginatedUsers{
		Data:       users,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (r *UserRepo) GetCommonFriends(userID1, userID2 uuid.UUID) ([]domain.User, error) {
	var friends []domain.User

	query := `
	SELECT u.id, u.name, u.email, u.gender, u.birth_date
	FROM users u
	JOIN user_friends uf1 ON uf1.friend_id = u.id
	JOIN user_friends uf2 ON uf2.friend_id = u.id
	WHERE uf1.user_id = $1 AND uf2.user_id = $2
	ORDER BY u.id
	`

	rows, err := r.db.Query(query, userID1.String(), userID2.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Gender, &u.BirthDate); err != nil {
			return nil, err
		}
		friends = append(friends, u)
	}

	return friends, nil
}
