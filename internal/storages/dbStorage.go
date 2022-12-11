package storages

import (
	"context"
	"database/sql"
	"log"

	"github.com/sandor-clegane/urlshortener/internal/common"
)

const (
	initQuery = "CREATE TABLE IF NOT EXISTS urls " +
		"(id varchar(255) PRIMARY KEY, " +
		"expand_url varchar(255), " +
		"user_id varchar(255))"
	getAllURLQuery = "SELECT id, expand_url " +
		"FROM urls " +
		"WHERE user_id=$1"
	getExpandURLQuery = "SELECT expand_url FROM urls " +
		"WHERE id=$1"
	insertURLQuery = "INSERT INTO urls (id, expand_url, user_id) " +
		"VALUES ($1, $2, $3)"
)

type dbStorage struct {
	dbConnection *sql.DB
}

func NewDBStorage(dbAddress string) Storage {
	return &dbStorage{
		dbConnection: connect(dbAddress),
	}
}

func connect(dbAddress string) *sql.DB {
	db, err := sql.Open("postgres", dbAddress)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(initQuery)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func (d *dbStorage) Insert(ctx context.Context, urlID, expandURL, userID string) {
	_, err := d.dbConnection.
		ExecContext(ctx, insertURLQuery, urlID, expandURL, userID)
	if err != nil {
		log.Fatal(err)
	}
}

func (d *dbStorage) InsertSome(ctx context.Context, expandURLwIDslice []common.PairURL, userID string) error {
	tx, err := d.dbConnection.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, insertURLQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range expandURLwIDslice {
		if _, err = stmt.Exec(p.ShortURL, p.ExpandURL, userID); err != nil {
			if err = tx.Rollback(); err != nil {
				log.Fatalf("update drivers: unable to rollback: %v", err)
			}
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("update drivers: unable to commit: %v", err)
		return err
	}

	return nil
}

func (d *dbStorage) lookUp(ctx context.Context, urlID string) (string, error) {
	var expandURL string
	err := d.dbConnection.
		QueryRowContext(ctx, getExpandURLQuery, urlID).
		Scan(&expandURL)
	if err != nil {
		return "", err
	}
	return expandURL, nil
}

func (d *dbStorage) LookUp(ctx context.Context, urlID string) (string, bool) {
	str, err := d.lookUp(ctx, urlID)
	if err != nil {
		return "", false
	}
	return str, true
}

func (d *dbStorage) GetPairsByID(ctx context.Context, userID string) ([]common.PairURL, bool) {
	pairs := make([]common.PairURL, 0)

	rows, err := d.dbConnection.QueryContext(ctx, getAllURLQuery, userID)
	if err != nil {
		return nil, false
	}
	defer rows.Close()

	var p common.PairURL
	for rows.Next() {
		err = rows.Scan(&p.ShortURL, &p.ExpandURL)
		if err != nil {
			return nil, false
		}
		pairs = append(pairs, p)
	}

	err = rows.Err()
	if err != nil {
		return nil, false
	}
	return pairs, true
}
