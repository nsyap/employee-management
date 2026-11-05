package main

import (
	"employee-management/internal/database"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Employee struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Department string `json:"department"`
}

var db *pgxpool.Pool

func main() {
	var err error

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Cannot load .env file:", err)
	}

	// Get database connection string
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	// Connect PostgreSQL
	db, err = database.Connect(databaseURL)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	if db == nil {
		log.Println("Database pool is nil!")
		return
	}
	defer db.Close()

	log.Println("Database connected successfully!")
	log.Println("Employee Management API started!")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/employees", employeesHandler)
	mux.HandleFunc("/api/employees/", employeeHandler)

	log.Println("API running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", cors(mux)); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func employeesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := db.Query(r.Context(), qGetEmployees)
		if err != nil {
			writeJSON(w, 500, ErrDatabase)
			return
		}
		log.Println("err:", err != nil)

		defer rows.Close()

		employees := []Employee{}
		for rows.Next() {
			var e Employee
			if err := rows.Scan(&e.ID, &e.Name, &e.Email, &e.Department); err != nil {
				writeJSON(w, 500, map[string]string{"error": "Read error"})
				return
			}
			employees = append(employees, e)
		}
		if err := rows.Err(); err != nil {
			writeJSON(w, 500, map[string]string{"error": "Read error"})
			return
		}
		writeJSON(w, 200, employees)

	case http.MethodPost:
		var e Employee
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&e); err != nil {
			writeJSON(w, 400, map[string]string{"error": "Invalid JSON"})
			return
		}

		e.Name = strings.TrimSpace(e.Name)
		e.Email = strings.TrimSpace(e.Email)
		e.Department = strings.TrimSpace(e.Department)

		if e.Name == "" || e.Email == "" || e.Department == "" ||
			!strings.Contains(e.Email, "@") {
			writeJSON(w, 400, map[string]string{
				"error": "Name, valid email and department are required",
			})
			return
		}

		err := db.QueryRow(r.Context(),
			qInsertEmployees,
			e.Name, e.Email, e.Department,
		).Scan(&e.ID, &e.Name, &e.Email, &e.Department)

		if err != nil {
			writeJSON(w, 409, map[string]string{
				"error": "Could not add employee; email may already exist",
			})
			return
		}
		writeJSON(w, 201, e)

	default:
		w.Header().Set("Allow", "GET, POST")
		writeJSON(w, 405, map[string]string{"error": "Method not allowed"})
	}
}

func employeeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", "DELETE")
		writeJSON(w, 405, map[string]string{"error": "Method not allowed"})
		return
	}

	idText := strings.TrimPrefix(r.URL.Path, "/api/employees/")
	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]string{"error": "Invalid employee ID"})
		return
	}

	tag, err := db.Exec(r.Context(),
		qDeleteEmployeesById, id)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "Delete failed"})
		return
	}
	if tag.RowsAffected() == 0 {
		writeJSON(w, 404, map[string]string{"error": "Employee not found"})
		return
	}

	writeJSON(w, 200, map[string]string{"message": "Employee deleted"})
}

/*
CORS - Cross-Origin Resource Sharing.
a browser security mechanism that controls
whether a frontend from one origin
(domain, port, or protocol) is allowed to access
resources from a different origin (backend server)
*/
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:5173" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
