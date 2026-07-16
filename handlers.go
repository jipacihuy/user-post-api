package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// ===== MIDDLEWARE =====

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondError(w, http.StatusUnauthorized, "invalid authorization format")
			return
		}
		tokenString := parts[1]
		userID, err := verifyToken(tokenString)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid token: "+err.Error())
			return
		}
		r.Header.Set("user_id", strconv.Itoa(userID))
		next(w, r)
	}
}

// ===== AUTH HANDLERS =====

func Register(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if user.Username == "" || user.Email == "" || user.Password == "" {
		respondError(w, http.StatusBadRequest, "username, email, and password required")
		return
	}
	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		// Ambil pesan error pertama
		errors := err.(validator.ValidationErrors)
		respondError(w, http.StatusBadRequest, errors[0].Translate(nil))
		return
	}
	// Check if email exists
	_, err := getUserByEmail(user.Email)
	if err == nil {
		respondError(w, http.StatusConflict, "email already exists")
		return
	}
	if err != sql.ErrNoRows {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	id, err := CreateUser(user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	user.ID = id
	user.Password = ""
	respondSuccess(w, http.StatusCreated, user)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		errors := err.(validator.ValidationErrors)
		respondError(w, http.StatusBadRequest, errors[0].Translate(nil))
		return
	}
	user, err := getUserByEmail(req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	// Compare hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	token, err := generateToken(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	user.Password = ""
	resp := LoginResponse{Token: token, User: user}
	respondSuccess(w, http.StatusOK, resp)
}

func GetProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("user_id")
	userID, _ := strconv.Atoi(userIDStr)
	user, err := getUserByID(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	user.Password = ""
	respondSuccess(w, http.StatusOK, user)
}

// ===== USER HANDLERS =====

func GetUsers(w http.ResponseWriter, r *http.Request) {
    cacheKey := "users:all"

    var users []User
    err := GetCache(cacheKey, &users)
    if err == nil {
        log.Println("✅ Cache HIT for users:all")
        respondSuccess(w, http.StatusOK, users) // <-- PAKAI respondSuccess!
        return
    }

    log.Println("❌ Cache MISS for users:all")

    users, err = getAllUsers()
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }

    _ = SetCache(cacheKey, users, 5*time.Minute)
    respondSuccess(w, http.StatusOK, users)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		errors := err.(validator.ValidationErrors)
		respondError(w, http.StatusBadRequest, errors[0].Translate(nil))
		return
	}

	id, err := CreateUser(user)
	if err != nil {
		respondError(w, http.StatusConflict, "email already exists or failed to create user")
		return
	}

	// Hapus cache users
	_ = DeleteCache("users:*")

	user.ID = id
	user.Password = ""
	respondSuccess(w, http.StatusCreated, user)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        respondError(w, http.StatusBadRequest, "invalid ID")
        return
    }

    cacheKey := fmt.Sprintf("user:%d", id)

    var user User
    err = GetCache(cacheKey, &user)
    if err == nil {
        log.Printf("✅ Cache HIT for user:%d", id)
        user.Password = ""
        respondSuccess(w, http.StatusOK, user) // <-- PAKAI respondSuccess!
        return
    }

    log.Printf("❌ Cache MISS for user:%d", id)

    user, err = getUserByID(id)
    if err != nil {
        respondError(w, http.StatusNotFound, "user not found")
        return
    }

    _ = SetCache(cacheKey, user, 10*time.Minute)
    user.Password = ""
    respondSuccess(w, http.StatusOK, user)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := updateUserDB(id, user); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	// Hapus cache users dan user detail
	_ = DeleteCache("users:*")
	_ = DeleteCache(fmt.Sprintf("user:%d", id))

	user.ID = id
	user.Password = ""
	respondSuccess(w, http.StatusOK, user)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	if err := deleteUserDB(id); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	// Hapus cache users dan user detail
	_ = DeleteCache("users:*")
	_ = DeleteCache(fmt.Sprintf("user:%d", id))

	respondSuccess(w, http.StatusOK, map[string]string{"message": "user deleted"})
}

// ===== POST HANDLERS =====

func GetPosts(w http.ResponseWriter, r *http.Request) {
	cacheKey := "posts:all"

	var posts []Post
	err := GetCache(cacheKey, &posts)
	if err == nil {
		respondSuccess(w, http.StatusOK, posts)
		return
	}

	posts, err = getAllPosts()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = SetCache(cacheKey, posts, 5*time.Minute)
	respondSuccess(w, http.StatusOK, posts)
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	var post Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	validate := validator.New()
	if err := validate.Struct(post); err != nil {
		errors := err.(validator.ValidationErrors)
		respondError(w, http.StatusBadRequest, errors[0].Translate(nil))
		return
	}

	id, err := createPostDB(post)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create post")
		return
	}

	// Hapus cache posts
	_ = DeleteCache("posts:*")

	post.ID = id
	respondSuccess(w, http.StatusCreated, post)
}

func GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	cacheKey := fmt.Sprintf("post:%d", id)

	var post Post
	err = GetCache(cacheKey, &post)
	if err == nil {
		respondSuccess(w, http.StatusOK, post)
		return
	}

	post, err = getPostByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "post not found")
		return
	}

	_ = SetCache(cacheKey, post, 10*time.Minute)
	respondSuccess(w, http.StatusOK, post)
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	if err := deletePostDB(id); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete post")
		return
	}

	// Hapus cache posts dan post detail
	_ = DeleteCache("posts:*")
	_ = DeleteCache(fmt.Sprintf("post:%d", id))

	respondSuccess(w, http.StatusOK, map[string]string{"message": "post deleted"})
}

