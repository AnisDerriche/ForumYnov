package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func insertUser(db *sql.DB, nom, prenom, email, mdp, text string) error {
	query := "INSERT INTO utilisateur (email, prenom, nom, mdp,contenu) VALUES (?, ?, ?, ?,?)"
	_, err := db.Exec(query, nom, prenom, email, mdp, text)
	if err != nil {
		return fmt.Errorf("could not insert user: %v", err)
	}
	return nil
}

func main() {
	// Connexion à SQLite
	db, err := sql.Open("sqlite3", "baseDonnee.db")
	if err != nil {
		log.Fatalf("Impossible de se connecter à la base de données : %v", err)
	}
	defer db.Close()

	fmt.Println("Connexion réussie à la base de données SQLite!")

	// Exemple d'insertion d'un utilisateur
	err = insertUser(db, "john.doe@example.com", "John", "Doe", "password123", "aezrtyuio")
	if err != nil {
		log.Fatalf("could not insert user: %v", err)
	}

	fmt.Println("User inserted successfully")
}
