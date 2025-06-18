package handler

import (
	"forum/database"
	"html/template"
	"net/http"
	"strconv"
)

func ModerationDashboardHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	user, err := database.GetUserWithRole(userID)
	if err != nil {
		http.Error(w, "Utilisateur non trouvé", http.StatusUnauthorized)
		return
	}
	if user.Role != "moderator" && user.Role != "admin" {
		http.Error(w, "Accès refusé", http.StatusForbidden)
		return
	}
	pendingPosts, err := database.GetPendingPosts()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des posts en attente", http.StatusInternalServerError)
		return
	}
	t, err := template.ParseFiles("templates/moderation.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template", http.StatusInternalServerError)
		return
	}
	data := struct {
		PendingPosts []database.Post
		User         database.User
	}{
		PendingPosts: pendingPosts,
		User:         user,
	}
	t.Execute(w, data)
}

func ApprovePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("user_id")
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	user, err := database.GetUserWithRole(userID)
	if err != nil || (user.Role != "moderator" && user.Role != "admin") {
		http.Error(w, "Accès refusé", http.StatusForbidden)
		return
	}
	postIDStr := r.FormValue("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "ID de post invalide", http.StatusBadRequest)
		return
	}
	err = database.SetPostModerationStatus(postID, "approved")
	if err != nil {
		http.Error(w, "Erreur lors de l'approbation du post", http.StatusInternalServerError)
		return
	}
	post, err := database.GetPostByID(postID)
	if err == nil {
		msg := "Votre post \"" + post.Title + "\" a été approuvé."
		_ = database.CreateNotification(post.UserID, msg, post.ID, 0)
	}
	http.Redirect(w, r, "/moderation", http.StatusSeeOther)
}

func RejectPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("user_id")
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	user, err := database.GetUserWithRole(userID)
	if err != nil || (user.Role != "moderator" && user.Role != "admin") {
		http.Error(w, "Accès refusé", http.StatusForbidden)
		return
	}
	postIDStr := r.FormValue("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "ID de post invalide", http.StatusBadRequest)
		return
	}
	err = database.SetPostModerationStatus(postID, "rejected")
	if err != nil {
		http.Error(w, "Erreur lors du rejet du post", http.StatusInternalServerError)
		return
	}
	post, err := database.GetPostByID(postID)
	if err == nil {
		msg := "Votre post \"" + post.Title + "\" a été rejeté."
		_ = database.CreateNotification(post.UserID, msg, post.ID, 0)
	}
	http.Redirect(w, r, "/moderation", http.StatusSeeOther)
}

func PromoteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("user_id")
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	adminID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	admin, err := database.GetUserWithRole(adminID)
	if err != nil || admin.Role != "admin" {
		http.Error(w, "Accès réservé aux administrateurs", http.StatusForbidden)
		return
	}
	targetUserIDStr := r.FormValue("user_id")
	targetUserID, err := strconv.Atoi(targetUserIDStr)
	if err != nil {
		http.Error(w, "ID d'utilisateur invalide", http.StatusBadRequest)
		return
	}
	err = database.UpdateUserRole(targetUserID, "moderator")
	if err != nil {
		http.Error(w, "Erreur lors de la promotion de l'utilisateur", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func DemoteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("user_id")
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	adminID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	admin, err := database.GetUserWithRole(adminID)
	if err != nil || admin.Role != "admin" {
		http.Error(w, "Accès réservé aux administrateurs", http.StatusForbidden)
		return
	}
	targetUserIDStr := r.FormValue("user_id")
	targetUserID, err := strconv.Atoi(targetUserIDStr)
	if err != nil {
		http.Error(w, "ID d'utilisateur invalide", http.StatusBadRequest)
		return
	}
	err = database.UpdateUserRole(targetUserID, "user")
	if err != nil {
		http.Error(w, "Erreur lors de la rétrogradation de l'utilisateur", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}
