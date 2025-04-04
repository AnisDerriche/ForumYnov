package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// Fonction pour se connecter à SQLite
func connectDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "baseDonnee.db")
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Fonction pour créer la table utilisateur
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
		email TEXT,
		comment TEXT,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(post_id) REFERENCES posts(id),
		FOREIGN KEY(email) REFERENCES utilisateur(email)
	);

	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT,
		contenu TEXT,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(email) REFERENCES utilisateur(email)
	);
	`
	_, err := db.Exec(tab)
	return err
}

// Fonction pour hacher un mot de passe
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// Fonction pour insérer un utilisateur
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

func insertPost(db *sql.DB, email, contenu string) error {
	tab := "INSERT INTO posts (email, contenu, created_at) VALUES (?, ?, datetime('now'))"
	_, err := db.Exec(tab, email, contenu)
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

// Fonction pour gérer la connexion d'un utilisateur
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

	// Connexion à la base de données
	db, err := connectDB()
	if err != nil {
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Vérifier si l'utilisateur existe
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

	// Réponse JSON si la connexion réussit
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Connexion réussie"})
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

	log.Printf("Prénom : %s, Nom : %s, Email : %s, Mot de passe : %s", prenom, nom, email, password)

	// Connexion à la base de données
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

	// Réponse JSON si l'inscription réussit
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Inscription réussie"})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "connect.html")
}

func main() {
	db, err := connectDB()
	if err != nil {
		log.Fatalf("Impossible de se connecter à la base de données : %v", err)
	}
	defer db.Close()

	// Création de la table si elle n'existe pas
	err = createTable(db)
	if err != nil {
		log.Fatalf("Erreur lors de la création de la table : %v", err)
	}

	fmt.Println("Serveur démarré sur http://localhost:8080")
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/signup", signupHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
