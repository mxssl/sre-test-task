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

func TestConvertTime(t *testing.T) {
	strDate := "2000-01-01"
	res, err := converStringToTime(strDate)
	if err != nil {
		t.Error(err)
	}
	expected := 1
	if res.Day() != 1 {
		t.Errorf("Wrong convertation: got %v want %v", res.Day(), expected)
	}
}

func TestCompareDate(t *testing.T) {
	recievedDate, _ := converStringToTime("2000-01-10")

	birthDate, _ := converStringToTime("2000-01-10")
	in5DaysDate, _ := converStringToTime("2018-01-05")
	anotherDate, _ := converStringToTime("2020-01-01")

	exp1 := 0
	if d1 := compareTime(recievedDate, birthDate); d1 != exp1 {
		t.Errorf("Wrong date comparison: got %v, want %v", d1, exp1)
	}

	exp2 := 5
	if d2 := compareTime(recievedDate, in5DaysDate); d2 != exp2 {
		t.Errorf("Wrong date comparison: got %v, want %v", d2, exp2)
	}

	exp3 := -1
	if d3 := compareTime(recievedDate, anotherDate); d3 != exp3 {
		t.Errorf("Wrong date comparison: got %v, want %v", d3, exp3)
	}
}
