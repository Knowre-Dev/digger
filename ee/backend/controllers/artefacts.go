package controllers

import (
	"fmt"
	"github.com/diggerhq/digger/backend/middleware"
	"github.com/diggerhq/digger/backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
)

func SetJobArtefact(c *gin.Context) {
	jobTokenValue, exists := c.Get(middleware.JOB_TOKEN_KEY)
	if !exists {
		c.String(http.StatusBadRequest, "missing value: bearer job token")
		return
	}

	jobToken, err := models.DB.GetJobToken(jobTokenValue)
	if err != nil {
		c.String(http.StatusBadRequest, "could not find job token")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
		return
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error opening file"})
		return
	}
	defer src.Close()

	// Read the content
	content, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading file content"})
		return
	}

	// Deleting existing artefacts
	err = models.DB.DeleteJobTokenArtefacts(jobToken.ID)
	if err != nil {
		log.Printf("could not delete artefacts: %v", err)
		c.JSON(http.StatusInternalServerError, "could not delete existing artefacts")
		return
	}

	log.Printf("contents of the file is: %v", string(content))
	// Create a new File record
	artefactRecord := models.JobArtefact{
		JobTokenID:  jobToken.ID,
		Filename:    uuid.NewString() + ".zip",
		Contents:    content,
		Size:        file.Size,
		ContentType: file.Header.Get("Content-Type"),
	}

	// Save the file to the database
	if result := models.DB.GormDB.Create(&artefactRecord); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving file to database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "id": artefactRecord.ID})

}

func DownloadJobArtefact(c *gin.Context) {
	jobTokenValue, exists := c.Get(middleware.JOB_TOKEN_KEY)
	if !exists {
		c.String(http.StatusBadRequest, "missing value: bearer job token")
		return
	}

	jobToken, err := models.DB.GetJobToken(jobTokenValue)
	if err != nil {
		c.String(http.StatusBadRequest, "could not find job token")
		return
	}

	artefact, err := models.DB.GetJobArtefact(jobToken.ID)
	if err != nil {
		c.String(http.StatusBadRequest, "could not find any artefacts for job token")
		return
	}

	// Set the appropriate headers
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", artefact.Filename))
	c.Header("Content-Type", artefact.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", artefact.Size))

	// Write the file contents to the response
	c.Data(http.StatusOK, artefact.ContentType, artefact.Contents)
}
