package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string `json:"id" storm:"id,index"`
	Username     string `json:"username" storm:"unique"`
	PasswordHash string `json:"passwordHash"`
}

type UserSession struct {
	Token     string    `storm:"id,index"`
	Username  string    `storm:"index"`
	ExpiresAt time.Time
}

func initDefaultUser() {
	var user User
	if db.One("Username", "root", &user) == nil {
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("nightlight"), bcrypt.DefaultCost)
	if err != nil {
		hclog.Default().Named("auth").Error("failed to hash default password", "err", err)
		return
	}
	user = User{
		ID:           uuid.New().String(),
		Username:     "root",
		PasswordHash: string(hash),
	}
	if err := db.Save(&user); err != nil {
		hclog.Default().Named("auth").Error("failed to save default user", "err", err)
		return
	}
	hclog.Default().Named("auth").Info("created default user root")
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var user User
	if err := db.One("Username", req.Username, &user); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token := uuid.New().String()
	session := UserSession{
		Token:     token,
		Username:  user.Username,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := db.Save(&session); err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((24 * time.Hour).Seconds()),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "username": user.Username})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		var session UserSession
		if db.One("Token", cookie.Value, &session) == nil {
			db.DeleteStruct(&session)
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	w.WriteHeader(http.StatusOK)
}

func CurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	session, ok := validSession(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"username": session.Username})
}

func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	session, ok := validSession(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.NewPassword) < 8 {
		http.Error(w, "new password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	var user User
	if err := db.One("Username", session.Username, &user); err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		http.Error(w, "current password is incorrect", http.StatusUnauthorized)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}
	user.PasswordHash = string(hash)
	if err := db.Save(&user); err != nil {
		http.Error(w, "failed to save password", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := validSession(r); !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func validSession(r *http.Request) (UserSession, bool) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return UserSession{}, false
	}
	var session UserSession
	if err := db.One("Token", cookie.Value, &session); err != nil {
		return UserSession{}, false
	}
	if time.Now().After(session.ExpiresAt) {
		db.DeleteStruct(&session)
		return UserSession{}, false
	}
	return session, true
}
