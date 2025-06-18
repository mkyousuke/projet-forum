package handler

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"forum/database"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func ConnexionHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		t, err := template.ParseFiles("templates/connexion.html")
		if err != nil {
			http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
			return
		}
		_ = t.Execute(w, nil)

	case http.MethodPost:
		identifier := r.FormValue("identifier")
		password := r.FormValue("password")
		if identifier == "" || password == "" {
			http.Error(w, "Tous les champs requis", http.StatusBadRequest)
			return
		}
		var user database.User
		var err error
		if strings.Contains(identifier, "@") {
			user, err = database.GetUserByEmail(identifier)
		} else {
			user, err = database.GetUserByUsername(identifier)
		}
		if err != nil {
			http.Error(w, "Utilisateur introuvable", http.StatusUnauthorized)
			return
		}
		if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			http.Error(w, "Identifiants invalides", http.StatusUnauthorized)
			return
		}

		sessionID := uuid.NewString()
		expiry := time.Now().Add(24 * time.Hour)
		if err := database.CreateSession(sessionID, user.ID, expiry); err != nil {
			http.Error(w, "Erreur création session", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Expires:  expiry,
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "user_id",
			Value:    strconv.Itoa(user.ID),
			Path:     "/",
			HttpOnly: true,
		})

		http.Redirect(w, r, "/index", http.StatusSeeOther)

	default:
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
	}
}
