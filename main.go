package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/kelseyhightower/envconfig"

	"github.com/gorilla/mux"
)

// PUT /hello/John { "dateOfBrith": "2000-01-01" }
// GET /hello/John { "message": "Hello, John! Your birthday in 5 days" }
// GET /hello/John { "message": "Hello, John! Happy Birthday!" }
// GET /hello/John { "message": "Hello, John!" }
// GET /health { "alive": "true" }

// User describes user's model
type User struct {
	ID          int `gorm:"AUTO_INCREMENT"`
	Name        string
	DateOfBirth string `json:"dateOfBirth"`
}

// Message describes message's model
type Message struct {
	Message string `json:"message"`
}

// Health describes health check message model
type Health struct {
	Alive bool `json:"alive"`
}

// Settings for connection to a db
type Settings struct {
	Host     string `envconfig:"DB_HOST" required:"true"`
	Port     string `envconfig:"DB_PORT" required:"true"`
	User     string `envconfig:"DB_USER" required:"true"`
	Name     string `envconfig:"DB_NAME" required:"true"`
	Password string `envconfig:"DB_PASSWORD" required:"true"`
}

// Db connection pointer
var db *gorm.DB

func main() {
	log.Println("App is starting...")

	// Process env variables. If there are no required variables throw a error
	var s Settings
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatalf("Cannot process env variables. Error: %v\n", err)
	}

	// Setup connection to postgresql db
	connString := fmt.Sprintf("host=%v port=%v user=%v dbname=%v password=%v sslmode=disable", s.Host, s.Port, s.User, s.Name, s.Password)

	var dbErr error
	db, dbErr = gorm.Open("postgres", connString)
	if dbErr != nil {
		log.Fatalf("Cannot connect to DB, error: %v", dbErr)
	}
	defer db.Close()
	log.Println("Connection to db established")

	// Create table for users
	log.Println("Apply migrations...")

	var user User

	db.AutoMigrate(&User{})

	// Setup mux router
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/hello/{name}", user.helloPutHandler).Methods("PUT")
	r.HandleFunc("/hello/{name}", user.helloGetHandler).Methods("GET")
	r.HandleFunc("/health", healthHandler).Methods("GET")

	// Configure http server
	srv := &http.Server{
		Addr: "0.0.0.0:80",

		// Setup timeouts
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,

		Handler: r, // Pass the instance of gorilla/mux router
	}

	// Run the http server in a goroutine
	go func() {
		log.Fatal(srv.ListenAndServe())
	}()

	log.Println("App is ready to accept connections!")

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)

	// Accept graceful shutdown when quit via SIGINT (Ctrl+C) signal
	signal.Notify(c, os.Interrupt)

	// Block until receive a signal
	<-c

	log.Println("Gracefully shutting down the app...")
	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Wait until the timeout deadline if there are any active connections
	err = srv.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

// PUT "/hello/{name}" handler
func (u User) helloPutHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request from: %v, Method: %v, URI: %v\n", r.RemoteAddr, r.Method, r.RequestURI)

	// Get {name} variable from uri
	vars := mux.Vars(r)

	u.Name = vars["name"]

	// Read and unmarshall request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Can't get body: %v", err)
		http.Error(w, "Can't get body", http.StatusInternalServerError)
	}
	if err := json.Unmarshal(body, &u); err != nil {
		log.Printf("Can't unmarshal: %v", err)
		http.Error(w, "Can't unmarshal json", http.StatusInternalServerError)
	}

	// Put user into a database
	if err := addUser(u); err != nil {
		log.Printf("Can't add user %s with error: %v", u.Name, err)
		http.Error(w, "Can't add user to db", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET "/hello/{name}" handler
func (u User) helloGetHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request from: %v, Method: %v, URI: %v\n", r.RemoteAddr, r.Method, r.RequestURI)

	// Get {name} variable from uri
	vars := mux.Vars(r)
	u.Name = vars["name"]

	u, err := getUser(u.Name)
	if err != nil {
		log.Printf("Can't get user: %s", u.Name)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Parse string to date
	t, err := converStringToTime(u.DateOfBirth)
	if err != nil {
		log.Printf("Cannot convert date. Error: %v\n", err)
	}

	// Send json encoded message depends on birth date
	var msg Message
	dob := compareTime(t, time.Now())

	if dob == 5 {
		msg.Message = fmt.Sprintf("Hello, %v! Your birthday in 5 days", u.Name)
		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Cannot marshall json: %v\n", err)
			http.Error(w, "Cannot marshall json", http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(data); err != nil {
			log.Printf("Cannot send responce. Error: %v", err)
		}
		return
	}

	if dob == 0 {
		msg.Message = fmt.Sprintf("Hello, %v! Happy Birthday!", u.Name)
		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Cannot marshall json: %v\n", err)
			http.Error(w, "Cannot marshall json", http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(data); err != nil {
			log.Printf("Cannot send responce. Error: %v", err)
		}
		return
	}

	msg.Message = fmt.Sprintf("Hello, %v!", u.Name)
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Cannot marshall json: %v\n", err)
		http.Error(w, "Cannot marshall json", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(data); err != nil {
		log.Printf("Cannot send responce. Error: %v", err)
	}
	return
}

// GET "/health" handler
func healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request from: %v, Method: %v, URI: %v\n", r.RemoteAddr, r.Method, r.RequestURI)
	w.Header().Set("Content-Type", "application/json")

	var msg Health

	// If db is unreachable return an error
	if err := db.DB().Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg.Alive = false
		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Cannot marshall json: %v\n", err)
			http.Error(w, "Cannot marshall json", http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(data); err != nil {
			log.Printf("Cannot send responce. Error: %v", err)
		}
		log.Println("Cannot ping a db")
		return
	}

	msg.Alive = true
	w.WriteHeader(http.StatusOK)
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Cannot marshall json: %v\n", err)
		http.Error(w, "Cannot marshall json", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(data); err != nil {
		log.Printf("Cannot send responce. Error: %v", err)
	}
}

// Add user to a database
func addUser(u User) error {
	if err := db.Create(&u).Error; err != nil {
		return err
	}
	return nil
}

// Get user from a database
func getUser(name string) (User, error) {
	var user User
	result := db.Find(&user, &User{Name: name})
	return user, result.Error
}

// Compare current date and birthday date
func compareTime(x time.Time, y time.Time) int {
	if (x.Day() - 5) == y.Day() {
		// Birthday in 5 days
		return 5
	} else if x.Day() == y.Day() {
		// Today is a birthday
		return 0
	} else {
		// In all other cases
		return -1
	}
}

func converStringToTime(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
