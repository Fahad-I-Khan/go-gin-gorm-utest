package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testRouter *gin.Engine

func resetDatabase(db *gorm.DB) {
    db.Exec("DELETE FROM users") // Clear all users
    db.Exec("DELETE FROM sqlite_sequence WHERE name='users'") // Reset auto-increment IDs (specific to SQLite)
}

func setupTestEnvironment() {
	// Use an in-memory SQLite database for testing
	db, _ = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.AutoMigrate(&User{})

	testRouter = gin.Default()
	initializeRoutes(testRouter)
}

func initializeRoutes(r *gin.Engine) {
	r.GET("/api/v1/users", getUsers)
	r.GET("/api/v1/users/:id", getUser)
	r.POST("/api/v1/users", createUser)
	r.PUT("/api/v1/users/:id", updateUser)
	r.DELETE("/api/v1/users/:id", deleteUser)
}

func TestGetUsers(t *testing.T) {
	setupTestEnvironment()

	// Seed the database
	db.Create(&User{Name: "Alice", Email: "alice@example.com"})
	db.Create(&User{Name: "Bob", Email: "bob@example.com"})

	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var users []User
	_ = json.Unmarshal(w.Body.Bytes(), &users)
	assert.Equal(t, 2, len(users))
}

func TestGetUser(t *testing.T) {
	resetDatabase(db)

	// Seed the database
	user := User{Name: "Charlie", Email: "charlie@example.com"}
	db.Create(&user)

	req, _ := http.NewRequest("GET", "/api/v1/users/1", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var fetchedUser User
	_ = json.Unmarshal(w.Body.Bytes(), &fetchedUser)
	assert.Equal(t, "Charlie", fetchedUser.Name)
}

func TestCreateUser(t *testing.T) {
	setupTestEnvironment()

	newUser := User{Name: "Dave", Email: "dave@example.com"}
	jsonData, _ := json.Marshal(newUser)

	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createdUser User
	// db.First(&createdUser, "email = ?", "dave@example.com")
	err := json.Unmarshal(w.Body.Bytes(), &createdUser)
	assert.NoError(t, err, "Response body should unmarshal correctly")
	assert.Equal(t, "Dave", createdUser.Name)
}

func TestUpdateUser(t *testing.T) {
	// Reset the database to ensure test independence
    resetDatabase(db)

	// Seed the database
	user := User{Name: "Eve", Email: "eve@example.com"}
	db.Create(&user)

	updatedUser := User{Name: "Eve Updated", Email: "eve.updated@example.com"}
	jsonData, _ := json.Marshal(updatedUser)

	req, _ := http.NewRequest("PUT", "/api/v1/users/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var fetchedUser User
	err := json.Unmarshal(w.Body.Bytes(), &fetchedUser)
	assert.NoError(t, err, "Response body should unmarshal correctly")
	// db.First(&fetchedUser, 1)
	assert.Equal(t, "Eve Updated", fetchedUser.Name)
}

func TestDeleteUser(t *testing.T) {
	// Reset the database to ensure test independence
    resetDatabase(db)

	// Seed the database
	user := User{Name: "Frank", Email: "frank@example.com"}
	db.Create(&user)

	req, _ := http.NewRequest("DELETE", "/api/v1/users/1", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var fetchedUser User
	err := db.First(&fetchedUser, 1).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}
