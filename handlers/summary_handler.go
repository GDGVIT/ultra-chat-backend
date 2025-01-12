package handlers

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"time"
	"ultra-chat-backend/repositories"
	"ultra-chat-backend/utils"
)

type SummaryHandler struct {
	repo *repositories.MongoSummaryRepository // Use pointer to MongoSummaryRepository
}

func NewSummaryHandler(summaryRepo *repositories.MongoSummaryRepository) *SummaryHandler {
	return &SummaryHandler{repo: summaryRepo} // Initialize with summaryRepo
}

func (h *SummaryHandler) CreateSummary(c echo.Context) error {
	type RequestBody struct {
		Content   string `json:"content"`
		ServerID  string `json:"server_id"`
		IsPrivate bool   `json:"is_private"`
		UserID    string `json:"user_id"`
	}

	var body RequestBody
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if body.Content == "" || body.ServerID == "" || body.UserID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing required fields"})
	}

	exists, dbErr := h.repo.CheckUserExists(body.UserID)
	if dbErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}
	if !exists {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: User not found"})
	}

	summaryID := uuid.New().String()

	createdAt := time.Now().Format(time.RFC3339)

	err := h.repo.AddSummary(summaryID, body.UserID, body.ServerID, body.IsPrivate, body.Content, createdAt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create summary"})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message":    "Summary created successfully",
		"summary_id": summaryID,
	})
}

func (h *SummaryHandler) GetSummaries(c echo.Context) error {
	userID := c.Request().Header.Get("ID")
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	filter := bson.M{"user_id": userID}
	summaries, err := h.repo.GetSummaries(filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve summaries"})
	}

	return c.JSON(http.StatusOK, summaries)
}

func (h *SummaryHandler) UpdateSummary(c echo.Context) error {
	type RequestBody struct {
		SummaryID string `json:"summary_id"`
		ServerID  string `json:"server_id"`
		IsPrivate bool   `json:"is_private"`
		Content   string `json:"content"`
	}

	var body RequestBody
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	userID := c.Request().Header.Get("ID")
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	if body.SummaryID == "" || body.ServerID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing required fields"})
	}

	err := h.repo.UpdateSummary(userID, body.ServerID, body.IsPrivate, body.Content)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update summary"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Summary updated successfully"})
}

func (h *SummaryHandler) DeleteSummary(c echo.Context) error {
	type RequestBody struct {
		SummaryID string `json:"summary_id"`
	}
	var body RequestBody
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	userID := c.Request().Header.Get("ID")
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	if err := h.repo.DeleteSummary(userID, body.SummaryID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete summary"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Summary deleted successfully"})
}

func (h *SummaryHandler) IsAuthenticated(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	fmt.Println(authHeader)
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: Missing token"})
	}

	accessToken := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		accessToken = authHeader[7:]
	} else {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: Invalid token format"})
	}

	userInfo, err := utils.FetchUserInfo(accessToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: Failed to validate token"})
	}

	userID, ok := userInfo["id"].(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: Invalid user data"})
	}

	c.Set("userID", userID)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":   "Authenticated",
		"user_id":   userID,
		"user_info": userInfo,
	})
}
