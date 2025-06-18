package handler

import (
	"net/http"
	"strconv"

	"forum/database"
)

func LikePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
		return
	}
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
	err = database.SetPostLike(userID, postID, 1)
	if err != nil {
		http.Error(w, "Erreur lors du like", http.StatusInternalServerError)
		return
	}
	post, err := database.GetPostByID(postID)
	if err == nil && post.UserID != userID {
		database.CreateNotification(post.UserID, "Quelqu'un a liké votre post", postID, 0)
	}
	http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
}

func DislikePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
		return
	}
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
	err = database.SetPostLike(userID, postID, -1)
	if err != nil {
		http.Error(w, "Erreur lors du dislike", http.StatusInternalServerError)
		return
	}
	post, err := database.GetPostByID(postID)
	if err == nil && post.UserID != userID {
		database.CreateNotification(post.UserID, "Quelqu'un a disliké votre post", postID, 0)
	}
	http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
}

func LikeCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
		return
	}
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
	commentIDStr := r.FormValue("comment_id")
	if commentIDStr == "" {
		http.Error(w, "ID de commentaire manquant", http.StatusBadRequest)
		return
	}
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "ID de commentaire invalide", http.StatusBadRequest)
		return
	}
	err = database.SetCommentLike(userID, commentID, 1)
	if err != nil {
		http.Error(w, "Erreur lors du like du commentaire", http.StatusInternalServerError)
		return
	}
	comment, err := database.GetCommentByID(commentID)
	if err == nil && comment.UserID != userID {
		database.CreateNotification(comment.UserID, "Quelqu'un a liké votre commentaire", comment.PostID, commentID)
	}
	postIDStr := r.FormValue("post_id")
	if postIDStr == "" {
		http.Error(w, "ID de post manquant", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/post?id="+postIDStr, http.StatusSeeOther)
}

func DislikeCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
		return
	}
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
	commentIDStr := r.FormValue("comment_id")
	if commentIDStr == "" {
		http.Error(w, "ID de commentaire manquant", http.StatusBadRequest)
		return
	}
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "ID de commentaire invalide", http.StatusBadRequest)
		return
	}
	err = database.SetCommentLike(userID, commentID, -1)
	if err != nil {
		http.Error(w, "Erreur lors du dislike du commentaire", http.StatusInternalServerError)
		return
	}
	comment, err := database.GetCommentByID(commentID)
	if err == nil && comment.UserID != userID {
		database.CreateNotification(comment.UserID, "Quelqu'un a disliké votre commentaire", comment.PostID, commentID)
	}
	postIDStr := r.FormValue("post_id")
	if postIDStr == "" {
		http.Error(w, "ID de post manquant", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/post?id="+postIDStr, http.StatusSeeOther)
}