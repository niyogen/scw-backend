package handler

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"delivery-backend/internal/service"
	"delivery-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	mediaSvc service.MediaService
}

func NewMediaHandler(mediaSvc service.MediaService) *MediaHandler {
	return &MediaHandler{
		mediaSvc: mediaSvc,
	}
}

// UploadPhoto handles single image upload (profile picture / avatar / service photo)
// POST /api/v1/upload/photo or POST /api/v1/upload/avatar
func (h *MediaHandler) UploadPhoto(c *gin.Context) {
	// Accept either "file", "photo", "image", or "avatar" form field
	fileHeader, err := c.FormFile("file")
	if err != nil {
		fileHeader, err = c.FormFile("photo")
	}
	if err != nil {
		fileHeader, err = c.FormFile("image")
	}
	if err != nil {
		fileHeader, err = c.FormFile("avatar")
	}

	if err != nil {
		utils.JSONBadRequest(c, "No image file provided in form data (expected 'file', 'photo', 'image', or 'avatar')", nil)
		return
	}

	// Max 10MB file size limit
	if fileHeader.Size > 10*1024*1024 {
		utils.JSONBadRequest(c, "File size exceeds 10MB limit", nil)
		return
	}

	// Validate content type & extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".gif":  true,
	}

	if !validExts[ext] {
		utils.JSONBadRequest(c, fmt.Sprintf("Invalid image format '%s'. Allowed: JPG, JPEG, PNG, WEBP, GIF", ext), nil)
		return
	}

	srcFile, err := fileHeader.Open()
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed opening uploaded file", err.Error())
		return
	}
	defer srcFile.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		switch ext {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".webp":
			contentType = "image/webp"
		case ".gif":
			contentType = "image/gif"
		}
	}

	mediaURL, objectPath, err := h.mediaSvc.UploadPhoto(c.Request.Context(), srcFile, fileHeader.Filename, contentType)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Upload failed", err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Image uploaded successfully to Google Cloud Storage", gin.H{
		"url":         mediaURL,
		"object_path": objectPath,
		"filename":    fileHeader.Filename,
		"size":        fileHeader.Size,
		"mime_type":   contentType,
	})
}

// ServeMedia streams an object from Google Cloud Storage or local fallback
// GET /api/v1/media/*filepath
func (h *MediaHandler) ServeMedia(c *gin.Context) {
	filepathParam := c.Param("filepath")
	filepathParam = strings.TrimPrefix(filepathParam, "/")

	if filepathParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filepath required"})
		return
	}

	reader, contentType, err := h.mediaSvc.GetMedia(c.Request.Context(), filepathParam)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media file not found"})
		return
	}
	defer reader.Close()

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("Access-Control-Allow-Origin", "*")

	_, _ = io.Copy(c.Writer, reader)
}
