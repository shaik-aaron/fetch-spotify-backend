package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type App struct {
	DB *sqlx.DB
}

type Request struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
	ClientID     string `json:"client_id"`
}

type Response struct {
	Message string `json:"message"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type SetTokenRequest struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (app *App) GetHello(c *gin.Context) {
	response := Response{
		Message: "Hello world",
	}
	c.JSON(http.StatusOK, response)
}

func (app *App) GetToken(c *gin.Context) {
	accessToken := ""
	query := `SELECT access_token FROM tokens WHERE id = 1`
	err := app.DB.Get(&accessToken, query)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No access token found"})
		return
	}
	c.JSON(http.StatusOK, accessToken)
}

func (app *App) GetCurrentlyPlayingTrack(c *gin.Context) {
	accessToken := ""
	query := `SELECT access_token FROM tokens WHERE id = 1`
	err := app.DB.Get(&accessToken, query)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No access token found"})
		return
	}
	track_url := "https://api.spotify.com/v1/me/player/currently-playing"

	client := &http.Client{}
	req, err := http.NewRequest("GET", track_url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	res, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make request"})
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNoContent {
		c.JSON(http.StatusOK, gin.H{"message": "No track currently playing"})
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// Parse the JSON to ensure it's valid before returning
	var trackData json.RawMessage
	err = json.Unmarshal(body, &trackData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid JSON response from Spotify"})
		return
	}

	c.JSON(http.StatusOK, trackData)
}

func (app *App) GetRecentlyPlayedTracks(c *gin.Context) {
	accessToken := ""
	query := `SELECT access_token FROM tokens WHERE id = 1`
	err := app.DB.Get(&accessToken, query)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No access token found"})
		return
	}
	track_url := "https://api.spotify.com/v1/me/player/recently-played"

	client := &http.Client{}
	req, err := http.NewRequest("GET", track_url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	res, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make request"})
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNoContent {
		c.JSON(http.StatusOK, gin.H{"message": "No recently played tracks"})
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// Parse the JSON to ensure it's valid before returning
	var trackData json.RawMessage
	err = json.Unmarshal(body, &trackData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid JSON response from Spotify"})
		return
	}

	c.JSON(http.StatusOK, trackData)
}

func (app *App) RefreshToken(c *gin.Context) {
	var refreshToken string
	query := `SELECT refresh_token FROM tokens WHERE id = 1`
	err := app.DB.Get(&refreshToken, query)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No refresh token found"})
		return
	}
	tokenURL := "https://accounts.spotify.com/api/token"

	// Prepare form data as URL-encoded string
	log.Println("refreshToken", refreshToken)
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", "b03179410679401796263b9d74a630cd")

	log.Println("data", data.Encode())

	post, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer post.Body.Close()

	// Log the HTTP status code
	log.Println("HTTP Status Code:", post.StatusCode)

	// Read the response body
	bodyBytes, err := io.ReadAll(post.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	// Log the raw response
	log.Println("Raw Response Body:", string(bodyBytes))

	// Parse the JSON response
	var response RefreshResponse
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		log.Println("JSON Parse Error:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response: " + err.Error()})
		return
	}

	// Log the parsed response
	log.Printf("Parsed Response: AccessToken=%s, RefreshToken=%s, ExpiresIn=%d",
		response.AccessToken, response.RefreshToken, response.ExpiresIn)
	query = `UPDATE tokens SET access_token = ?, refresh_token = ? WHERE id = 1`
	_, err = app.DB.Exec(query, response.AccessToken, response.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
	})
}

// Admin API to set token through API
func (app *App) SetToken(c *gin.Context) {
	var req SetTokenRequest

	// Bind JSON payload to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Check if a token record exists
	var count int
	countQuery := `SELECT COUNT(*) FROM tokens WHERE id = 1`
	err := app.DB.Get(&count, countQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing tokens: " + err.Error()})
		return
	}

	var query string
	var result interface{}

	if count > 0 {
		// Update existing token
		query = `UPDATE tokens SET access_token = ?, refresh_token = ?, updated_at = CURRENT_TIMESTAMP WHERE id = 1`
		result, err = app.DB.Exec(query, req.AccessToken, req.RefreshToken)
	} else {
		// Insert new token
		query = `INSERT INTO tokens (id, access_token, refresh_token) VALUES (1, ?, ?)`
		result, err = app.DB.Exec(query, req.AccessToken, req.RefreshToken)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save tokens: " + err.Error()})
		return
	}

	log.Printf("Tokens set successfully. Result: %+v", result)
	c.JSON(http.StatusOK, gin.H{
		"message":       "Tokens set successfully",
		"access_token":  req.AccessToken[:10] + "...",  // Only show first 10 chars for security
		"refresh_token": req.RefreshToken[:10] + "...", // Only show first 10 chars for security
	})
}
