package main

import (
	"goplayground/backend/internal/config"
	"goplayground/backend/internal/handlers"
	"goplayground/backend/internal/middleware"
	"goplayground/backend/internal/models"
	"goplayground/backend/internal/repository"
	"goplayground/backend/internal/runner"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")
	cfg := config.Load()
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	if err := db.AutoMigrate(&models.User{}, &models.File{}); err != nil {
		panic("migration failed: " + err.Error())
	}

	userRepo := repository.NewUserRepository(db)
	if cfg.AdminEmail != "" && cfg.AdminPassword != "" {
		existing, _ := userRepo.FindByEmail(cfg.AdminEmail)
		if existing == nil {
			hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), 10)
			if err == nil {
				admin := &models.User{
					Email:        cfg.AdminEmail,
					Fullname:     cfg.AdminFullname,
					PasswordHash: string(hash),
					Role:         "admin",
				}
				_ = userRepo.Create(admin)
			}
		}
	}

	dockerRunner, err := runner.NewDockerRunner(cfg.GoRunnerImage, cfg.PythonRunnerImage, cfg.RunTimeout)
	if err != nil {
		log.Fatal("failed to create docker runner: ", err)
	}
	defer dockerRunner.Close()

	fileRepo := repository.NewFileRepository(db)
	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWTSecret, cfg.AdminEmail)
	filesHandler := handlers.NewFilesHandler(fileRepo)
	usersHandler := handlers.NewUsersHandler(userRepo)
	runHandler := handlers.NewRunHandler(dockerRunner, cfg.RunTimeout)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	api := r.Group("/api")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		protected := api.Group("")
		protected.Use(middleware.Auth(cfg.JWTSecret))
		{
			protected.GET("/auth/me", authHandler.Me)
			protected.GET("/files", filesHandler.List)
			protected.POST("/files", filesHandler.Create)
			protected.GET("/files/:id", filesHandler.Get)
			protected.PUT("/files/:id", filesHandler.Update)
			protected.DELETE("/files/:id", filesHandler.Delete)
			protected.GET("/users", usersHandler.List)
			protected.PUT("/users/:id", usersHandler.Update)
			protected.POST("/run", runHandler.Run)
		}
	}

	r.Run(":" + cfg.Port)
}
