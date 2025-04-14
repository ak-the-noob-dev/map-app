package server

// TODO: Need to update test after update the tileserver

// import (
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"server/infra/logger"
// 	"server/internal/database"
// 	"server/internal/tileserver"

// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/assert"
// )

// func setupTestServer(t *testing.T) *gin.Engine {
// 	gin.SetMode(gin.TestMode)
// 	r := gin.New()

// 	// Create a test server instance
// 	logger := logger.NewLogger()
// 	s := &Server{
// 		port:       8080,
// 		db:         database.New(logger),
// 		tileserver: tileserver.NewTileServerController(logger),
// 	}

// 	// Register routes
// 	s.RegisterRoutes().ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
// 	return r
// }

// func TestHelloWorldHandler(t *testing.T) {
// 	r := setupTestServer(t)
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/", nil)
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	var response map[string]string
// 	err := json.Unmarshal(w.Body.Bytes(), &response)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "Hello World", response["message"])
// }

// func TestHealthHandler(t *testing.T) {
// 	r := setupTestServer(t)
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/health", nil)
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)
// 	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
// }

// func TestTileServerRoutes(t *testing.T) {
// 	r := setupTestServer(t)

// 	tests := []struct {
// 		name           string
// 		url            string
// 		expectedStatus int
// 		expectedType   string
// 	}{
// 		{
// 			name:           "Valid tile request",
// 			url:            "/tiles/12/2048/1361",
// 			expectedStatus: http.StatusOK,
// 			expectedType:   "application/x-protobuf",
// 		},
// 		{
// 			name:           "Invalid tile request - missing coordinates",
// 			url:            "/tiles/12",
// 			expectedStatus: http.StatusBadRequest,
// 			expectedType:   "application/json",
// 		},
// 		{
// 			name:           "Invalid tile request - invalid coordinates",
// 			url:            "/tiles/12/invalid/1361",
// 			expectedStatus: http.StatusBadRequest,
// 			expectedType:   "application/json",
// 		},
// 		{
// 			name:           "Style request",
// 			url:            "/style.json",
// 			expectedStatus: http.StatusOK,
// 			expectedType:   "application/json",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			w := httptest.NewRecorder()
// 			req, _ := http.NewRequest("GET", tt.url, nil)
// 			r.ServeHTTP(w, req)

// 			assert.Equal(t, tt.expectedStatus, w.Code)
// 			assert.Equal(t, tt.expectedType, w.Header().Get("Content-Type"))

// 			if tt.expectedStatus == http.StatusOK && tt.expectedType == "application/json" {
// 				var response map[string]interface{}
// 				err := json.Unmarshal(w.Body.Bytes(), &response)
// 				assert.NoError(t, err)

// 				if tt.url == "/style.json" {
// 					// Verify style.json structure
// 					assert.Contains(t, response, "version")
// 					assert.Contains(t, response, "sources")
// 					assert.Contains(t, response, "layers")

// 					sources, ok := response["sources"].(map[string]interface{})
// 					assert.True(t, ok)
// 					assert.Contains(t, sources, "roads")

// 					roads, ok := sources["roads"].(map[string]interface{})
// 					assert.True(t, ok)
// 					assert.Equal(t, "vector", roads["type"])
// 				}
// 			}
// 		})
// 	}
// }

// func TestCorsMiddleware(t *testing.T) {
// 	r := setupTestServer(t)
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("OPTIONS", "/", nil)
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusNoContent, w.Code)
// 	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
// 	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS, PATCH", w.Header().Get("Access-Control-Allow-Methods"))
// 	assert.Equal(t, "Accept, Authorization, Content-Type, X-CSRF-Token", w.Header().Get("Access-Control-Allow-Headers"))
// 	assert.Equal(t, "false", w.Header().Get("Access-Control-Allow-Credentials"))
// }
