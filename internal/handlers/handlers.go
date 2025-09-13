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

func (app *App) GetHello(c *gin.Context) {
	response := Response{
		Message: "Hello world",
	}
	c.JSON(http.StatusOK, response)
}

func (app *App) GetToken(c *gin.Context) {
	var accessToken string
	query := `SELECT access_token FROM tokens WHERE id = 1`
	err := app.DB.Get(&accessToken, query)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No access token found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
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
	c.JSON(http.StatusOK, response)
}
