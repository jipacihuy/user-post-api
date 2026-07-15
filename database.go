package main

import (
	"database/sql"
	"log"
	"fmt"     
	"os" 

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func initDB() {
	var err error
	// Ambil dari environment variables (Railway bakal kasih ini otomatis)
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "zifa05"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "user_post_api"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("Failed to connect:", err)
	}
	log.Println("✅ PostgreSQL connected")
	createTables()
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(100) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS posts (
			id SERIAL PRIMARY KEY,
			user_id INT REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(200) NOT NULL,
			content TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Fatal("Failed to create table:", err)
		}
	}
	log.Println("Tables created/verified.")
}

// --- USER ---

func getAllUsers() ([]User, error) {
	rows, err := db.Query("SELECT id, username, email FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func getUserByID(id int) (User, error) {
	var u User
	err := db.QueryRow("SELECT id, username, email FROM users WHERE id = $1", id).
		Scan(&u.ID, &u.Username, &u.Email)
	return u, err
}

func getUserByEmail(email string) (User, error) {
	var u User
	err := db.QueryRow("SELECT id, username, email, password FROM users WHERE email = $1", email).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password)
	return u, err
}

func CreateUser(user User) (int, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("❌ Hashing error: %v", err)
		return 0, err
	}
	var id int
	query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id`
	err = db.QueryRow(query, user.Username, user.Email, string(hashed)).Scan(&id)
	if err != nil {
		log.Printf("❌ DB Insert error: %v", err)
		return 0, err
	}
	return id, nil
}

func updateUserDB(id int, user User) error {
	_, err := db.Exec("UPDATE users SET username = $1, email = $2 WHERE id = $3",
		user.Username, user.Email, id)
	return err
}

func deleteUserDB(id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

// --- POST ---

func getAllPosts() ([]Post, error) {
	rows, err := db.Query("SELECT id, title, content, user_id FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.UserID); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func getPostByID(id int) (Post, error) {
	var p Post
	err := db.QueryRow("SELECT id, title, content, user_id FROM posts WHERE id = $1", id).
		Scan(&p.ID, &p.Title, &p.Content, &p.UserID)
	return p, err
}

func createPostDB(post Post) (int, error) {
	var id int
	err := db.QueryRow(`INSERT INTO posts (title, content, user_id) VALUES ($1, $2, $3) RETURNING id`,
		post.Title, post.Content, post.UserID).Scan(&id)
	return id, err
}

func deletePostDB(id int) error {
	_, err := db.Exec("DELETE FROM posts WHERE id = $1", id)
	return err
}