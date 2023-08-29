package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

const ctxAccountKey CtxKey = "account"

type ApiError struct {
	Error string `json:"error"`
}

type apiFuc func(http.ResponseWriter, *http.Request) error

func makeHTTPHandleFunc(f apiFuc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}

	}
}

type APIServer struct {
	listenAddr string
	store      Storage
}

// APIServer constructor
func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

// register routes and start server
func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin))
	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))
	router.HandleFunc("/account/me", s.withJWTAuth(makeHTTPHandleFunc(s.handleMyAccount)))
	router.HandleFunc("/transfer", s.withJWTAuth(makeHTTPHandleFunc(s.handleTransfer)))
	router.HandleFunc("/deposit", s.withJWTAuth(makeHTTPHandleFunc(s.handleDeplosit)))

	log.Printf("Server is running on port: %v\n", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

// handle incoming requests to /account
func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		return s.handleGetAccount(w, r)
	}

	if r.Method == http.MethodPost {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("method not allowed: %s", r.Method)
}

// handle incoming requests to /account/{id}
func (s *APIServer) handleMyAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		return s.handleGetMyAccount(w, r)
	}

	if r.Method == http.MethodDelete {
		return s.handleDeleteMyAccount(w, r)
	}

	return fmt.Errorf("method not allowed: %s", r.Method)
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, accounts)
}

// func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {

// 	id, err := getIDFromPathParam(r)
// 	if err != nil {
// 		return err
// 	}

// 	account, err := s.store.GetAccountByID(id)
// 	if err != nil {
// 		return err
// 	}

// 	return WriteJSON(w, http.StatusOK, account)
// }

func (s *APIServer) handleDeplosit(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("method not allowed: %s", r.Method)
	}
	depReq := &DepositeRequest{}
	err := json.NewDecoder(r.Body).Decode(depReq)
	if err != nil {
		return err
	}

	accountCtx, ok := r.Context().Value(ctxAccountKey).(*Account)
	
	if !ok {
		accountCtx = &Account{}
	}

	err = s.store.Deposite(accountCtx, depReq.Amount)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusNoContent, nil)
}

func (s *APIServer) handleGetMyAccount(w http.ResponseWriter, r *http.Request) error {

	accountCtx, ok := r.Context().Value(ctxAccountKey).(*Account)
	
	if !ok {
		accountCtx = &Account{}
	}
	

	account, err := s.store.GetAccountByID(accountCtx.ID)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountReq := &CreateAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}
	defer r.Body.Close()

	account, err := NewAccount(createAccountReq.FirstName, createAccountReq.LastName, createAccountReq.Password)
	if err != nil {
		return err
	}

	account, err = s.store.CreateAccount(account)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, account)
}

func (s *APIServer) handleDeleteMyAccount(w http.ResponseWriter, r *http.Request) error {
	accountCtx, ok := r.Context().Value(ctxAccountKey).(*Account)
	
	if !ok {
		accountCtx = &Account{}
	}

	if err := s.store.DeleteAccount(accountCtx.ID); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusNoContent, nil)
}


func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("method not allowed: %s", r.Method)
	}
	loginReq := &LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(loginReq); err != nil{
		return err
	}
	defer r.Body.Close()

	acc, err := s.store.GetAccountByNumber(loginReq.AccountNumber)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(loginReq.Password))
	if err != nil {
		return WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Incorrect credentials"})
	}

	token, err := createJWTToken(acc)
	if err != nil {
		return err
	}


	return WriteJSON(w, http.StatusOK, map[string]interface{}{"token": token, "account": acc})
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("method not allowed: %s", r.Method)
	}
	transferReq := &TransferRequest{}
	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return err
	}
	defer r.Body.Close()

	accountCtx, ok := r.Context().Value(ctxAccountKey).(*Account)
	
	if !ok {
		accountCtx = &Account{}
	}

	to, err := s.store.GetAccountByNumber(transferReq.ToAccount)
	if err != nil {
		return err
	}

	transaction, err := s.store.CreateTransaction(accountCtx, to, int64(transferReq.Amount))
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, transaction)
}

func (s *APIServer) withJWTAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("x-jwt-token")

		token, err := validateJWT(tokenStr)
		if err != nil && !token.Valid {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Invalid token"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		id, ok := (claims["id"].(float64))
		if !ok {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Invalid token"})
			return
		}

		

		account, err := s.store.GetAccountByID(int(id))
		if err != nil {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Invalid token"})
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), ctxAccountKey, account))

		handlerFunc(w, r)
	}
}
