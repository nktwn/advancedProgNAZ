package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Model struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type User struct {
	Model
	Name     string
	Email    string
	Age      int
	Username string
	Password string
}

type JsonRequest struct {
	Person struct {
		ID       uint   `json:"id"`
		Name     string `json:"name"`
		Age      int    `json:"age"`
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"person"`
	Status string `json:"status"`
}

type JsonResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

const registrationHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Registration</title>
</head>
<body>
    <h2>Registration Form</h2>
    <form action="/register" method="post">
        <label for="name">Name:</label>
        <input type="text" id="name" name="name" required><br>

        <label for="email">Email:</label>
        <input type="email" id="email" name="email" required><br>

        <label for="username">Username:</label>
        <input type="text" id="username" name="username" required><br>

        <label for="password">Password:</label>
        <input type="password" id="password" name="password" required><br>

        <label for="confirmPassword">Confirm Password:</label>
        <input type="password" id="confirmPassword" name="confirmPassword" required><br>

        <input type="submit" value="Register">
    </form>
</body>
</html>
`

var db *gorm.DB

func main() {
	dsn := "user=postgres dbname=advancedProg password=12345678 sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	db.AutoMigrate(&User{})

	router := mux.NewRouter()

	router.HandleFunc("/person", handlePostRequest).Methods("POST")
	router.HandleFunc("/person", handleGetRequest).Methods("GET")
	router.HandleFunc("/register", handleRegistrationPage).Methods("GET")
	router.HandleFunc("/register", handleRegistration).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	var requestData JsonRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestData)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, JsonResponse{
			Status:  "400",
			Message: "Invalid JSON message",
		})
		return
	}

	newUser := User{
		Name:     requestData.Person.Name,
		Email:    requestData.Person.Email,
		Age:      requestData.Person.Age,
		Username: requestData.Person.Username,
		Password: requestData.Person.Password,
	}
	if err := createUser(&newUser); err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, JsonResponse{
			Status:  "500",
			Message: "Error creating user",
		})
		return
	}

	SendJSONResponse(w, http.StatusOK, JsonResponse{
		Status:  "success",
		Message: "Data successfully received",
	})
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	userID := uint(1)

	userByID, err := getUserByID(userID)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, JsonResponse{
			Status:  "500",
			Message: "Error getting user by ID",
		})
		return
	}

	SendJSONResponse(w, http.StatusOK, userByID)
}

func handleRegistrationPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, registrationHTML)
}

func handleRegistration(w http.ResponseWriter, r *http.Request) {
	SendJSONResponse(w, http.StatusOK, JsonResponse{
		Status:  "success",
		Message: "Registration successfully completed",
	})
}

func SendJSONResponse(w http.ResponseWriter, status int, message interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(message)
}

func createUser(user *User) error {
	return db.Create(user).Error
}

func getUserByID(userID uint) (User, error) {
	var user User
	result := db.First(&user, userID)
	return user, result.Error
}
