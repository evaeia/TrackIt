package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GetRequest struct {
	Authorization string `json:"Authorization"`
	Container_id  int    `json:"Container_id"`
}

func InventoryGet(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestBody := GetRequest{}
		if err := c.BindJSON(&requestBody); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Verify that the token is valid.
		var username string
		if username := isValidToken(requestBody.Authorization, db); username != "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Check if the container belongs to the user.
		var cont Container
		if result := db.Table("items").Where("LocID = ? AND username = ?", requestBody.Container_id, username).First(&cont); result.Error != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid container"})
			return
		}

		// Get all containers that have the requested container as their parent.
		var containers []Container
		if result := db.Table("Containers").Where("parentID = ? ", requestBody.Container_id).Find(&containers); result.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get containers"})
			return
		}

		// Get all items that are in the requested container.
		var items []Item
		if result := db.Table("items").Where("locID = ?", requestBody.Container_id).Find(&items); result.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get items"})
			return
		}

		// Merge the containers and items into a single slice.
		var results []interface{}
		for _, container := range containers {
			results = append(results, container)
		}
		for _, item := range items {
			results = append(results, item)
		}

		c.JSON(http.StatusOK, results)
	}
}

func isValidToken(authHeader string, db *gorm.DB) string {

	token := strings.TrimPrefix(authHeader, "Bearer ")
	// Query the database for a user with the given token.
	var user Account
	if err := db.Table("Accounts").Where("token = ?", token).First(&user).Error; err != nil {
		// If no user with the token is found, return false.
		return ""
	}

	return user.Username
}
