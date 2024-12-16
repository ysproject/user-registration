package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

type FormData struct {
	Username     string `json:"username"`
	JenisBank    string `json:"jenisBank"`
	NoRekening   string `json:"noRekening"`
	NamaRekening string `json:"namaRekening"`
	Server       string `json:"server"`
}

var db *sql.DB

func initDB() {
	var err error

	err = godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using system environment variables")
	}

	// Supabase PostgreSQL connection string
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping the database: %v", err)
	}

	fmt.Println("Connected to the database successfully!")
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data FormData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Update the query to use ? placeholders for MySQL
	query := `
	INSERT INTO "User" ("Username", "JenisBank", "NoRekening", "NamaRekening", "Server") 
	VALUES ($1, $2, $3, $4, $5)`
	_, err = db.Exec(query, data.Username, data.JenisBank, data.NoRekening, data.NamaRekening, data.Server)
	if err != nil {
		log.Printf("Failed to insert data: %v", err)
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Data saved successfully"))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func main() {
	initDB()
	defer db.Close()

	// Enable CORS for all routes
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Allow all origins (you can restrict this)
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Content-Type"},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/submit", submitHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", c.Handler(mux))) // Wrap the handler with CORS
}
