package handler

import (
	"fmt"
	"forum/database"
	"html/template"
	"log" // Importer log
	"net/http"
	"strconv"
)

type AdminUsersData struct {
	Users []database.User
	Admin database.User // Informations sur l'admin connecté pour la page
}

// Handler pour afficher la page de gestion des utilisateurs
func AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
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
	admin, err := database.GetUserWithRole(adminID) // Utiliser GetUserWithRole
	if err != nil || admin.Role != "admin" {
		http.Error(w, "Accès réservé aux administrateurs", http.StatusForbidden)
		return
	}

	users, err := getAllUsers() // Récupère tous les utilisateurs
	if err != nil {
		log.Printf("Erreur récupération utilisateurs: %v", err)
		http.Error(w, "Erreur lors de la récupération des utilisateurs", http.StatusInternalServerError)
		return
	}

	// Filtrer l'admin actuel de la liste affichée pour éviter l'auto-action
	filteredUsers := []database.User{}
	for _, u := range users {
		if u.ID != adminID {
			filteredUsers = append(filteredUsers, u)
		}
	}

	t, err := template.ParseFiles("templates/admin_users.html")
	if err != nil {
		log.Printf("Erreur parsing template admin_users.html: %v", err)
		http.Error(w, "Erreur lors du chargement du template", http.StatusInternalServerError)
		return
	}
	data := AdminUsersData{Users: filteredUsers, Admin: admin}
	if err := t.Execute(w, data); err != nil {
		log.Printf("Erreur template execution admin_users.html: %v", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

// Handler pour traiter les actions (promote, demote, delete)
func AdminUsersUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
		return
	}

	// Vérification droits admin
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

	// Récupération des données du formulaire
	targetIDStr := r.FormValue("user_id")
	action := r.FormValue("action")

	if targetIDStr == "" || action == "" {
		http.Error(w, "Données manquantes (user_id ou action)", http.StatusBadRequest)
		return
	}
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		http.Error(w, "ID utilisateur invalide", http.StatusBadRequest)
		return
	}

	// Interdiction d'auto-action
	if targetID == adminID {
		http.Error(w, "Vous ne pouvez pas effectuer cette action sur vous-même.", http.StatusBadRequest)
		return
	}

	// Traitement de l'action demandée
	switch action {
	case "promote":
		targetUser, err := database.GetUserWithRole(targetID)
		if err != nil {
			http.Error(w, "Utilisateur cible introuvable.", http.StatusNotFound)
            return
		}
		if targetUser.Role != "user" {
			http.Error(w, "L'utilisateur n'a pas le rôle 'user' requis pour être promu.", http.StatusBadRequest)
            return
		}
		if err := database.UpdateUserRole(targetID, "moderator"); err != nil {
			log.Printf("Erreur promotion user %d par admin %d: %v", targetID, adminID, err)
			http.Error(w, "Erreur lors de la promotion", http.StatusInternalServerError)
			return
		}
		log.Printf("Admin %d promoted user %d to moderator", adminID, targetID)

	case "demote":
		targetUser, err := database.GetUserWithRole(targetID)
		if err != nil {
			http.Error(w, "Utilisateur cible introuvable.", http.StatusNotFound)
            return
		}
		if targetUser.Role != "moderator" {
			http.Error(w, "L'utilisateur n'est pas un modérateur.", http.StatusBadRequest)
            return
		}
		if err := database.UpdateUserRole(targetID, "user"); err != nil {
			log.Printf("Erreur rétrogradation user %d par admin %d: %v", targetID, adminID, err)
			http.Error(w, "Erreur lors de la rétrogradation", http.StatusInternalServerError)
			return
		}
		log.Printf("Admin %d demoted user %d to user", adminID, targetID)

	case "delete":
		// Sécurité supplémentaire : ne pas supprimer un autre admin ? (A ajuster selon vos règles)
		targetUser, err := database.GetUserWithRole(targetID)
		if err != nil {
			http.Error(w, "Utilisateur cible introuvable.", http.StatusNotFound)
            return
		}
		if targetUser.Role == "admin" {
			http.Error(w, "La suppression d'un administrateur n'est pas autorisée via cette interface.", http.StatusForbidden)
			return
		}

		log.Printf("Admin %d initiating deletion for user %d (%s)", adminID, targetID, targetUser.Username)
		// Appel de la nouvelle fonction de suppression en cascade
		err = database.DeleteUserAndAssociatedData(targetID)
		if err != nil {
			log.Printf("Error deleting user %d by admin %d: %v", targetID, adminID, err)
			http.Error(w, fmt.Sprintf("Erreur lors de la suppression de l'utilisateur : %v", err), http.StatusInternalServerError)
			return
		}
		log.Printf("Admin %d successfully deleted user %d (%s)", adminID, targetID, targetUser.Username)

	default:
		http.Error(w, "Action non valide", http.StatusBadRequest)
		return
	}

	// Redirection vers la page de gestion après succès
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// Fonction utilitaire pour récupérer tous les utilisateurs (peut rester privée)
func getAllUsers() ([]database.User, error) {
	rows, err := database.DB.Query(
		// Ne pas sélectionner le mot de passe ici !
		"SELECT id, username, email, created_at, photo, role FROM users ORDER BY id ASC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []database.User
	for rows.Next() {
		var u database.User
		if err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.CreatedAt,
			&u.Photo,
			&u.Role,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}