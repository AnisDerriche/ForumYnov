package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type Post struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Email     string `json:"email"`
	Contenu   string `json:"contenu"`
	Category  string `json:"category"`
	CreatedAt string `json:"created_at"`
}

type Commentaire struct {
	ID        int    `json:"id"`
	PostID    int    `json:"post_id"`
	Email     string `json:"email"`
	Contenu   string `json:"contenu"`
	CreatedAt string `json:"created_at"`
}

func connectDB() (*sql.DB, error) {
	return sql.Open("sqlite3", "baseDonnee.db")
}

// Crée les tables nécessaires si elles n'existent pas
func createTable(db *sql.DB) error {
	tab := `
	CREATE TABLE IF NOT EXISTS utilisateur (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		prenom TEXT NOT NULL,
		nom TEXT NOT NULL,
		mdp TEXT NOT NULL,
		contenu TEXT,
		likes INTEGER DEFAULT 0
	);
	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER,
		title TEXT,
		email TEXT,
		comment TEXT,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(post_id) REFERENCES posts(id),
		FOREIGN KEY(email) REFERENCES utilisateur(email)
	);
	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT,
		title TEXT,
		contenu TEXT,
		category TEXT,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(email) REFERENCES utilisateur(email)
	);`
	_, err := db.Exec(tab)
	return err
}

// Met à jour le schéma de la base de données pour ajouter la colonne category
func updateDatabaseSchema(db *sql.DB) error {
	_, err := db.Exec("ALTER TABLE posts ADD COLUMN category TEXT DEFAULT 'general'")
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}
	return nil
}

var store = sessions.NewCookieStore([]byte("secret-key"))

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   120,
		HttpOnly: true,
	}
}

// Récupère la session utilisateur depuis la requête HTTP
func getSession(r *http.Request) (*sessions.Session, error) {
	return store.Get(r, "user-session")
}

// Hash le mot de passe fourni
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Insère un nouvel utilisateur dans la base
func insertUser(db *sql.DB, prenom, nom, email, password string) error {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}
	tab := "INSERT INTO utilisateur (email, prenom, nom, mdp, contenu, likes) VALUES (?, ?, ?, ?, ?, 0)"
	_, err = db.Exec(tab, email, prenom, nom, hashedPassword, "")
	return err
}

// Insère un nouveau post dans la base
func insertPost(db *sql.DB, title, email, contenu, category string) error {
	tab := "INSERT INTO posts (title, email, contenu, category, created_at) VALUES (?, ?, ?, ?, datetime('now'))"
	_, err := db.Exec(tab, title, email, contenu, category)
	return err
}

// Insère un nouveau commentaire
func insertComment(db *sql.DB, postID int, email, comment string) error {
	tab := "INSERT INTO comments (post_id, email, comment, created_at) VALUES (?, ?, ?, datetime('now'))"
	_, err := db.Exec(tab, postID, email, comment)
	return err
}

// Récupère tous les posts (optionnellement par catégorie)
func getPostsHandler(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	if category == "" {
		category = "all"
	}
	db, err := connectDB()
	if err != nil {
		http.Error(w, "Erreur de connexion", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var rows *sql.Rows
	if category == "all" {
		rows, err = db.Query("SELECT id, title, email, contenu, created_at, category FROM posts ORDER BY created_at DESC")
	} else {
		rows, err = db.Query("SELECT id, title, email, contenu, created_at, category FROM posts WHERE category = ? ORDER BY created_at DESC", category)
	}
	if err != nil {
		http.Error(w, "Erreur lors de la requête", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Email, &post.Contenu, &post.CreatedAt, &post.Category)
		if err != nil {
			http.Error(w, "Erreur lors du scan", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// Récupère les posts selon une catégorie donnée
func getPostsByCategoryHandler(w http.ResponseWriter, r *http.Request, category string) {
	db, err := connectDB()
	if err != nil {
		http.Error(w, "Erreur de connexion", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, title, email, contenu, created_at FROM posts WHERE category = ? ORDER BY created_at DESC", category)
	if err != nil {
		http.Error(w, "Erreur lors de la requête", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Email, &post.Contenu, &post.CreatedAt)
		if err != nil {
			http.Error(w, "Erreur lors du scan", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// Gère la connexion des utilisateurs et crée une session
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")

	db, err := connectDB()
	if err != nil {
		http.Error(w, "Erreur BDD", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var hashedPassword string
	err = db.QueryRow("SELECT mdp FROM utilisateur WHERE email = ?", email).Scan(&hashedPassword)
	if err == sql.ErrNoRows || bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) != nil {
		http.Error(w, "Identifiants invalides", http.StatusUnauthorized)
		return
	}

	session, _ := getSession(r)
	session.Values["email"] = email
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Gère l'inscription des utilisateurs et crée une session
func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()
	prenom := r.FormValue("prenom")
	nom := r.FormValue("nom")
	email := r.FormValue("new_email")
	password := r.FormValue("new_password")

	db, err := connectDB()
	if err != nil {
		http.Error(w, "Erreur BDD", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = insertUser(db, prenom, nom, email, password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session, _ := getSession(r)
	session.Values["email"] = email
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Affiche la page principale si l'utilisateur est connecté
func indexHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := getSession(r)
	if session.Values["email"] == nil {
		http.Redirect(w, r, "/connect.html", http.StatusSeeOther)
		return
	}
	http.ServeFile(w, r, "index.html")
}

// Gère la création de nouveaux posts
func createPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	session, _ := getSession(r)
	if session.Values["email"] == nil {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}
	email := session.Values["email"].(string)
	title := r.FormValue("title")
	contenu := r.FormValue("contenu")
	category := r.FormValue("category")
	if title == "" || contenu == "" || category == "" {
		http.Error(w, "Champs requis", http.StatusBadRequest)
		return
	}
	db, err := connectDB()
	if err != nil {
		http.Error(w, "Erreur BDD", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	insertPost(db, title, email, contenu, category)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Gère l'ajout d'un commentaire à un post
func addCommentHandler(w http.ResponseWriter, r *http.Request) {
	var c Commentaire
	json.NewDecoder(r.Body).Decode(&c)
	if c.Email == "" || c.Contenu == "" || c.PostID == 0 {
		http.Error(w, "Champs requis", http.StatusBadRequest)
		return
	}
	db, err := connectDB()
	if err != nil {
		http.Error(w, "Erreur BDD", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	db.Exec("INSERT INTO comments(post_id, email, comment, created_at) VALUES (?, ?, ?, datetime('now'))", c.PostID, c.Email, c.Contenu)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Commentaire ajouté avec succès"})
}

// Récupère les commentaires d'un post
func getCommentsHandler(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("post_id")
	if postID == "" {
		http.Error(w, "ID requis", http.StatusBadRequest)
		return
	}
	db, err := connectDB()
	if err != nil {
		http.Error(w, "Erreur BDD", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	rows, err := db.Query("SELECT email, comment, created_at FROM comments WHERE post_id = ? ORDER BY created_at DESC", postID)
	if err != nil {
		http.Error(w, "Erreur de requête", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var comments []Commentaire
	for rows.Next() {
		var c Commentaire
		rows.Scan(&c.Email, &c.Contenu, &c.CreatedAt)
		comments = append(comments, c)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func main() {
	db, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	createTable(db)
	updateDatabaseSchema(db)

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/create_post", createPostHandler)
	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/posts", getPostsHandler)
	http.HandleFunc("/posts/cyber", func(w http.ResponseWriter, r *http.Request) {
		getPostsByCategoryHandler(w, r, "cyber")
	})
	http.HandleFunc("/posts/info", func(w http.ResponseWriter, r *http.Request) {
		getPostsByCategoryHandler(w, r, "info")
	})
	http.HandleFunc("/posts/anglais", func(w http.ResponseWriter, r *http.Request) {
		getPostsByCategoryHandler(w, r, "anglais")
	})
	http.HandleFunc("/comments", addCommentHandler)
	http.HandleFunc("/comments/{id}", getCommentsHandler)
	http.Handle("/connect.html", http.FileServer(http.Dir(".")))
	http.Handle("/createPost.html", http.FileServer(http.Dir(".")))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
