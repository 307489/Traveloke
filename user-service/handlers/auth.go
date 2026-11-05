package handlers

import (
	"net/http"
	"os"
	"time"
	"user-service/database"
	"user-service/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required,min=3,max=30,alphanum"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6,max=72"`
		FullName string `json:"full_name" binding:"omitempty,max=100"`
		Phone    string `json:"phone" binding:"omitempty,numeric,min=8,max=15"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		ID:       uuid.New().String(),
		Username: input.Username,
		Email:    input.Email,
		Password: string(hashed),
		FullName: input.FullName,
		Phone:    input.Phone,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user_id": user.ID})
}

func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"is_admin": user.IsAdmin,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": signed})
}

func GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	var input struct {
		FullName        *string `json:"full_name" binding:"omitempty,max=100"`
		Phone           *string `json:"phone" binding:"omitempty,numeric,min=8,max=15"`
		Email           *string `json:"email" binding:"omitempty,email"`
		CurrentPassword *string `json:"current_password"`
		NewPassword     *string `json:"new_password" binding:"omitempty,min=6,max=72"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
		return
	}

	if input.FullName == nil && input.Phone == nil && input.Email == nil && input.NewPassword == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if input.FullName != nil {
		user.FullName = *input.FullName
	}

	if input.Phone != nil {
		user.Phone = *input.Phone
	}

	if input.Email != nil && *input.Email != user.Email {
		var count int64
		database.DB.Model(&models.User{}).
			Where("email = ? AND id <> ?", *input.Email, userID).
			Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already in use"})
			return
		}
		user.Email = *input.Email
	}

	if input.NewPassword != nil {
		if input.CurrentPassword == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is required to change password"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(*input.CurrentPassword)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
			return
		}

		if *input.CurrentPassword == *input.NewPassword {
			c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be different from current password"})
			return
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(*input.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = string(hashed)
	}

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated", "user": user})
}