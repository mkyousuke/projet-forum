package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"forum/database"
	"forum/handler"
	"forum/middleware"

	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/twitter"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/crypto/bcrypt"
)

func seedDefaultUsers() {
	defaults := []struct {
		username string
		email    string
		pwd      string
		role     string
	}{
		{"admin", "admin@example.com", "admin", "admin"},
		{"moderateur", "mod@example.com", "moderateur", "moderator"},
	}

	for _, u := range defaults {
		user, err := database.GetUserByUsername(u.username)
		if err != nil {
			hash, _ := bcrypt.GenerateFromPassword([]byte(u.pwd), bcrypt.DefaultCost)
			_ = database.CreateUser(u.username, u.email, string(hash))
			user, _ = database.GetUserByUsername(u.username)
			fmt.Printf("⚙️  Utilisateur %q créé\n", u.username)
		}
		if err := database.UpdateUserRole(user.ID, u.role); err != nil {
			fmt.Printf("❌ Impossible de définir le rôle de %q: %v\n", u.username, err)
		} else {
			fmt.Printf("✅ Rôle de %q défini sur %q\n", u.username, u.role)
		}
	}
}

func StartServer() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Pas de fichier .env trouvé, on compte sur les vars d'environnement")
	}

	domain := os.Getenv("DOMAIN")
	scheme := "https"
	port := os.Getenv("PORT")
	if port == "" {
		port = "443"
	}
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "80"
	}

	useHTTP := false
	if os.Getenv("CERT_FILE") == "" && os.Getenv("KEY_FILE") == "" && domain == "" {
		scheme = "http"
		if port == "443" {
			port = "2020"
		}
		domain = "localhost:" + port
		useHTTP = true
	} else if domain == "" {
		domain = "localhost"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, domain)
	if scheme == "https" && port != "443" {
		baseURL = fmt.Sprintf("%s://%s:%s", scheme, domain, port)
	} else if scheme == "http" && port != "80" {
		baseURL = fmt.Sprintf("%s://%s:%s", scheme, domain, port)
	}

	googleKey := os.Getenv("GOOGLE_KEY")
	googleSecret := os.Getenv("GOOGLE_SECRET")
	facebookKey := os.Getenv("FACEBOOK_KEY")
	facebookSecret := os.Getenv("FACEBOOK_SECRET")
	githubKey := os.Getenv("GITHUB_KEY")
	githubSecret := os.Getenv("GITHUB_SECRET")
	twitterKey := os.Getenv("TWITTER_KEY")
	twitterSecret := os.Getenv("TWITTER_SECRET")

	goth.UseProviders(
		google.New(googleKey, googleSecret, baseURL+"/auth/google/callback", "email", "profile"),
		facebook.New(facebookKey, facebookSecret, baseURL+"/auth/facebook/callback", "email", "public_profile"),
		github.New(githubKey, githubSecret, baseURL+"/auth/github/callback", "user", "user:email"),
		twitter.New(twitterKey, twitterSecret, baseURL+"/auth/twitter/callback"),
	)

	if err := database.InitDB("./forum.db"); err != nil {
		log.Fatal(err)
	}
	defer database.CloseDB()

	seedDefaultUsers()

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", handler.RedirectToIndex)
	mux.HandleFunc("/index", handler.IndexHandler)
	mux.HandleFunc("/inscription", handler.InscriptionHandler)
	mux.HandleFunc("/connexion", handler.ConnexionHandler)
	mux.HandleFunc("/deconnexion", handler.DeconnexionHandler)
	mux.HandleFunc("/profil", handler.ProfilHandler)
	mux.HandleFunc("/modify-profil", handler.ModifyProfileHandler)
	mux.HandleFunc("/api-tmdb", handler.TmdbHandler)
	mux.HandleFunc("/actualites", handler.ActualitesHandler)
	mux.HandleFunc("/nouveau-post", handler.NewPostHandler)
	mux.HandleFunc("/posts", handler.PostsHandler)
	mux.HandleFunc("/post", handler.PostDetailHandler)
	mux.HandleFunc("/delete-post", handler.DeletePostHandler)
	mux.HandleFunc("/edit-post", handler.EditPostHandler)
	mux.HandleFunc("/add-comment", handler.AddCommentHandler)
	mux.HandleFunc("/delete-comment", handler.DeleteCommentHandler)
	mux.HandleFunc("/notifications", handler.NotificationsHandler)
	mux.HandleFunc("/notifications-page", handler.NotificationsPageHandler)
	mux.HandleFunc("/notifications/mark-read", handler.MarkNotificationsAsReadHandler)
	mux.HandleFunc("/like-post", handler.LikePostHandler)
	mux.HandleFunc("/dislike-post", handler.DislikePostHandler)
	mux.HandleFunc("/like-comment", handler.LikeCommentHandler)
	mux.HandleFunc("/dislike-comment", handler.DislikeCommentHandler)
	mux.HandleFunc("/auth/google", handler.GoogleAuthHandler)
	mux.HandleFunc("/auth/google/callback", handler.GoogleCallbackHandler)
	mux.HandleFunc("/auth/facebook", handler.FacebookAuthHandler)
	mux.HandleFunc("/auth/facebook/callback", handler.FacebookCallbackHandler)
	mux.HandleFunc("/auth/github", handler.GithubAuthHandler)
	mux.HandleFunc("/auth/github/callback", handler.GithubCallbackHandler)
	mux.HandleFunc("/auth/twitter", handler.TwitterAuthHandler)
	mux.HandleFunc("/auth/twitter/callback", handler.TwitterCallbackHandler)
	mux.HandleFunc("/moderation", handler.ModerationDashboardHandler)
	mux.HandleFunc("/moderation/approve", handler.ApprovePostHandler)
	mux.HandleFunc("/moderation/reject", handler.RejectPostHandler)
	mux.HandleFunc("/admin/promote", handler.PromoteUserHandler)
	mux.HandleFunc("/admin/demote", handler.DemoteUserHandler)
	mux.HandleFunc("/admin/users", handler.AdminUsersHandler)
	mux.HandleFunc("/admin/users/update", handler.AdminUsersUpdateHandler)
	mux.HandleFunc("/report-post", handler.ReportPostHandler)
	mux.HandleFunc("/admin/reports", handler.AdminReportsHandler)
	mux.HandleFunc("/admin/reports/respond", handler.RespondReportHandler)
	mux.HandleFunc("/gemini-chat", handler.GeminiChatPage)
	mux.HandleFunc("/api/gemini-chat", handler.GeminiChatAPI)

	handlerWithRate := middleware.RateLimit(middleware.GzipAndCacheMiddleware(mux))

	serverAddr := ":" + port

	commonTLSConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
		PreferServerCipherSuites: true,
	}

	srv := &http.Server{
		Addr:              serverAddr,
		Handler:           handlerWithRate,
		TLSConfig:         commonTLSConfig,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	certFile := os.Getenv("CERT_FILE")
	keyFile := os.Getenv("KEY_FILE")

	if useHTTP {
		urlToDisplay := baseURL
		if scheme == "http" && port == "80" {
			urlToDisplay = fmt.Sprintf("http://%s", domain)
		} else if scheme == "http" && domain == "localhost:"+port {
			urlToDisplay = fmt.Sprintf("http://localhost:%s", port)
		}

		fmt.Printf("✅ Serveur HTTP démarré. Ouvrez ce lien dans votre navigateur :\n")
		fmt.Printf("\x1b]8;;%s\x07%s\x1b]8;;\x07\n", urlToDisplay, urlToDisplay)
		log.Fatal(srv.ListenAndServe())
	} else if certFile != "" && keyFile != "" {
		urlToDisplay := baseURL
		if port == "443" {
			urlToDisplay = fmt.Sprintf("https://%s", domain)
		}

		fmt.Printf("✅ Serveur HTTPS (certificats locaux) démarré. Ouvrez ce lien dans votre navigateur :\n")
		fmt.Printf("\x1b]8;;%s\x07%s\x1b]8;;\x07\n", urlToDisplay, urlToDisplay)
		log.Fatal(srv.ListenAndServeTLS(certFile, keyFile))
	} else if domain != "" && domain != "localhost" {
		fmt.Println("🚀 Configuration avec Autocert pour le domaine:", domain)
		m := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(domain),
			Cache:      autocert.DirCache("cert-cache"),
		}

		srv.TLSConfig.GetCertificate = m.GetCertificate

		go func() {
			httpSrv := &http.Server{
				Addr:              ":" + httpPort,
				Handler:           m.HTTPHandler(nil),
				ReadHeaderTimeout: 5 * time.Second,
				WriteTimeout:      10 * time.Second,
				IdleTimeout:       120 * time.Second,
			}
			log.Printf("🚀 Démarrage du serveur HTTP sur le port %s pour le challenge ACME...\n", httpPort)
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("❌ Erreur serveur HTTP pour ACME: %v", err)
			}
		}()

		urlToDisplay := baseURL
		if port == "443" {
			urlToDisplay = fmt.Sprintf("https://%s", domain)
		}

		fmt.Printf("✅ Serveur HTTPS (Autocert) démarré pour %s. Accédez via :\n", domain)
		fmt.Printf("\x1b]8;;%s\x07%s\x1b]8;;\x07\n", urlToDisplay, urlToDisplay)
		log.Fatal(srv.ListenAndServeTLS("", ""))
	} else {
		fmt.Println("⚠️ Configuration HTTPS ambiguë ou incomplète. Veuillez vérifier les variables d'environnement CERT_FILE, KEY_FILE, et DOMAIN.")
		fmt.Println("ℹ️  Démarrage sur HTTP sur localhost:2020 par défaut comme solution de repli.")
		srv.Addr = "localhost:2020"
		srv.TLSConfig = nil
		urlToDisplay := "http://localhost:2020"
		fmt.Printf("✅ Serveur HTTP (repli) démarré. Ouvrez ce lien dans votre navigateur :\n")
		fmt.Printf("\x1b]8;;%s\x07%s\x1b]8;;\x07\n", urlToDisplay, urlToDisplay)
		log.Fatal(srv.ListenAndServe())
	}
}