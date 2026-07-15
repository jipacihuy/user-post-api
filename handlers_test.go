package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
)

// ===== SETUP =====

func TestMain(m *testing.M) {
	// Inisialisasi database (sama dengan yang dipakai di main)
	initDB()
	// Jalankan semua test
	code := m.Run()
	// Exit dengan status code
	os.Exit(code)
}

// Helper: hapus user berdasarkan email (untuk cleanup)
func deleteUserByEmail(email string) {
	db.Exec("DELETE FROM users WHERE email = $1", email)
}

// ===== TEST REGISTER =====

func TestRegister(t *testing.T) {
	// Buat data unik biar ga bentrok
	email := "test_" + strconv.FormatInt(time.Now().UnixNano(), 10) + "@test.com"
	username := "test_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	password := "password123"

	// Cleanup setelah test selesai
	defer deleteUserByEmail(email)

	// --- Kasus 1: Register Sukses ---
	t.Run("Register Success", func(t *testing.T) {
		user := map[string]string{
			"username": username,
			"email":    email,
			"password": password,
		}
		body, _ := json.Marshal(user)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		Register(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected 201, got %d", w.Code)
		}
	})

	// --- Kasus 2: Register dengan Email Duplikat (harus 409 Conflict) ---
	t.Run("Register Duplicate Email", func(t *testing.T) {
		user := map[string]string{
			"username": username,
			"email":    email,
			"password": password,
		}
		body, _ := json.Marshal(user)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		Register(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Expected 409 Conflict, got %d", w.Code)
		}
	})

	// --- Kasus 3: Register dengan Username Kosong (harus 400 Bad Request) ---
	t.Run("Register Empty Username", func(t *testing.T) {
		invalidUser := map[string]string{
			"username": "",
			"email":    "test@test.com",
			"password": "password123",
		}
		body, _ := json.Marshal(invalidUser)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		Register(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", w.Code)
		}
	})

	// --- Kasus 4: Register dengan Password Pendek (harus 400) ---
	t.Run("Register Short Password", func(t *testing.T) {
		invalidUser := map[string]string{
			"username": "testuser",
			"email":    "test@test.com",
			"password": "123",
		}
		body, _ := json.Marshal(invalidUser)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		Register(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", w.Code)
		}
	})
}

// ===== TEST LOGIN =====

func TestLogin(t *testing.T) {
	// Buat user dulu
	email := "test_login_" + strconv.FormatInt(time.Now().UnixNano(), 10) + "@test.com"
	username := "testlogin_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	password := "password123"
	defer deleteUserByEmail(email)

	// Register user (sebagai setup)
	user := map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	Register(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Setup failed: register returned %d", w.Code)
	}

	// --- Kasus 1: Login Sukses ---
	t.Run("Login Success", func(t *testing.T) {
		loginReq := map[string]string{
			"email":    email,
			"password": password,
		}
		loginBody, _ := json.Marshal(loginReq)
		reqLogin := httptest.NewRequest("POST", "/login", bytes.NewBuffer(loginBody))
		reqLogin.Header.Set("Content-Type", "application/json")
		wLogin := httptest.NewRecorder()

		Login(wLogin, reqLogin)

		if wLogin.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", wLogin.Code)
		}

		// Cek apakah token ada
		var resp LoginResponse
		json.NewDecoder(wLogin.Body).Decode(&resp)
		if resp.Token == "" {
			t.Error("Token should not be empty")
		}
	})

	// --- Kasus 2: Login dengan Password Salah (harus 401) ---
	t.Run("Login Wrong Password", func(t *testing.T) {
		loginReq := map[string]string{
			"email":    email,
			"password": "wrongpassword",
		}
		loginBody, _ := json.Marshal(loginReq)
		reqLogin := httptest.NewRequest("POST", "/login", bytes.NewBuffer(loginBody))
		reqLogin.Header.Set("Content-Type", "application/json")
		wLogin := httptest.NewRecorder()

		Login(wLogin, reqLogin)

		if wLogin.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", wLogin.Code)
		}
	})

	// --- Kasus 3: Login dengan Email Tidak Terdaftar (harus 401) ---
	t.Run("Login Non-Existent Email", func(t *testing.T) {
		loginReq := map[string]string{
			"email":    "notexist@test.com",
			"password": "password123",
		}
		loginBody, _ := json.Marshal(loginReq)
		reqLogin := httptest.NewRequest("POST", "/login", bytes.NewBuffer(loginBody))
		reqLogin.Header.Set("Content-Type", "application/json")
		wLogin := httptest.NewRecorder()

		Login(wLogin, reqLogin)

		if wLogin.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", wLogin.Code)
		}
	})
}

// ===== TEST PROFILE (Protected Route) =====

func TestProfile(t *testing.T) {
	// Buat user
	email := "test_profile_" + strconv.FormatInt(time.Now().UnixNano(), 10) + "@test.com"
	username := "testprofile_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	password := "password123"
	defer deleteUserByEmail(email)

	// Register user
	user := map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	Register(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Setup failed: register returned %d", w.Code)
	}

	// Login untuk dapat token
	loginReq := map[string]string{
		"email":    email,
		"password": password,
	}
	loginBody, _ := json.Marshal(loginReq)
	reqLogin := httptest.NewRequest("POST", "/login", bytes.NewBuffer(loginBody))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	Login(wLogin, reqLogin)
	if wLogin.Code != http.StatusOK {
		t.Fatalf("Login failed: %d", wLogin.Code)
	}
	var loginResp LoginResponse
	json.NewDecoder(wLogin.Body).Decode(&loginResp)
	token := loginResp.Token

	// --- Kasus 1: Profile dengan Token Valid (harus 200) ---
	t.Run("Profile With Valid Token", func(t *testing.T) {
		reqProfile := httptest.NewRequest("GET", "/profile", nil)
		reqProfile.Header.Set("Authorization", "Bearer "+token)
		wProfile := httptest.NewRecorder()

		// Panggil handler dengan middleware
		authMiddleware(GetProfile)(wProfile, reqProfile)

		if wProfile.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", wProfile.Code)
		}

		// Cek apakah username sesuai
		var profile User
		json.NewDecoder(wProfile.Body).Decode(&profile)
		if profile.Username != username {
			t.Errorf("Expected username %s, got %s", username, profile.Username)
		}
	})

	// --- Kasus 2: Profile Tanpa Token (harus 401) ---
	t.Run("Profile Without Token", func(t *testing.T) {
		reqProfile := httptest.NewRequest("GET", "/profile", nil)
		wProfile := httptest.NewRecorder()

		authMiddleware(GetProfile)(wProfile, reqProfile)

		if wProfile.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", wProfile.Code)
		}
	})

	// --- Kasus 3: Profile dengan Token Invalid (harus 401) ---
	t.Run("Profile With Invalid Token", func(t *testing.T) {
		reqProfile := httptest.NewRequest("GET", "/profile", nil)
		reqProfile.Header.Set("Authorization", "Bearer invalidtoken123")
		wProfile := httptest.NewRecorder()

		authMiddleware(GetProfile)(wProfile, reqProfile)

		if wProfile.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", wProfile.Code)
		}
	})
}