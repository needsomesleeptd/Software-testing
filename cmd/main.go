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
var hasPassedAuth = make(map[string]bool)

func main() {
	err := godotenv.Load("../app.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	adminEmail = os.Getenv("USER_EMAIL")
	adminPassword = os.Getenv("USER_PASSWORD")
	verificationCodes[adminEmail] = os.Getenv("USER_SECRET")
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

func generateQRCode(secret string, email string, qrFilename string) error {
	// Create the TOTP key based on the secret and the account
	opts := totp.GenerateOpts{

		Issuer: "My Web Application",

		AccountName: email,

		Secret: []byte(secret),
	}
	key, err := totp.Generate(opts)
	if err != nil {
		return fmt.Errorf("failed to create TOTP key: %w", err)
	}

	// Create the QR Code image
	buf := new(bytes.Buffer)
	img, err := key.Image(200, 200)
	if err != nil {
		return fmt.Errorf("failed to generate QR code image: %w", err)
	}

	// Encode the image to PNG
	if err := png.Encode(buf, img); err != nil {
		return fmt.Errorf("failed to encode QR code image: %w", err)
	}

	// Write the QR code to a file
	if err := os.WriteFile(qrFilename+".png", buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write QR code to file: %w", err)
	}

	return nil
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

	// Generate the QR code and save it to a file
	if err := generateQRCode(key.Secret(), email, qrFilename); err != nil {
		return "", err
	}

	// Return the TOTP secret
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
		hasPassedAuth[request.Email] = true
		c.JSON(http.StatusOK, gin.H{"message": "Totp Qr saved on the disk", "totp_secret": verificationCodes[request.Email]})
	} else {
		hasPassedAuth[request.Email] = true
		generateQRCode(verificationCodes[request.Email], request.Email, request.Email+".png")
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

	if !hasPassedAuth[request.Email] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Need to login first"})
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
