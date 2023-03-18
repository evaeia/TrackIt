package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DeleteRequest struct {
	Token string `json:"token"`
	Type  string `json:"type"`
	ID    int    `json:"id"`
}

func InventoryDelete(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		requestBody := DeleteRequest{}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Verify that the token is valid.
		if !isValidToken(requestBody.Token, db) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		if requestBody.Type != "item" && requestBody.Type != "container" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid type"})
			return
		}

		if requestBody.Type == "item" {
			if err := deleteItem(db, requestBody.ID, requestBody.Token); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else if requestBody.Type == "container" {
			// Delete all items and sub-containers associated with the container.
			if err := DestroyContainer(db, requestBody.ID); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

		}

		c.Status(http.StatusNoContent)

	}
}

func deleteItem(db *gorm.DB, id int, token string) error {
	// Check if the item belongs to the user.
	var item Item
	if result := db.Table("items").Where("id = ? AND username = ?", id, getUsernameFromToken(token, db)).First(&item); result.Error != nil {
		return result.Error
	}

	// Delete the item.
	if result := db.Table("items").Delete(&item); result.Error != nil {
		return result.Error
	}

	return nil
}

func DestroyContainer(db *gorm.DB, locID int) error {
	// Check if the container belongs to the user.//////////////////////////////////////////////////////////// TODO IMPROVE
	var container Container
	if result := db.Table("containers").Where("LocID = ? AND ParentID = 0", locID).First(&container); result.Error != nil {
		return result.Error
	}

	// Delete all items inside the container and any sub-containers.
	if result := db.Table("items").Where("LocID = ? OR LocID IN (SELECT LocID FROM containers WHERE ParentID = ?)", locID, locID).Delete(&Item{}); result.Error != nil {
		return result.Error
	}

	// Delete all containers inside the container and any sub-containers.
	if result := db.Table("containers").Where("ParentID = ?", locID).Delete(&Container{}); result.Error != nil {
		return result.Error
	}

	// Delete the container.
	if result := db.Table("containers").Delete(&container); result.Error != nil {
		return result.Error
	}

	return nil
}

func destroyContainer(db *gorm.DB, locID int) error {
	// Check if the container belongs to the user. //////////////////////////////////////////////////////////// TODO IMPROVE
	var container Container
	if result := db.Table("containers").Where("LocID = ? AND ParentID = 0", locID).First(&container); result.Error != nil {
		return result.Error
	}

	// Use a recursive CTE to delete all containers and sub-containers in a single query.
	query := `
		WITH RECURSIVE cte AS (
			SELECT LocID FROM containers WHERE LocID = ?
			UNION ALL
			SELECT LocID FROM containers WHERE ParentID IN (SELECT LocID FROM cte)
		)
		DELETE FROM items WHERE LocID IN (SELECT LocID FROM cte);
		DELETE FROM containers WHERE LocID IN (SELECT LocID FROM cte);
	`

	if result := db.Exec(query, locID); result.Error != nil {
		return result.Error
	}

	return nil
}

/*
The DestroyContainer function expects the ID of a top-level container to be passed in as an argument, but there is no validation that the container is actually a top-level container. If an ID of a non-top-level container is passed in, the function will delete all items and sub-containers associated with that container, but leave the container itself intact.

The DestroyContainer function deletes all items and sub-containers associated with the specified container without checking if they belong to the user. If an attacker has access to another user's token, they could use this endpoint to delete items and containers belonging to that user.

*/
