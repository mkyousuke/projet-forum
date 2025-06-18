package handler

import (
	"html/template"
	"net/http"
	"strconv"

	"forum/database"
)

func AdminReportsHandler(w http.ResponseWriter, r *http.Request) {
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
	notifs, err := database.GetNotificationsByUserID(adminID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des notifications", http.StatusInternalServerError)
		return
	}
	t, err := template.ParseFiles("templates/admin_reports.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template", http.StatusInternalServerError)
		return
	}
	data := struct {
		Notifications []database.Notification
		Admin         database.User
	}{
		Notifications: notifs,
		Admin:         admin,
	}
	t.Execute(w, data)
}

func RespondReportHandler(w http.ResponseWriter, r *http.Request) {
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
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
		return
	}
	notifIDStr := r.FormValue("notif_id")
	if notifIDStr == "" {
		http.Error(w, "ID de notification manquant", http.StatusBadRequest)
		return
	}
	_, err = strconv.Atoi(notifIDStr)
	if err != nil {
		http.Error(w, "ID de notification invalide", http.StatusBadRequest)
		return
	}
	response := r.FormValue("response")
	if response == "" {
		http.Error(w, "Réponse vide", http.StatusBadRequest)
		return
	}
	_ = database.CreateNotification(adminID, "Votre réponse au report ("+notifIDStr+"): "+response, 0, 0)
	http.Redirect(w, r, "/admin/reports", http.StatusSeeOther)
}
