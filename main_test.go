// main_test.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/kelseyhightower/envconfig"
)

type DOB struct {
	DateOfBirth string `json:"dateOfBirth"`
}

func TestDBConnection(t *testing.T) {

	var s Settings
	err := envconfig.Process("", &s)
	if err != nil {
		t.Errorf("Cannot process env variables. Error: %v\n", err)
	}

	// Setup connection to postgresql db
	connString := fmt.Sprintf("host=%v port=%v user=%v dbname=%v password=%v sslmode=disable", s.Host, s.Port, s.User, s.Name, s.Password)

	var dbErr error
	db, dbErr = gorm.Open("postgres", connString)
	if dbErr != nil {
		t.Errorf("Cannot connect to the db. Please provide access to the database to continue testing. Error: %v", dbErr)
	}
}

func TestDBMigrations(t *testing.T) {
	db.AutoMigrate(&User{})
	hasTable := db.HasTable(&User{})
	if hasTable != true {
		t.Errorf("DB has not table users")
	}
}

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `{"alive":true}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

}

func TestHelloPutHandler(t *testing.T) {

	var user User

	msg := &DOB{
		DateOfBirth: "2000-11-10",
	}
	jsonMsg, _ := json.Marshal(msg)
	fmt.Println(string(jsonMsg))

	req, err := http.NewRequest("PUT", "/hello/testuser", bytes.NewBuffer(jsonMsg))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/hello/{name}", user.helloPutHandler)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}
}

func TestHelloGetHandler(t *testing.T) {

	var user User

	req, err := http.NewRequest("GET", "/hello/testuser", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(user.helloGetHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `{"message":"Hello, testuser!"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
