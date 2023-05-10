package database

import (
	"database/sql"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

func Load(fp string) *sql.DB {
	db, err := sql.Open("sqlite3", fp)

	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS exchange (
			uuid TEXT NOT NULL,
			keyUserA TEXT UNIQUE NOT NULL,
			keyUserB TEXT UNIQUE,
		);
		CREATE INDEX IF NOT EXISTS indexExchangeUuid on exchange(uuid);
	`)

	if err != nil {
		panic(err)
	}

	return db
}

// Add a new exchange and return the unique id of the created exchange.
func MakeNewExchange(db *sql.DB, keyUserA string) (uid string, err error) {
	tx, err := db.Begin()

	if err != nil {
		return uid, err
	}

	defer tx.Rollback()

	uid = uuid.New().String()
	stmt := "INSERT INTO exchange keyUserA, uuid VALUES(?, ?)"
	_, err = tx.Exec(stmt, keyUserA, uid)

	if err != nil {
		return uid, err
	}

	return uid, tx.Commit()
}

// Insert User B public key into existing exchange, returns User A public key.
func UserBSwapKey(db *sql.DB, uid string, keyUserB string) (keyUserA string, err error) {
	tx, err := db.Begin()

	if err != nil {
		return keyUserA, err
	}

	defer tx.Rollback()

	// add key b
	stmt := "INSERT INTO exchange keyUserB values(?) WHERE uuid=?"
	_, err = tx.Exec(stmt, keyUserB, uid)

	if err != nil {
		return keyUserA, err
	}

	query := "SELECT keyUserA FROM exchange WHERE uuid=?"
	row := tx.QueryRow(query, uid)
	err = row.Scan(&keyUserA)

	if err != nil {
		return keyUserA, err
	}

	return keyUserA, tx.Commit()
}

// Retrieve User B public key and delete exchange.
func FinishExchange(db *sql.DB, uid string) (keyUserB string, err error) {
	tx, err := db.Begin()
	if err != nil {
		return keyUserB, err
	}
	defer tx.Rollback()

	query := "SELECT keyUserB FROM exchange WHERE uuid=?"
	row := tx.QueryRow(query, uid)
	err = row.Scan(&keyUserB)

	if err != nil {
		return keyUserB, err
	}

	if err = deleteExchange(db, tx, uid); err != nil {
		return keyUserB, err
	}

	return keyUserB, tx.Commit()
}

func deleteExchange(db *sql.DB, tx *sql.Tx, uid string) error {
	stmt := "DELETE FROM exchange WHERE uuid=?"
	_, err := tx.Exec(stmt, uid)

	return err
}
