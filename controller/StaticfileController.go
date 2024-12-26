package controller

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func StaticImage(c *gin.Context) {
	filepath := c.Param("filepath")
	// Check if the filepath has a valid image extension
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
	isValid := false
	for _, ext := range validExtensions {
		if strings.HasSuffix(filepath, ext) {
			isValid = true
			break
		}
	}
	if !isValid {
		c.AbortWithStatusJSON(404, gin.H{"error": "File not found"})
		return
	}
	// Serve the image if valid
	c.File("images/" + filepath)
}
