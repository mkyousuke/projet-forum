package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"forum/database"
)


func ReportPostHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	reporterID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	postIDStr := r.FormValue("post_id")
	if postIDStr == "" {
		http.Error(w, "ID de post manquant", http.StatusBadRequest)
		return
	}
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "ID de post invalide", http.StatusBadRequest)
		return
	}
	post, err := database.GetPostByID(postID)
	if err != nil {
		http.Error(w, "Post introuvable", http.StatusNotFound)
		return
	}
	reporter, err := database.GetUserWithRole(reporterID)
	if err != nil {
		http.Error(w, "Utilisateur introuvable", http.StatusUnauthorized)
		return
	}
	message := fmt.Sprintf("Le post \"%s\" (ID:%d) a été signalé par %s (ID:%d)", post.Title, post.ID, reporter.Username, reporter.ID)
	mods, err := database.GetModeratorsAndAdmins()
	if err == nil {
		for _, mod := range mods {
			_ = database.CreateNotification(mod.ID, message, post.ID, 0)
		}
	}
	_ = database.CreateNotification(reporter.ID, "Votre signalement a été envoyé aux modérateurs.", post.ID, 0)
	http.Redirect(w, r, "/post?id="+strconv.Itoa(post.ID), http.StatusSeeOther)
}
