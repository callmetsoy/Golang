package v1

import (
	"net/http"
	"practice-7/internal/entity"
	"practice-7/internal/usecase"
	"practice-7/pkg/logger"
	"practice-7/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type userRoutes struct {
	t usecase.UserInterface
	l logger.Interface
}

func NewUserRoutes(handler *gin.RouterGroup, t usecase.UserInterface, l logger.Interface, rl *utils.RateLimiter) {
	r := &userRoutes{t, l}
	h := handler.Group("/users")
	h.Use(utils.RateLimiterMiddleware(rl))
	{
		h.POST("/", r.RegisterUser)
		h.POST("/login", r.LoginUser)
		protected := h.Group("/")
		protected.Use(utils.JWTAuthMiddleware())
		{
			protected.GET("/me", r.GetMe)
			protected.PATCH("/promote/:id", utils.RoleMiddleware("admin"), r.PromoteUser)
			protected.GET("/protected/hello", r.ProtectedFunc)
		}
	}
}

func (r *userRoutes) RegisterUser(c *gin.Context) {
    var createUserDTO entity.CreateUserDTO
    if err := c.ShouldBindJSON(&createUserDTO); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    hashedPassword, err := utils.HashPassword(createUserDTO.Password)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
        return
    }
    user := entity.User{
        Username: createUserDTO.Username,
        Email:    createUserDTO.Email,
        Password: hashedPassword,
    }
    createdUser, sessionID, err := r.t.RegisterUser(&user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    // Return user without password
    c.JSON(http.StatusCreated, gin.H{
        "message":    "User registered successfully. Please check your email for verification code.",
        "session_id": sessionID,
        "user": gin.H{
            "id":       createdUser.ID,
            "username": createdUser.Username,
            "email":    createdUser.Email,
            "role":     createdUser.Role,
        },
    })
}

func (r *userRoutes) LoginUser(c *gin.Context) {
	var input entity.LoginUserDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := r.t.LoginUser(&input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (r *userRoutes) GetMe(c *gin.Context) {
    userIDStr, ok := c.Get("userID")
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
        return
    }

    userID, err := uuid.Parse(userIDStr.(string))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    user, err := r.t.GetUserByID(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "id":       user.ID,
        "username": user.Username,
        "email":    user.Email,
        "role":     user.Role,
    })
}

func (r *userRoutes) PromoteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	err = r.t.PromoteUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User promoted to admin"})
}


func (r *userRoutes) ProtectedFunc(c *gin.Context) {
	c.JSON(200, gin.H{"message": "OK"})
}