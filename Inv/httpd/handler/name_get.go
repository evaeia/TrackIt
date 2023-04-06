package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NameGet(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			return
		}

		// Verify that the token is valid.
		var username string
		if username = IsValidToken(token, db); username == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		containerIDStr := c.Query("Container_id")
		if containerIDStr == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing container ID"})
			return
		}

		containerID, err := strconv.Atoi(containerIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"error": "Invalid container ID"})
			return
		}

		// Retrieve the container with the specified ID and username.
		var container Container
		if result := db.Table("Containers").Where("LocID = ? AND username = ?", containerID, username).First(&container); result.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get container"})
			return
		}

		// Add the name of the current container to the response.
		names := container.Name
		/*
			// Traverse the parent containers until ParentID equals 1.
			for container.ParentID != 0 {
				if result := db.Table("Containers").Where("ParentID = ? AND username = ?", container.ParentID, username).First(&container); result.Error != nil {
					c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Failed to get container"})
					return
				}
				names = container.Name + "/" + names
			}
		*/
		c.JSON(http.StatusOK, names)
	}
}
