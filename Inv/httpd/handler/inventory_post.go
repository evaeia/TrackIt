package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InvRequest struct {
	Authorization string `json:"Authorization"`
	Kind          string `json:"Kind"` // container or item?
	ID            int    `json:"ID"`
	Cont          int    `json:"Cont"`
	Name          string `json:"Name"`
	Type          string `json:"Type"`
	Count         int    `json:"Count"` //for items
}

func InventoryPost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestBody := InvRequest{}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Verify that the token is valid.
		var username string
		if username = IsValidToken(requestBody.Authorization, db); username == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		switch requestBody.Kind {
		case "container":
			{
				newContainer := Container{
					LocID:    requestBody.ID,
					Name:     requestBody.Name,
					ParentID: requestBody.Cont,
					User:     username,
				}

				if result := db.Table("containers").Create(&newContainer); result.Error != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create container"})
					return
				}
			}
		case "item":
			{
				newItem := Item{
					ItemID:   requestBody.ID,
					User:     username,
					ItemName: requestBody.Name,
					LocID:    requestBody.Cont,
					Count:    requestBody.Count,
				}

				if result := db.Table("items").Create(&newItem); result.Error != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item"})
					return
				}
			}
		default:
			{
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
				return
			}
		}

		c.Status(http.StatusNoContent)
	}
}
