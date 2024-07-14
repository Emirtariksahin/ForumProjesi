package models

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

type ModeratorRequest struct {
	ID       int
	Username string
	Status   string
}

func HandleAdminPanel(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./Back-end/database/forum.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT mr.id, u.username, mr.status FROM moderator_requests mr JOIN users u ON mr.user_id = u.id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var requests []ModeratorRequest
	for rows.Next() {
		var req ModeratorRequest
		if err := rows.Scan(&req.ID, &req.Username, &req.Status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		requests = append(requests, req)
	}

	tmpl, ok := tmplCache["panel"]
	if !ok {
		http.Error(w, "Could not load admin panel template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Requests []ModeratorRequest
	}{
		Requests: requests,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func HandleApproveRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		requestID := r.FormValue("requestId")

		db, err := sql.Open("sqlite3", "./Back-end/database/forum.db")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		_, err = db.Exec("UPDATE moderator_requests SET status = 'approved' WHERE id = ?", requestID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var userID int
		err = db.QueryRow("SELECT user_id FROM moderator_requests WHERE id = ?", requestID).Scan(&userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("UPDATE users SET is_moderator = 1 WHERE id = ?", userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/panel", http.StatusSeeOther)
	}
}

func HandleRejectRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		requestID := r.FormValue("requestId")

		db, err := sql.Open("sqlite3", "./Back-end/database/forum.db")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		_, err = db.Exec("UPDATE moderator_requests SET status = 'rejected' WHERE id = ?", requestID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var userID int
		err = db.QueryRow("SELECT user_id FROM moderator_requests WHERE id = ?", requestID).Scan(&userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("UPDATE users SET is_moderator = 0 WHERE id = ?", userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/panel", http.StatusSeeOther)
	}
}

func HandleRevokeModerator(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		requestIDStr := r.FormValue("requestId")
		requestID, err := strconv.Atoi(requestIDStr)
		if err != nil {
			http.Error(w, "Invalid request ID", http.StatusBadRequest)
			return
		}

		log.Printf("Revoke request for request ID: %d", requestID) // Debugging log

		db, err := sql.Open("sqlite3", "./Back-end/database/forum.db")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Get the user ID associated with the request ID
		var userID int
		err = tx.QueryRow("SELECT user_id FROM moderator_requests WHERE id = ?", requestID).Scan(&userID)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Request ID not found", http.StatusNotFound)
			return
		}

		// Remove moderator status
		_, err = tx.Exec("UPDATE users SET is_moderator = 0 WHERE id = ?", userID)
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Optionally remove admin status if needed
		_, err = tx.Exec("UPDATE users SET is_admin = 0 WHERE id = ?", userID)
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update the request status to "revoked"
		_, err = tx.Exec("UPDATE moderator_requests SET status = 'revoked' WHERE id = ?", requestID)
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/panel", http.StatusSeeOther)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

