package handler

import (
	"fmt"
	"html/template"
	"net/http"

	"forum/database"
	"golang.org/x/crypto/bcrypt"
)

func InscriptionHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		t, err := template.ParseFiles("templates/inscription.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, nil)
		if err != nil {
			http.Error(w, "Erreur lors de l'exécution du template", http.StatusInternalServerError)
		}
	case http.MethodPost:
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		if username == "" || email == "" || password == "" {
			http.Error(w, "Tous les champs sont requis", http.StatusBadRequest)
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Erreur lors du hachage du mot de passe", http.StatusInternalServerError)
			return
		}
		err = database.CreateUser(username, email, string(hashedPassword))
		if err != nil {
			http.Error(w, fmt.Sprintf("Erreur lors de la création de l'utilisateur: %v", err), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
	default:
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
	}
}
