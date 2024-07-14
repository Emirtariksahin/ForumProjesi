package models

import (
	"database/sql"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

type Post struct {
	ID            int
	UserID        int
	Title         string
	Content       string
	Image         string
	CategoryID    int
	CreatedAt     time.Time
	TotalLikes    int
	TotalDislikes int
}

type ModeratorPanelData struct {
	Posts []Post
}

func HandleModeratorPanel(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./Back-end/database/forum.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, user_id, title, content, image, category_id, created_at, total_likes, total_dislikes FROM posts")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.Image, &post.CategoryID, &post.CreatedAt, &post.TotalLikes, &post.TotalDislikes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	tmpl, ok := tmplCache["moderator_panel"]
	if !ok {
		http.Error(w, "Could not load moderator panel template", http.StatusInternalServerError)
		return
	}

	data := ModeratorPanelData{
		Posts: posts,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func HandleModeratorDeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		postID := r.FormValue("postId")

		db, err := sql.Open("sqlite3", "./Back-end/database/forum.db")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		_, err = db.Exec("DELETE FROM posts WHERE id = ?", postID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/moderatorPanel", http.StatusSeeOther)
	}
}

func HandleReportPost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		postID := r.FormValue("postId")
		reason := r.FormValue("reason")
		moderatorID := getCurrentModeratorID(r) // Assume this function retrieves the logged-in moderator ID

		db, err := sql.Open("sqlite3", "./Back-end/database/forum.db")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		_, err = db.Exec("INSERT INTO reports (post_id, moderator_id, reason) VALUES (?, ?, ?)", postID, moderatorID, reason)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/moderatorPanel", http.StatusSeeOther)
	}
}

func getCurrentModeratorID(r *http.Request) int {
	// Placeholder function. Replace with actual implementation to get the logged-in moderator's ID.
	return 1
}
