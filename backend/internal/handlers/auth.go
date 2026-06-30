package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"goplayground/backend/internal/models"
	"goplayground/backend/internal/repository"
)

type AuthHandler struct {
	userRepo   *repository.UserRepository
	jwtSecret  string
	adminEmail string
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtSecret string, adminEmail string) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, jwtSecret: jwtSecret, adminEmail: adminEmail}
}

type RegisterRequest struct {
	Fullname string `json:"fullname" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	AccessToken string `json:"accessToken"`
	Role        string `json:"role"`
	Fullname    string `json:"fullname"`
	Email       string `json:"email"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	existing, _ := h.userRepo.FindByEmail(req.Email)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}
	role := "student"
	if h.adminEmail != "" && req.Email == h.adminEmail {
		role = "admin"
	}
	user := &models.User{
		Fullname:     req.Fullname,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         role,
	}
	if err := h.userRepo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}
	token, err := h.createToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}
	c.JSON(http.StatusCreated, AuthResponse{
		AccessToken: token,
		Role:        role,
		Fullname:    user.Fullname,
		Email:       user.Email,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	role := user.Role
	if role == "" {
		role = "student"
	}
	token, err := h.createToken(user.ID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}
	c.JSON(http.StatusOK, AuthResponse{
		AccessToken: token,
		Role:        role,
		Fullname:    user.Fullname,
		Email:       user.Email,
	})
}

type MeResponse struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Fullname string `json:"fullname"`
	Role     string `json:"role"`
}

func (h *AuthHandler) Me(c *gin.Context) {
	v, _ := c.Get("userID")
	var userID uint
	switch id := v.(type) {
	case float64:
		userID = uint(id)
	case uint:
		userID = id
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	user, err := h.userRepo.FindByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	role := user.Role
	if role == "" {
		role = "student"
	}
	c.JSON(http.StatusOK, MeResponse{
		ID:       user.ID,
		Email:    user.Email,
		Fullname: user.Fullname,
		Role:     role,
	})
}

func (h *AuthHandler) createToken(userID uint, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  float64(userID),
		"role": role,
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}
