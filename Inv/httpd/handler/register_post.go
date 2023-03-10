package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	//current encrypt
	"golang.org/x/crypto/bcrypt"
	/*/ salting
	        "crypto/rand"
	        "crypto/sha256"
	        "encoding/base64"
	  // encrypting
	        "crypto/aes"
	        "crypto/cipher"
	        "crypto/rand"
	        "encoding/base64"
	        "fmt"
	        "io"*/)

type RegisterRequest struct {
	Username             string `json:"username"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

// Hash and Salt password
func hashAndSalt(password []byte) string {

	// Use GenerateFromPassword to hash & salt pwd
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	// Hash the password using the salt
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.MinCost)
	if err != nil {
		println(err)
		return ""
	}
	// Convert the hash to a string and return it
	return string(hash)
}

func RegisterPost(DB *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse the request body.
		var request RegisterRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Check if the user exists.
		var existingUser Account
		if result := DB.Table("accounts").Where("username = ?", request.Username).First(&existingUser); result.Error == nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
			return
		}

		// Check if the password is correct.
		if request.Password != request.PasswordConfirmation {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Password and password confirmation do not match"})
			return
		}

		// Create a new user object with the provided username and password.
		newUser := Account{
			Username: request.Username,
			Password: hashAndSalt([]byte(request.Password)), //replaced with hash and salt password,
			Token:    generateToken(),
		}
		// Create a new container object with a unique LocID.
		var maxLocID int64
		err := DB.Table("containers").Select("MAX(LocID)").Row().Scan(&maxLocID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get max LocID"})
			return
		}

		newContainer := Container{
			LocID:    int(maxLocID) + 1,
			Name:     newUser.Username + "'s container",
			ParentID: 0, // Assuming it's a top-level container.
		}

		// Start a new transaction to ensure atomicity.
		tx := DB.Begin()

		// Create the new user and container objects in the database.
		if result := tx.Table("accounts").Create(&newUser); result.Error != nil {
			tx.Rollback()
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		if result := tx.Table("containers").Create(&newContainer); result.Error != nil {
			tx.Rollback()
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create container"})
			return
		}

		// Update the new user's RootLoc to the LocID of the new container.
		if result := tx.Table("accounts").Where("username = ?", newUser.Username).Update("rootLoc", newContainer.LocID); result.Error != nil {
			tx.Rollback()
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user's RootLoc"})
			return
		}

		// Commit the transaction.
		tx.Commit()

		// Return the token to the user.
		response := LoginResponse{Token: newUser.Token, RootLoc: newContainer.LocID}
		c.JSON(http.StatusOK, response)
	}
}
