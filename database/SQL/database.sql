CREATE TABLE IF NOT EXISTS users (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    username    TEXT    NOT NULL UNIQUE,
    email       TEXT    NOT NULL UNIQUE,
    password    TEXT    NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    photo       TEXT    DEFAULT 'profil.png',
    role        TEXT    DEFAULT 'user'
);

CREATE TABLE IF NOT EXISTS posts (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id           INTEGER NOT NULL,
    title             TEXT    NOT NULL,
    content           TEXT    NOT NULL,
    original_content  TEXT,
    image_path        TEXT,
    moderation_status TEXT    DEFAULT 'pending',
    created_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
    modified_at       DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id)
);


CREATE TABLE IF NOT EXISTS comments (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id    INTEGER NOT NULL,
    user_id    INTEGER NOT NULL,
    content    TEXT    NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS likes (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL,
    post_id    INTEGER,
    comment_id INTEGER,
    value      INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id)  REFERENCES users(id),
    FOREIGN KEY (post_id)  REFERENCES posts(id),
    FOREIGN KEY (comment_id) REFERENCES comments(id)
);

CREATE TABLE IF NOT EXISTS categories (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS post_categories (
    post_id     INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    PRIMARY KEY (post_id, category_id),
    FOREIGN KEY (post_id)     REFERENCES posts(id),
    FOREIGN KEY (category_id) REFERENCES categories(id)
);

CREATE TABLE IF NOT EXISTS notifications (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL,
    message    TEXT    NOT NULL,
    post_id    INTEGER,
    comment_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id)    REFERENCES users(id),
    FOREIGN KEY (post_id)    REFERENCES posts(id),
    FOREIGN KEY (comment_id) REFERENCES comments(id)
);

CREATE TABLE IF NOT EXISTS sessions (
    session_id TEXT PRIMARY KEY,
    user_id    INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

INSERT OR IGNORE INTO categories (name) VALUES
    ('Action'), ('Aventure'), ('Drame'), ('Comédie'), ('Science-fiction'),
    ('Horreur'), ('Thriller'), ('Romance'), ('Animation'), ('Fantastique'),
    ('Super-héros'),

    ('Films'), ('Séries'), ('Documentaires'), ('Dessins animés'),
    ('Courts-métrages'),

    ('Marvel'), ('DC'), ('Star Wars'), ('Harry Potter'),
    ('Game of Thrones'), ('Stranger Things'),
    ('The Walking Dead'), ('The Witcher'),

    ('Théories & Spoilers'), ('Critiques & avis'),
    ('Recommandations'), ('Questions / débats'),
    ('Easter eggs & références'), ('Scènes cultes'),

    ('Nouveautés'), ('Classiques'),
    ('À venir (teasers/bandes-annonces)'),

    ('Cinéma américain'), ('Cinéma européen'),
    ('Cinéma asiatique'), ('Autres (indépendant, etc.)');
