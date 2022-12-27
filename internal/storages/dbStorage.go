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
		"user_id VARCHAR(255)), " +
		"is_deleted BOOLEAN)"
	insertURLQuery = "INSERT INTO urls (id, expand_url, user_id, is_deleted) " +
		"VALUES ($1, $2, $3, $4)"
	insertURLQueryWithConstraint = "INSERT INTO urls (id, expand_url, user_id, is_deleted) " +
		"VALUES ($1, $2, $3, $4) " +
		"ON CONFLICT DO NOTHING"
	deleteURLQuery = "UPDATE urls " +
		"SET is_deleted = $1 " +
		"WHERE id = $2 AND user_id = $3"
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
	WorkersCount int = 10
)

type dbStorage struct {
	dbConnection *sql.DB
	deletedChan  chan common.DeletableURL
	done         chan struct{}
	wg           sync.WaitGroup
	once         sync.Once
}

func NewDBStorage(dbAddress string) (*dbStorage, error) {
	connection, err := connect(dbAddress)
	if err != nil {
		return nil, err
	}
	storage := &dbStorage{
		dbConnection: connection,
		deletedChan:  make(chan common.DeletableURL),
		done:         make(chan struct{}),
	}
	storage.runDeletionWorkerPool()
	return storage, nil
}

func connect(dbAddress string) (*sql.DB, error) {
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

func (d *dbStorage) StopWorkerPool() {
	d.once.Do(func() {
		close(d.done)
		close(d.deletedChan)
	})
	d.wg.Wait()
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

func (d *dbStorage) RemoveSomeURL(_ context.Context, delSliceURL []common.DeletableURL) error {
	for _, ud := range delSliceURL {
		err := d.Push(ud)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *dbStorage) Push(url common.DeletableURL) error {
	select {
	case <-d.done:
		return errors.New("workers stopped")
	case d.deletedChan <- url:
		return nil
	}
}

func (d *dbStorage) runDeletionWorkerPool() {
	for i := 0; i < WorkersCount; i++ {
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			ctx := context.Background()
			for {
				select {
				case <-d.done:
					log.Println("Exiting")
					return
				case ud, ok := <-d.deletedChan:
					if !ok {
						return
					}
					_, err := d.dbConnection.
						ExecContext(ctx, deleteURLQuery, ud.IsDeleted, ud.ShortURL, ud.UserID)
					if err != nil {
						log.Printf("Delete error %v", err)
						return
					}
				}
			}
		}()
	}
}
