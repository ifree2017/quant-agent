package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"quant-agent/internal/store"
)

var testAuthStore *store.Store

// init sets up the auth test store.
func init() {
	gin.SetMode(gin.TestMode)

	// 初始化测试用数据库连接
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@127.0.0.1:5432/quant_agent?sslmode=disable"
	}

	ctx := context.Background()
	var err error
	testAuthStore, err = store.NewStore(ctx, connStr)
	if err != nil {
		// 如果无法连接数据库，testAuthStore 为 nil
		return
	}

	// 注入 user store
	SetUserStore(NewStoreAuth(testAuthStore))
}

func cleanupTestUsers() {
	if testAuthStore == nil {
		return
	}
	ctx := context.Background()
	testAuthStore.DB().Exec(ctx, "DELETE FROM users WHERE username LIKE 'testuser%' OR username = 'duplicate' OR username = 'user1' OR username = 'user2' OR username = 'loginuser' OR username = 'wrongpass'")
}

// Helper to call Register handler
func callRegister(t *testing.T, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	Register()(c)
	return w
}

// Helper to call Login handler
func callLogin(t *testing.T, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	Login()(c)
	return w
}

// TestRegisterSuccess - 注册成功
func TestRegisterSuccess(t *testing.T) {
	if testAuthStore == nil {
		t.Skip("database not available")
	}
	defer cleanupTestUsers()

	body := `{"username":"testuser","email":"test@example.com","password":"TestPass123"}`
	w := callRegister(t, body)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != CodeSuccess {
		t.Errorf("expected code %d, got %d, message: %s", CodeSuccess, resp.Code, resp.Message)
	}
}

// TestRegisterDuplicateUsername - 用户名重复
func TestRegisterDuplicateUsername(t *testing.T) {
	if testAuthStore == nil {
		t.Skip("database not available")
	}
	defer cleanupTestUsers()

	// 先注册一个用户
	body := `{"username":"duplicate","email":"dup1@example.com","password":"TestPass123"}`
	callRegister(t, body)

	// 再用相同用户名注册
	body2 := `{"username":"duplicate","email":"dup2@example.com","password":"TestPass123"}`
	w2 := callRegister(t, body2)

	if w2.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for duplicate, got %d", http.StatusBadRequest, w2.Code)
	}

	var resp AuthResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != CodeUsernameExists {
		t.Errorf("expected code %d (用户名已存在), got %d", CodeUsernameExists, resp.Code)
	}
}

// TestRegisterDuplicateEmail - 邮箱重复
func TestRegisterDuplicateEmail(t *testing.T) {
	if testAuthStore == nil {
		t.Skip("database not available")
	}
	defer cleanupTestUsers()

	// 先注册一个用户
	body := `{"username":"user1","email":"dup_email@example.com","password":"TestPass123"}`
	callRegister(t, body)

	// 再用相同邮箱注册
	body2 := `{"username":"user2","email":"dup_email@example.com","password":"TestPass123"}`
	w2 := callRegister(t, body2)

	if w2.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for duplicate email, got %d", http.StatusBadRequest, w2.Code)
	}

	var resp AuthResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != CodeEmailExists {
		t.Errorf("expected code %d (邮箱已注册), got %d", CodeEmailExists, resp.Code)
	}
}

// TestRegisterInvalidInput - 无效输入
func TestRegisterInvalidInput(t *testing.T) {
	if testAuthStore == nil {
		t.Skip("database not available")
	}
	defer cleanupTestUsers()

	tests := []struct {
		name string
		body string
	}{
		{"empty username", `{"username":"","email":"test@example.com","password":"TestPass123"}`},
		{"empty email", `{"username":"testuser","email":"","password":"TestPass123"}`},
		{"empty password", `{"username":"testuser","email":"test@example.com","password":""}`},
		{"invalid email", `{"username":"testuser","email":"notanemail","password":"TestPass123"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := callRegister(t, tt.body)
			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status %d for %s, got %d", http.StatusBadRequest, tt.name, w.Code)
			}
		})
	}
}

// TestLoginSuccess - 登录成功
func TestLoginSuccess(t *testing.T) {
	if testAuthStore == nil {
		t.Skip("database not available")
	}
	defer cleanupTestUsers()

	// 先注册
	regBody := `{"username":"loginuser","email":"login@example.com","password":"TestPass123"}`
	regW := callRegister(t, regBody)
	if regW.Code != http.StatusCreated {
		t.Fatalf("setup failed: register returned %d", regW.Code)
	}

	// 再登录
	loginBody := `{"username":"loginuser","password":"TestPass123"}`
	loginW := callLogin(t, loginBody)

	if loginW.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, loginW.Code)
	}

	var resp AuthResponse
	if err := json.Unmarshal(loginW.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != CodeSuccess {
		t.Errorf("expected code %d, got %d, message: %s", CodeSuccess, resp.Code, resp.Message)
	}
	if resp.Data == nil || resp.Data.Token == "" {
		t.Error("expected token in response")
	}
}

// TestLoginWrongPassword - 密码错误
func TestLoginWrongPassword(t *testing.T) {
	if testAuthStore == nil {
		t.Skip("database not available")
	}
	defer cleanupTestUsers()

	// 注册
	regBody := `{"username":"wrongpass","email":"wrong@example.com","password":"CorrectPass"}`
	regW := callRegister(t, regBody)
	if regW.Code != http.StatusCreated {
		t.Fatalf("setup failed: register returned %d", regW.Code)
	}

	// 用错误密码登录
	loginBody := `{"username":"wrongpass","password":"WrongPass"}`
	loginW := callLogin(t, loginBody)

	if loginW.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, loginW.Code)
	}

	var resp AuthResponse
	if err := json.Unmarshal(loginW.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != CodeWrongPassword {
		t.Errorf("expected code %d (密码错误), got %d", CodeWrongPassword, resp.Code)
	}
}

// TestLoginUserNotFound - 用户不存在
func TestLoginUserNotFound(t *testing.T) {
	if testAuthStore == nil {
		t.Skip("database not available")
	}
	defer cleanupTestUsers()

	loginBody := `{"username":"nonexistent","password":"TestPass123"}`
	loginW := callLogin(t, loginBody)

	if loginW.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, loginW.Code)
	}

	var resp AuthResponse
	if err := json.Unmarshal(loginW.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != CodeUserNotFound {
		t.Errorf("expected code %d (用户不存在), got %d", CodeUserNotFound, resp.Code)
	}
}

// TestLoginInvalidInput - 登录无效输入
func TestLoginInvalidInput(t *testing.T) {
	if testAuthStore == nil {
		t.Skip("database not available")
	}
	defer cleanupTestUsers()

	tests := []struct {
		name string
		body string
	}{
		{"empty username", `{"username":"","password":"TestPass123"}`},
		{"empty password", `{"username":"testuser","password":""}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := callLogin(t, tt.body)
			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status %d for %s, got %d", http.StatusBadRequest, tt.name, w.Code)
			}
		})
	}
}
