package repository

import (
	"database/sql"

	"user-post-api/internal/model"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) GetAll() ([]model.User, error) {
	rows, err := r.DB.Query("SELECT id, username, email FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) GetByID(id int) (model.User, error) {
	var u model.User
	err := r.DB.QueryRow("SELECT id, username, email FROM users WHERE id = $1", id).
		Scan(&u.ID, &u.Username, &u.Email)
	return u, err
}

func (r *UserRepository) GetByEmail(email string) (model.User, error) {
	var u model.User
	err := r.DB.QueryRow("SELECT id, username, email, password FROM users WHERE email = $1", email).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password)
	return u, err
}

func (r *UserRepository) Create(user model.User) (int, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	var id int
	query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id`
	err = r.DB.QueryRow(query, user.Username, user.Email, string(hashed)).Scan(&id)
	return id, err
}

func (r *UserRepository) Update(id int, user model.User) error {
	_, err := r.DB.Exec("UPDATE users SET username = $1, email = $2 WHERE id = $3",
		user.Username, user.Email, id)
	return err
}

func (r *UserRepository) Delete(id int) error {
	_, err := r.DB.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}
