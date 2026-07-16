package repository

import (
	"database/sql"
	"user-post-api/internal/model"
)

type PostRepository struct {
	DB *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{DB: db}
}

func (r *PostRepository) GetAll() ([]model.Post, error) {
	rows, err := r.DB.Query("SELECT id, title, content, user_id FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []model.Post
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.UserID); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *PostRepository) GetByID(id int) (model.Post, error) {
	var p model.Post
	err := r.DB.QueryRow("SELECT id, title, content, user_id FROM posts WHERE id = $1", id).
		Scan(&p.ID, &p.Title, &p.Content, &p.UserID)
	return p, err
}

func (r *PostRepository) Create(post model.Post) (int, error) {
	var id int
	err := r.DB.QueryRow(`INSERT INTO posts (title, content, user_id) VALUES ($1, $2, $3) RETURNING id`,
		post.Title, post.Content, post.UserID).Scan(&id)
	return id, err
}

func (r *PostRepository) Delete(id int) error {
	_, err := r.DB.Exec("DELETE FROM posts WHERE id = $1", id)
	return err
}