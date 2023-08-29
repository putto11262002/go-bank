package main

import (
	// "math/rand"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)


type DepositeRequest struct {
	Amount int64 `json:"amount"`
}


type LoginRequest struct {
	AccountNumber int64 `json:"accountNumber"`
	Password string `json:"password"`
}

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	Password string `json:"password"`
}

type TransferRequest struct {
	ToAccount int64 `json:"toAccount"`
	Amount int64 `json:"amount"`
}

type Transaction struct {
	ID int `json:"id"`
	From int64 `json:"from"`
	To int64 `json:"to"`
	Amount int64 `json:"amount"`
	CreatedAt time.Time `json:"createdAt"`
}
func NewTransaction(from, to, amount int64) (*Transaction){
	return &Transaction{
		From: from,
		To: to,
		Amount: amount,
		CreatedAt: time.Now().UTC(),
	}
}

type Account struct {
	ID        int `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Number    int64 `json:"number"`
	Balance   int64 `json:"balance"`
	Password string `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}



func NewAccount(firstName, lastName, password string) (*Account, error) {
	encpw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Account{
		// ID: rand.Intn(1000),
		FirstName: firstName,
		LastName: lastName,
		Number: int64(rand.Intn(1000)),
		Password: string(encpw),
		Balance: 0,
		CreatedAt: time.Now().UTC(),
	}, nil
}

type CtxKey string


