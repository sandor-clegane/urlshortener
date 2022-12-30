package storages

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/sandor-clegane/urlshortener/internal/common"
	"github.com/sandor-clegane/urlshortener/internal/common/myerrors"
)

//modifying queries
const (
	initQuery = "CREATE TABLE IF NOT EXISTS urls " +
		"(id VARCHAR(255) PRIMARY KEY, " +
		"expand_url VARCHAR(255) UNIQUE, " +
		"user_id VARCHAR(255), " +
		"is_deleted boolean)"
	insertURLQuery = "INSERT INTO urls (id, expand_url, user_id, is_deleted) " +
		"VALUES ($1, $2, $3, $4)"
	insertURLQueryWithConstraint = "INSERT INTO urls (id, expand_url, user_id, is_deleted) " +
		"VALUES ($1, $2, $3, $4) " +
		"ON CONFLICT DO NOTHING"
	deleteURLQuery = "UPDATE urls " +
		"SET is_deleted=$1 " +
		"WHERE id=$2 AND user_id=$3"
)

//non-mod queries
const (
	getAllURLQuery = "SELECT id, expand_url " +
		"FROM urls " +
		"WHERE user_id=$1"
	getExpandURLQuery = "SELECT expand_url,is_deleted FROM urls " +
		"WHERE id=$1"
)

//utility constants
const (
	workersCount int = 10
)

type dbStorage struct {
	dbConnection  *sql.DB
	deletionBatch chan common.DeletableURL
	sync          *syncObj
}

type syncObj struct {
	done chan struct{}
	wg   sync.WaitGroup
	once sync.Once
}

func NewDBStorage(dbAddress string) (*dbStorage, error) {
	connection, err := connectAndInit(dbAddress)
	if err != nil {
		return nil, err
	}
	storage := &dbStorage{
		dbConnection:  connection,
		deletionBatch: make(chan common.DeletableURL),
		sync:          &syncObj{done: make(chan struct{})},
	}
	storage.runWorkerPool()
	return storage, nil
}

func connectAndInit(dbAddress string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbAddress)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(initQuery)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (d *dbStorage) Insert(ctx context.Context, urlID, expandURL, userID string) error {
	res, err := d.dbConnection.
		ExecContext(ctx, insertURLQueryWithConstraint, urlID, expandURL, userID, false)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("URL %s already exists", expandURL)
	}
	return nil
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
		if _, err = stmt.Exec(p.ShortURL, p.ExpandURL, userID, false); err != nil {
			if err = tx.Rollback(); err != nil {
				log.Printf("update drivers: unable to rollback: %v", err)
				return err
			}
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("update drivers: unable to commit: %v", err)
		return err
	}

	return nil
}

func (d *dbStorage) LookUp(ctx context.Context, urlID string) (string, error) {
	var u common.DeletableURL
	err := d.dbConnection.
		QueryRowContext(ctx, getExpandURLQuery, urlID).
		Scan(&u.ExpandURL, &u.IsDeleted)
	if err != nil {
		return "", err
	}
	if u.IsDeleted {
		return "", myerrors.NewDeleteViolation(u.ExpandURL, nil)
	}

	return u.ExpandURL, nil
}

func (d *dbStorage) GetPairsByID(ctx context.Context, userID string) ([]common.PairURL, error) {
	pairs := make([]common.PairURL, 0)

	rows, err := d.dbConnection.QueryContext(ctx, getAllURLQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var p common.PairURL
	for rows.Next() {
		err = rows.Scan(&p.ShortURL, &p.ExpandURL)
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, p)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return pairs, nil
}

func (d *dbStorage) DeleteMultipleURLs(_ context.Context, delSliceURL []common.DeletableURL) error {
	for _, delURL := range delSliceURL {
		err := d.push(delURL)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *dbStorage) push(dURL common.DeletableURL) error {
	select {
	case d.deletionBatch <- dURL:
		return nil
	case <-d.sync.done:
		return errors.New("storage closed")
	}
}

func (d *dbStorage) runWorkerPool() {
	for i := 0; i < workersCount; i++ {
		d.sync.wg.Add(1)
		go func() {
			defer d.sync.wg.Done()
			ctx := context.Background()
			for {
				select {
				case delURL, ok := <-d.deletionBatch:
					if !ok {
						return
					}
					_, err := d.dbConnection.ExecContext(ctx, deleteURLQuery,
						delURL.IsDeleted, delURL.ShortURL, delURL.UserID)
					if err != nil {
						log.Printf("%v", err)
						return
					}
				case <-d.sync.done:
					return
				}
			}
		}()
	}
}

func (d *dbStorage) Stop() {
	d.sync.once.Do(func() {
		close(d.sync.done)
		close(d.deletionBatch)
	})
	d.sync.wg.Wait()
}

func (d *dbStorage) Ping(ctx context.Context) error {
	return d.dbConnection.PingContext(ctx)
}
