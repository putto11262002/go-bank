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
	CreateTransaction(*Account, *Account, int64) (*Transaction, error)
	Deposite(*Account, int64) error
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
	if err := s.createAccountTable(); err != nil {
		return err
	}
	if err := s.createTransactionTable(); err != nil {
		return err
	}
	return nil
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


func (s *PostgresStore) createTransactionTable() error {
	query := `
	create table if not exists transaction (
		id serial primary key,
		fromAcc serial,
		toAcc serial,
		amount int,
		created_at timestamp
	)
	`
	_, err := s.db.Exec(query)
	return err
}


func (s *PostgresStore) Deposite(account *Account, amount int64) error {
	_, err  := s.db.Exec("update account set balance = balance + $1", amount)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) CreateTransaction(from *Account, to *Account, amount int64) (*Transaction, error) {
	tx, err := s.db.Begin()

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	row := tx.QueryRow("select balance from account where id = $1", from.ID)
	
	err = row.Scan(&from.Balance)
	if err != nil {
		return nil, err
	}

	if from.Balance < amount {
		return nil, fmt.Errorf("insufficiant balance")
	}

	_, err = tx.Exec("update account set balance = $1 where id = $2", from.Balance - amount, from.ID)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec("update account set balance = $1 where id = $2", to.Balance + amount, to.ID)
	if err != nil {
		return nil, err
	}

	transaction := NewTransaction(from.Number, to.Number, amount)

	row = tx.QueryRow(`
	insert into transaction 
	(fromAcc, toAcc, amount, created_at)
	values
	($1, $2, $3, $4)
	returning id, fromAcc, toAcc, amount, created_at
	`,
	transaction.From,
	transaction.To,
	transaction.Amount,
	transaction.CreatedAt)

	transaction, err = scanIntoTransaction(row)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return transaction, nil
}

func (s *PostgresStore) CreateAccount(acc *Account) (*Account, error) {
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

func scanIntoTransaction(scanable Scannable)(*Transaction, error){
	transaction := Transaction{}
	err := scanable.Scan(
		&transaction.ID,
		&transaction.From,
		&transaction.To,
		&transaction.Amount,
		&transaction.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &transaction, nil
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