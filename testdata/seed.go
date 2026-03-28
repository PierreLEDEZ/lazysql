//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbPath := filepath.Join(filepath.Dir(os.Args[0]), "demo.db")
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	os.Remove(dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    price REAL NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0,
    category TEXT NOT NULL
);

CREATE TABLE orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    total REAL NOT NULL,
    ordered_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO users (email, name, role, created_at) VALUES
    ('alice@example.com', 'Alice Martin', 'admin', '2024-01-15 09:30:00'),
    ('bob@example.com', 'Bob Dupont', 'user', '2024-02-03 14:22:00'),
    ('charlie@example.com', 'Charlie Moreau', 'user', '2024-03-12 11:05:00'),
    ('diana@example.com', 'Diana Leroy', 'editor', '2024-04-08 16:45:00'),
    ('etienne@example.com', 'Etienne Bernard', 'user', '2024-05-21 08:10:00'),
    ('fiona@example.com', 'Fiona Petit', 'admin', '2024-06-14 10:30:00'),
    ('gaston@example.com', 'Gaston Roux', 'user', '2024-07-02 13:15:00'),
    ('helene@example.com', 'Hélène Duval', 'editor', '2024-08-19 17:00:00'),
    ('ivan@example.com', 'Ivan Thomas', 'user', '2024-09-05 09:45:00'),
    ('julie@example.com', 'Julie Garcia', 'user', '2024-10-30 12:20:00');

INSERT INTO products (name, price, stock, category) VALUES
    ('Clavier mécanique', 89.99, 45, 'Périphériques'),
    ('Souris ergonomique', 49.50, 120, 'Périphériques'),
    ('Écran 27 pouces', 349.00, 18, 'Moniteurs'),
    ('Câble USB-C', 12.99, 500, 'Accessoires'),
    ('Hub USB 7 ports', 34.90, 75, 'Accessoires'),
    ('Casque Bluetooth', 79.00, 60, 'Audio'),
    ('Webcam HD', 59.99, 35, 'Périphériques'),
    ('Tapis de souris XL', 24.50, 200, 'Accessoires');

INSERT INTO orders (user_id, product_id, quantity, total, ordered_at) VALUES
    (1, 1, 1, 89.99, '2024-03-01 10:00:00'),
    (2, 3, 1, 349.00, '2024-03-05 14:30:00'),
    (1, 4, 3, 38.97, '2024-03-10 09:15:00'),
    (3, 6, 1, 79.00, '2024-04-02 11:45:00'),
    (5, 2, 2, 99.00, '2024-04-15 16:20:00'),
    (4, 7, 1, 59.99, '2024-05-08 13:00:00'),
    (2, 5, 1, 34.90, '2024-05-22 10:30:00'),
    (7, 1, 1, 89.99, '2024-06-01 08:45:00'),
    (6, 8, 2, 49.00, '2024-06-18 15:10:00'),
    (8, 3, 1, 349.00, '2024-07-04 12:00:00'),
    (9, 6, 1, 79.00, '2024-08-12 09:30:00'),
    (10, 4, 5, 64.95, '2024-09-25 14:15:00');
`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Base créée : %s\n", dbPath)

	var count int
	db.QueryRow("SELECT count(*) FROM users").Scan(&count)
	fmt.Printf("  users: %d\n", count)
	db.QueryRow("SELECT count(*) FROM products").Scan(&count)
	fmt.Printf("  products: %d\n", count)
	db.QueryRow("SELECT count(*) FROM orders").Scan(&count)
	fmt.Printf("  orders: %d\n", count)
}
