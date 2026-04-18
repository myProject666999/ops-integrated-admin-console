package runtime

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

var testServer *server
var testRouter *gin.Engine

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Setenv("AD_API_URL", "http://localhost:8081")
	os.Setenv("PRINT_API_URL", "http://localhost:8082")
	os.Setenv("VPN_SSH_ADDR", "localhost:22")
	os.Setenv("FIREWALL_SSH_ADDR", "localhost:22")
	os.Setenv("CREDENTIAL_SECRET", "test-secret-key-for-testing")
	os.Setenv("PROJECT_CACHE_TTL_MINUTES", "10")
	os.Setenv("SESSION_IDLE_TTL_MINUTES", "60")

	cfg := appConfig{
		ADAPIURL:        "http://localhost:8081",
		PrintAPIURL:     "http://localhost:8082",
		VPNSshAddr:      "localhost:22",
		FirewallSSHAddr: "localhost:22",
		CredentialKey:   "test-secret-key-for-testing",
		ProjectCacheTTL: 10 * time.Minute,
		SessionIdleTTL:  60 * time.Minute,
		StaticDir:       "./static",
	}
	runtimeCfg = cfg

	dbPath := "./test_db/test.db"
	os.MkdirAll("./test_db", 0o755)

	db, err := initTestDB(dbPath, cfg)
	if err != nil {
		panic("failed to init test db: " + err.Error())
	}

	testServer = &server{
		db:                 db,
		tokenTTL:           24 * time.Hour,
		cfg:                cfg,
		jobs:               make(map[string]*asyncOperateJob),
		projectSessions:    newProjectSessionManager(),
		browserCloseStates: make(map[string]*browserCloseState),
	}

	testRouter = setupRouter(testServer)

	code := m.Run()

	db.Close()
	os.RemoveAll("./test_db")
	os.Exit(code)
}

func initTestDB(path string, cfg appConfig) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err = initDB(db, cfg); err != nil {
		return nil, err
	}
	return db, nil
}

func TestHealthEndpoint(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

func TestRegisterEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		payload    map[string]string
		wantStatus int
	}{
		{
			name:       "valid registration",
			payload:    map[string]string{"username": "testuser1", "password": "password123"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty username",
			payload:    map[string]string{"username": "", "password": "password123"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty password",
			payload:    map[string]string{"username": "testuser2", "password": ""},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "short password",
			payload:    map[string]string{"username": "testuser3", "password": "short"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status code %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestLoginEndpoint(t *testing.T) {
	registerBody, _ := json.Marshal(map[string]string{
		"username": "loginuser",
		"password": "password123",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(registerBody))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	tests := []struct {
		name       string
		payload    map[string]string
		wantStatus int
	}{
		{
			name:       "valid login",
			payload:    map[string]string{"username": "loginuser", "password": "password123"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid password",
			payload:    map[string]string{"username": "loginuser", "password": "wrongpassword"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "non-existent user",
			payload:    map[string]string{"username": "nonexistent", "password": "password123"},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status code %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestProtectedEndpoints(t *testing.T) {
	registerBody, _ := json.Marshal(map[string]string{
		"username": "protecteduser",
		"password": "password123",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(registerBody))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	loginBody, _ := json.Marshal(map[string]string{
		"username": "protecteduser",
		"password": "password123",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	var loginResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	token := loginResp["token"].(string)

	t.Run("me endpoint with valid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		testRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("me endpoint without token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/auth/me", nil)
		testRouter.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("logout endpoint", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		testRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestChangePassword(t *testing.T) {
	registerBody, _ := json.Marshal(map[string]string{
		"username": "pwduser",
		"password": "oldpassword",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(registerBody))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	loginBody, _ := json.Marshal(map[string]string{
		"username": "pwduser",
		"password": "oldpassword",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	var loginResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	token := loginResp["token"].(string)

	tests := []struct {
		name       string
		payload    map[string]string
		wantStatus int
	}{
		{
			name:       "valid password change",
			payload:    map[string]string{"old_password": "oldpassword", "new_password": "newpassword123"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "wrong old password",
			payload:    map[string]string{"old_password": "wrongpassword", "new_password": "newpassword123"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "short new password",
			payload:    map[string]string{"old_password": "oldpassword", "new_password": "short"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/auth/change-password", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			testRouter.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status code %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestProjectCredentials(t *testing.T) {
	registerBody, _ := json.Marshal(map[string]string{
		"username": "creduser",
		"password": "password123",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(registerBody))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	loginBody, _ := json.Marshal(map[string]string{
		"username": "creduser",
		"password": "password123",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	var loginResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	token := loginResp["token"].(string)

	t.Run("get credentials", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/projects/credentials", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		testRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("update credential", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"account":  "testaccount",
			"password": "testpassword",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/projects/credentials/ad", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		testRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})
}
