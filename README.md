# 🎬 CinéForum - Forum collaboratif sur le cinéma et les séries

Bienvenue sur **CinéForum**, un forum web moderne et complet où les utilisateurs peuvent discuter de films et séries, interagir avec une IA, recevoir des actualités cinéma en temps réel, et bien plus encore.

---

## 🌟 Fonctionnalités principales

### 🧑‍💻 Authentification & Gestion des utilisateurs
- Inscription et connexion classique
- Connexion via **OAuth** (Google, GitHub, Facebook, Twitter)
- Déconnexion sécurisée
- Profil utilisateur personnalisé :
  - Avatar sélectionnable
  - Statistiques personnelles (posts likés, commentaires, dernières activités)

### 📝 Forum & Publications
- Création, modification et suppression de **posts**
- Ajout et suppression de **commentaires**
- Affichage des publications avec discussions associées

### 👍 Interactions sociales
- **Like/Dislike** sur les posts **et** commentaires
- **Notifications** dynamiques :
  - Pour les interactions reçues
  - Pour les nouveaux commentaires
- Bouton **"Marquer comme lu"** pour gérer les notifications

### 📢 Actualités cinéma
- Page d’**actualités cinéma** alimentée par l’API **NewsAPI**
- Affichage sous forme de **cartes défilables** avec images

### 🤖 Assistant IA (Gemini)
- Chat interactif avec une IA spécialisée dans le cinéma et les séries
- Réponses contextualisées dans une interface type chatbot

### 🛡️ Modération & Administration
- Interface **modération** : suppression de contenu signalé
- Interface **admin** :
  - Gestion des utilisateurs
  - Suppression ou bannissement
- Visibilité des rôles pour le personnel (modérateur/admin)

### 🚨 Signalement
- Système de **report** pour les contenus inappropriés
- Suivi des signalements dans l’interface d’administration

### 🖼️ Multimédia & UI/UX
- Sélection de **photos de profil**
- Design moderne et interface intuitive

### 🚀 Optimisations & Sécurité
- Chargement optimisé avec **lazy loading** et **preload**
- Utilisation de **HTTPS** avec certificats auto-signés (`cert.pem`, `key.pem`)
- **Rate limiting** pour prévenir les abus

---

## 🗂️ Structure du projet

└── forum/
    ├── README.md
    ├── forum.db
    ├── go.mod
    ├── go.sum
    ├── main.go
    ├── certs/
    │   ├── cert.pem
    │   ├── key.pem
    │   └── .DS_Store
    ├── database/
    │   ├── database.go
    │   └── SQL/
    │       └── database.sql
    ├── handler/
    │   ├── actualites.go
    │   ├── admin_reports.go
    │   ├── admin_users.go
    │   ├── comment.go
    │   ├── connexion.go
    │   ├── deconnexion.go
    │   ├── gemini_chat.go
    │   ├── handlers.go
    │   ├── inscription.go
    │   ├── like.go
    │   ├── moderation.go
    │   ├── notifications.go
    │   ├── oauth.go
    │   ├── post.go
    │   ├── profil.go
    │   ├── report.go
    │   └── TMDB.go
    ├── middleware/
    │   ├── rate_limiter.go
    │   └── static.go
    ├── server/
    │   └── server.go
    ├── static/
    │   ├── .DS_Store
    │   ├── css/
    │   │   ├── actualites.css
    │   │   ├── admin.css
    │   │   ├── API.css
    │   │   ├── connexion.css
    │   │   ├── gemini.css
    │   │   ├── index.css
    │   │   ├── inscription.css
    │   │   ├── main.css
    │   │   ├── modify_profil.css
    │   │   ├── new_post.css
    │   │   ├── notifications.css
    │   │   ├── post_detail.css
    │   │   ├── posts.css
    │   │   ├── profil.css
    │   │   └── theoriesSpoilers.css
    │   ├── images/
    │   │   ├── background.jpg.webp
    │   │   ├── .DS_Store
    │   │   └── profil/
    │   │       └── .DS_Store
    │   ├── js/
    │   │   ├── api.js
    │   │   └── gemini_chat.js
    │   └── uploads/
    └── templates/
        ├── actualites.html
        ├── admin_users.html
        ├── API.html
        ├── connexion.html
        ├── edit_post.html
        ├── gemini_chat.html
        ├── index.html
        ├── inscription.html
        ├── moderation.html
        ├── modify_profil.html
        ├── new_post.html
        ├── notifications.html
        ├── post_detail.html
        ├── posts.html
        ├── profil.html
        └── .DS_Store

---

## 🔧 Installation & Lancement

### Prérequis
- Go 1.20+
- SQLite
- Ngrok (optionnel pour test externe)

### Installation

```bash
git clone https://ytrack.learn.ynov.com/git/lomaxime/forum.git
cd forum
go mod tidy
```

### Lancement en HTTPS local

```bash
go run .
```
> Accès via https://localhost

### Modération 

Identifiants administrateur et modérateurs : 

Nom d'utilisateur de l'administrateur : admin
Mot de passe de l'administrateur : admin

Nom d'utilisateur du modérateur : moderateur
Mot de passe du modérateur : moderateur

## 📌 Remarques

- Les photos de profil peuvent parfois prendre du temps à charger, n'hésitez pas à recharger la page plusieurs fois. 
- Environnement totalement local, sans dépendance cloud
- Base de données SQLite pré-initialisée dans `forum.db`
---

## 📘 License & Attributions

Projet développé par LOPRIN Maxime dans le cadre d’un exercice à Val d'Europe Ynov Campus - 2025   

This project uses the Google Gemini API.  
© 2024 Google LLC. All rights reserved.  
Licensed under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0).

This project uses the NewsAPI to fetch news content.  
News data provided by [NewsAPI.org](https://newsapi.org).  
Please review [NewsAPI's terms of use](https://newsapi.org/terms) for usage restrictions.

This product uses the TMDB API but is not endorsed or certified by TMDB.  
Data provided by [The Movie Database (TMDB)](https://www.themoviedb.org/).

All logos, images, and brand assets used (e.g., Netflix, Marvel, DC Universe, Star Wars) are the property of their respective owners.  
This project is for educational or non-commercial use only and is not affiliated with or endorsed by any of these entities.
