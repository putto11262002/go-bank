package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) (*Account, error)
	DeleteAccount(int) error
	GetAccountByID(int) (*Account, error)
	GetAccounts() ([]*Account, error)
	GetAccountByNumber(int64) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
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

func (s *PostgresStore) init() error {
	return s.createAccountTable()
}

func (s *PostgresStore) createAccountTable() error {
	query := `create table if not exists account (
		id serial primary key,
		first_name varchar(50),
		last_name varchar(50),
		number serial,
		balance int, 
		password varchar(200),
		created_at timestamp
		)`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateAccount(acc *Account) (*Account, error) {
	fmt.Printf("%+v", acc)
	query := `
	insert into account 
	(first_name, last_name, number, balance, password, created_at)
	values  
	($1, $2, $3, $4, $5, $6)

	returning id, first_name, last_name, number, balance, password, created_at
	`
	row := s.db.QueryRow(query,
		acc.FirstName,
		acc.LastName,
		acc.Number,
		acc.Balance,
		acc.Password,
		acc.CreatedAt,
	)

	account, err := scanIntoAccount(row)
	if err != nil {
		return nil, err
	}

	return account, nil
}


func (s *PostgresStore) DeleteAccount(id int) error {
	_, err := s.db.Query("delete from account where id = $1", id)
	return err
}

func (s *PostgresStore) GetAccountByID(id int) (*Account, error) {
	rows, err := s.db.Query("select * from account where id = $1", id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account %d not found", id)
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("select * from account")
	if err != nil {
		return nil, err
	}
	accounts := []*Account{}
	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil

}

func (s * PostgresStore) GetAccountByNumber(accountNumber int64)(*Account, error){
	rows, err := s.db.Query("select * from account where number = $1", accountNumber)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account %d not found", accountNumber)
}

type Scannable interface {
	Scan(dest ...interface{}) error
}


func scanIntoAccount(scanable Scannable) (*Account, error) {
	account := Account{}
	err := scanable.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.Balance,
		&account.Password,
		&account.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &account, nil
}