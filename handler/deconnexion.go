package handler

import (
	"net/http"

	"forum/database"
)

func DeconnexionHandler(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session_id"); err == nil {
		_ = database.DeleteSession(cookie.Value)
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
		})
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/index", http.StatusSeeOther)
}
