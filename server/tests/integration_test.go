package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kokp520/banking-system/server/internal/handler"
	"github.com/kokp520/banking-system/server/internal/service"
	"github.com/kokp520/banking-system/server/internal/storage"
	"github.com/kokp520/banking-system/server/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupRouter 設置測試路由
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	// 初始化測試環境的logger
	logger.Init("info", "json", "")

	memoryStorage := storage.NewMemoryStorage()
	accountService := service.NewAccountService(memoryStorage)
	accountHandler := handler.NewAccountHandler(accountService)

	r := gin.New()
	r.Use(gin.Recovery()) // 添加recovery中間件
	v1 := r.Group("/v1")
	{
		account := v1.Group("/account")
		{
			account.POST("", accountHandler.CreateAccount)
			account.GET("/:id", accountHandler.GetAccount)
			account.POST("/:id/deposit", accountHandler.Deposit)
			account.POST("/:id/withdraw", accountHandler.Withdraw)
			account.POST("/:id/transfer", accountHandler.Transfer)
		}
	}

	return r
}

// TestCreateAccountAPI 測試創建帳戶API
func TestCreateAccountAPI(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name: "create account successfully",
			requestBody: map[string]interface{}{
				"name":            "test case1",
				"initial_balance": "100.50",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "missing name error",
			requestBody: map[string]interface{}{
				"initial_balance": "100.00",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "balance 0",
			requestBody: map[string]interface{}{
				"name":            "test case3 ",
				"initial_balance": "0",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/v1/account", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 如果期望錯誤但得到了錯誤的狀態碼，打印調試信息
			if tt.expectError && w.Code != tt.expectedStatus {
				t.Logf("test user: %s", tt.name)
				t.Logf("Expected status %d, got %d", tt.expectedStatus, w.Code)
				t.Logf("Response body: %s", w.Body.String())
			}

			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				data := response["data"].(map[string]interface{})
				assert.NotNil(t, data["id"])
				assert.Equal(t, tt.requestBody["name"], data["name"])
			}
		})
	}
}

// TestGetAccountAPI 測試查詢帳戶API
func TestGetAccountAPI(t *testing.T) {
	router := setupRouter()

	// 先創建一個帳戶
	createReq := map[string]interface{}{
		"name":            "test user",
		"initial_balance": "250.75",
	}
	jsonBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/v1/account", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	accountID := createResponse["data"].(map[string]interface{})["id"].(float64)

	// 測試查詢帳戶
	t.Run("get account", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/v1/account/%d", int(accountID)), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		assert.Equal(t, accountID, data["id"])
		assert.Equal(t, "test user", data["name"])
		assert.Equal(t, "250.75", data["balance"])
	})

	t.Run("get non-existing account", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/account/9999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestDepositAPI 測試存款API
func TestDepositAPI(t *testing.T) {
	router := setupRouter()

	accountID := createTestAccount(t, router, "test user", "100.00")

	tests := []struct {
		name           string
		accountID      string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name:      "Valid deposit",
			accountID: fmt.Sprintf("%d", accountID),
			requestBody: map[string]interface{}{
				"amount": "50.25",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:      "Zero deposit",
			accountID: fmt.Sprintf("%d", accountID),
			requestBody: map[string]interface{}{
				"amount": "0",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:      "Negative deposit",
			accountID: fmt.Sprintf("%d", accountID),
			requestBody: map[string]interface{}{
				"amount": "-10.00",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},

		{
			name:      "Non-existing account",
			accountID: "9999",
			requestBody: map[string]interface{}{
				"amount": "50.00",
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", fmt.Sprintf("/v1/account/%s/deposit", tt.accountID), bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "deposit successful", response["data"].(map[string]interface{})["message"])
			}
		})
	}
}

// TestWithdrawAPI 測試提款API
func TestWithdrawAPI(t *testing.T) {
	router := setupRouter()

	// 創建測試帳戶
	accountID := createTestAccount(t, router, "Withdraw Test User", "100.00")

	tests := []struct {
		name           string
		accountID      string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name:      "Valid withdrawal",
			accountID: fmt.Sprintf("%d", accountID),
			requestBody: map[string]interface{}{
				"amount": "30.00",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:      "Insufficient balance",
			accountID: fmt.Sprintf("%d", accountID),
			requestBody: map[string]interface{}{
				"amount": "200.00",
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:      "Zero withdrawal",
			accountID: fmt.Sprintf("%d", accountID),
			requestBody: map[string]interface{}{
				"amount": "0",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", fmt.Sprintf("/v1/account/%s/withdraw", tt.accountID), bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestTransferAPI 測試轉帳API
func TestTransferAPI(t *testing.T) {
	router := setupRouter()

	// 創建測試帳戶
	fromAccountID := createTestAccount(t, router, "From User", "200.00")
	toAccountID := createTestAccount(t, router, "To User", "50.00")

	tests := []struct {
		name           string
		fromAccountID  string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name:          "Valid transfer",
			fromAccountID: fmt.Sprintf("%d", fromAccountID),
			requestBody: map[string]interface{}{
				"to_account_id": toAccountID,
				"amount":        "75.50",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:          "Transfer to same account",
			fromAccountID: fmt.Sprintf("%d", fromAccountID),
			requestBody: map[string]interface{}{
				"to_account_id": fromAccountID,
				"amount":        "10.00",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:          "Insufficient balance",
			fromAccountID: fmt.Sprintf("%d", fromAccountID),
			requestBody: map[string]interface{}{
				"to_account_id": toAccountID,
				"amount":        "300.00",
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", fmt.Sprintf("/v1/account/%s/transfer", tt.fromAccountID), bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestConcurrentAPIOperations 測試併發API操作
func TestConcurrentAPIOperations(t *testing.T) {
	router := setupRouter()

	// 創建多個測試帳戶
	accountIDs := make([]int, 10)
	for i := 0; i < 10; i++ {
		accountIDs[i] = createTestAccount(t, router, fmt.Sprintf("Concurrent User %d", i), "1000.00")
	}

	var wg sync.WaitGroup
	goroutineCount := 100
	wg.Add(goroutineCount)

	// 併發執行多種操作
	for i := 0; i < goroutineCount; i++ {
		go func(index int) {
			defer wg.Done()

			accountID := accountIDs[index%len(accountIDs)]

			switch index % 4 {
			case 0:
				// 查詢操作
				req, _ := http.NewRequest("GET", fmt.Sprintf("/v1/account/%d", accountID), nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

			case 1:
				// 存款操作
				requestBody := map[string]interface{}{"amount": "10.00"}
				jsonBody, _ := json.Marshal(requestBody)
				req, _ := http.NewRequest("POST", fmt.Sprintf("/v1/account/%d/deposit", accountID), bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

			case 2:
				// 提款操作
				requestBody := map[string]interface{}{"amount": "5.00"}
				jsonBody, _ := json.Marshal(requestBody)
				req, _ := http.NewRequest("POST", fmt.Sprintf("/v1/account/%d/withdraw", accountID), bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

			case 3:
				// 轉帳操作
				toAccountID := accountIDs[(index+1)%len(accountIDs)]
				if accountID != toAccountID {
					requestBody := map[string]interface{}{
						"to_account_id": toAccountID,
						"amount":        "1.00",
					}
					jsonBody, _ := json.Marshal(requestBody)
					req, _ := http.NewRequest("POST", fmt.Sprintf("/v1/account/%d/transfer", accountID), bytes.NewBuffer(jsonBody))
					req.Header.Set("Content-Type", "application/json")
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)
				}
			}
		}(i)
	}

	wg.Wait()

	// 驗證所有帳戶仍然可以正常查詢
	for _, accountID := range accountIDs {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/v1/account/%d", accountID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}

// TestCompleteWorkflow 測試完整的銀行業務流程
func TestCompleteWorkflow(t *testing.T) {
	router := setupRouter()

	// 1. 創建三個帳戶
	AID := createTestAccount(t, router, "A", "1000.00")
	BID := createTestAccount(t, router, "B", "500.00")
	CID := createTestAccount(t, router, "C", "0.00")

	// a deposit 200
	depositRequest := map[string]interface{}{"amount": "200.00"}
	jsonBody, _ := json.Marshal(depositRequest)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/v1/account/%d/deposit", AID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// b withdraw 100
	withdrawRequest := map[string]interface{}{"amount": "100.00"}
	jsonBody, _ = json.Marshal(withdrawRequest)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/v1/account/%d/withdraw", BID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// A transfer to C 300
	transferRequest := map[string]interface{}{
		"to_account_id": CID,
		"amount":        "300.00",
	}
	jsonBody, _ = json.Marshal(transferRequest)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/v1/account/%d/transfer", AID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// check
	expectedBalances := map[int]string{
		AID: "900.00", // 1000 + 200 - 300
		BID: "400.00", // 500 - 100
		CID: "300.00", // 0 + 300
	}

	for accountID, expectedBalance := range expectedBalances {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/v1/account/%d", accountID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		actualBalance := data["balance"].(string)
		assert.Equal(t, expectedBalance, actualBalance, "Account %d balance mismatch", accountID)
	}
}

func createTestAccount(t *testing.T, router *gin.Engine, name, initialBalance string) int {
	createReq := map[string]interface{}{
		"name":            name,
		"initial_balance": initialBalance,
	}

	jsonBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/v1/account", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	return int(data["id"].(float64))
}
