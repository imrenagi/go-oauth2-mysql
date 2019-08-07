package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"
)

type ClientStore struct {
	db        *sqlx.DB
	tableName string

	initTableDisabled bool
	maxLifetime       time.Duration
	maxOpenConns      int
	maxIdleConns      int
}

// ClientStoreItem data item
type ClientStoreItem struct {
	ID     string `db:"id"`
	Secret string `db:"secret"`
	Domain string `db:"domain"`
	Data   string `db:"data"`
}

// NewClientStore creates PostgreSQL store instance
func NewClientStore(db *sqlx.DB, options ...ClientStoreOption) (*ClientStore, error) {

	store := &ClientStore{
		db:           db,
		tableName:    "oauth2_clients",
		maxLifetime:  time.Hour * 2,
		maxOpenConns: 50,
		maxIdleConns: 25,
	}

	for _, o := range options {
		o(store)
	}

	var err error
	if !store.initTableDisabled {
		err = store.initTable()
	}

	if err != nil {
		return store, err
	}

	store.db.SetMaxOpenConns(store.maxOpenConns)
	store.db.SetMaxIdleConns(store.maxIdleConns)
	store.db.SetConnMaxLifetime(store.maxLifetime)

	return store, err
}

func (s *ClientStore) initTable() error {

	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id VARCHAR(255) NOT NULL PRIMARY KEY,
		secret VARCHAR(255) NOT NULL,
		domain VARCHAR(255) NOT NULL,
		data TEXT NOT NULL	
	  );
`, s.tableName)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	return nil
}

func (s *ClientStore) toClientInfo(data string) (oauth2.ClientInfo, error) {
	var cm models.Client
	err := jsoniter.Unmarshal([]byte(data), &cm)
	return &cm, err
}

// GetByID retrieves and returns client information by id
func (s *ClientStore) GetByID(id string) (oauth2.ClientInfo, error) {
	if id == "" {
		return nil, nil
	}

	var item ClientStoreItem
	err := s.db.QueryRowx(fmt.Sprintf("SELECT * FROM %s WHERE id = ?", s.tableName), id).StructScan(&item)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	}

	return s.toClientInfo(item.Data)
}

// Create creates and stores the new client information
func (s *ClientStore) Create(info oauth2.ClientInfo) error {
	data, err := jsoniter.Marshal(info)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(fmt.Sprintf("INSERT INTO %s (id, secret, domain, data) VALUES (?,?,?,?)", s.tableName),
		info.GetID(),
		info.GetSecret(),
		info.GetDomain(),
		string(data),
	)
	if err != nil {
		return err
	}
	return nil
}
