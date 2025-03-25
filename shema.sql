CREATE TABLE utilisateur (
    email TEXT PRIMARY KEY,
    prenom TEXT NOT NULL,
    nom TEXT NOT NULL,
    mdp TEXT NOT NULL,
    contenu TEXT
);

CREATE TABLE like (
    email TEXT PRIMARY KEY,
    likes INT
);


CREATE TABLE cybersecurite (
    email TEXT NOT NULL,
    titre TEXT NOT NULL,
    contenu TEXT NOT NULL,
    heure DATETIME NOT NULL,
    FOREIGN KEY (email) REFERENCES utilisateur(email)
);

CREATE TABLE informatique(
    email TEXT NOT NULL,
    titre TEXT NOT NULL,
    contenu TEXT NOT NULL,
    heure DATETIME NOT NULL,
    FOREIGN KEY (email) REFERENCES utilisateur(email)
);

CREATE TABLE francais(
    email TEXT NOT NULL,
    titre TEXT NOT NULL,
    contenu TEXT NOT NULL,
    heure DATETIME NOT NULL,
    FOREIGN KEY (email) REFERENCES utilisateur(email)
);

CREATE TABLE anglais(
    email TEXT NOT NULL,
    titre TEXT NOT NULL,
    contenu TEXT NOT NULL,
    heure DATETIME NOT NULL,
    FOREIGN KEY (email) REFERENCES utilisateur(email)
);
