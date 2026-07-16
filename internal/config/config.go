package config

import (
	"context"
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

var DB *sql.DB
var RDB *redis.Client
var Ctx = context.Background()

func InitDB() *sql.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:zifa05@localhost:5432/user_post_api?sslmode=disable"
	}
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to open DB:", err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping DB:", err)
	}
	log.Println("✅ PostgreSQL connected")
	return DB
}

func InitRedis() *redis.Client {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://127.0.0.1:6379"
	}
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Println("⚠️ Redis not available, running without cache")
		return nil
	}
	rdb := redis.NewClient(opt)
	_, err = rdb.Ping(Ctx).Result()
	if err != nil {
		log.Println("⚠️ Redis not available, running without cache")
		return nil
	}
	log.Println("✅ Redis connected")
	RDB = rdb
	return rdb
}

func CreateTables(db *sql.DB) {
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
	log.Println("✅ Tables created/verified.")
}
