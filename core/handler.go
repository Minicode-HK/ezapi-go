package core 

import (
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
)


// Get returns a handler that retrieves all items from the database
func Get[T any](db *DBWrapper[T]) gin.HandlerFunc {
    return func(c *gin.Context) {
		var res []T = db.GetAll()
		
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
		paginateRes := db.PaginateWith(&res, page, pageSize)

		SendSuccess(c, paginateRes.Data, gin.H{
            "pagination": gin.H{
                "total":       paginateRes.Total,
                "page":        paginateRes.Page,
                "page_size":   paginateRes.PageSize,
                "total_pages": paginateRes.TotalPages,
                "has_more":    paginateRes.HasMore,
            },
        })
    }
}

// GetById returns a handler that retrieves a single item by ID
func GetById[T any](db *DBWrapper[T]) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
		item := db.GetById(id)

		if item == nil {
			SendError(c, 404, "Item not found")
			return
		}
		
		SendSuccess(c, item)
    }
}

// Post returns a handler that creates a new item
func Post[T any](db *DBWrapper[T]) gin.HandlerFunc {
    return func(c *gin.Context) {
        var newItem T

        if !ValidateBind(c, &newItem) {
            return
        }

        db.Add(&newItem)

		SendSuccess(c, newItem)
    }
}

// Put returns a handler that updates an existing item by ID
func Put[T any](db *DBWrapper[T]) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        var updatedItem T

		if !ValidateBind(c, &updatedItem) {
			return
		}

		success := db.Update(id, &updatedItem)
		if !success {
			SendError(c, 404, "Item not found")
			return
		}
		
		SendSuccess(c, updatedItem)
    }
}

// Delete returns a handler that deletes an item by ID
func Delete[T any](db *DBWrapper[T]) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")

        db.mu.Lock() // Write lock
        defer db.mu.Unlock()

        for i, item := range *db.data {
            idField := reflect.ValueOf(item).FieldByName("Id")
            if idField.IsValid() && idField.String() == id {
                *db.data = append((*db.data)[:i], (*db.data)[i+1:]...)
                SendSuccess(c, gin.H{"message": "Item deleted successfully"})
                return
            }
        }
        SendError(c, 404, "Item not found")
    }
}

func ValidateBind(c *gin.Context, obj interface{}) bool {
    if err := c.ShouldBind(obj); err != nil {
        var errorMessages []string
        
        if validationErrors, ok := err.(validator.ValidationErrors); ok {
            for _, e := range validationErrors {
                errorMessages = append(errorMessages, e.Field()+" is "+e.Tag())
            }
        } else {
            errorMessages = append(errorMessages, "Invalid request format")
        }
        
        SendErrorWithDetails(c, 400, "Validation failed", errorMessages)
        return false
    }
    return true
}
