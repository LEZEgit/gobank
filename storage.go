package main

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccountByID(int) (*Account, error)
	GetAccounts() ([]*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostGresStore() (*PostgresStore, error) {
	connStr := "postgres://gobankuser:gobankpassword@localhost:5432/gobank?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.createTableAccount()
}

func (s *PostgresStore) createTableAccount() error {
	query := `create table if not exists account (
		id serial primary key, 
		first_name varchar(50) not null,
		last_name varchar(50) not null,
		number int, 
		balance int, 
		created_at timestamp
	);`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateAccount(acc *Account) error {
	query := `
	insert into account 
	(first_name, last_name, number, balance, created_at)
	values
	($1, $2, $3, $4, $5);
	`
	_, err := s.db.Exec(
		query,
		acc.FirstName,
		acc.LastName,
		acc.Number,
		acc.Balance,
		acc.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) UpdateAccount(*Account) error {
	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	return nil
}

func (s *PostgresStore) GetAccountByID(id int) (*Account, error) {
	query := `
	select * 
	from account
	where account.id = $1
	`
	accountRows, err := s.db.Query(query, id)

	if err != nil {
		return nil, err
	}
	defer accountRows.Close()
	accounts, err := scanIntoAccounts(accountRows)
	if err != nil {
		return nil, err
	}
	return accounts[0], nil
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("select * from account")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts, err := scanIntoAccounts(rows)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func scanIntoAccounts(rows *sql.Rows) ([]*Account, error) {
	accounts := []*Account{}
	for rows.Next() {
		account := new(Account)
		err := rows.Scan(
			&account.ID,
			&account.FirstName,
			&account.LastName,
			&account.Number,
			&account.Balance,
			&account.CreatedAt,
		)

		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}
