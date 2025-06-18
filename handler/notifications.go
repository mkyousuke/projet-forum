package handler

import (
	"encoding/json"
	"forum/database"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func NotificationsHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}
	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Error(w, "Cookie invalide", http.StatusBadRequest)
		return
	}
	notifs, err := database.GetNotificationsByUserID(userID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des notifications", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifs)
}

type NotificationView struct {
	Message   string
	PostLink  string
	CreatedAt time.Time
}

func NotificationsPageHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	notifs, err := database.GetNotificationsByUserID(userID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des notifications", http.StatusInternalServerError)
		return
	}
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		loc = time.UTC
	}
	var views []NotificationView
	for _, n := range notifs {
		nv := NotificationView{}
		if n.CommentID != 0 {
			comment, err := database.GetCommentByID(n.CommentID)
			if err == nil {
				nv.Message = "Quelqu'un a liké votre commentaire"
				if comment.UserID != 0 {
				}
			} else {
				nv.Message = n.Message
			}
			nv.PostLink = "/post?id=" + strconv.Itoa(n.PostID)
		} else {
			nv.Message = n.Message
			nv.PostLink = "/post?id=" + strconv.Itoa(n.PostID)
		}
		nv.CreatedAt = n.CreatedAt.In(loc)
		views = append(views, nv)
	}
	templatePath := filepath.Join("templates", "notifications.html")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		http.Error(w, "Template introuvable", http.StatusNotFound)
		return
	}
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, views)
	if err != nil {
		http.Error(w, "Erreur lors de l'exécution du template", http.StatusInternalServerError)
		return
	}
}

func MarkNotificationsAsReadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("user_id")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}
	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Error(w, "Cookie invalide", http.StatusBadRequest)
		return
	}
	err = database.DeleteNotificationsByUserID(userID)
	if err != nil {
		http.Error(w, "Erreur lors de la suppression des notifications", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/notifications-page", http.StatusSeeOther)
}