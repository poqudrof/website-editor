package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ContentRequest struct {
	Content         string `json:"content"`          // The edited content
	OriginalContent string `json:"original_content"` // Original HTML content (sent on first edit)
}

func GetContent(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		var content Content
		result := db.First(&content, "id = ?", id)

		if result.Error != nil {
			// Return empty/not found
			return c.JSON(fiber.Map{
				"id":        id,
				"content":   "",
				"is_edited": false,
			})
		}

		// Return edited content if exists, otherwise original
		displayContent := content.EditedContent
		if !content.IsEdited {
			displayContent = content.OriginalContent
		}

		return c.JSON(fiber.Map{
			"id":               content.ID,
			"content":          displayContent,
			"original_content": content.OriginalContent,
			"edited_content":   content.EditedContent,
			"is_edited":        content.IsEdited,
			"updated_at":       content.UpdatedAt,
		})
	}
}

func PutContent(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		var req ContentRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		var content Content
		result := db.First(&content, "id = ?", id)

		if result.Error != nil {
			// First time - create new record with original content
			content = Content{
				ID:              id,
				OriginalContent: req.OriginalContent,
				EditedContent:   req.Content,
				IsEdited:        true,
				UpdatedAt:       time.Now().Unix(),
			}
		} else {
			// Update existing - only update edited content
			content.EditedContent = req.Content
			content.IsEdited = true
			content.UpdatedAt = time.Now().Unix()

			// Set original content if provided and not already set
			if req.OriginalContent != "" && content.OriginalContent == "" {
				content.OriginalContent = req.OriginalContent
			}
		}

		db.Save(&content)

		return c.JSON(fiber.Map{
			"id":               content.ID,
			"content":          content.EditedContent,
			"original_content": content.OriginalContent,
			"edited_content":   content.EditedContent,
			"is_edited":        content.IsEdited,
			"updated_at":       content.UpdatedAt,
		})
	}
}
