package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

var DB *sql.DB

func initDB() {
	connStr := "postgres://todo_app_db_cs8i_user:sxaUqhe89HO3FOKqkAPtB7Yh4WLgSJMT@dpg-d2cu5abuibrs738sjdvg-a.oregon-postgres.render.com:5432/todo_app_db_cs8i?sslmode=require"

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	err = DB.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to DB!")
}

func main() {
	initDB()
	defer DB.Close()

	http.HandleFunc("/todo", withCORS(todoHandler))
	log.Println(" Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow frontend to call backend
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func todoHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getAllTodosHandler(w, r)
	case http.MethodPost:
		addTodoHandler(w, r)
	case http.MethodPut:
		idStr := r.URL.Query().Get("id")
		id, _ := strconv.Atoi(idStr)
		updateTodoHandler(w, r, id)
	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		id, _ := strconv.Atoi(idStr)
		deleteTodoHandler(w, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getAllTodosHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query("SELECT id, title, completed FROM todos")
	if err != nil {
		http.Error(w, "Error fetching todos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	todos := []Todo{}

	for rows.Next() {
		var t Todo
		err := rows.Scan(&t.ID, &t.Title, &t.Completed)
		if err != nil {
			http.Error(w, "Error scanning row", http.StatusInternalServerError)
			return
		}
		todos = append(todos, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func addTodoHandler(w http.ResponseWriter, r *http.Request) {
	var t Todo
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	err = DB.QueryRow("INSERT INTO todos (title, completed) VALUES ($1, $2) RETURNING id", t.Title, t.Completed).Scan(&t.ID)
	if err != nil {
		http.Error(w, "Error adding todo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func updateTodoHandler(w http.ResponseWriter, r *http.Request, id int) {
	var t Todo
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	_, err = DB.Exec("UPDATE todos SET title=$1, completed=$2 WHERE id=$3", t.Title, t.Completed, id)
	if err != nil {
		http.Error(w, "Error updating todo", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteTodoHandler(w http.ResponseWriter, id int) {
	_, err := DB.Exec("DELETE FROM todos WHERE id=$1", id)
	if err != nil {
		http.Error(w, "Error deleting todo", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
