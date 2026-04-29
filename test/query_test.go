package test

import (
	"reflect"
	"testing"

	"github.com/Minicode-HK/ezapi-go/ez"
)

type User struct {
	FirstName string
	LastName  string
}

func TestCreateQuery(t *testing.T) {

	query := ez.Query(&[]User{})

	// assert query is not nil
	if query == nil {
		t.Fatal("Expected query to be not nil")
	}

}

// get
func TestWhereQuery(t *testing.T) {
	users := []User{
		{FirstName: "John", LastName: "Doe"},
	}

	res := ez.Query(&users).Where("FirstName", "=", "John").Get()

	if res == nil {
		t.Fatal("Expected query to be not nil")
	}

	if len(res) != 1 && res[0].FirstName != "John" && res[0].LastName != "Doe" {
		t.Fatalf("Expected 1 result, got %d", len(res))
	}
}

func TestWhereQueryWith2Record(t *testing.T) {
	users := []User{
		{FirstName: "John", LastName: "Doe"},
		{FirstName: "John", LastName: "Doe"},
	}

	res := ez.Query(&users).Where("FirstName", "=", "John").Get()

	if res == nil {
		t.Fatal("Expected query to be not nil")
	}

	if len(res) != 2 {
		t.Fatalf("Expected 1 result, got %d", len(res))
	}
}

func TestWhereQueryWith2RecordAndShouldReturn1RecordOnly(t *testing.T) {
	users := []User{
		{FirstName: "John", LastName: "Doe"},
		{FirstName: "John", LastName: "Doe2"},
	}

	res := ez.Query(&users).
		Where("FirstName", "=", "John").Where("LastName", "=", "Doe").
		Get()

	if res == nil {
		t.Fatal("Expected query to be not nil")
	}

	if len(res) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(res))
	}
}

func TestWhereQueryReturnSameObject(t *testing.T) {
	users := []User{
		{FirstName: "John", LastName: "Doe"},
	}

	res := ez.Query(&users).Where("FirstName", "=", "John").Get()
	if res == nil {
		t.Fatal("Expected query to be not nil")
	}

	if len(res) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(res))
	}

	if reflect.TypeOf(res[0]) != reflect.TypeOf(&User{}) {
		t.Fatalf("Expected User, got %v", reflect.TypeOf(res[0]))
	}

	// expecting the result from query is still pointing to the same object
	if res[0] != &users[0] {
		t.Fatalf("Expected user, got %v", res[0])
	}
}

func TestWhereQueryReturnSameObjectAfterUpdateOriginalArray(t *testing.T) {
	users := []User{
		{FirstName: "John", LastName: "Doe"},
	}

	res := ez.Query(&users).Where("FirstName", "=", "John").Get()

	users = append(users, User{FirstName: "John2", LastName: "Doe2"})

	if res == nil {
		t.Fatal("Expected query to be not nil")
	}

	if len(res) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(res))
	}

	if reflect.TypeOf(res[0]) != reflect.TypeOf(&User{}) {
		t.Fatalf("Expected User, got %v", reflect.TypeOf(res[0]))
	}

	res[0].FirstName = "John2"
	res[0].LastName = "John2"
	t.Logf("%v", users)

	// expecting the result from query is still pointing to the same object
	if res[0] != &users[0] {
		t.Fatalf("Expecting user %p, got %p", &users[0], res[0])
	}
}

func TestLimitQuery(t *testing.T) {
	users := []User{
		{FirstName: "John", LastName: "Doe"},
	}

	res := ez.Query(&users).Where("FirstName", "=", "John").Limit(1).Get()

	if len(res) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(res))
	}
}

// add
func TestInsertQuery(t *testing.T) {
	users := []User{
		{FirstName: "John", LastName: "Doe"},
	}
	query := ez.Query(&users)

	if query == nil {
		t.Fatal("Expected query to be not nil")
	}

	query.Insert(&User{FirstName: "Jane", LastName: "Smith"})

	if len(users) != 2 {
		t.Fatalf("Expected 2 users, got %d", len(users))
	}

}

// update
func TestUpdateQuery(t *testing.T) {
	users := []User{
		{FirstName: "John", LastName: "Doe"},
	}
	query := ez.Query(&users)

	if query == nil {
		t.Fatal("Expected query to be not nil")
	}

	query.Update("FirstName", "John")

	if users[0].FirstName != "John" {
		t.Fatalf("Expected FirstName to be 'John', got '%s'", users[0].FirstName)
	}

	query.Update("LastName", "Smith")
	if users[0].FirstName != "John" && users[0].LastName != "Smith" {
		t.Fatalf("Expected FirstName to be 'John' or 'Smith', got '%s'", users[1].FirstName)
	}
}
