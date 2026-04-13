package runtime

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

func setupTestServer(t *testing.T) *ginServer {
	gin.SetMode(gin.TestMode)

	tmpDir, err := os.MkdirTemp("", "test_db")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}

	cfg := appConfig{
		CredentialKey:   "test_key_12345678901234567890",
		ProjectCacheTTL: 3600,
		SessionIdleTTL:  1800,
		StaticDir:       "./static",
	}

	if err = initDB(db, cfg); err != nil {
		t.Fatalf("初始化测试数据库失败: %v", err)
	}

	srv := &server{
		db:                 db,
		tokenTTL:           24 * time.Hour,
		cfg:                cfg,
		jobs:               make(map[string]*asyncOperateJob),
		projectSessions:    newProjectSessionManager(),
		browserCloseStates: make(map[string]*browserCloseState),
	}

	return newGinServer(srv)
}

func TestHealthEndpoint(t *testing.T) {
	gs := setupTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	gs.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("解析响应失败: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("期望状态 'ok', 实际 '%s'", response["status"])
	}
}

func TestRegisterEndpoint(t *testing.T) {
	gs := setupTestServer(t)

	registerData := map[string]string{
		"username": "testuser",
		"password": "testpass123",
	}
	jsonData, _ := json.Marshal(registerData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	gs.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d, 响应: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestLoginEndpoint(t *testing.T) {
	gs := setupTestServer(t)

	registerData := map[string]string{
		"username": "loginuser",
		"password": "loginpass123",
	}
	jsonData, _ := json.Marshal(registerData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	gs.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("注册失败: %s", w.Body.String())
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	gs.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际 %d, 响应: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("解析响应失败: %v", err)
	}

	if _, ok := response["token"]; !ok {
		t.Error("登录响应中缺少 token")
	}
}

func TestLoginWithInvalidCredentials(t *testing.T) {
	gs := setupTestServer(t)

	registerData := map[string]string{
		"username": "validuser",
		"password": "validpass123",
	}
	jsonData, _ := json.Marshal(registerData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	gs.router.ServeHTTP(w, req)

	invalidData := map[string]string{
		"username": "validuser",
		"password": "wrongpassword",
	}
	jsonData, _ = json.Marshal(invalidData)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	gs.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddlewareWithoutToken(t *testing.T) {
	gs := setupTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/auth/me", nil)
	gs.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望状态码 %d, 实际 %d", http.StatusUnauthorized, w.Code)
	}
}

func TestProjectTypeValidation(t *testing.T) {
	tests := []struct {
		name     string
		projType string
		expected bool
	}{
		{"有效类型 ad", "ad", true},
		{"有效类型 print", "print", true},
		{"有效类型 vpn", "vpn", true},
		{"无效类型 unknown", "unknown", false},
		{"空字符串", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validProjectType(tt.projType)
			if result != tt.expected {
				t.Errorf("validProjectType(%s) = %v, 期望 %v", tt.projType, result, tt.expected)
			}
		})
	}
}

func TestCredentialProjectTypeValidation(t *testing.T) {
	tests := []struct {
		name     string
		projType string
		expected bool
	}{
		{"有效类型 ad", "ad", true},
		{"有效类型 print", "print", true},
		{"有效类型 vpn", "vpn", true},
		{"有效类型 vpn_firewall", "vpn_firewall", true},
		{"无效类型 unknown", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validCredentialProjectType(tt.projType)
			if result != tt.expected {
				t.Errorf("validCredentialProjectType(%s) = %v, 期望 %v", tt.projType, result, tt.expected)
			}
		})
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		expected   string
	}{
		{"有效 Bearer token", "Bearer abc123", "abc123"},
		{"小写 bearer", "bearer def456", "def456"},
		{"无前缀", "justtoken", ""},
		{"空字符串", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBearerToken(tt.authHeader)
			if result != tt.expected {
				t.Errorf("extractBearerToken(%q) = %q, 期望 %q", tt.authHeader, result, tt.expected)
			}
		})
	}
}

func TestToBool(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"float64 1", float64(1), true},
		{"float64 0", float64(0), false},
		{"int 1", 1, true},
		{"int 0", 0, false},
		{"string true", "true", true},
		{"string false", "false", false},
		{"string 1", "1", true},
		{"string 0", "0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toBool(tt.input)
			if result != tt.expected {
				t.Errorf("toBool(%v) = %v, 期望 %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{"短字符串不截断", "hello", 10, "hello"},
		{"长字符串截断", "hello world", 5, "hello"},
		{"空字符串", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.max)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, 期望 %q", tt.input, tt.max, result, tt.expected)
			}
		})
	}
}
