package database

import (
	"database/sql"
	"fmt"
	"log" 
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dbFilePath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbFilePath+"?_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	if _, err := DB.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Printf("Warning: Could not enable foreign key support: %v", err)
	}
	if err = createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	return nil
}

func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

func createTables() error {
	sqlPath := filepath.Join("database", "SQL", "database.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}
	if _, err := DB.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute SQL commands: %w", err)
	}
	return nil
}

type User struct {
	ID        int
	Username  string
	Email     string
	Password  string
	CreatedAt string
	Photo     string
	Role      string
}

type Category struct {
	ID   int
	Name string
}

type Post struct {
	ID              int
	UserID          int
	Username        string
	Title           string
	Content         string
	OriginalContent string
	ImagePath       string
	CreatedAt       time.Time
	ModifiedAt      time.Time
	Likes           int
	Dislikes        int
	Categories      []string
}

type Comment struct {
	ID        int
	PostID    int
	UserID    int
	Username  string
	Content   string
	CreatedAt time.Time
	Likes     int
	Dislikes  int
	Photo     string
}

type Notification struct {
	ID        int
	UserID    int
	Message   string
	PostID    int
	CommentID int
	CreatedAt time.Time
}

func CreateUser(username, email, password string) error {
	_, err := DB.Exec(
		"INSERT INTO users (username, email, password, photo) VALUES (?, ?, ?, ?);",
		username, email, password, "profil.png")
	return err
}

func GetUserByEmail(email string) (User, error) {
	var u User
	err := DB.QueryRow(
		"SELECT id, username, email, password, created_at, photo, role FROM users WHERE email = ?;",
		email).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.CreatedAt, &u.Photo, &u.Role)
	return u, err
}

func GetUserByUsername(username string) (User, error) {
	var u User
	err := DB.QueryRow(
		"SELECT id, username, email, password, created_at, photo, role FROM users WHERE username = ?;",
		username).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.CreatedAt, &u.Photo, &u.Role)
	return u, err
}

func GetUserWithRole(id int) (User, error) {
	var u User
	err := DB.QueryRow(
		"SELECT id, username, email, password, created_at, photo, role FROM users WHERE id = ?;",
		id).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.CreatedAt, &u.Photo, &u.Role)
	if err == nil && u.Photo == "" {
		u.Photo = "default.png" 
	}
	return u, err
}

func GetUserByID(id int) (User, error) {
	return GetUserWithRole(id)
}


func UpdateUserRole(userID int, role string) error {
	_, err := DB.Exec("UPDATE users SET role = ? WHERE id = ?;", role, userID)
	return err
}

func GetModeratorsAndAdmins() ([]User, error) {
	rows, err := DB.Query(
		"SELECT id, username, email, password, created_at, photo, role FROM users WHERE role IN ('admin','moderator');")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.CreatedAt, &u.Photo, &u.Role); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, rows.Err()
}

func GetAllCategories() ([]Category, error) {
	rows, err := DB.Query("SELECT id, name FROM categories ORDER BY name;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func GetCategoriesByPostID(postID int) ([]string, error) {
	rows, err := DB.Query(
		`SELECT c.name FROM categories c
         JOIN post_categories pc ON pc.category_id = c.id
         WHERE pc.post_id = ?;`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, rows.Err()
}

func CreatePost(userID int, title, content, imagePath string, categoryIDs []int) error {
	user, err := GetUserWithRole(userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user role: %w", err)
	}
	status := "pending"
	if user.Role == "admin" || user.Role == "moderator" {
		status = "approved"
	}

	res, err := DB.Exec(
		`INSERT INTO posts (user_id, title, content, original_content, image_path, moderation_status)
         VALUES (?, ?, ?, ?, ?, ?);`,
		userID, title, content, content, imagePath, status)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}

	if len(categoryIDs) > 0 {
		postID, _ := res.LastInsertId()
		stmt, err := DB.Prepare("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?);")
		if err != nil {
			log.Printf("Error preparing statement for post_categories: %v", err)
			return fmt.Errorf("failed to prepare statement for categories: %w", err)
		}
		defer stmt.Close()
		for _, cid := range categoryIDs {
			if _, err := stmt.Exec(postID, cid); err != nil {
				log.Printf("Error inserting post_category link (%d, %d): %v", postID, cid, err)
			}
		}
	}
	return nil
}

func parseSQLiteDate(ts string) time.Time {
    if ts == "" {
        return time.Time{} 
    }
	parsed, err := time.Parse(time.RFC3339, ts) 
	if err != nil {
		parsed, err = time.Parse("2006-01-02 15:04:05", ts)
        if err != nil {
             log.Printf("Failed to parse date string '%s': %v", ts, err)
             return time.Time{}
        }
	}

    return parsed 
}


func enrichPostWithMeta(p *Post) {
	p.Likes, _ = CountPostLikes(p.ID)
	p.Dislikes, _ = CountPostDislikes(p.ID)
	p.Categories, _ = GetCategoriesByPostID(p.ID)
}

func GetAllPosts() ([]Post, error) {
	rows, err := DB.Query(
		`SELECT p.id, p.user_id, u.username, p.title, p.content,
                p.original_content, p.image_path, p.created_at, p.modified_at
           FROM posts p
           JOIN users u ON p.user_id = u.id
           WHERE p.moderation_status = 'approved'
           ORDER BY p.created_at DESC;`)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}
	defer rows.Close()

	var list []Post
	for rows.Next() {
		var p Post
		var created, modified sql.NullString
		if err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content,
			&p.OriginalContent, &p.ImagePath, &created, &modified); err != nil {
			return nil, fmt.Errorf("failed to scan post row: %w", err)
		}
		if created.Valid {
			p.CreatedAt = parseSQLiteDate(created.String)
		}
		if modified.Valid {
			p.ModifiedAt = parseSQLiteDate(modified.String)
		}
		enrichPostWithMeta(&p)
		list = append(list, p)
	}
	return list, rows.Err()
}

func GetRecentPosts(limit int) ([]Post, error) {
	rows, err := DB.Query(
		`SELECT p.id, p.user_id, u.username, p.title, p.content,
                p.original_content, p.image_path, p.created_at, p.modified_at
           FROM posts p
           JOIN users u ON p.user_id = u.id
           WHERE p.moderation_status = 'approved'
           ORDER BY p.created_at DESC
           LIMIT ?;`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent posts: %w", err)
	}
	defer rows.Close()

	var list []Post
	for rows.Next() {
		var p Post
		var created, modified sql.NullString
		if err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content,
			&p.OriginalContent, &p.ImagePath, &created, &modified); err != nil {
			return nil, fmt.Errorf("failed to scan recent post row: %w", err)
		}
		if created.Valid {
            p.CreatedAt = parseSQLiteDate(created.String)
        }
		if modified.Valid {
			p.ModifiedAt = parseSQLiteDate(modified.String)
		}
		enrichPostWithMeta(&p)
		list = append(list, p)
	}
	return list, rows.Err()
}

func GetPostByID(id int) (Post, error) {
	row := DB.QueryRow(
		`SELECT p.id, p.user_id, u.username, p.title, p.content,
                p.original_content, p.image_path, p.created_at, p.modified_at
           FROM posts p
           JOIN users u ON p.user_id = u.id
           WHERE p.id = ? AND p.moderation_status = 'approved';`, id)
	var p Post
	var created, modified sql.NullString
	if err := row.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content,
		&p.OriginalContent, &p.ImagePath, &created, &modified); err != nil {
		// Retourner sql.ErrNoRows si non trouvé, ou une erreur formatée sinon
		if err == sql.ErrNoRows {
			return p, err
		}
		return p, fmt.Errorf("failed to get post %d: %w", id, err)
	}
	if created.Valid {
        p.CreatedAt = parseSQLiteDate(created.String)
    }
	if modified.Valid {
		p.ModifiedAt = parseSQLiteDate(modified.String)
	}
	enrichPostWithMeta(&p)
	return p, nil
}


func GetPostByIDAny(id int) (Post, error) {
	row := DB.QueryRow(
		`SELECT p.id, p.user_id, u.username, p.title, p.content,
                p.original_content, p.image_path, p.created_at, p.modified_at
           FROM posts p
           JOIN users u ON p.user_id = u.id
           WHERE p.id = ?;`, id)
	var p Post
	var created, modified sql.NullString
	if err := row.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content,
		&p.OriginalContent, &p.ImagePath, &created, &modified); err != nil {
        if err == sql.ErrNoRows {
			return p, err
		}
		return p, fmt.Errorf("failed to get post any status %d: %w", id, err)
	}
	if created.Valid {
        p.CreatedAt = parseSQLiteDate(created.String)
    }
	if modified.Valid {
		p.ModifiedAt = parseSQLiteDate(modified.String)
	}
	enrichPostWithMeta(&p)
	return p, nil
}

func GetPendingPosts() ([]Post, error) {
	rows, err := DB.Query(
		`SELECT p.id, p.user_id, u.username, p.title, p.content,
                p.original_content, p.image_path, p.created_at, p.modified_at
           FROM posts p
           JOIN users u ON p.user_id = u.id
           WHERE p.moderation_status = 'pending'
           ORDER BY p.created_at DESC;`)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending: %w", err)
	}
	defer rows.Close()
	var list []Post
	for rows.Next() {
		var p Post
		var created, modified sql.NullString
		if err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content,
			&p.OriginalContent, &p.ImagePath, &created, &modified); err != nil {
			return nil, fmt.Errorf("failed to scan pending: %w", err)
		}
		if created.Valid {
            p.CreatedAt = parseSQLiteDate(created.String)
        }
		if modified.Valid {
			p.ModifiedAt = parseSQLiteDate(modified.String)
		}
		enrichPostWithMeta(&p)
		list = append(list, p)
	}
	return list, rows.Err()
}

func GetLastPostForUser(userID int) (Post, error) {
	row := DB.QueryRow(
		`SELECT p.id, p.user_id, u.username, p.title, p.content,
                p.original_content, p.image_path, p.created_at, p.modified_at
           FROM posts p
           JOIN users u ON p.user_id = u.id
           WHERE p.user_id = ?
           ORDER BY p.created_at DESC
           LIMIT 1;`, userID)
	var p Post
	var created, modified sql.NullString
	err := row.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content,
		&p.OriginalContent, &p.ImagePath, &created, &modified)
	if err != nil {
		return p, err 
	}
	if created.Valid {
        p.CreatedAt = parseSQLiteDate(created.String)
    }
	if modified.Valid {
		p.ModifiedAt = parseSQLiteDate(modified.String)
	}
	enrichPostWithMeta(&p)
	return p, nil
}

func UpdatePost(postID, userID int, title, content, imagePath string) error {
    var oldContent string
	var oldOriginal sql.NullString
	err := DB.QueryRow(
		"SELECT content, original_content FROM posts WHERE id = ? AND user_id = ?;",
		postID, userID).Scan(&oldContent, &oldOriginal)
    if err != nil {
		return fmt.Errorf("cannot retrieve post %d for update: %w", postID, err)
	}
	if !oldOriginal.Valid || oldOriginal.String == "" {
		oldOriginal.String = oldContent
        oldOriginal.Valid = true
	}
	_, err = DB.Exec(
		`UPDATE posts SET content = ?, modified_at = CURRENT_TIMESTAMP, original_content = ?
           WHERE id = ? AND user_id = ?;`,
		content, oldOriginal.String, postID, userID)

    if err != nil {
        return fmt.Errorf("failed to update post %d: %w", postID, err)
    }
	return nil
}


func DeletePost(postID, userID int) error {
	tx, err := DB.Begin()
	if err != nil { return fmt.Errorf("delete post: begin tx: %w", err)}
	defer tx.Rollback() 

	_, err = tx.Exec("DELETE FROM likes WHERE post_id = ?;", postID)
	if err != nil { return fmt.Errorf("delete post likes: %w", err)}
	_, err = tx.Exec("DELETE FROM comments WHERE post_id = ?;", postID)
	if err != nil { return fmt.Errorf("delete post comments: %w", err)}
	_, err = tx.Exec("DELETE FROM post_categories WHERE post_id = ?;", postID)
	if err != nil { return fmt.Errorf("delete post categories: %w", err)}
	_, err = tx.Exec("DELETE FROM notifications WHERE post_id = ?;", postID)
	if err != nil { return fmt.Errorf("delete post notifications: %w", err)}

	result, err := tx.Exec("DELETE FROM posts WHERE id = ? AND user_id = ?;", postID, userID)
    if err != nil { return fmt.Errorf("delete post record: %w", err)}

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return fmt.Errorf("post %d not found or not owned by user %d", postID, userID)
    }

	return tx.Commit()
}

func AdminDeletePost(postID int) error {
	tx, err := DB.Begin()
	if err != nil { return fmt.Errorf("admin delete post: begin tx: %w", err)}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM likes WHERE post_id = ?;", postID)
	if err != nil { return fmt.Errorf("admin delete post likes: %w", err)}
	_, err = tx.Exec("DELETE FROM comments WHERE post_id = ?;", postID)
	if err != nil { return fmt.Errorf("admin delete post comments: %w", err)}
	_, err = tx.Exec("DELETE FROM post_categories WHERE post_id = ?;", postID)
	if err != nil { return fmt.Errorf("admin delete post categories: %w", err)}
	_, err = tx.Exec("DELETE FROM notifications WHERE post_id = ?;", postID)
	if err != nil { return fmt.Errorf("admin delete post notifications: %w", err)}

	result, err := tx.Exec("DELETE FROM posts WHERE id = ?;", postID)
    if err != nil { return fmt.Errorf("admin delete post record: %w", err)}

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return fmt.Errorf("admin delete: post %d not found", postID)
    }

	return tx.Commit()
}


func SetPostModerationStatus(postID int, status string) error {
	_, err := DB.Exec("UPDATE posts SET moderation_status = ? WHERE id = ?;", status, postID)
	return err
}

func CreateComment(postID, userID int, content string) error {
	_, err := DB.Exec(
		"INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?);",
		postID, userID, content)
	return err
}

func DeleteComment(commentID, userID int) error {
	tx, err := DB.Begin()
	if err != nil { return fmt.Errorf("delete comment: begin tx: %w", err)}
	defer tx.Rollback()

	var exists int
	err = tx.QueryRow("SELECT 1 FROM comments WHERE id = ? AND user_id = ?", commentID, userID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("comment %d not found or not owned by user %d", commentID, userID)
		}
		return fmt.Errorf("check comment ownership: %w", err)
	}

	_, err = tx.Exec("DELETE FROM likes WHERE comment_id = ?;", commentID)
	if err != nil { return fmt.Errorf("delete comment likes: %w", err)}
	_, err = tx.Exec("DELETE FROM notifications WHERE comment_id = ?;", commentID)
	if err != nil { return fmt.Errorf("delete comment notifications: %w", err)}

	result, err := tx.Exec("DELETE FROM comments WHERE id = ? AND user_id = ?;", commentID, userID)
	if err != nil { return fmt.Errorf("delete comment record: %w", err)}

	rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 { 
        return fmt.Errorf("comment %d not found or not owned by user %d during delete", commentID, userID)
    }

	return tx.Commit()
}

func AdminDeleteComment(commentID int) error {
	tx, err := DB.Begin()
	if err != nil { return fmt.Errorf("admin delete comment: begin tx: %w", err)}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM likes WHERE comment_id = ?;", commentID)
	if err != nil { return fmt.Errorf("admin delete comment likes: %w", err)}
	_, err = tx.Exec("DELETE FROM notifications WHERE comment_id = ?;", commentID)
	if err != nil { return fmt.Errorf("admin delete comment notifications: %w", err)}

	result, err := tx.Exec("DELETE FROM comments WHERE id = ?;", commentID)
	if err != nil { return fmt.Errorf("admin delete comment record: %w", err)}

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return fmt.Errorf("admin delete: comment %d not found", commentID)
    }

	return tx.Commit()
}


func GetCommentsByPostID(postID int) ([]Comment, error) {
	rows, err := DB.Query(
		`SELECT c.id, c.post_id, c.user_id, u.username, c.content,
                c.created_at, u.photo
           FROM comments c
           JOIN users u ON c.user_id = u.id
           WHERE c.post_id = ?
           ORDER BY c.created_at ASC;`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Comment
	for rows.Next() {
		var c Comment
		var created string
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username,
			&c.Content, &created, &c.Photo); err != nil {
			return nil, err
		}
        c.CreatedAt = parseSQLiteDate(created)
		c.Likes, _ = CountCommentLikes(c.ID)
		c.Dislikes, _ = CountCommentDislikes(c.ID)
        if c.Photo == "" {
            c.Photo = "default.png" 
        }
		list = append(list, c)
	}
	return list, rows.Err()
}

func GetCommentByID(commentID int) (Comment, error) {
	row := DB.QueryRow(
		`SELECT c.id, c.post_id, c.user_id, u.username, c.content,
                c.created_at, u.photo
           FROM comments c
           JOIN users u ON c.user_id = u.id
           WHERE c.id = ?;`, commentID)
	var c Comment
	var created string
	if err := row.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username,
		&c.Content, &created, &c.Photo); err != nil {
		return c, err 
	}
	c.CreatedAt = parseSQLiteDate(created)
	c.Likes, _ = CountCommentLikes(c.ID)
	c.Dislikes, _ = CountCommentDislikes(c.ID)
    if c.Photo == "" {
       c.Photo = "default.png"
    }
	return c, nil
}

func SetPostLike(userID, postID, value int) error {
	_, err := DB.Exec(
		`INSERT INTO likes (user_id, post_id, value, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)
         ON CONFLICT(user_id, post_id) DO UPDATE SET value = excluded.value, created_at = excluded.created_at;`,
		userID, postID, value)
	if err != nil {
		log.Printf("Error setting post like (user: %d, post: %d, value: %d): %v", userID, postID, value, err)
	}
	return err
}

func SetCommentLike(userID, commentID, value int) error {
	_, err := DB.Exec(
		`INSERT INTO likes (user_id, comment_id, value, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)
         ON CONFLICT(user_id, comment_id) DO UPDATE SET value = excluded.value, created_at = excluded.created_at;`,
		userID, commentID, value)
	if err != nil {
		log.Printf("Error setting comment like (user: %d, comment: %d, value: %d): %v", userID, commentID, value, err)
	}
	return err
}


func CountPostLikes(postID int) (int, error) {
	var c int
	err := DB.QueryRow(
		"SELECT COUNT(*) FROM likes WHERE post_id = ? AND value = 1;", postID).Scan(&c)
	return c, err
}

func CountPostDislikes(postID int) (int, error) {
	var c int
	err := DB.QueryRow(
		"SELECT COUNT(*) FROM likes WHERE post_id = ? AND value = -1;", postID).Scan(&c)
	return c, err
}

func CountCommentLikes(commentID int) (int, error) {
	var c int
	err := DB.QueryRow(
		"SELECT COUNT(*) FROM likes WHERE comment_id = ? AND value = 1;", commentID).Scan(&c)
	return c, err
}

func CountCommentDislikes(commentID int) (int, error) {
	var c int
	err := DB.QueryRow(
		"SELECT COUNT(*) FROM likes WHERE comment_id = ? AND value = -1;", commentID).Scan(&c)
	return c, err
}

func CreateNotification(userID int, message string, postID, commentID int) error {
    var postIDNullable sql.NullInt64
	var commentIDNullable sql.NullInt64

	if postID != 0 {
		postIDNullable = sql.NullInt64{Int64: int64(postID), Valid: true}
	}
	if commentID != 0 {
		commentIDNullable = sql.NullInt64{Int64: int64(commentID), Valid: true}
	}

	_, err := DB.Exec(
		"INSERT INTO notifications (user_id, message, post_id, comment_id) VALUES (?, ?, ?, ?);",
		userID, message, postIDNullable, commentIDNullable)
	return err
}


func GetNotificationsByUserID(userID int) ([]Notification, error) {
	rows, err := DB.Query(
		`SELECT id, user_id, message, post_id, comment_id, created_at
           FROM notifications
           WHERE user_id = ?
           ORDER BY created_at DESC;`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Notification
	for rows.Next() {
		var n Notification
		var created string
        var postIDNullable, commentIDNullable sql.NullInt64
		if err := rows.Scan(&n.ID, &n.UserID, &n.Message, &postIDNullable, &commentIDNullable, &created); err != nil {
			return nil, err
		}
        n.CreatedAt = parseSQLiteDate(created)
        if postIDNullable.Valid {
            n.PostID = int(postIDNullable.Int64)
        }
        if commentIDNullable.Valid {
            n.CommentID = int(commentIDNullable.Int64)
        }
		list = append(list, n)
	}
	return list, rows.Err()
}

func DeleteNotificationsByUserID(userID int) error {
	_, err := DB.Exec("DELETE FROM notifications WHERE user_id = ?;", userID)
	return err
}

func CreateSession(sessionID string, userID int, expiresAt time.Time) error {
	_, err := DB.Exec(
		"INSERT INTO sessions (session_id, user_id, expires_at) VALUES (?, ?, ?);",
		sessionID, userID, expiresAt.Format("2006-01-02 15:04:05"))
	return err
}

func GetUserIDBySession(sessionID string) (int, error) {
	var uid int
	var exp string
	err := DB.QueryRow(
		"SELECT user_id, expires_at FROM sessions WHERE session_id = ?;", sessionID).Scan(&uid, &exp)
	if err != nil {
		return 0, err 
	}
	e, err := time.Parse("2006-01-02 15:04:05", exp)
    if err != nil {
        log.Printf("Invalid expiry date format for session %s: %v", sessionID, err)
         _ = DeleteSession(sessionID)
        return 0, fmt.Errorf("session date invalide")
    }
	if time.Now().After(e) {
		_ = DeleteSession(sessionID)
		return 0, fmt.Errorf("session expirée")
	}
	return uid, nil
}

func DeleteSession(sessionID string) error {
	_, err := DB.Exec("DELETE FROM sessions WHERE session_id = ?;", sessionID)
	return err
}

func GetUserStats(userID int) (int, int, error) {
	var liked, comments int
	err := DB.QueryRow(
		"SELECT COUNT(*) FROM likes WHERE user_id = ? AND post_id IS NOT NULL AND value = 1;",
		userID).Scan(&liked)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get liked posts count: %w", err)
	}
	err = DB.QueryRow(
		"SELECT COUNT(*) FROM comments WHERE user_id = ?;", userID).Scan(&comments)
    if err != nil {
        return liked, 0, fmt.Errorf("failed to get comments count: %w", err)
    }
	return liked, comments, nil
}


func GetLastPostDate(userID int) (time.Time, error) {
	var ts sql.NullString 
	err := DB.QueryRow("SELECT MAX(created_at) FROM posts WHERE user_id = ?;", userID).Scan(&ts)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, nil 
		}
		return time.Time{}, err
	}
    if !ts.Valid || ts.String == "" {
        return time.Time{}, nil 
    }
	return parseSQLiteDate(ts.String), nil
}

func GetLastCommentDate(userID int) (time.Time, error) {
	var ts sql.NullString
	err := DB.QueryRow("SELECT MAX(created_at) FROM comments WHERE user_id = ?;", userID).Scan(&ts)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
    if !ts.Valid || ts.String == "" {
        return time.Time{}, nil
    }
	return parseSQLiteDate(ts.String), nil
}

func GetLastLikeDate(userID int) (time.Time, error) {
	var ts sql.NullString
	err := DB.QueryRow("SELECT MAX(created_at) FROM likes WHERE user_id = ?;", userID).Scan(&ts)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
     if !ts.Valid || ts.String == "" {
        return time.Time{}, nil
    }
	return parseSQLiteDate(ts.String), nil
}


func GetLastActivityDate(userID int) (time.Time, error) {
	p1, _ := GetLastCommentDate(userID) 
	p2, _ := GetLastLikeDate(userID)

	if p1.IsZero() && p2.IsZero() {
		return time.Time{}, fmt.Errorf("no activity found") 
	}
	if p1.After(p2) {
		return p1, nil
	}
	return p2, nil
}

func GetLastConnection(userID int) (time.Time, error) {
    var ts sql.NullString
    err := DB.QueryRow(
        "SELECT MAX(created_at) FROM sessions WHERE user_id = ? ORDER BY created_at DESC LIMIT 1;",
        userID).Scan(&ts)
	if err != nil || !ts.Valid || ts.String == "" {
		return time.Time{}, err
	}
	return parseSQLiteDate(ts.String), nil
}

func DeleteUserAndAssociatedData(userID int) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("delete user: failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			log.Printf("Rolling back transaction for user %d deletion due to error: %v", userID, err)
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				log.Printf("Error committing transaction for user %d deletion: %v", userID, err)
			} else {
				log.Printf("Successfully deleted user %d and associated data.", userID)
			}
		}
	}()

	log.Printf("Starting deletion process for user ID: %d", userID)

	log.Printf("Deleting sessions for user %d...", userID)
	_, err = tx.Exec("DELETE FROM sessions WHERE user_id = ?;", userID)
	if err != nil { err = fmt.Errorf("delete sessions: %w", err); return err }

	log.Printf("Deleting likes given by user %d...", userID)
	_, err = tx.Exec("DELETE FROM likes WHERE user_id = ?;", userID)
	if err != nil { err = fmt.Errorf("delete user likes: %w", err); return err }

	log.Printf("Deleting likes on comments by user %d...", userID)
	_, err = tx.Exec(`DELETE FROM likes WHERE comment_id IN (SELECT id FROM comments WHERE user_id = ?);`, userID)
	if err != nil { err = fmt.Errorf("delete likes on user comments: %w", err); return err }

	log.Printf("Deleting likes on posts by user %d...", userID)
	_, err = tx.Exec(`DELETE FROM likes WHERE post_id IN (SELECT id FROM posts WHERE user_id = ?);`, userID)
	if err != nil { err = fmt.Errorf("delete likes on user posts: %w", err); return err }

	log.Printf("Deleting notifications for user %d...", userID)
	_, err = tx.Exec("DELETE FROM notifications WHERE user_id = ?;", userID)
	if err != nil { err = fmt.Errorf("delete user notifications: %w", err); return err }

	log.Printf("Deleting comments by user %d...", userID)
	_, err = tx.Exec("DELETE FROM comments WHERE user_id = ?;", userID)
	if err != nil { err = fmt.Errorf("delete user comments: %w", err); return err }

	log.Printf("Deleting comments on posts by user %d...", userID)
	_, err = tx.Exec(`DELETE FROM comments WHERE post_id IN (SELECT id FROM posts WHERE user_id = ?);`, userID)
	if err != nil { err = fmt.Errorf("delete comments on user posts: %w", err); return err }

	log.Printf("Deleting post categories for user %d posts...", userID)
	_, err = tx.Exec(`DELETE FROM post_categories WHERE post_id IN (SELECT id FROM posts WHERE user_id = ?);`, userID)
	if err != nil { err = fmt.Errorf("delete post categories for user posts: %w", err); return err }

	log.Printf("Deleting posts by user %d...", userID)
	_, err = tx.Exec("DELETE FROM posts WHERE user_id = ?;", userID)
	if err != nil { err = fmt.Errorf("delete user posts: %w", err); return err }

	log.Printf("Deleting user %d record...", userID)
	result, err := tx.Exec("DELETE FROM users WHERE id = ?;", userID)
	if err != nil { err = fmt.Errorf("delete user record: %w", err); return err }

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		err = fmt.Errorf("user with ID %d not found for deletion (maybe already deleted?)", userID)
		return err
	}

	log.Printf("Deletion transaction for user %d prepared successfully.", userID)
	return nil
}