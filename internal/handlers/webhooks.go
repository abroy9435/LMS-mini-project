package handlers

import (
	"LMS-mini-project-backend/internal/config"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	svix "github.com/svix/svix-webhooks/go"
)

// ClerkPayload represents the structure of the Clerk webhook JSON
type ClerkPayload struct {
	Type string `json:"type"`
	Data struct {
		ID             string `json:"id"`
		FirstName      string `json:"first_name"`
		LastName       string `json:"last_name"`
		EmailAddresses []struct {
			EmailAddress string `json:"email_address"`
		} `json:"email_addresses"`
	} `json:"data"`
}

func HandleClerkWebhook(c *gin.Context) {
	// 1. Get the Webhook Secret from your environment variables
	secret := os.Getenv("CLERK_WEBHOOK_SECRET")
	if secret == "" {
		log.Println("Error: CLERK_WEBHOOK_SECRET is not set")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server misconfiguration"})
		return
	}

	// 2. Read the raw request body (Required for Svix verification)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not read body"})
		return
	}

	// 3. Verify the Webhook Signature
	wh, err := svix.NewWebhook(secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create webhook object"})
		return
	}

	err = wh.Verify(payload, c.Request.Header)
	if err != nil {
		log.Println("Webhook verification failed:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// 4. Parse the verified JSON payload
	var event ClerkPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// 5. Handle the "user.created" event
	if event.Type == "user.created" || event.Type == "user.updated" {
		clerkID := event.Data.ID
		email := ""
		if len(event.Data.EmailAddresses) > 0 {
			email = event.Data.EmailAddresses[0].EmailAddress
		}

		firstName := event.Data.FirstName
		lastName := event.Data.LastName

		// 6. Update the Database using GORM
		// Make sure to import "LMS-mini-project-backend/internal/config" at the top

		result := config.DB.Table("users").
			Where("email = ?", email).
			Updates(map[string]interface{}{
				"clerk_id":   clerkID,
				"first_name": firstName,
				"last_name":  lastName,
			})

		if result.Error != nil {
			log.Println("Database error syncing user:", result.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync user to database"})
			return
		}

		if result.RowsAffected == 0 {
			log.Printf("Warning: Webhook fired for %s, but email not found in DB\n", email)
		} else {
			log.Printf("Successfully synced user %s with Clerk ID %s\n", email, clerkID)
		}
	}

	// Always return a 200 OK to acknowledge receipt
	c.JSON(http.StatusOK, gin.H{"message": "Webhook processed successfully"})
}
