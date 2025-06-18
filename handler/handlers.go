package handler

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"forum/database"
)

func renderTemplate(w http.ResponseWriter, r *http.Request, templateName string, data interface{}) {
	templatePath := filepath.Join("templates", templateName)
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		http.Error(w, "Template "+templateName+" introuvable", http.StatusNotFound)
		return
	}
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Erreur chargement template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Erreur exécution template: "+err.Error(), http.StatusInternalServerError)
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	type IndexData struct {
		IsLoggedIn  bool
		PhotoURL    string
		RecentPosts []database.Post
	}
	data := IndexData{
		IsLoggedIn:  false,
		PhotoURL:    "/static/images/profil/profil.png", 
		RecentPosts: nil,
	}

	if c, err := r.Cookie("user_id"); err == nil && c.Value != "" {
		if uid, err2 := strconv.Atoi(c.Value); err2 == nil {
			if u, err3 := database.GetUserByID(uid); err3 == nil && u.Photo != "" {
				data.IsLoggedIn = true
				if strings.HasPrefix(u.Photo, "http") {
					data.PhotoURL = u.Photo
				} else {
					data.PhotoURL = "/static/images/profil/" + u.Photo
				}
			}
		}
	}

	if sc, err := r.Cookie("session_id"); err == nil {
		if uid, err2 := database.GetUserIDBySession(sc.Value); err2 == nil {
			if u, err3 := database.GetUserByID(uid); err3 == nil && u.Photo != "" {
				data.IsLoggedIn = true
				if strings.HasPrefix(u.Photo, "http") {
					data.PhotoURL = u.Photo
				} else {
					data.PhotoURL = "/static/images/profil/" + u.Photo
				}
			}
		}
	}

	if posts, err := database.GetRecentPosts(3); err == nil {
		data.RecentPosts = posts
	}

	renderTemplate(w, r, "index.html", data)
}

func TheoriesSpoilersHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "theoriesSpoilers.html", nil)
}

func RedirectToIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/index", http.StatusSeeOther)
}
