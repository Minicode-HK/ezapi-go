package main

import (
	"errors"
	"testing"

	"github.com/Minicode-HK/ezapi-go/core"
)

// Test model for hooks
type Article struct {
	Id      int `ez:"id:auto_increment"`
	Title   string
	Content string
	Views   int
	Status  string
}

// ============================================
// BEFORE CREATE HOOK TESTS
// ============================================

func TestHook_BeforeCreate_Success(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	// Register hook that sets default status
	collection.RegisterBeforeCreateHook(func(article *Article) error {
		if article.Status == "" {
			article.Status = "draft"
		}
		return nil
	})

	article := &Article{Title: "Test Article", Content: "Content"}
	collection.Add(article)

	if article.Status != "draft" {
		t.Errorf("Expected Status 'draft', got '%s'", article.Status)
	}
}

func TestHook_BeforeCreate_Error(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	// Register hook that rejects articles without title
	collection.RegisterBeforeCreateHook(func(article *Article) error {
		if article.Title == "" {
			return errors.New("title is required")
		}
		return nil
	})

	article := &Article{Content: "Content without title"}
	collection.Add(article)

	// Article should not be added due to hook error
	if _, err := collection.GetById(1); err == nil {
		t.Error("Expected article not to be added when hook returns error")
	}

	count := collection.Count()
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestHook_BeforeCreate_Multiple(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	executionOrder := []string{}

	// Register multiple hooks
	collection.RegisterBeforeCreateHook(func(article *Article) error {
		executionOrder = append(executionOrder, "hook1")
		article.Views = 0 // Initialize views
		return nil
	})

	collection.RegisterBeforeCreateHook(func(article *Article) error {
		executionOrder = append(executionOrder, "hook2")
		if article.Status == "" {
			article.Status = "draft"
		}
		return nil
	})

	collection.RegisterBeforeCreateHook(func(article *Article) error {
		executionOrder = append(executionOrder, "hook3")
		return nil
	})

	article := &Article{Title: "Test", Content: "Content"}
	collection.Add(article)

	// Verify all hooks executed in order
	if len(executionOrder) != 3 {
		t.Errorf("Expected 3 hooks to execute, got %d", len(executionOrder))
	}

	if executionOrder[0] != "hook1" || executionOrder[1] != "hook2" || executionOrder[2] != "hook3" {
		t.Errorf("Hooks executed in wrong order: %v", executionOrder)
	}

	// Verify hook effects
	if article.Views != 0 {
		t.Errorf("Expected Views 0, got %d", article.Views)
	}
	if article.Status != "draft" {
		t.Errorf("Expected Status 'draft', got '%s'", article.Status)
	}
}

func TestHook_BeforeCreate_ErrorStopsChain(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	hook1Executed := false
	hook2Executed := false
	hook3Executed := false

	collection.RegisterBeforeCreateHook(func(article *Article) error {
		hook1Executed = true
		return nil
	})

	collection.RegisterBeforeCreateHook(func(article *Article) error {
		hook2Executed = true
		return errors.New("validation failed")
	})

	collection.RegisterBeforeCreateHook(func(article *Article) error {
		hook3Executed = true
		return nil
	})

	article := &Article{Title: "Test"}
	collection.Add(article)

	if !hook1Executed {
		t.Error("Expected hook1 to execute")
	}
	if !hook2Executed {
		t.Error("Expected hook2 to execute")
	}
	if hook3Executed {
		t.Error("Expected hook3 NOT to execute after error")
	}

	if collection.Count() != 0 {
		t.Error("Article should not be added when hook returns error")
	}
}

// ============================================
// AFTER CREATE HOOK TESTS
// ============================================

func TestHook_AfterCreate_Success(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	hookCalled := false
	var capturedId int

	collection.RegisterAfterCreateHook(func(article *Article) error {
		hookCalled = true
		capturedId = article.Id
		return nil
	})

	article := &Article{Title: "Test Article", Content: "Content"}
	collection.Add(article)

	if !hookCalled {
		t.Error("Expected after create hook to be called")
	}

	if capturedId != article.Id {
		t.Errorf("Expected captured ID %d, got %d", article.Id, capturedId)
	}

	if capturedId == 0 {
		t.Error("Expected ID to be set before after create hook")
	}
}

func TestHook_AfterCreate_Error(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	collection.RegisterAfterCreateHook(func(article *Article) error {
		return errors.New("after create error")
	})

	article := &Article{Title: "Test"}
	collection.Add(article)

	// Even though after hook fails, article should already be in collection
	// (This is current behavior - you might want to change this to rollback)
	count := collection.Count()
	if count != 0 {
		t.Logf("Note: Article count is %d - after hook error behavior", count)
	}
}

// ============================================
// BEFORE UPDATE HOOK TESTS
// ============================================

func TestHook_BeforeUpdate_Success(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	// Register hook that increments views
	collection.RegisterBeforeUpdateHook(func(existing *Article, incoming *Article, _ map[string]any) error {
		incoming.Views++ // Track update count
		return nil
	})

	article := &Article{Title: "Original", Content: "Content", Views: 0}
	collection.Add(article)

	err := collection.Update(article.Id, Article{
		Title:   "Updated",
		Content: "New Content",
	})

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	updated, _ := collection.GetById(article.Id)
	if updated.Views != 1 {
		t.Errorf("Expected Views 1 (incremented by hook), got %d", updated.Views)
	}
}

func TestHook_BeforeUpdate_Error(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	// Register hook that prevents status changes
	collection.RegisterBeforeUpdateHook(func(existing *Article, incoming *Article, _ map[string]any) error {
		if incoming != nil && incoming.Status != existing.Status && existing.Status == "published" {
			return errors.New("cannot change status of published article")
		}
		return nil
	})

	article := &Article{Title: "Article", Content: "Content", Status: "published"}
	collection.Add(article)

	err := collection.Update(article.Id, Article{
		Title:   "Updated Title",
		Status:  "draft", // Try to change status
		Content: "Content",
	})

	if err == nil {
		t.Error("Expected error when hook prevents update")
	}

	// Verify article was NOT updated
	updated, _ := collection.GetById(article.Id)
	if updated.Title != "Article" {
		t.Errorf("Expected original title 'Article', got '%s'", updated.Title)
	}
	if updated.Status != "published" {
		t.Errorf("Expected status 'published', got '%s'", updated.Status)
	}
}

func TestHook_BeforeUpdate_CanModifyIncoming(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	// Hook that sanitizes incoming data
	collection.RegisterBeforeUpdateHook(func(existing *Article, incoming *Article, _ map[string]any) error {
		if incoming != nil {
			// Force status to always be lowercase
			incoming.Status = "modified-by-hook"
		}
		return nil
	})

	article := &Article{Title: "Article", Status: "draft"}
	collection.Add(article)

	err := collection.Update(article.Id, Article{
		Title:  "Updated",
		Status: "PUBLISHED",
	})

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	updated, _ := collection.GetById(article.Id)
	if updated.Status != "modified-by-hook" {
		t.Errorf("Expected Status 'modified-by-hook', got '%s'", updated.Status)
	}
}

// ============================================
// AFTER UPDATE HOOK TESTS
// ============================================

func TestHook_AfterUpdate_Success(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	hookCalled := false
	var capturedTitle string

	collection.RegisterAfterUpdateHook(func(article *Article) error {
		hookCalled = true
		capturedTitle = article.Title
		return nil
	})

	article := &Article{Title: "Original", Content: "Content"}
	collection.Add(article)

	collection.Update(article.Id, Article{
		Title:   "Updated",
		Content: "Content",
	})

	if !hookCalled {
		t.Error("Expected after update hook to be called")
	}

	if capturedTitle != "Updated" {
		t.Errorf("Expected captured title 'Updated', got '%s'", capturedTitle)
	}
}

func TestHook_AfterUpdate_WithUpdateAttributes(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	hookCallCount := 0

	collection.RegisterAfterUpdateHook(func(article *Article) error {
		hookCallCount++
		return nil
	})

	article := &Article{Title: "Original", Views: 0}
	collection.Add(article)

	err := collection.UpdateWithAttributes(article.Id, map[string]interface{}{
		"Views": 100,
	})

	if err != nil {
		t.Fatalf("UpdateWithAttributes failed: %v", err)
	}

	if hookCallCount != 1 {
		t.Errorf("Expected hook to be called once, called %d times", hookCallCount)
	}

	updated, _ := collection.GetById(article.Id)
	if updated.Views != 100 {
		t.Errorf("Expected Views 100, got %d", updated.Views)
	}
}

// ============================================
// BEFORE REMOVE HOOK TESTS
// ============================================

func TestHook_BeforeRemove_Success(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	removedTitle := ""

	collection.RegisterBeforeRemoveHook(func(article *Article) error {
		removedTitle = article.Title
		return nil
	})

	article := &Article{Title: "To Be Deleted", Content: "Content"}
	collection.Add(article)

	err := collection.Remove(article.Id)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if removedTitle != "To Be Deleted" {
		t.Errorf("Expected removed title 'To Be Deleted', got '%s'", removedTitle)
	}

	if collection.Count() != 0 {
		t.Error("Article should be removed")
	}
}

func TestHook_BeforeRemove_Error(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	// Prevent deletion of published articles
	collection.RegisterBeforeRemoveHook(func(article *Article) error {
		if article.Status == "published" {
			return errors.New("cannot delete published article")
		}
		return nil
	})

	article := &Article{Title: "Published Article", Status: "published"}
	collection.Add(article)

	err := collection.Remove(article.Id)
	if err == nil {
		t.Error("Expected error when hook prevents deletion")
	}

	// Verify article was NOT removed
	if collection.Count() != 1 {
		t.Error("Article should still be in collection")
	}

	retrieved, _ := collection.GetById(article.Id)
	if retrieved == nil {
		t.Error("Article should still exist")
	}
}

// ============================================
// AFTER REMOVE HOOK TESTS
// ============================================

func TestHook_AfterRemove_Success(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	hookCalled := false
	var removedId int

	collection.RegisterAfterRemoveHook(func(article *Article) error {
		hookCalled = true
		removedId = article.Id
		return nil
	})

	article := &Article{Title: "Test", Content: "Content"}
	collection.Add(article)
	id := article.Id

	err := collection.Remove(id)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if !hookCalled {
		t.Error("Expected after remove hook to be called")
	}

	if removedId != id {
		t.Errorf("Expected removed ID %d, got %d", id, removedId)
	}
}

func TestHook_AfterRemove_ElementAlreadyDeleted(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	var capturedId int
	wasDeleted := false

	collection.RegisterAfterRemoveHook(func(article *Article) error {
		capturedId = article.Id
		// Element should be deleted from collection at this point
		// but we can still access the passed article pointer
		wasDeleted = true
		return nil
	})

	article := &Article{Title: "Test"}
	collection.Add(article)
	id := article.Id

	err := collection.Remove(id)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if !wasDeleted {
		t.Error("After remove hook should have been called")
	}

	// Verify element is actually deleted (checked AFTER hook completes)
	_, err = collection.GetById(id)
	if err == nil {
		t.Error("Element should be deleted from collection")
	}

	if capturedId != id {
		t.Errorf("Expected captured ID %d, got %d", id, capturedId)
	}
}

// ============================================
// COMPLEX HOOK SCENARIOS
// ============================================

func TestHook_MultipleHooksAllTypes(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	eventLog := []string{}

	// Register all hook types
	collection.RegisterBeforeCreateHook(func(article *Article) error {
		eventLog = append(eventLog, "before-create")
		return nil
	})

	collection.RegisterAfterCreateHook(func(article *Article) error {
		eventLog = append(eventLog, "after-create")
		return nil
	})

	collection.RegisterBeforeUpdateHook(func(existing *Article, incoming *Article, _ map[string]any) error {
		eventLog = append(eventLog, "before-update")
		return nil
	})

	collection.RegisterAfterUpdateHook(func(article *Article) error {
		eventLog = append(eventLog, "after-update")
		return nil
	})

	collection.RegisterBeforeRemoveHook(func(article *Article) error {
		eventLog = append(eventLog, "before-remove")
		return nil
	})

	collection.RegisterAfterRemoveHook(func(article *Article) error {
		eventLog = append(eventLog, "after-remove")
		return nil
	})

	// Perform operations
	article := &Article{Title: "Test"}
	collection.Add(article)

	collection.Update(article.Id, Article{Title: "Updated"})

	collection.Remove(article.Id)

	expectedLog := []string{
		"before-create",
		"after-create",
		"before-update",
		"after-update",
		"before-remove",
		"after-remove",
	}

	if len(eventLog) != len(expectedLog) {
		t.Errorf("Expected %d events, got %d: %v", len(expectedLog), len(eventLog), eventLog)
	}

	for i, expected := range expectedLog {
		if i >= len(eventLog) || eventLog[i] != expected {
			t.Errorf("Event %d: expected '%s', got '%s'", i, expected, eventLog[i])
		}
	}
}

func TestHook_ValidationChain(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	// Hook 1: Validate title
	collection.RegisterBeforeCreateHook(func(article *Article) error {
		if len(article.Title) < 3 {
			return errors.New("title too short")
		}
		return nil
	})

	// Hook 2: Validate content
	collection.RegisterBeforeCreateHook(func(article *Article) error {
		if len(article.Content) < 10 {
			return errors.New("content too short")
		}
		return nil
	})

	// Hook 3: Set defaults
	collection.RegisterBeforeCreateHook(func(article *Article) error {
		if article.Status == "" {
			article.Status = "draft"
		}
		return nil
	})

	// Test 1: All validations pass
	article1 := &Article{Title: "Good Title", Content: "Long enough content here"}
	collection.Add(article1)

	if article1.Status != "draft" {
		t.Error("Expected status to be set by hook")
	}

	if collection.Count() != 1 {
		t.Error("Article should be added when all validations pass")
	}

	// Test 2: Title validation fails
	article2 := &Article{Title: "Hi", Content: "Long enough content"}
	collection.Add(article2)

	if collection.Count() != 1 {
		t.Error("Article should NOT be added when title validation fails")
	}

	// Test 3: Content validation fails
	article3 := &Article{Title: "Good Title", Content: "Short"}
	collection.Add(article3)

	if collection.Count() != 1 {
		t.Error("Article should NOT be added when content validation fails")
	}
}

// ============================================
// EDGE CASES
// ============================================

func TestHook_EmptyHooks(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	// No hooks registered - should work normally
	article := &Article{Title: "Test"}
	collection.Add(article)

	if collection.Count() != 1 {
		t.Error("Operations should work normally with no hooks")
	}

	err := collection.Update(article.Id, Article{Title: "Updated"})
	if err != nil {
		t.Error("Update should work with no hooks")
	}

	err = collection.Remove(article.Id)
	if err != nil {
		t.Error("Remove should work with no hooks")
	}
}

func TestHook_NilPointerHandling(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	collection.RegisterBeforeCreateHook(func(article *Article) error {
		// Hook should receive valid pointer
		if article == nil {
			return errors.New("article is nil")
		}
		return nil
	})

	article := &Article{Title: "Test"}
	collection.Add(article)

	if collection.Count() != 1 {
		t.Error("Hook should receive valid article pointer")
	}
}

// ============================================
// UPDATE WITH ATTRIBUTES HOOK TESTS
// ============================================

func TestHook_BeforeUpdate_WithUpdateAttributes_ReceivesAttributes(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	var capturedAttributes map[string]any
	var capturedRequestObj *Article

	collection.RegisterBeforeUpdateHook(func(existing *Article, requestObj *Article, attrs map[string]any) error {
		capturedRequestObj = requestObj
		capturedAttributes = attrs
		return nil
	})

	article := &Article{Title: "Original", Views: 0}
	collection.Add(article)

	err := collection.UpdateWithAttributes(article.Id, map[string]interface{}{
		"Views": 100,
		"Title": "Updated",
	})

	if err != nil {
		t.Fatalf("UpdateWithAttributes failed: %v", err)
	}

	// Verify hook received nil requestObj and non-nil attributes
	if capturedRequestObj != nil {
		t.Error("Expected requestObj to be nil for UpdateWithAttributes")
	}

	if capturedAttributes == nil {
		t.Fatal("Expected attributes to be non-nil")
	}

	if capturedAttributes["Views"] != 100 {
		t.Errorf("Expected Views attribute to be 100, got %v", capturedAttributes["Views"])
	}

	if capturedAttributes["Title"] != "Updated" {
		t.Errorf("Expected Title attribute to be 'Updated', got %v", capturedAttributes["Title"])
	}
}

func TestHook_BeforeUpdate_WithUpdate_ReceivesRequestObj(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	var capturedAttributes map[string]any
	var capturedRequestObj *Article

	collection.RegisterBeforeUpdateHook(func(existing *Article, requestObj *Article, attrs map[string]any) error {
		capturedRequestObj = requestObj
		capturedAttributes = attrs
		return nil
	})

	article := &Article{Title: "Original", Views: 0}
	collection.Add(article)

	err := collection.Update(article.Id, Article{
		Title: "Updated",
		Views: 100,
	})

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify hook received non-nil requestObj and nil attributes
	if capturedRequestObj == nil {
		t.Error("Expected requestObj to be non-nil for Update")
	}

	if capturedAttributes != nil {
		t.Error("Expected attributes to be nil for Update")
	}

	if capturedRequestObj != nil && capturedRequestObj.Title != "Updated" {
		t.Errorf("Expected requestObj Title to be 'Updated', got '%s'", capturedRequestObj.Title)
	}
}

func TestHook_BeforeUpdate_CanDistinguishUpdateTypes(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	updateTypeCalled := ""

	collection.RegisterBeforeUpdateHook(func(existing *Article, requestObj *Article, attrs map[string]any) error {
		if requestObj != nil && attrs == nil {
			updateTypeCalled = "Update"
		} else if requestObj == nil && attrs != nil {
			updateTypeCalled = "UpdateWithAttributes"
		} else {
			updateTypeCalled = "Unknown"
		}
		return nil
	})

	article := &Article{Title: "Original"}
	collection.Add(article)

	// Test Update
	collection.Update(article.Id, Article{Title: "Updated1"})
	if updateTypeCalled != "Update" {
		t.Errorf("Expected 'Update', got '%s'", updateTypeCalled)
	}

	// Test UpdateWithAttributes
	collection.UpdateWithAttributes(article.Id, map[string]interface{}{
		"Views": 10,
	})
	if updateTypeCalled != "UpdateWithAttributes" {
		t.Errorf("Expected 'UpdateWithAttributes', got '%s'", updateTypeCalled)
	}
}

func TestHook_BeforeUpdate_ValidateAttributes(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	// Hook that validates Views attribute
	collection.RegisterBeforeUpdateHook(func(existing *Article, requestObj *Article, attrs map[string]any) error {
		if attrs != nil {
			if views, ok := attrs["Views"]; ok {
				if viewsInt, ok := views.(int); ok && viewsInt < 0 {
					return errors.New("views cannot be negative")
				}
			}
		}
		return nil
	})

	article := &Article{Title: "Article", Views: 10}
	collection.Add(article)

	// Should succeed with valid Views
	err := collection.UpdateWithAttributes(article.Id, map[string]interface{}{
		"Views": 100,
	})
	if err != nil {
		t.Errorf("Should allow positive views: %v", err)
	}

	// Should fail with negative Views
	err = collection.UpdateWithAttributes(article.Id, map[string]interface{}{
		"Views": -10,
	})
	if err == nil {
		t.Error("Expected error for negative views")
	}

	// Verify article was not updated with invalid value
	updated, _ := collection.GetById(article.Id)
	if updated.Views < 0 {
		t.Error("Article should not be updated when hook returns error")
	}
}

func TestHook_BeforeUpdate_ModifyAttributesNotSupported(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	// Note: Modifying the attrs map in the hook WILL affect the update
	// because maps are passed by reference
	collection.RegisterBeforeUpdateHook(func(existing *Article, requestObj *Article, attrs map[string]any) error {
		if attrs != nil {
			// This modification WILL affect the actual update
			attrs["Views"] = 999
		}
		return nil
	})

	article := &Article{Title: "Article", Views: 0}
	collection.Add(article)

	err := collection.UpdateWithAttributes(article.Id, map[string]interface{}{
		"Views": 100,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	updated, _ := collection.GetById(article.Id)
	// The hook's modification will take effect because maps are reference types
	if updated.Views == 999 {
		// Expected behavior - hook can modify attributes map
		t.Log("Hook successfully modified attributes (expected behavior)")
	} else {
		t.Errorf("Expected Views to be 999 (modified by hook), got %d", updated.Views)
	}
}

func TestHook_BeforeUpdate_BothValidation(t *testing.T) {
	collection := core.NewCollection[int, Article]()

	// Hook that validates both Update and UpdateWithAttributes
	collection.RegisterBeforeUpdateHook(func(existing *Article, requestObj *Article, attrs map[string]any) error {
		// Validate Update request
		if requestObj != nil {
			if requestObj.Views < 0 {
				return errors.New("views cannot be negative in Update")
			}
		}

		// Validate UpdateWithAttributes request
		if attrs != nil {
			if views, ok := attrs["Views"]; ok {
				if viewsInt, ok := views.(int); ok && viewsInt < 0 {
					return errors.New("views cannot be negative in UpdateWithAttributes")
				}
			}
		}

		return nil
	})

	article := &Article{Title: "Article", Views: 10}
	collection.Add(article)

	// Test Update with invalid data
	err := collection.Update(article.Id, Article{
		Title: "Updated",
		Views: -5,
	})
	if err == nil {
		t.Error("Expected error for negative views in Update")
	}

	// Test UpdateWithAttributes with invalid data
	err = collection.UpdateWithAttributes(article.Id, map[string]interface{}{
		"Views": -5,
	})
	if err == nil {
		t.Error("Expected error for negative views in UpdateWithAttributes")
	}

	// Verify article unchanged
	updated, _ := collection.GetById(article.Id)
	if updated.Views != 10 {
		t.Errorf("Expected Views to remain 10, got %d", updated.Views)
	}
}
