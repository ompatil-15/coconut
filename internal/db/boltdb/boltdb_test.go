package boltdb

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewBoltStore(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "boltdb-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")

	// Test creating new store
	store, err := NewBoltStore(dbPath)
	if err != nil {
		t.Fatalf("NewBoltStore failed: %v", err)
	}
	defer store.Close()

	if store == nil {
		t.Fatal("NewBoltStore returned nil")
	}

	// Verify file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestBoltStore_CreateBucket(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "boltdb-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	store, err := NewBoltStore(dbPath)
	if err != nil {
		t.Fatalf("NewBoltStore failed: %v", err)
	}
	defer store.Close()

	// Test creating bucket
	err = store.CreateBucket("test-bucket")
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	// Test creating same bucket again (should not error)
	err = store.CreateBucket("test-bucket")
	if err != nil {
		t.Fatalf("CreateBucket should not fail for existing bucket: %v", err)
	}
}

func TestBoltStore_PutGet(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "boltdb-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	store, err := NewBoltStore(dbPath)
	if err != nil {
		t.Fatalf("NewBoltStore failed: %v", err)
	}
	defer store.Close()

	bucket := "test-bucket"
	err = store.CreateBucket(bucket)
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	// Test put and get
	key := "test-key"
	value := []byte("test-value")

	err = store.Put(bucket, key, value)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	retrieved, err := store.Get(bucket, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != string(value) {
		t.Errorf("Expected '%s', got '%s'", string(value), string(retrieved))
	}
}

func TestBoltStore_GetNonExistent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "boltdb-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	store, err := NewBoltStore(dbPath)
	if err != nil {
		t.Fatalf("NewBoltStore failed: %v", err)
	}
	defer store.Close()

	bucket := "test-bucket"
	err = store.CreateBucket(bucket)
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	// Test getting non-existent key
	_, err = store.Get(bucket, "non-existent")
	if err == nil {
		t.Error("Get should fail for non-existent key")
	}
}

func TestBoltStore_Delete(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "boltdb-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	store, err := NewBoltStore(dbPath)
	if err != nil {
		t.Fatalf("NewBoltStore failed: %v", err)
	}
	defer store.Close()

	bucket := "test-bucket"
	err = store.CreateBucket(bucket)
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	key := "test-key"
	value := []byte("test-value")

	// Put value
	err = store.Put(bucket, key, value)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Verify it exists
	_, err = store.Get(bucket, key)
	if err != nil {
		t.Fatalf("Get failed before delete: %v", err)
	}

	// Delete it
	err = store.Delete(bucket, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	_, err = store.Get(bucket, key)
	if err == nil {
		t.Error("Get should fail after delete")
	}
}

func TestBoltStore_ListKeys(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "boltdb-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	store, err := NewBoltStore(dbPath)
	if err != nil {
		t.Fatalf("NewBoltStore failed: %v", err)
	}
	defer store.Close()

	bucket := "test-bucket"
	err = store.CreateBucket(bucket)
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	// Add multiple keys
	testKeys := []string{"key1", "key2", "key3"}
	for _, key := range testKeys {
		err = store.Put(bucket, key, []byte("value-"+key))
		if err != nil {
			t.Fatalf("Put failed for key %s: %v", key, err)
		}
	}

	// List keys
	keys, err := store.ListKeys(bucket)
	if err != nil {
		t.Fatalf("ListKeys failed: %v", err)
	}

	if len(keys) != len(testKeys) {
		t.Errorf("Expected %d keys, got %d", len(testKeys), len(keys))
	}

	// Verify all keys are present
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	for _, expectedKey := range testKeys {
		if !keyMap[expectedKey] {
			t.Errorf("Expected key '%s' not found in list", expectedKey)
		}
	}
}

func TestBoltStore_NonExistentBucket(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "boltdb-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	store, err := NewBoltStore(dbPath)
	if err != nil {
		t.Fatalf("NewBoltStore failed: %v", err)
	}
	defer store.Close()

	// Test operations on non-existent bucket
	_, err = store.Get("non-existent", "key")
	if err == nil {
		t.Error("Get should fail for non-existent bucket")
	}

	err = store.Put("non-existent", "key", []byte("value"))
	if err == nil {
		t.Error("Put should fail for non-existent bucket")
	}

	err = store.Delete("non-existent", "key")
	if err == nil {
		t.Error("Delete should fail for non-existent bucket")
	}

	_, err = store.ListKeys("non-existent")
	if err == nil {
		t.Error("ListKeys should fail for non-existent bucket")
	}
}

func TestBoltStore_Close(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "boltdb-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	store, err := NewBoltStore(dbPath)
	if err != nil {
		t.Fatalf("NewBoltStore failed: %v", err)
	}

	// Close the store
	err = store.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Operations should fail after close
	err = store.CreateBucket("test")
	if err == nil {
		t.Error("CreateBucket should fail after close")
	}
}

func TestBoltStore_ConcurrentAccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "boltdb-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	store, err := NewBoltStore(dbPath)
	if err != nil {
		t.Fatalf("NewBoltStore failed: %v", err)
	}
	defer store.Close()

	bucket := "test-bucket"
	err = store.CreateBucket(bucket)
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	// Test concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := sprintf("key-%d", id)
			value := []byte(sprintf("value-%d", id))
			err := store.Put(bucket, key, value)
			if err != nil {
				t.Errorf("Concurrent put failed for %s: %v", key, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all keys were written
	keys, err := store.ListKeys(bucket)
	if err != nil {
		t.Fatalf("ListKeys failed: %v", err)
	}

	if len(keys) != 10 {
		t.Errorf("Expected 10 keys, got %d", len(keys))
	}
}

// Simple sprintf implementation for tests
// Simple sprintf implementation for tests
func sprintf(format string, args ...interface{}) string {
	result := format
	for _, arg := range args {
		switch v := arg.(type) {
		case int:
			if v == 0 {
				result = strings.Replace(result, "%d", "0", 1)
			} else {
				digits := ""
				num := v
				if num < 0 {
					digits = "-"
					num = -num
				}
				for num > 0 {
					digits = string(rune('0'+num%10)) + digits
					num /= 10
				}
				result = strings.Replace(result, "%d", digits, 1)
			}
		case string:
			result = strings.Replace(result, "%s", v, 1)
		}
	}
	return result
}

