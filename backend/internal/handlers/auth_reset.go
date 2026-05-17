package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	netmail "net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/albertson/albertsonmissionplan/internal/auth"
	"github.com/albertson/albertsonmissionplan/internal/mail"
	"github.com/albertson/albertsonmissionplan/internal/store"
)

func passwordResetTTL() time.Duration {
	h := strings.TrimSpace(os.Getenv("PASSWORD_RESET_HOURS"))
	if h == "" {
		return 24 * time.Hour
	}
	n, err := strconv.Atoi(h)
	if err != nil || n <= 0 || n > 720 {
		return 24 * time.Hour
	}
	return time.Duration(n) * time.Hour
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// APIAuthForgotPassword queues a reset email for approved leader accounts. Always returns the same
// success shape when the email parses, to avoid leaking which addresses exist.
func (s *Server) APIAuthForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed."})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var req forgotPasswordRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body."})
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if _, err := netmail.ParseAddress(email); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Please enter a valid email address."})
		return
	}

	okPayload := map[string]any{
		"ok":      true,
		"message": "If that email has an approved leader account, check your inbox for reset instructions.",
	}

	_, err := store.LeaderCredentialHash(ctx, s.DB, email)
	switch {
	case errors.Is(err, sql.ErrNoRows), errors.Is(err, store.ErrLeaderAccountPending):
		writeJSON(w, http.StatusOK, okPayload)
		return
	case err != nil:
		log.Printf("forgot-password lookup: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not process request."})
		return
	}

	if s.Mail == nil {
		log.Printf("forgot-password: mail disabled, no email sent for %s", email)
		writeJSON(w, http.StatusOK, okPayload)
		return
	}
	base := s.trimmedPublicBaseURL()
	if base == "" {
		log.Printf("forgot-password: PUBLIC_BASE_URL empty, cannot email reset link for %s", email)
		writeJSON(w, http.StatusOK, okPayload)
		return
	}

	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		log.Printf("forgot-password rand: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not process request."})
		return
	}
	tokenHex := hex.EncodeToString(secret)
	hashHex := store.PasswordResetTokenHashHex(secret)
	expires := time.Now().Add(passwordResetTTL()).Unix()

	link := fmt.Sprintf("%s/admin/reset-password?t=%s", base, tokenHex)
	expiryH := int(passwordResetTTL().Hours())
	subject, plain, html, err := mail.RenderPasswordResetEmail(email, link, expiryH)
	if err != nil {
		log.Printf("forgot-password render mail: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not process request."})
		return
	}

	if err := store.DeletePasswordResetTokensForEmail(ctx, s.DB, email); err != nil {
		log.Printf("forgot-password delete old tokens: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not process request."})
		return
	}
	if err := store.InsertPasswordResetToken(ctx, s.DB, email, hashHex, expires); err != nil {
		log.Printf("forgot-password insert token: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not process request."})
		return
	}

	if err := s.Mail.SendAlternative(ctx, []string{email}, subject, plain, html); err != nil {
		log.Printf("forgot-password send mail: %v", err)
	}
	writeJSON(w, http.StatusOK, okPayload)
}

// APIAuthResetPassword consumes a token from the forgot-password email and sets a new password.
func (s *Server) APIAuthResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed."})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var req resetPasswordRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body."})
		return
	}
	tokenHex := strings.TrimSpace(strings.ToLower(req.Token))
	if tokenHex == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Reset link is missing or incomplete."})
		return
	}

	newHash, err := auth.HashPassword(req.Password)
	switch {
	case errors.Is(err, auth.ErrPasswordEmpty):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Please enter a password."})
		return
	case errors.Is(err, auth.ErrPasswordTooLong):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Password is too long (max 72 characters — bcrypt's hard limit)."})
		return
	case err != nil:
		log.Printf("reset-password hash: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not reset password."})
		return
	}

	err = store.ResetLeaderPasswordWithToken(ctx, s.DB, tokenHex, newHash)
	if errors.Is(err, store.ErrPasswordResetInvalidToken) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "This reset link is invalid or has expired. Request a new one from the sign-in page."})
		return
	}
	if err != nil {
		log.Printf("reset-password store: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not reset password."})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
