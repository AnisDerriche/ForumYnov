package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

func connectDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "baseDonnee.db")
	if err != nil {
		return nil, err
	}
	return db, nil
}

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
	);
	`
	_, err := db.Exec(tab)
	return err
}

func updateDatabaseSchema(db *sql.DB) error {
	_, err := db.Exec("ALTER TABLE posts ADD COLUMN category TEXT DEFAULT 'general'")
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}
	return nil
}

var store = sessions.NewCookieStore([]byte("secret-key")) // Clé secrète pour les sessions

// Fonction pour récupérer la session
func getSession(r *http.Request) (*sessions.Session, error) {
	session, err := store.Get(r, "user-session")
	if err != nil {
		return nil, err
	}
	return session, nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func insertUser(db *sql.DB, prenom, nom, email, password string) error {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("erreur lors du hash du mot de passe : %v", err)
	}

	tab := "INSERT INTO utilisateur (email, prenom, nom, mdp, contenu, likes) VALUES (?, ?, ?, ?, ?, 0)"
	_, err = db.Exec(tab, email, prenom, nom, hashedPassword, "")
	if err != nil {
		return fmt.Errorf("erreur lors de l'insertion de l'utilisateur : %v", err)
	}
	return nil
}

func insertPost(db *sql.DB, title, email, contenu, category string) error {
	fmt.Printf("Inserting post: %s | %s | %s | %s\n", title, email, contenu, category)
	tab := "INSERT INTO posts (title, email, contenu, category, created_at) VALUES (?, ?, ?, ?, datetime('now'))"
	_, err := db.Exec(tab, title, email, contenu, category)
	if err != nil {
		return fmt.Errorf("could not insert post: %v", err)
	}
	return nil
}

func insertComment(db *sql.DB, postID int, email, comment string) error {
	tab := "INSERT INTO comments (post_id, email, comment, created_at) VALUES (?, ?, ?, datetime('now'))"
	_, err := db.Exec(tab, postID, email, comment)
	if err != nil {
		return fmt.Errorf("could not insert comment: %v", err)
	}
	return nil
}

func getPostsHandler(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	if category == "" {
		category = "all" // Par défaut, afficher tous les posts
	}

	db, err := connectDB()
	if err != nil {
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var rows *sql.Rows
	if category == "all" {
		// Récupérer tous les posts
		rows, err = db.Query("SELECT id, title, email, contenu, created_at, category FROM posts ORDER BY created_at DESC")
	} else {
		// Récupérer les posts filtrés par catégorie
		rows, err = db.Query("SELECT id, title, email, contenu, created_at, category FROM posts WHERE category = ? ORDER BY created_at DESC", category)
	}

	if err != nil {
		http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Email, &post.Contenu, &post.CreatedAt, &post.Category)
		if err != nil {
			http.Error(w, "Erreur lors du scan des posts", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func getPostsByCategoryHandler(w http.ResponseWriter, r *http.Request, category string) {
	db, err := connectDB()
	if err != nil {
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Requête pour récupérer les posts filtrés par catégorie
	rows, err := db.Query("SELECT id, title, email, contenu, created_at FROM posts WHERE category = ? ORDER BY created_at DESC", category)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Email, &post.Contenu, &post.CreatedAt)
		if err != nil {
			http.Error(w, "Erreur lors du scan des posts", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Requête reçue pour /login")

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Erreur de lecture du formulaire", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	log.Printf("Email : %s, Mot de passe : %s", email, password)

	db, err := connectDB()
	if err != nil {
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var hashedPassword string
	err = db.QueryRow("SELECT mdp FROM utilisateur WHERE email = ?", email).Scan(&hashedPassword)
	if err == sql.ErrNoRows {
		http.Error(w, "Utilisateur introuvable", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		http.Error(w, "Mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	// Création de la session pour l'utilisateur
	session, err := getSession(r)
	if err != nil {
		http.Error(w, "Erreur de session", http.StatusInternalServerError)
		return
	}
	session.Values["email"] = email // Stocke l'email dans la session
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Erreur lors de la sauvegarde de la session", http.StatusInternalServerError)
		return
	}

	// Redirige vers la page d'accueil après la connexion réussie
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Requête reçue pour /signup")

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Erreur de lecture du formulaire", http.StatusBadRequest)
		return
	}

	prenom := r.FormValue("prenom")
	nom := r.FormValue("nom")
	email := r.FormValue("new_email")
	password := r.FormValue("new_password")

	log.Printf("Prénom : %s, Nom : %s, Email : %s", prenom, nom, email)

	db, err := connectDB()
	if err != nil {
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = insertUser(db, prenom, nom, email, password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur lors de l'inscription : %v", err), http.StatusInternalServerError)
		return
	}

	// Création automatique de la session après inscription
	session, err := getSession(r)
	if err != nil {
		http.Error(w, "Erreur de session", http.StatusInternalServerError)
		return
	}
	session.Values["email"] = email // Stocke l'email dans la session
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Erreur lors de la sauvegarde de la session", http.StatusInternalServerError)
		return
	}

	// Redirige vers la page d'accueil après l'inscription et la connexion
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifie si l'utilisateur est connecté (session)
	session, err := getSession(r)
	if err != nil || session.Values["email"] == nil {
		// Redirige vers la page de connexion si l'utilisateur n'est pas connecté
		http.Redirect(w, r, "/connect.html", http.StatusSeeOther)
		return
	}

	// Si l'utilisateur est connecté, affiche la page des posts
	http.ServeFile(w, r, "index.html")
}

func createPostHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Requête reçue pour /create_post")

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier si l'utilisateur est connecté
	session, err := getSession(r)
	if err != nil || session.Values["email"] == nil {
		http.Error(w, "Vous devez être connecté pour créer un post", http.StatusUnauthorized)
		return
	}

	email := session.Values["email"].(string) // Récupérer l'email de la session
	title := r.FormValue("title")
	contenu := r.FormValue("contenu")
	category := r.FormValue("category") // Récupérer la catégorie choisie

	if title == "" || contenu == "" || category == "" {
		http.Error(w, "Tous les champs doivent être remplis", http.StatusBadRequest)
		return
	}

	log.Printf("title : %s, Email : %s, Contenu : %s, Category: %s", title, email, contenu, category)

	db, err := connectDB()
	if err != nil {
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Insérer le post avec la catégorie
	err = insertPost(db, title, email, contenu, category)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur lors de l'insertion du post : %v", err), http.StatusInternalServerError)
		return
	}

	// Rediriger vers la page d'accueil après l'insertion
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	db, err := connectDB()
	if err != nil {
		log.Fatalf("Impossible de se connecter à la base de données : %v", err)
	}
	defer db.Close()

	err = createTable(db)
	if err != nil {
		log.Fatalf("Erreur lors de la création de la table : %v", err)
	}

	err = updateDatabaseSchema(db)
	if err != nil {
		log.Printf("Erreur lors de la mise à jour du schéma : %v", err)
	}

	fmt.Println("Serveur démarré sur http://localhost:8080")

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
	http.Handle("/connect.html", http.FileServer(http.Dir(".")))
	http.Handle("/createPost.html", http.FileServer(http.Dir(".")))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
