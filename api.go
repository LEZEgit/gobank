package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/login", makeHTTPHandlerFunc(s.handleLogin))
	router.HandleFunc("/account", makeHTTPHandlerFunc(s.handleAccount))
	router.HandleFunc("/account/transfer", makeHTTPHandlerFunc(s.handleTransfer))
	router.HandleFunc("/account/{id}", s.authMiddleware(s.handleAccountByID))

	log.Println("JSON API Server running on PORT", s.listenAddr)

	if err := http.ListenAndServe(s.listenAddr, router); err != nil {
		log.Println("Error starting the server: ", err.Error())
	}
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("Method not allowed: %s", r.Method)
}

func (s *APIServer) handleAccountByID(w http.ResponseWriter, r *http.Request, client_account *Account) error {
	if r.Method == "GET" {
		return s.handleGetAccountByID(w, r, client_account)
	}
	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}

	return fmt.Errorf("Method not allowed: %s", r.Method)
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request, client_account *Account) error {
	id, err := getID(r)
	if err != nil {
		return err
	}
	if id != client_account.ID {
		return WriteJSON(w, http.StatusForbidden, ApiError{Error: "Can't access other's account"})
	}

	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	loginRequest := new(LoginRequest)
	if err := json.NewDecoder(r.Body).Decode(loginRequest); err != nil {
		return err
	}
	defer r.Body.Close()

	id := loginRequest.ID
	log.Printf("User with id: %v tried logging in", id)

	_, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return fmt.Errorf("JWT_SECRET is not configured")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": strconv.Itoa(id),
		"exp": time.Now().Add(time.Minute).Unix(),
	})
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]string{"token": signedToken})
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountReq := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}
	defer r.Body.Close()

	account := NewAccount(
		createAccountReq.FirstName,
		createAccountReq.LastName,
	)
	fmt.Printf("%+v", account)

	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}
	err = s.store.DeleteAccount(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, map[string]int{"Deleted Account with id:": id})
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferReq := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return err
	}
	defer r.Body.Close()

	return nil
}

func WriteJSON(w http.ResponseWriter, statusCode int, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(payload)
}

type ApiError struct {
	Error string `json:"error"`
}

type apiFunc func(http.ResponseWriter, *http.Request) error

// this is going to decorate our apiFunc to an HTTP HandlerFunc
func makeHTTPHandlerFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func getID(r *http.Request) (int, error) {
	idString := mux.Vars(r)["id"]
	if idString == "" {
		return -1, fmt.Errorf("[LOG] No ID passed in the URL")
	}
	id, err := strconv.Atoi(idString)
	if err != nil {
		return -1, fmt.Errorf("[LOG] Couldn't parse URL Parameters in GET /account/id")
	}
	return id, nil
}
