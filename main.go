package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
)

var store = sessions.NewCookieStore([]byte("secret-key"))

func createTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS utilisateur (
		email TEXT PRIMARY KEY,
		prenom TEXT,
		nom TEXT,
		mdp TEXT,
		contenu TEXT,
		likes INTEGER DEFAULT 0
	)`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("could not create table: %v", err)
	}
	return nil
}

func insertUser(db *sql.DB, email, prenom, nom, mdp, text string, likes int) error {
	query := "INSERT INTO utilisateur (email, prenom, nom, mdp, contenu, likes) VALUES (?, ?, ?, ?, ?, 0)"
	_, err := db.Exec(query, email, prenom, nom, mdp, text)
	if err != nil {
		return fmt.Errorf("could not insert user: %v", err)
	}
	return nil
}

func addLike(db *sql.DB, email string) error {
	query := "UPDATE utilisateur SET likes = likes + 1 WHERE email = ?"
	_, err := db.Exec(query, email)
	if err != nil {
		return fmt.Errorf("could not add like: %v", err)
	}
	return nil
}

func removeLike(db *sql.DB, email string) error {
	query := "UPDATE utilisateur SET likes = likes - 1 WHERE email = ? AND likes > 0"
	_, err := db.Exec(query, email)
	if err != nil {
		return fmt.Errorf("could not remove like: %v", err)
	}
	return nil
}

func getUserLikes(db *sql.DB, email string) (int, error) {
	var likes int
	query := "SELECT likes FROM utilisateur WHERE email = ?"
	err := db.QueryRow(query, email).Scan(&likes)
	if err != nil {
		return 0, fmt.Errorf("could not retrieve likes: %v", err)
	}
	return likes, nil
}

func createPostsTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT,
		contenu TEXT,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(email) REFERENCES utilisateur(email)
	)`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("could not create posts table: %v", err)
	}
	return nil
}

func createCommentsTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER,
		email TEXT,
		comment TEXT,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(post_id) REFERENCES posts(id),
		FOREIGN KEY(email) REFERENCES utilisateur(email)
	)`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("could not create comments table: %v", err)
	}
	return nil
}

func insertPost(db *sql.DB, email, contenu string) error {
	query := "INSERT INTO posts (email, contenu, created_at) VALUES (?, ?, datetime('now'))"
	_, err := db.Exec(query, email, contenu)
	if err != nil {
		return fmt.Errorf("could not insert post: %v", err)
	}
	return nil
}

func insertComment(db *sql.DB, postID int, email, comment string) error {
	query := "INSERT INTO comments (post_id, email, comment, created_at) VALUES (?, ?, ?, datetime('now'))"
	_, err := db.Exec(query, postID, email, comment)
	if err != nil {
		return fmt.Errorf("could not insert comment: %v", err)
	}
	return nil
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Vérifie l'existence de l'utilisateur et le mot de passe
	var storedPassword string
	err := db.QueryRow("SELECT mdp FROM utilisateur WHERE email = ?", email).Scan(&storedPassword)
	if err != nil {
		http.Error(w, "Utilisateur non trouvé", http.StatusUnauthorized)
		return
	}

	if storedPassword != password {
		http.Error(w, "Mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	// Créer la session
	session, _ := store.Get(r, "session")
	session.Values["email"] = email
	session.Save(r, w)

	fmt.Fprintln(w, "Connexion réussie")
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Options.MaxAge = -1 // Supprime la session
	session.Save(r, w)

	fmt.Fprintln(w, "Déconnexion réussie")
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		_, ok := session.Values["email"]
		if !ok {
			http.Error(w, "Non autorisé", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Connexion à SQLite
	db, err := sql.Open("sqlite3", "baseDonnee.db")
	if err != nil {
		log.Fatalf("Impossible de se connecter à la base de données : %v", err)
	}
	defer db.Close()

	fmt.Println("Connexion réussie à la base de données SQLite!")

	err = createTable(db)
	if err != nil {
		log.Fatalf("could not create table: %v", err)
	}

	err = createPostsTable(db) // 🔹 Crée la table des posts
	if err != nil {
		log.Fatalf("could not create posts table: %v", err)
	}

	err = createCommentsTable(db) // 🔹 Crée la table des commentaires
	if err != nil {
		log.Fatalf("could not create comments table: %v", err)
	}

	// Exemple d'insertion d'un utilisateur
	err = insertUser(db, "john.doe@example.com", "John", "Doe", "password123", "Mon premier post", 0)
	if err != nil {
		log.Fatalf("could not insert user: %v", err)
	}

	fmt.Println("User inserted successfully")

	err = insertPost(db, "john.doe@example.com", "Ceci est mon premier post !")
	if err != nil {
		log.Fatalf("could not insert post: %v", err)
	}
	fmt.Println("Post ajouté avec succès")

	err = insertComment(db, 1, "john.doe@example.com", "Ceci est mon premier commentaire !")
	if err != nil {
		log.Fatalf("could not insert comment: %v", err)
	}
	fmt.Println("Commentaire ajouté avec succès")

	// Ajouter un like à l'utilisateur
	err = addLike(db, "john.doe@example.com")
	if err != nil {
		log.Fatalf("could not add like: %v", err)
	}

	fmt.Println("Like ajouté avec succès")

	// Enlever un like à l'utilisateur
	err = removeLike(db, "john.doe@example.com")
	if err != nil {
		log.Fatalf("could not remove like: %v", err)
	}

	fmt.Println("Like retiré avec succès")

	// Récupérer le nombre de likes
	likes, err := getUserLikes(db, "john.doe@example.com")
	if err != nil {
		log.Fatalf("could not get likes: %v", err)
	}

	fmt.Printf("Nombre de likes de John Doe : %d\n", likes)

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		loginHandler(w, r, db)
	})

	http.HandleFunc("/logout", logoutHandler)

	http.Handle("/moncompte", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Bienvenue sur votre compte privé !")
	})))

	fmt.Println("Serveur lancé sur : http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erreur serveur HTTP : %v", err)
	}

}
