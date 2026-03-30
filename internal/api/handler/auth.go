package handler

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"quant-agent/internal/store"
)

// ---------------------------------------------------------------------------
// 错误码定义
// ---------------------------------------------------------------------------
const (
	CodeSuccess           = 0
	CodeUsernameExists    = 1001
	CodeEmailExists       = 1002
	CodeWrongPassword     = 1003
	CodeUserNotFound      = 1004
	CodeInvalidInput      = 1005
)

// ---------------------------------------------------------------------------
// 请求/响应结构
// ---------------------------------------------------------------------------

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserResponse 用户信息响应
type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthData 认证数据
type AuthData struct {
	Token string        `json:"token"`
	User  UserResponse  `json:"user"`
}

// AuthResponse 统一响应结构
type AuthResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    *AuthData  `json:"data,omitempty"`
}

// ---------------------------------------------------------------------------
// JWT 配置
// ---------------------------------------------------------------------------

var jwtSecret = []byte("quant-agent-secret-key-2024")

// Claims JWT Claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// generateToken 生成 JWT token
func generateToken(userID, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 7 * time.Hour)), // 7天过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "quant-agent",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ---------------------------------------------------------------------------
// 辅助函数
// ---------------------------------------------------------------------------

// 邮箱格式验证
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// isValidUsername 检查用户名格式 (字母数字下划线，3-30位)
func isValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 30 {
		return false
	}
	match, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	return match
}

// hashPassword 哈希密码
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// checkPassword 验证密码
func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ---------------------------------------------------------------------------
// User Store Interface（支持数据库操作）
// ---------------------------------------------------------------------------

// UserStoreInterface 定义用户存储接口
type UserStoreInterface interface {
	CreateUser(user *User) error
	GetUserByUsername(username string) (*User, error)
	GetUserByEmail(email string) (*User, error)
}

// globalUserStore 全局用户存储
var globalUserStore UserStoreInterface

// SetUserStore 设置用户存储
func SetUserStore(s UserStoreInterface) {
	globalUserStore = s
}

// User 用户模型
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// StoreAuth 注入 store 以支持 auth 操作
type StoreAuth struct {
	store *store.Store
}

// NewStoreAuth 创建 StoreAuth
func NewStoreAuth(s *store.Store) *StoreAuth {
	return &StoreAuth{store: s}
}

// CreateUser 创建用户
func (s *StoreAuth) CreateUser(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.store.DB().Exec(ctx, `
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)

	return err
}

// GetUserByUsername 根据用户名获取用户
func (s *StoreAuth) GetUserByUsername(username string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user User
	err := s.store.DB().QueryRow(ctx, `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users WHERE username = $1
	`, username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *StoreAuth) GetUserByEmail(email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user User
	err := s.store.DB().QueryRow(ctx, `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// Register 注册处理器
func Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, AuthResponse{
				Code:    CodeInvalidInput,
				Message: "无效的请求参数: " + err.Error(),
			})
			return
		}

		// 输入验证
		if !isValidUsername(req.Username) {
			c.JSON(http.StatusBadRequest, AuthResponse{
				Code:    CodeInvalidInput,
				Message: "用户名格式不正确（需3-30位字母数字下划线）",
			})
			return
		}
		if !isValidEmail(req.Email) {
			c.JSON(http.StatusBadRequest, AuthResponse{
				Code:    CodeInvalidInput,
				Message: "邮箱格式不正确",
			})
			return
		}
		if len(req.Password) < 6 {
			c.JSON(http.StatusBadRequest, AuthResponse{
				Code:    CodeInvalidInput,
				Message: "密码至少6位",
			})
			return
		}

		// 检查用户名是否存在
		if globalUserStore != nil {
			existingUser, err := globalUserStore.GetUserByUsername(req.Username)
			if err == nil && existingUser != nil {
				c.JSON(http.StatusBadRequest, AuthResponse{
					Code:    CodeUsernameExists,
					Message: "用户名已存在",
				})
				return
			}

			// 检查邮箱是否存在
			existingEmail, err := globalUserStore.GetUserByEmail(req.Email)
			if err == nil && existingEmail != nil {
				c.JSON(http.StatusBadRequest, AuthResponse{
					Code:    CodeEmailExists,
					Message: "邮箱已被注册",
				})
				return
			}

			// 创建用户
			passwordHash, err := hashPassword(req.Password)
			if err != nil {
				c.JSON(http.StatusInternalServerError, AuthResponse{
					Code:    1,
					Message: "密码加密失败",
				})
				return
			}

			now := time.Now()
			user := &User{
				ID:           uuid.New().String(),
				Username:     req.Username,
				Email:        req.Email,
				PasswordHash: passwordHash,
				CreatedAt:    now,
				UpdatedAt:    now,
			}

			if err := globalUserStore.CreateUser(user); err != nil {
				c.JSON(http.StatusInternalServerError, AuthResponse{
					Code:    1,
					Message: "创建用户失败: " + err.Error(),
				})
				return
			}

			// 生成 token
			token, err := generateToken(user.ID, user.Username)
			if err != nil {
				c.JSON(http.StatusInternalServerError, AuthResponse{
					Code:    1,
					Message: "生成token失败",
				})
				return
			}

			c.JSON(http.StatusCreated, AuthResponse{
				Code:    CodeSuccess,
				Message: "注册成功",
				Data: &AuthData{
					Token: token,
					User: UserResponse{
						ID:        user.ID,
						Username:  user.Username,
						Email:     user.Email,
						CreatedAt: user.CreatedAt,
					},
				},
			})
			return
		}

		// 如果没有设置 store，返回错误
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    1,
			Message: "用户存储未初始化",
		})
	}
}

// Login 登录处理器
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, AuthResponse{
				Code:    CodeInvalidInput,
				Message: "无效的请求参数: " + err.Error(),
			})
			return
		}

		// 输入验证
		if req.Username == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, AuthResponse{
				Code:    CodeInvalidInput,
				Message: "用户名和密码不能为空",
			})
			return
		}

		if globalUserStore != nil {
			// 获取用户
			user, err := globalUserStore.GetUserByUsername(req.Username)
			if err != nil {
				c.JSON(http.StatusUnauthorized, AuthResponse{
					Code:    CodeUserNotFound,
					Message: "用户不存在",
				})
				return
			}

			// 验证密码
			if !checkPassword(req.Password, user.PasswordHash) {
				c.JSON(http.StatusUnauthorized, AuthResponse{
					Code:    CodeWrongPassword,
					Message: "密码错误",
				})
				return
			}

			// 生成 token
			token, err := generateToken(user.ID, user.Username)
			if err != nil {
				c.JSON(http.StatusInternalServerError, AuthResponse{
					Code:    1,
					Message: "生成token失败",
				})
				return
			}

			c.JSON(http.StatusOK, AuthResponse{
				Code:    CodeSuccess,
				Message: "登录成功",
				Data: &AuthData{
					Token: token,
					User: UserResponse{
						ID:        user.ID,
						Username:  user.Username,
						Email:     user.Email,
						CreatedAt: user.CreatedAt,
					},
				},
			})
			return
		}

		// 如果没有设置 store，返回错误
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    1,
			Message: "用户存储未初始化",
		})
	}
}
