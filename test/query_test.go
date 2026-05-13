package main

import (
	"testing"

	"github.com/Minicode-HK/ezapi-go/core"
)

// Test models for query builder
type Employee struct {
	Id         int `ez:"id:auto_increment"`
	Name       string
	Age        int
	Department string
	Salary     float64
	Active     bool
}

// ============================================
// BASIC WHERE TESTS
// ============================================

func TestQuery_Where_Equals(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		Where("Department", "=", "Engineering").
		Get()

	if len(results) != 3 {
		t.Errorf("Expected 3 engineers, got %d", len(results))
	}

	for _, emp := range results {
		if emp.Department != "Engineering" {
			t.Errorf("Expected Department 'Engineering', got '%s'", emp.Department)
		}
	}
}

func TestQuery_Where_NotEquals(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		Where("Department", "!=", "Engineering").
		Get()

	if len(results) != 2 {
		t.Errorf("Expected 2 non-engineers, got %d", len(results))
	}

	for _, emp := range results {
		if emp.Department == "Engineering" {
			t.Error("Should not include Engineering department")
		}
	}
}

func TestQuery_Where_GreaterThan(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		Where("Age", ">", 30).
		Get()

	if len(results) != 3 {
		t.Errorf("Expected 3 employees over 30, got %d", len(results))
	}

	for _, emp := range results {
		if emp.Age <= 30 {
			t.Errorf("Expected Age > 30, got %d", emp.Age)
		}
	}
}

func TestQuery_Where_LessThan(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		Where("Age", "<", 30).
		Get()

	if len(results) != 2 {
		t.Errorf("Expected 2 employees under 30, got %d", len(results))
	}

	for _, emp := range results {
		if emp.Age >= 30 {
			t.Errorf("Expected Age < 30, got %d", emp.Age)
		}
	}
}

func TestQuery_Where_GreaterThanOrEqual(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		Where("Salary", ">=", 80000.0).
		Get()

	if len(results) != 2 {
		t.Errorf("Expected 2 employees with salary >= 80000, got %d", len(results))
	}

	for _, emp := range results {
		if emp.Salary < 80000.0 {
			t.Errorf("Expected Salary >= 80000, got %f", emp.Salary)
		}
	}
}

func TestQuery_Where_LessThanOrEqual(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		Where("Salary", "<=", 70000.0).
		Get()

	if len(results) != 2 {
		t.Errorf("Expected 2 employees with salary <= 70000, got %d", len(results))
	}

	for _, emp := range results {
		if emp.Salary > 70000.0 {
			t.Errorf("Expected Salary <= 70000, got %f", emp.Salary)
		}
	}
}

func TestQuery_Where_Contains(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		Where("Name", "contains", "Smith").
		Get()

	if len(results) != 2 {
		t.Errorf("Expected 2 employees with 'Smith' in name, got %d", len(results))
	}

	for _, emp := range results {
		if !contains(emp.Name, "Smith") {
			t.Errorf("Expected Name to contain 'Smith', got '%s'", emp.Name)
		}
	}
}

// ============================================
// MULTIPLE WHERE CONDITIONS
// ============================================

func TestQuery_MultipleWhere(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		Where("Department", "=", "Engineering").
		Where("Age", ">", 25).
		Get()

	if len(results) != 2 {
		t.Errorf("Expected 2 engineers over 25, got %d", len(results))
	}

	for _, emp := range results {
		if emp.Department != "Engineering" || emp.Age <= 25 {
			t.Errorf("Expected Engineering and Age > 25, got %s, age %d", emp.Department, emp.Age)
		}
	}
}

func TestQuery_MultipleWhere_Complex(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		Where("Department", "=", "Engineering").
		Where("Salary", ">=", 80000.0).
		Where("Active", "=", true).
		Get()

	if len(results) == 0 {
		t.Error("Expected to find matching employees")
	}

	for _, emp := range results {
		if emp.Department != "Engineering" || emp.Salary < 80000 || !emp.Active {
			t.Error("Employee does not match all conditions")
		}
	}
}

// ============================================
// WHEREWITH (CUSTOM FILTER) TESTS
// ============================================

func TestQuery_WhereWith_Single(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		WhereWith(func(emp *Employee) bool {
			return emp.Age >= 25 && emp.Age <= 35
		}).
		Get()

	if len(results) != 4 {
		t.Errorf("Expected 4 employees aged 25-35, got %d", len(results))
	}

	for _, emp := range results {
		if emp.Age < 25 || emp.Age > 35 {
			t.Errorf("Expected Age 25-35, got %d", emp.Age)
		}
	}
}

func TestQuery_WhereWith_Multiple(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		WhereWith(func(emp *Employee) bool {
			return emp.Age > 25
		}).
		WhereWith(func(emp *Employee) bool {
			return emp.Salary > 70000
		}).
		Get()

	for _, emp := range results {
		if emp.Age <= 25 || emp.Salary <= 70000 {
			t.Error("Employee does not match all custom filters")
		}
	}
}

func TestQuery_Where_And_WhereWith_Combined(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		Where("Department", "=", "Engineering").
		WhereWith(func(emp *Employee) bool {
			return emp.Salary > 75000
		}).
		Get()

	for _, emp := range results {
		if emp.Department != "Engineering" || emp.Salary <= 75000 {
			t.Error("Employee does not match combined conditions")
		}
	}
}

// ============================================
// LIMIT AND OFFSET TESTS
// ============================================

func TestQuery_Limit(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		Limit(2).
		Get()

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestQuery_Limit_ExceedsTotal(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		Limit(100).
		Get()

	// Should return all available (5 in our test data)
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
}

func TestQuery_Offset(t *testing.T) {
	collection := setupEmployeeCollection()

	// Get all results ordered by ID
	allResults := collection.Query().
		OrderBy("Id", "asc").
		Get()

	// Get results with offset 2
	offsetResults := collection.Query().
		OrderBy("Id", "asc").
		Offset(2).
		Get()

	if len(offsetResults) != len(allResults)-2 {
		t.Errorf("Expected %d results, got %d", len(allResults)-2, len(offsetResults))
	}

	// Verify first element of offset results matches third element of all results
	if offsetResults[0].Id != allResults[2].Id {
		t.Error("Offset did not skip correct number of elements")
	}
}

func TestQuery_Limit_And_Offset(t *testing.T) {
	collection := setupEmployeeCollection()

	// Pagination: page 2, size 2
	results := collection.Query().
		OrderBy("Id", "asc").
		Limit(2).
		Offset(2).
		Get()

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestQuery_Offset_ExceedsTotal(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		Offset(100).
		Get()

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

// ============================================
// ORDERBY TESTS
// ============================================

func TestQuery_OrderBy_Asc(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		OrderBy("Age", "asc").
		Get()

	// Verify ascending order
	for i := 1; i < len(results); i++ {
		if results[i].Age < results[i-1].Age {
			t.Errorf("Results not in ascending order: %d before %d", results[i-1].Age, results[i].Age)
		}
	}
}

func TestQuery_OrderBy_Desc(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		OrderBy("Age", "desc").
		Get()

	// Verify descending order
	for i := 1; i < len(results); i++ {
		if results[i].Age > results[i-1].Age {
			t.Errorf("Results not in descending order: %d before %d", results[i-1].Age, results[i].Age)
		}
	}
}

func TestQuery_OrderBy_WithFilter(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		Where("Department", "=", "Engineering").
		OrderBy("Salary", "desc").
		Get()

	// Verify all are engineers
	for _, emp := range results {
		if emp.Department != "Engineering" {
			t.Error("Non-engineer in results")
		}
	}

	// Verify descending salary order
	for i := 1; i < len(results); i++ {
		if results[i].Salary > results[i-1].Salary {
			t.Errorf("Salaries not in descending order: %f before %f", results[i-1].Salary, results[i].Salary)
		}
	}
}

// ============================================
// TERMINAL METHOD TESTS
// ============================================

func TestQuery_Get_All(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().Get()

	if len(results) != 5 {
		t.Errorf("Expected 5 total employees, got %d", len(results))
	}
}

func TestQuery_Get_NoMatches(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		Where("Age", ">", 100).
		Get()

	if len(results) != 0 {
		t.Errorf("Expected no results, got %d", len(results))
	}
}

func TestQuery_First_Success(t *testing.T) {
	collection := setupEmployeeCollection()

	emp, found := collection.Query().
		Where("Department", "=", "Sales").
		First()

	if !found {
		t.Fatal("Expected to find employee")
	}

	if emp.Department != "Sales" {
		t.Errorf("Expected Sales department, got %s", emp.Department)
	}
}

func TestQuery_First_NoMatch(t *testing.T) {
	collection := setupEmployeeCollection()

	_, found := collection.Query().
		Where("Age", ">", 100).
		First()

	if found {
		t.Error("Expected not to find any employee")
	}
}

func TestQuery_Count_All(t *testing.T) {
	collection := setupEmployeeCollection()

	count := collection.Query().Count()

	if count != 5 {
		t.Errorf("Expected count 5, got %d", count)
	}
}

func TestQuery_Count_WithFilter(t *testing.T) {
	collection := setupEmployeeCollection()

	count := collection.Query().
		Where("Department", "=", "Engineering").
		Count()

	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

func TestQuery_Count_NoMatches(t *testing.T) {
	collection := setupEmployeeCollection()

	count := collection.Query().
		Where("Age", ">", 100).
		Count()

	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestQuery_Exists_True(t *testing.T) {
	collection := setupEmployeeCollection()

	exists := collection.Query().
		Where("Department", "=", "Engineering").
		Exists()

	if !exists {
		t.Error("Expected exists to be true")
	}
}

func TestQuery_Exists_False(t *testing.T) {
	collection := setupEmployeeCollection()

	exists := collection.Query().
		Where("Age", ">", 100).
		Exists()

	if exists {
		t.Error("Expected exists to be false")
	}
}

// ============================================
// COMPLEX QUERY TESTS
// ============================================

func TestQuery_ComplexQuery(t *testing.T) {
	collection := setupEmployeeCollection()

	results := collection.Query().
		Where("Active", "=", true).
		Where("Salary", ">=", 70000.0).
		WhereWith(func(emp *Employee) bool {
			return len(emp.Name) > 5
		}).
		OrderBy("Age", "desc").
		Limit(3).
		Get()

	if len(results) > 3 {
		t.Errorf("Expected max 3 results, got %d", len(results))
	}

	for _, emp := range results {
		if !emp.Active || emp.Salary < 70000 || len(emp.Name) <= 5 {
			t.Error("Employee does not match all conditions")
		}
	}

	// Verify ordering
	for i := 1; i < len(results); i++ {
		if results[i].Age > results[i-1].Age {
			t.Error("Results not in descending age order")
		}
	}
}

func TestQuery_Pagination_Scenario(t *testing.T) {
	collection := setupEmployeeCollection()

	// Get page 1 (items 0-1)
	page1 := collection.Query().
		OrderBy("Id", "asc").
		Limit(2).
		Offset(0).
		Get()

	// Get page 2 (items 2-3)
	page2 := collection.Query().
		OrderBy("Id", "asc").
		Limit(2).
		Offset(2).
		Get()

	// Get page 3 (items 4+)
	page3 := collection.Query().
		OrderBy("Id", "asc").
		Limit(2).
		Offset(4).
		Get()

	// add two more record
	collection.Add(&Employee{
		Id:     1,
		Name:   "John",
		Age:    10,
		Salary: 70000.0,
	})
	collection.Add(&Employee{
		Id:     2,
		Name:   "Jane",
		Age:    10,
		Salary: 70000.0,
	})

	page4 := collection.Query().
		OrderBy("Id", "asc").
		Page(4, 2).
		Get()

	if len(page1) != 2 {
		t.Errorf("Expected page1 to have 2 items, got %d", len(page1))
	}
	if len(page2) != 2 {
		t.Errorf("Expected page2 to have 2 items, got %d", len(page2))
	}
	if len(page3) != 1 {
		t.Errorf("Expected page3 to have 1 item, got %d", len(page3))
	}
	if len(page4) != 1 {
		t.Errorf("Expected page4 to have 1 item, got %d", len(page4))
	}

	// Verify no overlap
	if page1[0].Id == page2[0].Id || page2[0].Id == page3[0].Id || page3[0].Id == page4[0].Id {
		t.Error("Pages should not overlap")
	}
}

// ============================================
// EDGE CASE TESTS
// ============================================

func TestQuery_EmptyCollection(t *testing.T) {
	collection := core.NewCollection[int, Employee]()

	results := collection.Query().
		Where("Age", ">", 20).
		Get()

	if len(results) != 0 {
		t.Errorf("Expected empty results, got %d", len(results))
	}

	count := collection.Query().Count()
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestQuery_InvalidFieldName(t *testing.T) {
	collection := setupEmployeeCollection()

	// Should not panic, just return no results
	results := collection.Query().
		Where("NonExistentField", "=", "value").
		Get()

	if len(results) != 0 {
		t.Errorf("Expected no results for invalid field, got %d", len(results))
	}
}

func TestQuery_BooleanField(t *testing.T) {
	collection := setupEmployeeCollection()

	activeEmployees := collection.Query().
		Where("Active", "=", true).
		Get()

	inactiveEmployees := collection.Query().
		Where("Active", "=", false).
		Get()

	for _, emp := range activeEmployees {
		if !emp.Active {
			t.Error("Expected only active employees")
		}
	}

	for _, emp := range inactiveEmployees {
		if emp.Active {
			t.Error("Expected only inactive employees")
		}
	}

	if len(activeEmployees)+len(inactiveEmployees) != 5 {
		t.Error("Active + Inactive should equal total employees")
	}
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// setupEmployeeCollection creates a test collection with sample data
func setupEmployeeCollection() *core.Collection[int, Employee] {
	collection := core.NewCollection[int, Employee]()

	employees := []*Employee{
		{Name: "John Smith", Age: 35, Department: "Engineering", Salary: 90000, Active: true},
		{Name: "Jane Smith", Age: 28, Department: "Engineering", Salary: 85000, Active: true},
		{Name: "Bob Jones", Age: 42, Department: "Sales", Salary: 70000, Active: true},
		{Name: "Alice Brown", Age: 25, Department: "Engineering", Salary: 75000, Active: false},
		{Name: "Charlie Wilson", Age: 31, Department: "HR", Salary: 65000, Active: true},
	}

	for _, emp := range employees {
		collection.Add(emp)
	}

	return collection
}

// contains helper for string contains check
func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) &&
		(haystack == needle || len(needle) == 0 ||
			(len(haystack) > 0 && len(needle) > 0 && stringContains(haystack, needle)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
