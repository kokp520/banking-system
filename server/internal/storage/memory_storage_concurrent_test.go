package storage

import (
	"sync"
	"testing"
	"time"

	"github.com/kokp520/banking-system/server/internal/model"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConcurrentDeposits 測試併發存款的安全性
func TestConcurrentDeposits(t *testing.T) {
	storage := NewMemoryStorage()
	
	// 創建帳戶
	account := &model.Account{
		Name:    "Concurrent Test User",
		Balance: decimal.NewFromFloat(100.0),
	}
	err := storage.CreateAccount(account)
	require.NoError(t, err)
	
	// 併發存款測試
	goroutineCount := 100
	depositAmount := decimal.NewFromFloat(1.0)
	
	var wg sync.WaitGroup
	wg.Add(goroutineCount)
	
	// 啟動100個goroutine同時存款
	for i := 0; i < goroutineCount; i++ {
		go func() {
			defer wg.Done()
			err := storage.Deposit(account.ID, depositAmount)
			assert.NoError(t, err)
		}()
	}
	
	wg.Wait()
	
	// 驗證最終餘額
	finalAccount, err := storage.GetAccountByID(account.ID)
	require.NoError(t, err)
	
	expectedBalance := decimal.NewFromFloat(100.0).Add(decimal.NewFromFloat(100.0)) // 100 + (100 * 1)
	assert.True(t, expectedBalance.Equal(finalAccount.Balance))
}

// TestConcurrentWithdraws 測試併發提款的安全性
func TestConcurrentWithdraws(t *testing.T) {
	storage := NewMemoryStorage()
	
	// 創建帳戶，餘額足夠進行併發提款
	account := &model.Account{
		Name:    "Concurrent Test User",
		Balance: decimal.NewFromFloat(1000.0),
	}
	err := storage.CreateAccount(account)
	require.NoError(t, err)
	
	goroutineCount := 100
	withdrawAmount := decimal.NewFromFloat(5.0)
	
	var wg sync.WaitGroup
	var successfulWithdraws int64
	var mutex sync.Mutex
	
	wg.Add(goroutineCount)
	
	// 啟動100個goroutine同時提款
	for i := 0; i < goroutineCount; i++ {
		go func() {
			defer wg.Done()
			err := storage.Withdraw(account.ID, withdrawAmount)
			if err == nil {
				mutex.Lock()
				successfulWithdraws++
				mutex.Unlock()
			}
		}()
	}
	
	wg.Wait()
	
	// 驗證最終餘額和成功提款次數
	finalAccount, err := storage.GetAccountByID(account.ID)
	require.NoError(t, err)
	
	expectedBalance := decimal.NewFromFloat(1000.0).Sub(decimal.NewFromFloat(float64(successfulWithdraws) * 5.0))
	assert.True(t, expectedBalance.Equal(finalAccount.Balance))
	
	// 確保不會出現負餘額
	assert.True(t, finalAccount.Balance.GreaterThanOrEqual(decimal.Zero))
}

// TestConcurrentTransfers 測試併發轉帳的安全性和死鎖預防
func TestConcurrentTransfers(t *testing.T) {
	storage := NewMemoryStorage()
	
	// 創建多個帳戶
	accountCount := 10
	accounts := make([]*model.Account, accountCount)
	
	for i := 0; i < accountCount; i++ {
		account := &model.Account{
			Name:    "User " + string(rune(i+'A')),
			Balance: decimal.NewFromFloat(1000.0),
		}
		err := storage.CreateAccount(account)
		require.NoError(t, err)
		accounts[i] = account
	}
	
	// 計算初始總餘額
	initialTotal := decimal.NewFromFloat(float64(accountCount) * 1000.0)
	
	goroutineCount := 200
	var wg sync.WaitGroup
	wg.Add(goroutineCount)
	
	// 啟動大量併發轉帳
	for i := 0; i < goroutineCount; i++ {
		go func(index int) {
			defer wg.Done()
			
			// 隨機選擇兩個不同的帳戶
			fromIdx := index % accountCount
			toIdx := (index + 1) % accountCount
			
			if fromIdx != toIdx {
				amount := decimal.NewFromFloat(10.0)
				storage.Transfer(accounts[fromIdx].ID, accounts[toIdx].ID, amount)
			}
		}(i)
	}
	
	wg.Wait()
	
	// 驗證總餘額保持不變（轉帳不會創造或銷毀金錢）
	var finalTotal decimal.Decimal
	for _, account := range accounts {
		finalAccount, err := storage.GetAccountByID(account.ID)
		require.NoError(t, err)
		finalTotal = finalTotal.Add(finalAccount.Balance)
	}
	
	assert.True(t, initialTotal.Equal(finalTotal), 
		"Total balance should remain unchanged after transfers. Initial: %s, Final: %s", 
		initialTotal.String(), finalTotal.String())
}

// TestConcurrentTransferDeadlockPrevention 測試轉帳死鎖預防
func TestConcurrentTransferDeadlockPrevention(t *testing.T) {
	storage := NewMemoryStorage()
	
	// 創建兩個帳戶用於測試死鎖場景
	account1 := &model.Account{Name: "User 1", Balance: decimal.NewFromFloat(1000.0)}
	account2 := &model.Account{Name: "User 2", Balance: decimal.NewFromFloat(1000.0)}
	
	storage.CreateAccount(account1)
	storage.CreateAccount(account2)
	
	goroutineCount := 100
	var wg sync.WaitGroup
	wg.Add(goroutineCount * 2) // 每個方向各100個goroutine
	
	// 同時進行雙向轉帳，測試死鎖預防機制
	for i := 0; i < goroutineCount; i++ {
		// A→B 轉帳
		go func() {
			defer wg.Done()
			storage.Transfer(account1.ID, account2.ID, decimal.NewFromFloat(1.0))
		}()
		
		// B→A 轉帳
		go func() {
			defer wg.Done()
			storage.Transfer(account2.ID, account1.ID, decimal.NewFromFloat(1.0))
		}()
	}
	
	// 設置超時，如果發生死鎖會超時
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()
	
	select {
	case <-done:
		// 測試通過，沒有死鎖
		t.Log("No deadlock detected")
	case <-time.After(10 * time.Second):
		t.Fatal("Deadlock detected - test timed out")
	}
	
	// 驗證總餘額保持不變
	finalAccount1, _ := storage.GetAccountByID(account1.ID)
	finalAccount2, _ := storage.GetAccountByID(account2.ID)
	totalBalance := finalAccount1.Balance.Add(finalAccount2.Balance)
	expectedTotal := decimal.NewFromFloat(2000.0)
	
	assert.True(t, expectedTotal.Equal(totalBalance))
}

// TestConcurrentReadWrite 測試讀寫併發安全性
func TestConcurrentReadWrite(t *testing.T) {
	storage := NewMemoryStorage()
	
	// 創建測試帳戶
	account := &model.Account{
		Name:    "Read Write Test User",
		Balance: decimal.NewFromFloat(1000.0),
	}
	err := storage.CreateAccount(account)
	require.NoError(t, err)
	
	var wg sync.WaitGroup
	readCount := 200
	writeCount := 50
	
	wg.Add(readCount + writeCount)
	
	// 啟動大量讀操作
	for i := 0; i < readCount; i++ {
		go func() {
			defer wg.Done()
			_, err := storage.GetAccountByID(account.ID)
			assert.NoError(t, err)
		}()
	}
	
	// 啟動一些寫操作
	for i := 0; i < writeCount; i++ {
		go func(index int) {
			defer wg.Done()
			if index%2 == 0 {
				storage.Deposit(account.ID, decimal.NewFromFloat(1.0))
			} else {
				storage.Withdraw(account.ID, decimal.NewFromFloat(1.0))
			}
		}(i)
	}
	
	wg.Wait()
	
	// 驗證帳戶仍然存在且可讀取
	finalAccount, err := storage.GetAccountByID(account.ID)
	assert.NoError(t, err)
	assert.NotNil(t, finalAccount)
}

// TestConcurrentAccountCreation 測試併發創建帳戶
func TestConcurrentAccountCreation(t *testing.T) {
	storage := NewMemoryStorage()
	
	goroutineCount := 100
	var wg sync.WaitGroup
	wg.Add(goroutineCount)
	
	// 併發創建帳戶
	for i := 0; i < goroutineCount; i++ {
		go func(index int) {
			defer wg.Done()
			account := &model.Account{
				Name:    "Concurrent User " + string(rune(index)),
				Balance: decimal.NewFromFloat(100.0),
			}
			err := storage.CreateAccount(account)
			assert.NoError(t, err)
			assert.True(t, account.ID > 0)
		}(i)
	}
	
	wg.Wait()
	
	// 驗證所有帳戶都被正確創建
	// 最後一個帳戶的ID應該等於goroutineCount
	lastAccount := &model.Account{
		Name:    "Final User",
		Balance: decimal.NewFromFloat(100.0),
	}
	err := storage.CreateAccount(lastAccount)
	require.NoError(t, err)
	assert.Equal(t, uint64(goroutineCount+1), lastAccount.ID)
}

// TestRaceConditionInTransfer 測試轉帳中的競態條件
func TestRaceConditionInTransfer(t *testing.T) {
	storage := NewMemoryStorage()
	
	// 創建帳戶，剛好夠進行一次轉帳
	account1 := &model.Account{Name: "User 1", Balance: decimal.NewFromFloat(50.0)}
	account2 := &model.Account{Name: "User 2", Balance: decimal.NewFromFloat(0.0)}
	
	storage.CreateAccount(account1)
	storage.CreateAccount(account2)
	
	var wg sync.WaitGroup
	var successCount int64
	var mutex sync.Mutex
	
	// 嘗試多次轉帳相同金額，只有一次應該成功
	goroutineCount := 10
	wg.Add(goroutineCount)
	
	for i := 0; i < goroutineCount; i++ {
		go func() {
			defer wg.Done()
			err := storage.Transfer(account1.ID, account2.ID, decimal.NewFromFloat(50.0))
			if err == nil {
				mutex.Lock()
				successCount++
				mutex.Unlock()
			}
		}()
	}
	
	wg.Wait()
	
	// 只有一次轉帳應該成功
	assert.Equal(t, int64(1), successCount)
	
	// 驗證最終餘額
	finalAccount1, _ := storage.GetAccountByID(account1.ID)
	finalAccount2, _ := storage.GetAccountByID(account2.ID)
	
	assert.True(t, decimal.Zero.Equal(finalAccount1.Balance))
	assert.True(t, decimal.NewFromFloat(50.0).Equal(finalAccount2.Balance))
}

// BenchmarkConcurrentOperations 併發操作性能基準測試
func BenchmarkConcurrentOperations(b *testing.B) {
	storage := NewMemoryStorage()
	
	// 創建測試帳戶
	accountCount := 1000
	for i := 0; i < accountCount; i++ {
		account := &model.Account{
			Name:    "Benchmark User",
			Balance: decimal.NewFromFloat(1000.0),
		}
		storage.CreateAccount(account)
	}
	
	b.ResetTimer()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 隨機選擇操作類型
			switch b.N % 4 {
			case 0:
				// 查詢操作
				storage.GetAccountByID(uint64((b.N%accountCount)+1))
			case 1:
				// 存款操作
				storage.Deposit(uint64((b.N%accountCount)+1), decimal.NewFromFloat(1.0))
			case 2:
				// 提款操作
				storage.Withdraw(uint64((b.N%accountCount)+1), decimal.NewFromFloat(1.0))
			case 3:
				// 轉帳操作
				from := uint64((b.N%accountCount)+1)
				to := uint64(((b.N+1)%accountCount)+1)
				if from != to {
					storage.Transfer(from, to, decimal.NewFromFloat(1.0))
				}
			}
		}
	})
}