package main

import (
	"bytes"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pquerna/otp/totp"
)

var adminEmail string
var adminPassword string
var verificationCodes = make(map[string]string)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	adminEmail = os.Getenv("USER_EMAIL")
	adminPassword = os.Getenv("USER_PASSWORD")

	if adminEmail == "" || adminPassword == "" {
		log.Fatalf("Invalid password and email")
	}

	r := gin.Default()

	r.GET("/status", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
	r.POST("/login", loginHandler)
	r.POST("/verify", verifyLoginHandler)
	r.POST("/reset_password", resetPasswordHandler)

	if err := r.Run(":8084"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func generateSecretTotp(email string, qrFilename string) (string, error) {
	options := totp.GenerateOpts{
		Issuer:      "My Web Application",
		AccountName: email,
	}
	key, err := totp.Generate(options)
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	buf := new(bytes.Buffer)
	img, err := key.Image(200, 200)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code image: %w", err)
	}

	if err := png.Encode(buf, img); err != nil {
		return "", fmt.Errorf("failed to encode QR code image: %w", err)
	}

	if err := os.WriteFile(qrFilename+".png", buf.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("failed to write QR code to file: %w", err)
	}

	return key.Secret(), nil
}

func loginHandler(c *gin.Context) {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if request.Email != adminEmail || request.Password != adminPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if verificationCodes[request.Email] == "" {
		totpCode, err := generateSecretTotp(request.Email, request.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to create secret totp for email"})
			return
		}
		verificationCodes[request.Email] = totpCode
		c.JSON(http.StatusOK, gin.H{"message": "Totp Qr saved on the disk", "totp_secret": verificationCodes[request.Email]})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Totp Qr saved on the disk, you must already have totp"})
	}
}

func verifyLoginHandler(c *gin.Context) {
	var request struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	expectedCode, exists := verificationCodes[request.Email]

	isValid := totp.Validate(request.Code, expectedCode)
	if !exists || !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid verification code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Auth success!"})
}

func resetPasswordHandler(c *gin.Context) {
	var request struct {
		Email       string `json:"email"`
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"` // Added NewPassword field
		TotpCode    string `json:"totpCode"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if request.Email != adminEmail || request.OldPassword != adminPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	userTotpSecret, exists := verificationCodes[request.Email]
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "TOTP secret not found"})
		return
	}

	isValid := totp.Validate(request.TotpCode, userTotpSecret)
	if !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid verification code"})
		return
	}

	adminPassword = request.NewPassword

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
