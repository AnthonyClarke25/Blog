package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Post struct {
	ID      int
	Title   string
	Content string
}

func init() {
	// Ensure the "Database" directory exists
	if _, err := os.Stat("./Database"); os.IsNotExist(err) {
		err = os.Mkdir("./Database", os.ModePerm)
		if err != nil {
			log.Fatal("Failed to create Database directory:", err)
		}
	}

	// Resolve absolute path for database file
	dbPath, err := filepath.Abs("./Database/blog.db")
	if err != nil {
		log.Fatal("Error resolving database path:", err)
	}
	log.Println("Resolved Database Path:", dbPath)

	// Open the SQLite database
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Error opening database:", err)
	}

	// Test the database connection
	if _, err = db.Exec("SELECT 1"); err != nil {
		log.Fatalf("Database connection test failed: %v", err)
	}

	// Create the table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatal("Error creating table:", err)
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("HomeHandler called")

	// Fetch all blog posts from the database
	rows, err := db.Query("SELECT id, title FROM posts")
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title); err != nil {
			log.Printf("Error scanning posts: %v", err)
			http.Error(w, "Error scanning posts", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	// Resolve the template path
	tmplPath, err := filepath.Abs("templates/index.html")
	log.Println("Resolved Template Path for Homepage:", tmplPath)

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("Error loading template %s: %v", tmplPath, err)
		http.Error(w, "Error loading template homepage", http.StatusInternalServerError)
		return
	}

	// Render the template
	if err := tmpl.Execute(w, posts); err != nil {
		log.Printf("Error rendering homepage template: %v", err)
		http.Error(w, "Error rendering homepage", http.StatusInternalServerError)
	}
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("PostHandler called")

	// Extract the post ID from the URL
	id := r.URL.Path[len("/blog/post/"):]
	log.Printf("Requested Post ID: %s", id)

	// Query the database for the post by ID
	row := db.QueryRow("SELECT title, content FROM posts WHERE id = ?", id)

	var post Post
	err := row.Scan(&post.Title, &post.Content)
	if err != nil {
		log.Printf("Error fetching post with ID %s: %v", id, err)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Resolve the template path
	tmplPath, err := filepath.Abs("/templates/post.html")
	log.Println("Resolved Template Path for Post:", tmplPath)

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("Error loading template %s: %v", tmplPath, err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	// Render the template
	if err := tmpl.Execute(w, post); err != nil {
		log.Printf("Error rendering post template: %v", err)
		http.Error(w, "Error rendering post", http.StatusInternalServerError)
	}
}

func main() {
	// Log the current working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Current Working Directory:", dir)

	// Serve static files
	staticPath, err := filepath.Abs("blog/static")
	if err != nil {
		log.Fatalf("Error resolving static files path: %v", err)
	}
	log.Println("Serving static files from:", staticPath)
	http.Handle("/blog/static/", http.StripPrefix("/blog/static/", http.FileServer(http.Dir(staticPath))))

	// Define routes
	http.HandleFunc("/blog/", HomeHandler)
	http.HandleFunc("/blog/post/", PostHandler)

	// Start the server
	log.Println("Starting blog server at http://localhost:8080/blog")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Error starting the server:", err)
	}
}
