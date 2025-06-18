package handler

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"forum/database"
)

func NewPostHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	userID, _ := strconv.Atoi(cookie.Value)

	switch r.Method {
	case http.MethodGet:
		cats, _ := database.GetAllCategories()
		t, _ := template.ParseFiles(filepath.Join("templates", "new_post.html"))
		_ = t.Execute(w, struct{ Categories []database.Category }{cats})

	case http.MethodPost:
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "form error", http.StatusBadRequest)
			return
		}
		title := r.FormValue("title")
		content := r.FormValue("content")
		if title == "" || content == "" {
			http.Error(w, "missing fields", http.StatusBadRequest)
			return
		}

		var imagePath string
		if file, header, err := r.FormFile("image"); err == nil {
			defer file.Close()
			dir := filepath.Join("static", "uploads")
			os.MkdirAll(dir, os.ModePerm)
			imagePath = filepath.Join(dir, header.Filename)
			dst, _ := os.Create(imagePath)
			defer dst.Close()
			_, _ = io.Copy(dst, file)
		}

		var catIDs []int
		for _, s := range r.Form["categories"] {
			if id, err := strconv.Atoi(s); err == nil {
				catIDs = append(catIDs, id)
			}
		}

		if err := database.CreatePost(userID, title, content, imagePath, catIDs); err != nil {
			http.Error(w, "create post error", http.StatusInternalServerError)
			return
		}

		if user, e := database.GetUserWithRole(userID); e == nil {
			if last, e2 := database.GetLastPostForUser(userID); e2 == nil {
				if user.Role == "admin" || user.Role == "moderator" {
					_ = database.CreateNotification(userID,
						fmt.Sprintf("Votre post « %s » a bien été publié.", title),
						last.ID, 0)
				} else {
					_ = database.CreateNotification(userID,
						fmt.Sprintf("Votre post « %s » est en attente de validation.", title),
						last.ID, 0)
					if mods, _ := database.GetModeratorsAndAdmins(); mods != nil {
						for _, m := range mods {
							_ = database.CreateNotification(m.ID,
								fmt.Sprintf("Nouveau post « %s » à vérifier.", title),
								last.ID, 0)
						}
					}
				}
			}
		}

		http.Redirect(w, r, "/posts", http.StatusSeeOther)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func PostsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := database.GetAllPosts()
	if err != nil {
		http.Error(w, "cannot fetch posts", http.StatusInternalServerError)
		return
	}

	isMod := false
	if c, err := r.Cookie("user_id"); err == nil {
		if uid, e := strconv.Atoi(c.Value); e == nil {
			if u, e2 := database.GetUserWithRole(uid); e2 == nil && (u.Role == "admin" || u.Role == "moderator") {
				isMod = true
			}
		}
	}

	t, _ := template.ParseFiles(filepath.Join("templates", "posts.html"))
	_ = t.Execute(w, struct {
		Posts       []database.Post
		IsModerator bool
	}{posts, isMod})
}

func PostDetailHandler(w http.ResponseWriter, r *http.Request) {
    idStr := r.URL.Query().Get("id")
    if idStr == "" {
        http.Error(w, "missing id", http.StatusBadRequest)
        return
    }
    id, _ := strconv.Atoi(idStr)

    // On identifie l'utilisateur courant et s'il est modo/admin
    currentUserID := 0
    isMod := false
    if c, err := r.Cookie("user_id"); err == nil {
        if uid, e := strconv.Atoi(c.Value); e == nil {
            currentUserID = uid
            if u, e2 := database.GetUserWithRole(uid); e2 == nil &&
                (u.Role == "admin" || u.Role == "moderator") {
                isMod = true
            }
        }
    }

    // Chargement du post (avec ou sans droits de modo)
    var post database.Post
    var err error
    if isMod {
        post, err = database.GetPostByIDAny(id)
    } else {
        post, err = database.GetPostByID(id)
        if err != nil && currentUserID != 0 {
            // Permettre à l'auteur de voir son post en pending
            if p, e := database.GetPostByIDAny(id); e == nil && p.UserID == currentUserID {
                post = p
                err = nil
            }
        }
    }
    if err != nil {
        http.Error(w, "post not found", http.StatusNotFound)
        return
    }

    // Peut-on modifier ?
    editable := currentUserID == post.UserID || isMod

    // Photo de l'auteur du post (pas de l'utilisateur courant)
    authorPhoto := "profil.png"
    if u, e := database.GetUserByID(post.UserID); e == nil && u.Photo != "" {
        authorPhoto = u.Photo
    }

    // Récupération des commentaires
    comments, _ := database.GetCommentsByPostID(post.ID)

    // Le post a-t-il été modifié ?
    modified := !post.ModifiedAt.IsZero() &&
        post.OriginalContent != "" &&
        post.OriginalContent != post.Content

    // Exécution du template avec AuthorPhoto à la place de UserPhoto
    t, _ := template.ParseFiles(filepath.Join("templates", "post_detail.html"))
    _ = t.Execute(w, struct {
        Post        database.Post
        Editable    bool
        Comments    []database.Comment
        AuthorPhoto string
        Modified    bool
    }{
        Post:        post,
        Editable:    editable,
        Comments:    comments,
        AuthorPhoto: authorPhoto,
        Modified:    modified,
    })
}

func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	userID, _ := strconv.Atoi(cookie.Value)

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	postID, _ := strconv.Atoi(idStr)

	user, _ := database.GetUserWithRole(userID)
	if user.Role == "admin" || user.Role == "moderator" {
		_ = database.AdminDeletePost(postID)
	} else {
		_ = database.DeletePost(postID, userID)
	}
	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}

func EditPostHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	userID, _ := strconv.Atoi(cookie.Value)

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	postID, _ := strconv.Atoi(idStr)

	switch r.Method {
	case http.MethodGet:
		post, err := database.GetPostByID(postID)
		if err != nil || post.UserID != userID {
			http.Error(w, "unauthorized", http.StatusForbidden)
			return
		}
		t, _ := template.ParseFiles(filepath.Join("templates", "edit_post.html"))
		_ = t.Execute(w, post)

	case http.MethodPost:
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "form error", http.StatusBadRequest)
			return
		}
		title := r.FormValue("title")
		content := r.FormValue("content")
		if content == "" {
			http.Error(w, "missing content", http.StatusBadRequest)
			return
		}

		var imagePath string
		if file, header, err := r.FormFile("image"); err == nil {
			defer file.Close()
			dir := filepath.Join("static", "uploads")
			os.MkdirAll(dir, os.ModePerm)
			imagePath = filepath.Join(dir, header.Filename)
			dst, _ := os.Create(imagePath)
			defer dst.Close()
			_, _ = io.Copy(dst, file)
		} else {
			if p, e := database.GetPostByID(postID); e == nil {
				imagePath = p.ImagePath
			}
		}

		_ = database.UpdatePost(postID, userID, title, content, imagePath)
		http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}