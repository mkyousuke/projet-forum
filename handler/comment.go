package handler

import (
	"net/http"
	"strconv"

	"forum/database"
)

func AddCommentHandler(w http.ResponseWriter, r *http.Request) {
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
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
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
	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Contenu du commentaire requis", http.StatusBadRequest)
		return
	}
	err = database.CreateComment(postID, userID, content)
	if err != nil {
		http.Error(w, "Erreur lors de l'ajout du commentaire: "+err.Error(), http.StatusInternalServerError)
		return
	}
	post, err := database.GetPostByID(postID)
	if err == nil {
		if post.UserID != userID {
			message := "Quelqu'un a commenté votre post."
			_ = database.CreateNotification(post.UserID, message, postID, 0)
		}
	}
	http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
}

func DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
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
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID de commentaire manquant", http.StatusBadRequest)
		return
	}
	commentID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID de commentaire invalide", http.StatusBadRequest)
		return
	}
	postIDStr := r.URL.Query().Get("post_id")
	if postIDStr == "" {
		http.Error(w, "ID de post manquant", http.StatusBadRequest)
		return
	}
	user, err := database.GetUserWithRole(userID)
	if err != nil {
		http.Error(w, "Utilisateur non trouvé", http.StatusUnauthorized)
		return
	}
	if user.Role == "admin" || user.Role == "moderator" {
		err = database.AdminDeleteComment(commentID)
	} else {
		err = database.DeleteComment(commentID, userID)
	}
	if err != nil {
		http.Error(w, "Erreur lors de la suppression du commentaire: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/post?id="+postIDStr, http.StatusSeeOther)
}
