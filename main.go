package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

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

	// Exemple d'insertion d'un utilisateur
	err = insertUser(db, "john.doe@example.com", "John", "Doe", "password123", "Mon premier post", 0)
	if err != nil {
		log.Fatalf("could not insert user: %v", err)
	}

	fmt.Println("User inserted successfully")

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
}
