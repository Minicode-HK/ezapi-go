package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Minicode-HK/ezapi-go/core"
)

// Model with UUID-based ID
type Model struct {
	Id string `ez:"id:uuid"`

	FirstName string
	LastName  string
	Age       int
}

// User with auto-increment ID
type User struct {
	Id    int `ez:"id:auto_increment"`
	Email string
	Name  string
}

// Product with auto-increment ID and more fields
type Product struct {
	Id          int `ez:"id:auto_increment"`
	Name        string
	Price       float64
	InStock     bool
	Description string
}

// BlogPost with timestamps for testing auto-tracking
type BlogPost struct {
	Id        int `ez:"id:auto_increment"`
	Title     string
	Content   string
	CreatedAt time.Time `ez:"time:created_at"`
	UpdatedAt time.Time `ez:"time:updated_at"`
}

// ============================================
// ADD OPERATION TESTS
// ============================================

func TestAdd_Single_UUID(t *testing.T) {
	collection := core.NewCollection[string, Model]()

	model := &Model{
		FirstName: "John",
		LastName:  "Doe",
		Age:       30,
	}

	collection.Add(model)

	// Verify ID was generated
	if model.Id == "" {
		t.Fatal("Expected ID to be generated, got empty string")
	}

	// Verify stored in collection
	retrieved, err := collection.GetById(model.Id)
	if err != nil {
		t.Fatalf("Expected to find added model in collection: %v", err)
	}

	if retrieved.FirstName != "John" {
		t.Errorf("Expected FirstName 'John', got '%s'", retrieved.FirstName)
	}
}

func TestAdd_Multiple_AutoIncrement(t *testing.T) {
	collection := core.NewCollection[int, User]()

	users := []*User{
		{Email: "user1@example.com", Name: "User 1"},
		{Email: "user2@example.com", Name: "User 2"},
		{Email: "user3@example.com", Name: "User 3"},
	}

	for _, user := range users {
		collection.Add(user)
	}

	// Verify IDs are sequential
	if users[0].Id != 1 {
		t.Errorf("Expected first user ID to be 1, got %d", users[0].Id)
	}
	if users[1].Id != 2 {
		t.Errorf("Expected second user ID to be 2, got %d", users[1].Id)
	}
	if users[2].Id != 3 {
		t.Errorf("Expected third user ID to be 3, got %d", users[2].Id)
	}

	// Verify all are stored
	for _, user := range users {
		retrieved, err := collection.GetById(user.Id)
		if err != nil {
			t.Errorf("Expected to find user with ID %d: %v", user.Id, err)
		}
		if retrieved.Email != user.Email {
			t.Errorf("Expected email '%s', got '%s'", user.Email, retrieved.Email)
		}
	}
}

func TestAdd_MultipleTypes(t *testing.T) {
	collection := core.NewCollection[int, Product]()

	products := []*Product{
		{Name: "Laptop", Price: 999.99, InStock: true, Description: "High-performance laptop"},
		{Name: "Mouse", Price: 29.99, InStock: true, Description: "Wireless mouse"},
		{Name: "Keyboard", Price: 79.99, InStock: false, Description: "Mechanical keyboard"},
	}

	for _, p := range products {
		collection.Add(p)
	}

	if collection.Count() != 3 {
		t.Errorf("Expected count 3, got %d", collection.Count())
	}
}

// ============================================
// GET OPERATION TESTS
// ============================================

func TestGetById_Success(t *testing.T) {
	collection := core.NewCollection[int, User]()
	user := &User{Email: "test@example.com", Name: "Test User"}
	collection.Add(user)

	retrieved, err := collection.GetById(user.Id)
	if err != nil {
		t.Fatalf("Expected to find user: %v", err)
	}

	if retrieved.Email != user.Email {
		t.Errorf("Expected email '%s', got '%s'", user.Email, retrieved.Email)
	}

	// Verify it's a pointer to the same object
	if retrieved != user {
		t.Error("Expected GetById to return pointer to same object")
	}
}

func TestGetById_NotFound(t *testing.T) {
	collection := core.NewCollection[int, User]()

	_, err := collection.GetById(999)
	if err == nil {
		t.Error("Expected error when finding non-existent user")
	}
}

func TestGetByIdWithPointer_Success(t *testing.T) {
	collection := core.NewCollection[string, Model]()
	model := &Model{FirstName: "Jane", LastName: "Smith", Age: 25}
	collection.Add(model)

	retrieved, err := collection.GetByIdWithPointer(model.Id)
	if err != nil {
		t.Fatalf("Expected to find model: %v", err)
	}

	if retrieved.FirstName != "Jane" {
		t.Errorf("Expected FirstName 'Jane', got '%s'", retrieved.FirstName)
	}
}

func TestGetByIdWithCopy_Success(t *testing.T) {
	collection := core.NewCollection[int, User]()
	user := &User{Email: "copy@example.com", Name: "Copy User"}
	collection.Add(user)

	// Get a copy
	copy, err := collection.GetByIdWithCopy(user.Id)
	if err != nil {
		t.Fatalf("Expected to find user: %v", err)
	}

	// Modify the copy
	copy.Name = "Modified Name"

	// Verify original is unchanged
	original, _ := collection.GetById(user.Id)
	if original.Name == "Modified Name" {
		t.Error("Expected original to be unchanged when copy is modified")
	}
	if original.Name != "Copy User" {
		t.Errorf("Expected original name 'Copy User', got '%s'", original.Name)
	}
}

func TestGetByIdWithCopy_NotFound(t *testing.T) {
	collection := core.NewCollection[int, User]()

	_, err := collection.GetByIdWithCopy(999)
	if err == nil {
		t.Error("Expected error when finding non-existent user")
	}
}

// ============================================
// UPDATE OPERATION TESTS
// ============================================

func TestUpdate_EntireElement_Success(t *testing.T) {
	collection := core.NewCollection[int, User]()
	user := &User{Email: "old@example.com", Name: "Old Name"}
	collection.Add(user)

	err := collection.Update(user.Id, User{
		Email: "new@example.com",
		Name:  "New Name",
	})

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	updated, err := collection.GetById(user.Id)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}
	if updated.Email != "new@example.com" {
		t.Errorf("Expected email 'new@example.com', got '%s'", updated.Email)
	}
	if updated.Name != "New Name" {
		t.Errorf("Expected name 'New Name', got '%s'", updated.Name)
	}

	// Verify ID unchanged
	if updated.Id != user.Id {
		t.Errorf("Expected ID to remain %d, got %d", user.Id, updated.Id)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	collection := core.NewCollection[int, User]()

	err := collection.Update(999, User{Email: "test@example.com", Name: "Test"})
	if err == nil {
		t.Error("Expected error when updating non-existent element")
	}
}

func TestUpdateWithAttributes_PartialUpdate_Success(t *testing.T) {
	collection := core.NewCollection[int, Product]()
	product := &Product{
		Name:        "Original",
		Price:       100.0,
		InStock:     true,
		Description: "Original description",
	}
	collection.Add(product)

	// Update only some fields
	err := collection.UpdateWithAttributes(product.Id, map[string]any{
		"Name":  "Updated Name",
		"Price": 150.0,
	})

	if err != nil {
		t.Fatalf("UpdateWithAttributes failed: %v", err)
	}

	// Verify updated fields
	updated, err := collection.GetById(product.Id)
	if err != nil {
		t.Fatalf("Failed to get updated product: %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("Expected Name 'Updated Name', got '%s'", updated.Name)
	}
	if updated.Price != 150.0 {
		t.Errorf("Expected Price 150.0, got %f", updated.Price)
	}

	// Verify unchanged fields
	if !updated.InStock {
		t.Error("Expected InStock to remain true")
	}
	if updated.Description != "Original description" {
		t.Errorf("Expected Description unchanged, got '%s'", updated.Description)
	}
}

func TestUpdateWithAttributes_NotFound(t *testing.T) {
	collection := core.NewCollection[int, User]()

	err := collection.UpdateWithAttributes(999, map[string]any{"Name": "Test"})
	if err == nil {
		t.Error("Expected error when updating non-existent element")
	}
}

func TestUpdateWithAttributes_InvalidFieldName(t *testing.T) {
	collection := core.NewCollection[int, User]()
	user := &User{Email: "test@example.com", Name: "Test"}
	collection.Add(user)

	err := collection.UpdateWithAttributes(user.Id, map[string]any{
		"NonExistentField": "value",
	})

	if err == nil {
		t.Error("Expected error when updating non-existent field")
	}
}

func TestUpdateWithAttributes_TypeMismatch(t *testing.T) {
	collection := core.NewCollection[int, Product]()
	product := &Product{Name: "Test", Price: 100.0}
	collection.Add(product)

	// Try to set Price (float64) with a string
	err := collection.UpdateWithAttributes(product.Id, map[string]any{
		"Price": "not a number",
	})

	if err == nil {
		t.Error("Expected error when type doesn't match")
	}
}

func TestUpdateWithAttributes_IDField_Error(t *testing.T) {
	collection := core.NewCollection[int, User]()
	user := &User{Email: "test@example.com", Name: "Test"}
	collection.Add(user)
	originalId := user.Id

	// Try to update ID field - should return error
	err := collection.UpdateWithAttributes(user.Id, map[string]any{
		"Id": 999,
	})

	if err == nil {
		t.Error("Expected error when attempting to update ID field")
	}

	// Verify ID did not change
	updated, err := collection.GetById(originalId)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}
	if updated.Id != originalId {
		t.Errorf("Expected ID to remain %d, got %d", originalId, updated.Id)
	}
}

func TestUpdateWithAttributes_SkipsIDButUpdatesOthers(t *testing.T) {
	collection := core.NewCollection[int, User]()
	user := &User{Email: "test@example.com", Name: "Test"}
	collection.Add(user)
	originalId := user.Id

	// Update other fields (ID in map should cause error)
	err := collection.UpdateWithAttributes(user.Id, map[string]any{
		"Name": "Updated Name",
	})

	if err != nil {
		t.Fatalf("UpdateWithAttributes failed: %v", err)
	}

	// Verify ID unchanged
	updated, err := collection.GetById(originalId)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}
	if updated.Id != originalId {
		t.Errorf("Expected ID to remain %d, got %d", originalId, updated.Id)
	}

	// Verify other fields updated
	if updated.Name != "Updated Name" {
		t.Errorf("Expected Name 'Updated Name', got '%s'", updated.Name)
	}
}

// ============================================
// REMOVE OPERATION TESTS
// ============================================

func TestRemove_Success(t *testing.T) {
	collection := core.NewCollection[int, User]()
	user := &User{Email: "delete@example.com", Name: "Delete Me"}
	collection.Add(user)

	// Verify exists
	if _, err := collection.GetById(user.Id); err != nil {
		t.Fatalf("Expected user to exist before removal: %v", err)
	}

	// Remove
	collection.Remove(user.Id)

	// Verify removed
	if _, err := collection.GetById(user.Id); err == nil {
		t.Error("Expected user to be removed")
	}
}

func TestRemove_NotFound(t *testing.T) {
	collection := core.NewCollection[int, User]()

	// Should not panic when removing non-existent element
	collection.Remove(999)

	// Collection should remain functional
	user := &User{Email: "test@example.com", Name: "Test"}
	collection.Add(user)
	if _, err := collection.GetById(user.Id); err != nil {
		t.Errorf("Collection should still work after removing non-existent element: %v", err)
	}
}

func TestRemove_Multiple(t *testing.T) {
	collection := core.NewCollection[int, User]()

	users := []*User{
		{Email: "user1@example.com", Name: "User 1"},
		{Email: "user2@example.com", Name: "User 2"},
		{Email: "user3@example.com", Name: "User 3"},
	}

	for _, u := range users {
		collection.Add(u)
	}

	// Remove middle element
	collection.Remove(users[1].Id)

	// Verify first and third still exist
	if _, err := collection.GetById(users[0].Id); err != nil {
		t.Errorf("Expected first user to still exist: %v", err)
	}
	if _, err := collection.GetById(users[2].Id); err != nil {
		t.Errorf("Expected third user to still exist: %v", err)
	}

	// Verify second is removed
	if _, err := collection.GetById(users[1].Id); err == nil {
		t.Error("Expected second user to be removed")
	}
}

// ============================================
// COUNT OPERATION TESTS
// ============================================

func TestCount_Empty(t *testing.T) {
	collection := core.NewCollection[int, User]()

	if collection.Count() != 0 {
		t.Errorf("Expected count 0, got %d", collection.Count())
	}
}

func TestCount_AfterAdds(t *testing.T) {
	collection := core.NewCollection[int, User]()

	for i := 0; i < 5; i++ {
		collection.Add(&User{Email: "test@example.com", Name: "Test"})
	}

	if collection.Count() != 5 {
		t.Errorf("Expected count 5, got %d", collection.Count())
	}
}

// ============================================
// CONCURRENT ACCESS TESTS
// ============================================

func TestConcurrentAdd(t *testing.T) {
	collection := core.NewCollection[int, User]()

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			user := &User{
				Email: "concurrent" + string(rune(index)) + "@example.com",
				Name:  "Concurrent User",
			}
			collection.Add(user)
		}(i)
	}

	wg.Wait()

	if collection.Count() != numGoroutines {
		t.Errorf("Expected count %d, got %d", numGoroutines, collection.Count())
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	collection := core.NewCollection[int, User]()

	// Add initial data
	for i := 0; i < 10; i++ {
		collection.Add(&User{Email: "user@example.com", Name: "User"})
	}

	var wg sync.WaitGroup

	// Concurrent readers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 1; j <= 10; j++ {
				collection.GetById(j) // Ignore errors in concurrent read test
			}
		}()
	}

	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			collection.Add(&User{Email: "new@example.com", Name: "New"})
		}()
	}

	wg.Wait()
	// Should not panic or deadlock
}

// ============================================
// EDGE CASE TESTS
// ============================================

func TestEdgeCase_EmptyStrings(t *testing.T) {
	collection := core.NewCollection[string, Model]()

	model := &Model{FirstName: "", LastName: "", Age: 0}
	collection.Add(model)

	retrieved, err := collection.GetById(model.Id)
	if err != nil {
		t.Fatalf("Expected to find model with empty strings: %v", err)
	}

	if retrieved.FirstName != "" {
		t.Error("Expected empty FirstName")
	}
}

func TestEdgeCase_ZeroValues(t *testing.T) {
	collection := core.NewCollection[int, Product]()

	product := &Product{Name: "Free Item", Price: 0.0, InStock: false}
	collection.Add(product)

	retrieved, err := collection.GetById(product.Id)
	if err != nil {
		t.Fatalf("Expected to find product with zero values: %v", err)
	}

	if retrieved.Price != 0.0 {
		t.Errorf("Expected Price 0.0, got %f", retrieved.Price)
	}
}

func TestEdgeCase_UpdateToZeroValues(t *testing.T) {
	collection := core.NewCollection[int, Product]()
	product := &Product{Name: "Item", Price: 100.0, InStock: true}
	collection.Add(product)

	err := collection.UpdateWithAttributes(product.Id, map[string]any{
		"Price":   0.0,
		"InStock": false,
	})

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	updated, err := collection.GetById(product.Id)
	if err != nil {
		t.Fatalf("Failed to get updated product: %v", err)
	}
	if updated.Price != 0.0 {
		t.Errorf("Expected Price 0.0, got %f", updated.Price)
	}
	if updated.InStock {
		t.Error("Expected InStock false")
	}
}

func TestUpdate_DeepCopyWithSlices(t *testing.T) {
	type ArticleWithTags struct {
		Id   int `ez:"id:auto_increment"`
		Name string
		Tags []string
	}

	collection := core.NewCollection[int, ArticleWithTags]()
	article := &ArticleWithTags{Name: "Original", Tags: []string{"go", "db"}}
	collection.Add(article)

	// Update with slices
	source := ArticleWithTags{
		Name: "Updated",
		Tags: []string{"golang", "database"},
	}
	collection.Update(article.Id, source)

	// Modify source slice after update
	source.Tags[0] = "rust"
	source.Tags[1] = "distributed"

	// Verify collection data is NOT affected (deep copy)
	updated, err := collection.GetById(article.Id)
	if err != nil {
		t.Fatalf("Failed to get article: %v", err)
	}

	if updated.Tags[0] != "golang" {
		t.Errorf("Expected 'golang', got '%s' - shallow copy detected!", updated.Tags[0])
	}
	if updated.Tags[1] != "database" {
		t.Errorf("Expected 'database', got '%s' - shallow copy detected!", updated.Tags[1])
	}
}

func TestUpdate_DeepCopyWithMaps(t *testing.T) {
	type Config struct {
		Id    int `ez:"id:auto_increment"`
		Name  string
		Props map[string]string
	}

	collection := core.NewCollection[int, Config]()
	config := &Config{Name: "Config1", Props: map[string]string{"env": "prod"}}
	collection.Add(config)

	// Update with map
	source := Config{
		Name:  "Updated",
		Props: map[string]string{"env": "staging", "version": "1.0"},
	}
	collection.Update(config.Id, source)

	// Modify source map after update
	source.Props["env"] = "dev"
	source.Props["version"] = "2.0"

	// Verify collection data is NOT affected
	updated, err := collection.GetById(config.Id)
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	if updated.Props["env"] != "staging" {
		t.Errorf("Expected 'staging', got '%s'", updated.Props["env"])
	}
	if updated.Props["version"] != "1.0" {
		t.Errorf("Expected '1.0', got '%s'", updated.Props["version"])
	}
}

func TestGetByIdWithCopy_DeepCopyWithSlices(t *testing.T) {
	type ArticleWithTags struct {
		Id   int `ez:"id:auto_increment"`
		Name string
		Tags []string
	}

	collection := core.NewCollection[int, ArticleWithTags]()
	article := &ArticleWithTags{Name: "Original", Tags: []string{"go", "db"}}
	collection.Add(article)

	// Get a copy
	copy, err := collection.GetByIdWithCopy(article.Id)
	if err != nil {
		t.Fatalf("Expected to find article: %v", err)
	}

	// Modify the copy's slice
	copy.Tags[0] = "rust"
	copy.Tags[1] = "distributed"

	// Verify original is NOT affected (true deep copy)
	original, _ := collection.GetById(article.Id)
	if original.Tags[0] != "go" {
		t.Errorf("Expected original Tags[0] to be 'go', got '%s' - shallow copy!", original.Tags[0])
	}
	if original.Tags[1] != "db" {
		t.Errorf("Expected original Tags[1] to be 'db', got '%s' - shallow copy!", original.Tags[1])
	}
}

func TestGetByIdWithCopy_DeepCopyWithMaps(t *testing.T) {
	type ConfigWithProps struct {
		Id    int `ez:"id:auto_increment"`
		Name  string
		Props map[string]string
	}

	collection := core.NewCollection[int, ConfigWithProps]()
	config := &ConfigWithProps{
		Name:  "Config1",
		Props: map[string]string{"env": "prod", "version": "1.0"},
	}
	collection.Add(config)

	// Get a copy
	copy, err := collection.GetByIdWithCopy(config.Id)
	if err != nil {
		t.Fatalf("Expected to find config: %v", err)
	}

	// Modify the copy's map
	copy.Props["env"] = "dev"
	copy.Props["version"] = "2.0"

	// Verify original is NOT affected
	original, _ := collection.GetById(config.Id)
	if original.Props["env"] != "prod" {
		t.Errorf("Expected original Props[env] to be 'prod', got '%s' - shallow copy!", original.Props["env"])
	}
	if original.Props["version"] != "1.0" {
		t.Errorf("Expected original Props[version] to be '1.0', got '%s' - shallow copy!", original.Props["version"])
	}
}

// ============================================
// SNAPSHOT TESTS
// ============================================

func TestSnapshot_BasicSnapshot(t *testing.T) {
	collection := core.NewCollection[int, User]()

	user1 := &User{Email: "user1@example.com", Name: "User 1"}
	user2 := &User{Email: "user2@example.com", Name: "User 2"}
	collection.Add(user1)
	collection.Add(user2)

	// Take snapshot
	snapshot := collection.SnapShot()

	if snapshot == nil {
		t.Fatal("Expected snapshot to be created")
	}

	// Snapshot is successfully created

	// Verify snapshot count matches collection
	if len(snapshot.Data) != collection.Count() {
		t.Errorf("Expected snapshot size %d, got %d", collection.Count(), len(snapshot.Data))
	}
}

func TestSnapshot_SnapshotIsolation(t *testing.T) {
	collection := core.NewCollection[int, User]()

	user1 := &User{Email: "user1@example.com", Name: "User 1"}
	collection.Add(user1)

	// Take snapshot
	snapshot := collection.SnapShot()

	// Modify collection after snapshot
	collection.Add(&User{Email: "user3@example.com", Name: "User 3"})
	collection.Update(user1.Id, User{Email: "updated@example.com", Name: "Updated"})

	// Snapshot should remain unchanged
	if len(snapshot.Data) != 1 {
		t.Errorf("Expected snapshot size 1, got %d", len(snapshot.Data))
	}

	// Verify collection was modified
	if collection.Count() != 2 {
		t.Errorf("Expected collection size 2, got %d", collection.Count())
	}
}

func TestSnapshot_RestoreBasic(t *testing.T) {
	collection := core.NewCollection[int, User]()

	user1 := &User{Email: "user1@example.com", Name: "User 1"}
	user2 := &User{Email: "user2@example.com", Name: "User 2"}
	collection.Add(user1)
	collection.Add(user2)

	originalCount := collection.Count()

	// Take snapshot
	snapshot := collection.SnapShot()

	// Modify collection
	collection.Remove(user1.Id)
	collection.Add(&User{Email: "user3@example.com", Name: "User 3"})

	if collection.Count() != originalCount {
		t.Logf("Collection modified - count changed to %d", collection.Count())
	}

	// Restore from snapshot
	collection.Restore(snapshot)

	// Verify collection restored
	if collection.Count() != originalCount {
		t.Errorf("Expected count %d after restore, got %d", originalCount, collection.Count())
	}

	// Verify specific items exist
	if _, err := collection.GetById(user1.Id); err != nil {
		t.Error("Expected user1 to exist after restore")
	}
	if _, err := collection.GetById(user2.Id); err != nil {
		t.Error("Expected user2 to exist after restore")
	}
}

func TestSnapshot_RestoreWithSlices(t *testing.T) {
	type ArticleWithTags struct {
		Id   int `ez:"id:auto_increment"`
		Name string
		Tags []string
	}

	collection := core.NewCollection[int, ArticleWithTags]()

	article := &ArticleWithTags{Name: "Original", Tags: []string{"go", "db"}}
	collection.Add(article)

	// Take snapshot
	snapshot := collection.SnapShot()

	// Modify article in collection
	collection.Update(article.Id, ArticleWithTags{
		Name: "Updated",
		Tags: []string{"rust", "distributed"},
	})

	// Restore
	collection.Restore(snapshot)

	// Verify restored state (slices should be deep copied)
	restored, _ := collection.GetById(article.Id)
	if restored.Name != "Original" {
		t.Errorf("Expected Name 'Original', got '%s'", restored.Name)
	}
	if len(restored.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(restored.Tags))
	}
	if restored.Tags[0] != "go" || restored.Tags[1] != "db" {
		t.Errorf("Expected tags ['go', 'db'], got %v", restored.Tags)
	}
}

func TestSnapshot_RestoreWithMaps(t *testing.T) {
	type Config struct {
		Id    int `ez:"id:auto_increment"`
		Name  string
		Props map[string]string
	}

	collection := core.NewCollection[int, Config]()

	config := &Config{
		Name:  "Config1",
		Props: map[string]string{"env": "prod", "version": "1.0"},
	}
	collection.Add(config)

	// Take snapshot
	snapshot := collection.SnapShot()

	// Modify config in collection
	collection.Update(config.Id, Config{
		Name:  "Updated",
		Props: map[string]string{"env": "dev", "version": "2.0"},
	})

	// Restore
	collection.Restore(snapshot)

	// Verify restored state (maps should be deep copied)
	restored, _ := collection.GetById(config.Id)
	if restored.Props["env"] != "prod" {
		t.Errorf("Expected env 'prod', got '%s'", restored.Props["env"])
	}
	if restored.Props["version"] != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", restored.Props["version"])
	}
}

func TestSnapshot_MultipleSnapshots(t *testing.T) {
	collection := core.NewCollection[int, User]()

	user1 := &User{Email: "user1@example.com", Name: "User 1"}
	collection.Add(user1)

	// Snapshot 1: 1 user
	snapshot1 := collection.SnapShot()

	collection.Add(&User{Email: "user2@example.com", Name: "User 2"})

	// Snapshot 2: 2 users
	snapshot2 := collection.SnapShot()

	collection.Add(&User{Email: "user3@example.com", Name: "User 3"})

	// Collection now has 3 users
	if collection.Count() != 3 {
		t.Errorf("Expected 3 users, got %d", collection.Count())
	}

	// Restore to snapshot2 (2 users)
	collection.Restore(snapshot2)
	if collection.Count() != 2 {
		t.Errorf("Expected 2 users after restore to snapshot2, got %d", collection.Count())
	}

	// Restore to snapshot1 (1 user)
	collection.Restore(snapshot1)
	if collection.Count() != 1 {
		t.Errorf("Expected 1 user after restore to snapshot1, got %d", collection.Count())
	}
}

func TestSnapshot_SnapshotPreservesData(t *testing.T) {
	collection := core.NewCollection[int, Product]()

	product := &Product{
		Name:        "Laptop",
		Price:       999.99,
		InStock:     true,
		Description: "High-performance laptop",
	}
	collection.Add(product)

	// Take snapshot
	snapshot := collection.SnapShot()

	// Modify original in collection
	collection.Update(product.Id, Product{
		Name:        "Updated Laptop",
		Price:       1299.99,
		InStock:     false,
		Description: "Updated description",
	})

	// Restore
	collection.Restore(snapshot)

	// Verify all fields restored correctly
	restored, _ := collection.GetById(product.Id)
	if restored.Name != "Laptop" {
		t.Errorf("Expected Name 'Laptop', got '%s'", restored.Name)
	}
	if restored.Price != 999.99 {
		t.Errorf("Expected Price 999.99, got %f", restored.Price)
	}
	if !restored.InStock {
		t.Error("Expected InStock to be true")
	}
	if restored.Description != "High-performance laptop" {
		t.Errorf("Expected original description, got '%s'", restored.Description)
	}
}

func TestSnapshot_RestoreNilSnapshot(t *testing.T) {
	collection := core.NewCollection[int, User]()
	collection.Add(&User{Email: "test@example.com", Name: "Test"})

	// Restore with nil snapshot - should handle gracefully
	collection.Restore(nil)

	// Collection should still be functional
	if collection.Count() != 1 {
		t.Error("Collection should still contain data after restore(nil)")
	}
}

func TestSnapshot_SnapshotAfterRemoval(t *testing.T) {
	collection := core.NewCollection[int, User]()

	user1 := &User{Email: "user1@example.com", Name: "User 1"}
	user2 := &User{Email: "user2@example.com", Name: "User 2"}
	user3 := &User{Email: "user3@example.com", Name: "User 3"}

	collection.Add(user1)
	collection.Add(user2)
	collection.Add(user3)

	// Take snapshot with 3 users
	snapshot := collection.SnapShot()

	// Remove a user
	collection.Remove(user2.Id)

	// Collection has 2 users
	if collection.Count() != 2 {
		t.Errorf("Expected 2 users after removal, got %d", collection.Count())
	}

	// Restore to snapshot
	collection.Restore(snapshot)

	// All 3 users should be back
	if collection.Count() != 3 {
		t.Errorf("Expected 3 users after restore, got %d", collection.Count())
	}

	// Verify all users exist
	for _, user := range []*User{user1, user2, user3} {
		if _, err := collection.GetById(user.Id); err != nil {
			t.Errorf("Expected user %s to exist after restore", user.Email)
		}
	}
}

func TestSnapshot_DeepCopyIsolation(t *testing.T) {
	type Article struct {
		Id   int `ez:"id:auto_increment"`
		Name string
		Tags []string
	}

	collection := core.NewCollection[int, Article]()

	article := &Article{Name: "Original", Tags: []string{"go", "db"}}
	collection.Add(article)

	// Take snapshot before any modifications
	snapshot := collection.SnapShot()

	// Verify snapshot content is correct before modifications
	snapshotData := snapshot.Data
	if len(snapshotData) != 1 {
		t.Errorf("Expected snapshot to contain 1 item, got %d", len(snapshotData))
	}

	// Modify collection significantly
	collection.Update(article.Id, Article{Name: "Updated", Tags: []string{"python", "ai"}})
	collection.Add(&Article{Name: "New Article", Tags: []string{"rust"}})

	// Collection should be different now
	if collection.Count() != 2 {
		t.Errorf("Expected collection to have 2 items after modifications, got %d", collection.Count())
	}

	// Restore from snapshot
	collection.Restore(snapshot)

	// Verify snapshot restored correctly
	if collection.Count() != 1 {
		t.Errorf("Expected 1 item after restore, got %d", collection.Count())
	}

	restored, _ := collection.GetById(article.Id)
	if restored.Name != "Original" {
		t.Errorf("Expected Name 'Original', got '%s'", restored.Name)
	}
	if restored.Tags[0] != "go" {
		t.Errorf("Expected 'go', got '%s' after restore", restored.Tags[0])
	}
	if restored.Tags[1] != "db" {
		t.Errorf("Expected 'db', got '%s' after restore", restored.Tags[1])
	}
}

func TestSnapshot_SaveAsJson(t *testing.T) {
	type Article struct {
		Id   int `ez:"id:auto_increment"`
		Name string
		Tags []string
	}

	collection := core.NewCollection[int, Article]()

	// Add test data
	article1 := &Article{Name: "First Article", Tags: []string{"golang", "testing"}}
	article2 := &Article{Name: "Second Article", Tags: []string{"database"}}

	collection.Add(article1)
	collection.Add(article2)

	// Create snapshot
	snapshot := collection.SnapShot()

	// Create temp file for testing
	tempFile := "/tmp/test_snapshot.json"
	defer os.Remove(tempFile) // Cleanup

	// Save snapshot as JSON
	err := snapshot.SaveAsJson(tempFile)
	if err != nil {
		t.Fatalf("Failed to save snapshot as JSON: %v", err)
	}

	// Verify file exists
	_, err = os.Stat(tempFile)
	if err != nil {
		t.Fatalf("JSON file was not created: %v", err)
	}

	// Read and verify JSON content
	data, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}

	// Unmarshal JSON to verify structure
	var savedSnapshot map[string]interface{}
	err = json.Unmarshal(data, &savedSnapshot)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify metadata exists
	if name, ok := savedSnapshot["name"]; !ok || name == "" {
		t.Error("Snapshot name metadata missing or empty")
	}

	if _, ok := savedSnapshot["time"]; !ok {
		t.Error("Snapshot timestamp metadata missing")
	}

	if counter, ok := savedSnapshot["counter"]; !ok {
		t.Error("Snapshot counter metadata missing")
	} else if c, ok := counter.(float64); ok && c != 2 {
		t.Errorf("Expected counter 2, got %v", c)
	}

	// Verify data exists and has correct items
	if savedData, ok := savedSnapshot["data"].(map[string]interface{}); !ok {
		t.Error("Data field missing or not a map")
	} else {
		if len(savedData) != 2 {
			t.Errorf("Expected 2 items in saved data, got %d", len(savedData))
		}

		// Verify first article exists
		if article1Data, ok := savedData["1"]; ok {
			if article1Map, ok := article1Data.(map[string]interface{}); ok {
				if name, ok := article1Map["Name"].(string); !ok || name != "First Article" {
					t.Errorf("Expected article 1 name 'First Article', got %v", name)
				}
			}
		} else {
			t.Error("Article with ID 1 not found in saved data")
		}

		// Verify second article exists
		if article2Data, ok := savedData["2"]; ok {
			if article2Map, ok := article2Data.(map[string]interface{}); ok {
				if name, ok := article2Map["Name"].(string); !ok || name != "Second Article" {
					t.Errorf("Expected article 2 name 'Second Article', got %v", name)
				}
			}
		} else {
			t.Error("Article with ID 2 not found in saved data")
		}
	}
}

func TestSnapshot_LoadFromJson(t *testing.T) {
	type Article struct {
		Id   int `ez:"id:auto_increment"`
		Name string
		Tags []string
	}

	collection := core.NewCollection[int, Article]()

	// Add test data
	article1 := &Article{Name: "First Article", Tags: []string{"golang", "testing"}}
	article2 := &Article{Name: "Second Article", Tags: []string{"database"}}

	collection.Add(article1)
	collection.Add(article2)

	// Create and save snapshot
	snapshot := collection.SnapShot()
	tempFile := "/tmp/test_snapshot_load.json"
	defer os.Remove(tempFile) // Cleanup

	// log the content of the json
	t.Logf("Snapshot JSON content:\n%s", func() string {
		data, err := json.MarshalIndent(snapshot, "", "  ")
		if err != nil {
			return fmt.Sprintf("Failed to marshal snapshot for logging: %v", err)
		}
		return string(data)
	}())

	err := snapshot.SaveAsJson(tempFile)
	if err != nil {
		t.Fatalf("Failed to save snapshot: %v", err)
	}

	// Clear collection to test loading
	collection.Clear()
	if collection.Count() != 0 {
		t.Errorf("Expected collection to be empty after Clear(), got %d items", collection.Count())
	}

	// Load snapshot from JSON into a new snapshot
	loadedSnapshot := &core.SnapShot[int, Article]{}
	err = loadedSnapshot.LoadFromJson(tempFile)
	if err != nil {
		t.Fatalf("Failed to load snapshot from JSON: %v", err)
	}

	// Verify loaded metadata
	if loadedSnapshot.Name != snapshot.Name {
		t.Errorf("Expected name '%s', got '%s'", snapshot.Name, loadedSnapshot.Name)
	}

	loadedData := loadedSnapshot.Data
	if len(loadedData) != 2 {
		t.Errorf("Expected 2 items in loaded data, got %d", len(loadedData))
	}

	// Verify first article
	if val, ok := loadedData[1]; ok {
		if val.Name != "First Article" {
			t.Errorf("Expected name 'First Article', got '%s'", val.Name)
		}
		if len(val.Tags) != 2 || val.Tags[0] != "golang" {
			t.Errorf("Expected tags ['golang', 'testing'], got %v", val.Tags)
		}
	} else {
		t.Error("Article with ID 1 not found in loaded data")
	}

	// Verify second article
	if val, ok := loadedData[2]; ok {
		if val.Name != "Second Article" {
			t.Errorf("Expected name 'Second Article', got '%s'", val.Name)
		}
		if len(val.Tags) != 1 || val.Tags[0] != "database" {
			t.Errorf("Expected tags ['database'], got %v", val.Tags)
		}
	} else {
		t.Error("Article with ID 2 not found in loaded data")
	}

	// Verify counter was restored
	if loadedSnapshot.Counter != snapshot.Counter {
		t.Errorf("Expected counter %d, got %d", snapshot.Counter, loadedSnapshot.Counter)
	}
}

// ============================================
// TIMESTAMP AUTO-TRACKING TESTS
// ============================================

func TestTimestamp_CreatedAtOnAdd(t *testing.T) {
	collection := core.NewCollection[int, BlogPost]()

	post := &BlogPost{
		Title:   "Test Post",
		Content: "Test Content",
	}

	beforeAdd := time.Now()
	err := collection.Add(post)
	afterAdd := time.Now()

	if err != nil {
		t.Fatalf("Failed to add post: %v", err)
	}

	// Verify CreatedAt was set
	if post.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set after Add()")
	}

	// Verify CreatedAt is within the expected time range
	if post.CreatedAt.Before(beforeAdd) || post.CreatedAt.After(afterAdd.Add(time.Second)) {
		t.Errorf("CreatedAt %v is not within expected time range [%v, %v]", post.CreatedAt, beforeAdd, afterAdd)
	}
}

func TestTimestamp_UpdatedAtOnAdd(t *testing.T) {
	collection := core.NewCollection[int, BlogPost]()

	post := &BlogPost{
		Title:   "Test Post",
		Content: "Test Content",
	}

	err := collection.Add(post)
	if err != nil {
		t.Fatalf("Failed to add post: %v", err)
	}

	// Verify UpdatedAt was set (should be same as CreatedAt on Add)
	if post.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set after Add()")
	}

	// Check if timestamps are approximately equal (within 1 millisecond)
	diff := post.UpdatedAt.Sub(post.CreatedAt)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Millisecond {
		t.Errorf("On Add(), CreatedAt and UpdatedAt should be approximately equal. Diff: %v", diff)
	}
}

func TestTimestamp_UpdatedAtOnUpdate(t *testing.T) {
	collection := core.NewCollection[int, BlogPost]()

	post := &BlogPost{
		Title:   "Test Post",
		Content: "Test Content",
	}

	err := collection.Add(post)
	if err != nil {
		t.Fatalf("Failed to add post: %v", err)
	}

	originalCreatedAt := post.CreatedAt
	originalUpdatedAt := post.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Update post
	updatedPost := &BlogPost{
		Title:   "Updated Title",
		Content: "Updated Content",
	}

	beforeUpdate := time.Now()
	err = collection.Update(post.Id, *updatedPost)
	afterUpdate := time.Now()

	if err != nil {
		t.Fatalf("Failed to update post: %v", err)
	}

	// Retrieve to check updated values
	retrieved, err := collection.GetById(post.Id)
	if err != nil {
		t.Fatalf("Failed to retrieve post: %v", err)
	}

	// Verify CreatedAt didn't change
	if retrieved.CreatedAt != originalCreatedAt {
		t.Errorf("CreatedAt should not change on Update(). Before: %v, After: %v", originalCreatedAt, retrieved.CreatedAt)
	}

	// Verify UpdatedAt changed
	if retrieved.UpdatedAt == originalUpdatedAt {
		t.Error("UpdatedAt should change on Update()")
	}

	// Verify UpdatedAt is after original
	if !retrieved.UpdatedAt.After(originalUpdatedAt) {
		t.Errorf("UpdatedAt should be later than original. Original: %v, New: %v", originalUpdatedAt, retrieved.UpdatedAt)
	}

	// Verify UpdatedAt is within expected range
	if retrieved.UpdatedAt.Before(beforeUpdate) || retrieved.UpdatedAt.After(afterUpdate.Add(time.Second)) {
		t.Errorf("UpdatedAt %v is not within expected time range [%v, %v]", retrieved.UpdatedAt, beforeUpdate, afterUpdate)
	}
}

func TestTimestamp_MultiplePosts(t *testing.T) {
	collection := core.NewCollection[int, BlogPost]()

	post1 := &BlogPost{Title: "Post 1", Content: "Content 1"}
	post2 := &BlogPost{Title: "Post 2", Content: "Content 2"}
	post3 := &BlogPost{Title: "Post 3", Content: "Content 3"}

	collection.Add(post1)
	time.Sleep(5 * time.Millisecond)

	collection.Add(post2)
	time.Sleep(5 * time.Millisecond)

	collection.Add(post3)

	// Verify all have different CreatedAt (or at least ordered correctly)
	if !post1.CreatedAt.Before(post2.CreatedAt) {
		t.Errorf("Post1 CreatedAt should be before Post2. P1: %v, P2: %v", post1.CreatedAt, post2.CreatedAt)
	}

	if !post2.CreatedAt.Before(post3.CreatedAt) {
		t.Errorf("Post2 CreatedAt should be before Post3. P2: %v, P3: %v", post2.CreatedAt, post3.CreatedAt)
	}

	// All should have approximately equal CreatedAt and UpdatedAt per post (no updates yet)
	// Check within 1 millisecond tolerance
	for i, post := range []*BlogPost{post1, post2, post3} {
		diff := post.UpdatedAt.Sub(post.CreatedAt)
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Millisecond {
			t.Errorf("Post %d: CreatedAt and UpdatedAt should be approximately equal before any updates. Diff: %v", i+1, diff)
		}
	}
}

func TestTimestamp_UpdateWithAttributes(t *testing.T) {
	collection := core.NewCollection[int, BlogPost]()

	post := &BlogPost{
		Title:   "Original Title",
		Content: "Original Content",
	}

	err := collection.Add(post)
	if err != nil {
		t.Fatalf("Failed to add post: %v", err)
	}

	originalCreatedAt := post.CreatedAt
	originalUpdatedAt := post.UpdatedAt

	time.Sleep(10 * time.Millisecond)

	// Update only Title using UpdateWithAttributes
	err = collection.UpdateWithAttributes(post.Id, map[string]interface{}{
		"Title": "New Title",
	})

	if err != nil {
		t.Fatalf("Failed to update post attributes: %v", err)
	}

	retrieved, err := collection.GetById(post.Id)
	if err != nil {
		t.Fatalf("Failed to retrieve post: %v", err)
	}

	// Verify Title was updated
	if retrieved.Title != "New Title" {
		t.Errorf("Expected title 'New Title', got '%s'", retrieved.Title)
	}

	// Verify CreatedAt didn't change
	if retrieved.CreatedAt != originalCreatedAt {
		t.Errorf("CreatedAt should not change on UpdateWithAttributes()")
	}

	// Verify UpdatedAt changed
	if retrieved.UpdatedAt == originalUpdatedAt {
		t.Error("UpdatedAt should change on UpdateWithAttributes()")
	}

	if !retrieved.UpdatedAt.After(originalUpdatedAt) {
		t.Errorf("UpdatedAt should be later after UpdateWithAttributes()")
	}
}

// ============================================
// SOFT DELETE TESTS
// ============================================

func TestSoftDelete_SoftRemoveMarksAsDeleted(t *testing.T) {
	collection := core.NewCollection[int, BlogPost]()

	post := &BlogPost{Title: "Test Post", Content: "Content"}
	collection.Add(post)

	// Verify post exists and is active
	if !collection.Contains(post.Id) {
		t.Error("Post should exist after Add()")
	}

	// Soft delete
	err := collection.SoftRemove(post.Id)
	if err != nil {
		t.Fatalf("Failed to soft delete: %v", err)
	}

	// Verify post still exists in data but is marked inactive
	if !collection.Contains(post.Id, true) {
		t.Error("Post should still exist in data after SoftRemove (when including deleted)")
	}

	// Verify post is not found in default queries
	if collection.Contains(post.Id) {
		t.Error("Post should not be found in default queries after SoftRemove()")
	}
}

func TestSoftDelete_GetByIdExcludesDeleted(t *testing.T) {
	collection := core.NewCollection[int, BlogPost]()

	post := &BlogPost{Title: "Test Post", Content: "Content"}
	collection.Add(post)

	// Get before deletion
	retrieved, err := collection.GetById(post.Id)
	if err != nil {
		t.Fatalf("Failed to get post before delete: %v", err)
	}
	if retrieved.Title != "Test Post" {
		t.Errorf("Expected title 'Test Post', got '%s'", retrieved.Title)
	}

	// Soft delete
	collection.SoftRemove(post.Id)

	// GetById with default (exclude deleted)
	_, err = collection.GetById(post.Id)
	if err == nil {
		t.Error("GetById() should return error for soft-deleted item")
	}

	// GetById with includeDeleted
	retrieved, err = collection.GetById(post.Id, true)
	if err != nil {
		t.Fatalf("GetById(true) should find soft-deleted item: %v", err)
	}
	if retrieved.Title != "Test Post" {
		t.Errorf("Expected to retrieve soft-deleted item")
	}
}

func TestSoftDelete_QueryExcludesDeleted(t *testing.T) {
	collection := core.NewCollection[int, BlogPost]()

	post1 := &BlogPost{Title: "Post 1", Content: "Content 1"}
	post2 := &BlogPost{Title: "Post 2", Content: "Content 2"}
	post3 := &BlogPost{Title: "Post 3", Content: "Content 3"}

	collection.Add(post1)
	collection.Add(post2)
	collection.Add(post3)

	// Soft delete middle one
	collection.SoftRemove(post2.Id)

	// Query without WithSoftDeleted
	results := collection.Query().Get()
	if len(results) != 2 {
		t.Errorf("Expected 2 results (deleted excluded), got %d", len(results))
	}

	for _, post := range results {
		if post.Title == "Post 2" {
			t.Error("Query().Get() should not include soft-deleted posts")
		}
	}

	// Query with WithSoftDeleted
	results = collection.Query().WithSoftDeleted().Get()
	if len(results) != 3 {
		t.Errorf("Expected 3 results (deleted included), got %d", len(results))
	}
}

func TestSoftDelete_ListExcludesDeleted(t *testing.T) {
	collection := core.NewCollection[int, BlogPost]()

	post1 := &BlogPost{Title: "Post 1", Content: "C1"}
	post2 := &BlogPost{Title: "Post 2", Content: "C2"}
	post3 := &BlogPost{Title: "Post 3", Content: "C3"}

	collection.Add(post1)
	collection.Add(post2)
	collection.Add(post3)

	collection.SoftRemove(post2.Id)

	// List without deleted
	results := collection.List()
	if len(results) != 2 {
		t.Errorf("Expected 2 items (deleted excluded), got %d", len(results))
	}

	// List with deleted
	results = collection.List(true)
	if len(results) != 3 {
		t.Errorf("Expected 3 items (deleted included), got %d", len(results))
	}
}

func TestSoftDelete_CountExcludesDeleted(t *testing.T) {
	collection := core.NewCollection[int, BlogPost]()

	post1 := &BlogPost{Title: "Post 1", Content: "C1"}
	post2 := &BlogPost{Title: "Post 2", Content: "C2"}
	post3 := &BlogPost{Title: "Post 3", Content: "C3"}

	collection.Add(post1)
	collection.Add(post2)
	collection.Add(post3)

	if collection.Count() != 3 {
		t.Errorf("Expected count 3, got %d", collection.Count())
	}

	collection.SoftRemove(post2.Id)

	// Count without deleted
	if collection.Count() != 2 {
		t.Errorf("Expected count 2 after soft delete, got %d", collection.Count())
	}

	// Count with deleted
	if collection.Count(true) != 3 {
		t.Errorf("Expected count 3 with deleted, got %d", collection.Count(true))
	}
}

func TestSoftDelete_ContainsExcludesDeleted(t *testing.T) {
	collection := core.NewCollection[int, BlogPost]()

	post := &BlogPost{Title: "Post", Content: "Content"}
	collection.Add(post)

	// Before deletion
	if !collection.Contains(post.Id) {
		t.Error("Contains() should return true before delete")
	}

	collection.SoftRemove(post.Id)

	// After soft delete
	if collection.Contains(post.Id) {
		t.Error("Contains() should return false for soft-deleted after SoftRemove()")
	}

	// With includeDeleted
	if !collection.Contains(post.Id, true) {
		t.Error("Contains(true) should return true for soft-deleted")
	}
}

func TestSoftDelete_UpdateFailsOnDeleted(t *testing.T) {
	collection := core.NewCollection[int, BlogPost]()

	post := &BlogPost{Title: "Original", Content: "Content"}
	collection.Add(post)

	collection.SoftRemove(post.Id)

	// Try to update without includeDeleted
	err := collection.Update(post.Id, BlogPost{Title: "Updated", Content: "New"})
	if err == nil {
		t.Error("Update() should fail on soft-deleted item")
	}

	// Try to update with includeDeleted
	err = collection.Update(post.Id, BlogPost{Title: "Updated", Content: "New"}, true)
	if err != nil {
		t.Fatalf("Update(true) should succeed on soft-deleted: %v", err)
	}
}

func TestSoftDelete_UpdateWithAttributesFailsOnDeleted(t *testing.T) {
	collection := core.NewCollection[int, BlogPost]()

	post := &BlogPost{Title: "Original", Content: "Content"}
	collection.Add(post)

	collection.SoftRemove(post.Id)

	// Try to update without includeDeleted
	err := collection.UpdateWithAttributes(post.Id, map[string]interface{}{"Title": "Updated"})
	if err == nil {
		t.Error("UpdateWithAttributes() should fail on soft-deleted")
	}

	// Try to update with includeDeleted
	err = collection.UpdateWithAttributes(post.Id, map[string]interface{}{"Title": "Updated"}, true)
	if err != nil {
		t.Fatalf("UpdateWithAttributes(true) should succeed: %v", err)
	}
}

func TestSoftDelete_HardDeleteRemovesCompletely(t *testing.T) {
	collection := core.NewCollection[int, BlogPost]()

	post := &BlogPost{Title: "Post", Content: "Content"}
	collection.Add(post)

	// Hard delete
	err := collection.Remove(post.Id)
	if err != nil {
		t.Fatalf("Remove() failed: %v", err)
	}

	// Should not exist at all
	if collection.Contains(post.Id) {
		t.Error("Hard-deleted item should not be found")
	}

	if collection.Contains(post.Id, true) {
		t.Error("Hard-deleted item should not exist even with includeDeleted=true")
	}

	// Query should not find it
	results := collection.Query().WithSoftDeleted().Get()
	if len(results) > 0 {
		t.Error("Query().WithSoftDeleted() should not find hard-deleted items")
	}
}

func TestSoftDelete_SnapshotPreservesSoftDeleteState(t *testing.T) {
	collection := core.NewCollection[int, BlogPost]()

	post1 := &BlogPost{Title: "Post 1", Content: "C1"}
	post2 := &BlogPost{Title: "Post 2", Content: "C2"}
	post3 := &BlogPost{Title: "Post 3", Content: "C3"}

	collection.Add(post1)
	collection.Add(post2)
	collection.Add(post3)

	collection.SoftRemove(post2.Id)

	// Take snapshot
	snapshot := collection.SnapShot()

	// Verify snapshot contains activeElement state
	if len(snapshot.Data) != 3 {
		t.Errorf("Snapshot should contain 3 items (including soft-deleted), got %d", len(snapshot.Data))
	}

	if len(snapshot.ActiveElement) != 3 {
		t.Errorf("Snapshot activeElement should have 3 entries, got %d", len(snapshot.ActiveElement))
	}

	if snapshot.ActiveElement[post2.Id] {
		t.Error("Snapshot should show post2 as inactive (soft-deleted)")
	}

	if !snapshot.ActiveElement[post1.Id] || !snapshot.ActiveElement[post3.Id] {
		t.Error("Snapshot should show post1 and post3 as active")
	}

	// Clear and restore
	collection.Clear()
	collection.Restore(snapshot)

	// Verify state is restored
	if collection.Count() != 2 {
		t.Errorf("After restore, count should be 2 (active only), got %d", collection.Count())
	}

	if collection.Count(true) != 3 {
		t.Errorf("After restore, count(true) should be 3, got %d", collection.Count(true))
	}

	if collection.Contains(post2.Id) {
		t.Error("After restore, post2 should still be soft-deleted")
	}
}

func TestSoftDelete_QueryWithFilter(t *testing.T) {
	collection := core.NewCollection[int, BlogPost]()

	// Add multiple posts
	for i := 1; i <= 5; i++ {
		post := &BlogPost{
			Title:   fmt.Sprintf("Post %d", i),
			Content: fmt.Sprintf("Content %d", i),
		}
		collection.Add(post)
	}

	// Soft delete posts 2 and 4
	collection.SoftRemove(2)
	collection.SoftRemove(4)

	// Query with Where - should exclude deleted
	results := collection.Query().
		Where("Title", "contains", "Post").
		Get()

	if len(results) != 3 {
		t.Errorf("Expected 3 results (deleted excluded), got %d", len(results))
	}

	// Query with Where + WithSoftDeleted - should include all
	results = collection.Query().
		Where("Title", "contains", "Post").
		WithSoftDeleted().
		Get()

	if len(results) != 5 {
		t.Errorf("Expected 5 results (with soft-deleted), got %d", len(results))
	}
}

func TestSoftDelete_ConcurrentOperations(t *testing.T) {
	collection := core.NewCollection[int, BlogPost]()

	for i := 1; i <= 10; i++ {
		post := &BlogPost{
			Title:   fmt.Sprintf("Post %d", i),
			Content: fmt.Sprintf("Content %d", i),
		}
		collection.Add(post)
	}

	// Concurrent writes
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if id%2 == 0 {
				errors <- collection.SoftRemove(id)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Errorf("Concurrent soft delete error: %v", err)
		}
	}

	// Should have 5 active (odd ids)
	if collection.Count() != 5 {
		t.Errorf("Expected 5 active posts after concurrent delete, got %d", collection.Count())
	}

	if collection.Count(true) != 10 {
		t.Errorf("Expected 10 total posts, got %d", collection.Count(true))
	}
}

// ============================================
// QUERY OPERATOR TESTS (In, NotIn, Between)
// ============================================

func TestQuery_In_SingleValue(t *testing.T) {
	collection := core.NewCollection[int, Product]()

	products := []*Product{
		{Name: "Laptop", Price: 999.99},
		{Name: "Phone", Price: 499.99},
		{Name: "Tablet", Price: 299.99},
		{Name: "Monitor", Price: 199.99},
	}

	for _, p := range products {
		collection.Add(p)
	}

	// Query for products with specific names
	results := collection.Query().
		In("Name", []interface{}{"Laptop", "Phone"}).
		Get()

	if len(results) != 2 {
		t.Errorf("Expected 2 products, got %d", len(results))
	}

	names := make(map[string]bool)
	for _, p := range results {
		names[p.Name] = true
	}

	if !names["Laptop"] || !names["Phone"] {
		t.Error("Expected Laptop and Phone in results")
	}

	if names["Tablet"] || names["Monitor"] {
		t.Error("Should not include Tablet or Monitor")
	}
}

func TestQuery_In_WithPrices(t *testing.T) {
	collection := core.NewCollection[int, Product]()

	prices := []float64{99.99, 199.99, 299.99, 399.99, 499.99}
	for i, p := range prices {
		product := &Product{
			Name:  fmt.Sprintf("Product%d", i+1),
			Price: p,
		}
		collection.Add(product)
	}

	// Query for products with specific prices
	results := collection.Query().
		In("Price", []interface{}{199.99, 299.99, 499.99}).
		Get()

	if len(results) != 3 {
		t.Errorf("Expected 3 products, got %d", len(results))
	}

	for _, p := range results {
		if p.Price != 199.99 && p.Price != 299.99 && p.Price != 499.99 {
			t.Errorf("Product price %f not in expected list", p.Price)
		}
	}
}

func TestQuery_NotIn(t *testing.T) {
	collection := core.NewCollection[int, Product]()

	products := []*Product{
		{Name: "Laptop", Price: 999.99},
		{Name: "Phone", Price: 499.99},
		{Name: "Tablet", Price: 299.99},
		{Name: "Monitor", Price: 199.99},
		{Name: "Keyboard", Price: 99.99},
	}

	for _, p := range products {
		collection.Add(p)
	}

	// Query for products NOT in the list
	results := collection.Query().
		NotIn("Name", []interface{}{"Laptop", "Phone"}).
		Get()

	if len(results) != 3 {
		t.Errorf("Expected 3 products (NotIn), got %d", len(results))
	}

	for _, p := range results {
		if p.Name == "Laptop" || p.Name == "Phone" {
			t.Errorf("NotIn should exclude %s", p.Name)
		}
	}

	names := make(map[string]bool)
	for _, p := range results {
		names[p.Name] = true
	}

	if !names["Tablet"] || !names["Monitor"] || !names["Keyboard"] {
		t.Error("Expected Tablet, Monitor, and Keyboard in results")
	}
}

func TestQuery_Between(t *testing.T) {
	collection := core.NewCollection[int, Product]()

	products := []*Product{
		{Name: "Budget Item", Price: 50.0},
		{Name: "Cheap Item", Price: 100.0},
		{Name: "Mid Item 1", Price: 250.0},
		{Name: "Mid Item 2", Price: 350.0},
		{Name: "Expensive Item", Price: 500.0},
		{Name: "Very Expensive", Price: 1000.0},
	}

	for _, p := range products {
		collection.Add(p)
	}

	// Query for products with price between 200 and 400
	results := collection.Query().
		Between("Price", 200.0, 400.0).
		Get()

	if len(results) != 2 {
		t.Errorf("Expected 2 products in price range, got %d", len(results))
	}

	for _, p := range results {
		if p.Price < 200.0 || p.Price > 400.0 {
			t.Errorf("Product price %f not in range [200, 400]", p.Price)
		}
	}
}

func TestQuery_Between_Boundaries(t *testing.T) {
	collection := core.NewCollection[int, Product]()

	products := []*Product{
		{Name: "Item 1", Price: 100.0},
		{Name: "Item 2", Price: 250.0},
		{Name: "Item 3", Price: 500.0},
		{Name: "Item 4", Price: 750.0},
		{Name: "Item 5", Price: 1000.0},
	}

	for _, p := range products {
		collection.Add(p)
	}

	// Query with min and max as boundaries (inclusive)
	results := collection.Query().
		Between("Price", 250.0, 750.0).
		Get()

	if len(results) != 3 {
		t.Errorf("Expected 3 products including boundaries, got %d", len(results))
	}

	// Verify boundaries are included
	hasMin := false
	hasMax := false
	for _, p := range results {
		if p.Price == 250.0 {
			hasMin = true
		}
		if p.Price == 750.0 {
			hasMax = true
		}
	}

	if !hasMin {
		t.Error("Between should include minimum boundary")
	}
	if !hasMax {
		t.Error("Between should include maximum boundary")
	}
}

func TestQuery_In_NotIn_Combined(t *testing.T) {
	collection := core.NewCollection[int, Product]()

	for i := 1; i <= 10; i++ {
		product := &Product{
			Name:    fmt.Sprintf("Product%d", i),
			Price:   float64(i * 100),
			InStock: i%2 == 0,
		}
		collection.Add(product)
	}

	// Query: In a set of prices AND NOT in another set
	results := collection.Query().
		In("Price", []interface{}{200.0, 300.0, 400.0, 500.0, 600.0}).
		NotIn("Name", []interface{}{"Product2", "Product6"}).
		Get()

	if len(results) != 3 {
		t.Errorf("Expected 3 products (In combined with NotIn), got %d", len(results))
	}

	for _, p := range results {
		validPrice := p.Price == 200.0 || p.Price == 300.0 || p.Price == 400.0 || p.Price == 500.0 || p.Price == 600.0
		if !validPrice {
			t.Errorf("Product price %f should be in [200, 300, 400, 500, 600]", p.Price)
		}

		if p.Name == "Product2" || p.Name == "Product6" {
			t.Errorf("Should not include %s", p.Name)
		}
	}
}

func TestQuery_In_Where_Between_Combined(t *testing.T) {
	collection := core.NewCollection[int, Product]()

	products := []*Product{
		{Name: "Laptop", Price: 999.99, InStock: true},
		{Name: "Phone", Price: 499.99, InStock: false},
		{Name: "Tablet", Price: 299.99, InStock: true},
		{Name: "Monitor", Price: 199.99, InStock: false},
		{Name: "Keyboard", Price: 79.99, InStock: true},
	}

	for _, p := range products {
		collection.Add(p)
	}

	// Complex query: In specific names AND between price range AND in stock
	results := collection.Query().
		In("Name", []interface{}{"Laptop", "Tablet", "Monitor", "Keyboard"}).
		Between("Price", 100.0, 500.0).
		Where("InStock", "=", true).
		Get()

	if len(results) != 1 {
		t.Errorf("Expected 1 products from complex query, got %d", len(results))
	}

	// Should be Tablet and Keyboard
	names := make(map[string]bool)
	for _, p := range results {
		names[p.Name] = true
	}

	if !names["Tablet"] {
		t.Error("Expected Tablet in results")
	}

	if names["Laptop"] || names["Phone"] || names["Monitor"] || names["Keyboard"] {
		t.Error("Should not include Laptop, Phone, Monitor, Keyboard")
	}
}

func TestQuery_In_EmptyList(t *testing.T) {
	collection := core.NewCollection[int, Product]()

	for i := 1; i <= 5; i++ {
		product := &Product{
			Name:  fmt.Sprintf("Product%d", i),
			Price: float64(i * 100),
		}
		collection.Add(product)
	}

	// Query with empty In list
	results := collection.Query().
		In("Name", []interface{}{}).
		Get()

	if len(results) != 0 {
		t.Errorf("Expected 0 products with empty In list, got %d", len(results))
	}
}

func TestQuery_NotIn_EmptyList(t *testing.T) {
	collection := core.NewCollection[int, Product]()

	for i := 1; i <= 5; i++ {
		product := &Product{
			Name:  fmt.Sprintf("Product%d", i),
			Price: float64(i * 100),
		}
		collection.Add(product)
	}

	// Query with empty NotIn list - should match all
	results := collection.Query().
		NotIn("Name", []interface{}{}).
		Get()

	if len(results) != 5 {
		t.Errorf("Expected 5 products with empty NotIn list, got %d", len(results))
	}
}
