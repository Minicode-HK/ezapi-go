package route

import (
    "reflect"

    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
)

func Get[T any](inMemoryDB *[]T) gin.HandlerFunc {
    return func(c *gin.Context) {
        SendSuccess(c, inMemoryDB)
    }
}

func GetById[T any](inMemoryDB *[]T) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        
        for _, item := range *inMemoryDB {
            idField := reflect.ValueOf(item).FieldByName("Id")
            if idField.IsValid() && idField.String() == id {
                SendSuccess(c, item)
                return
            }
        }
        SendError(c, 404, "Item not found")
    }
}

func Post[T any](inMemoryDB *[]T) gin.HandlerFunc {
    return func(c *gin.Context) {
        var newItem T
        
        if !ValidateBind(c, &newItem) {
            return
        }
        
        *inMemoryDB = append(*inMemoryDB, newItem)
        SendSuccess(c, newItem)
    }
}

func Delete[T any](inMemoryDB *[]T) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        
        for i, item := range *inMemoryDB {
            // make sure the struct has an Id field
            idField := reflect.ValueOf(item).FieldByName("Id")
            if idField.IsValid() && idField.String() == id {
                *inMemoryDB = append((*inMemoryDB)[:i], (*inMemoryDB)[i+1:]...)
                SendSuccess(c, gin.H{"message": "Item deleted successfully"})
                return
            }
        }
        SendError(c, 404, "Item not found")
    }
}

func Put[T any](inMemoryDB *[]T) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        var updatedItem T
        
        if !ValidateBind(c, &updatedItem) {
            return
        }
        
        for i, item := range *inMemoryDB {
            // make sure the struct has an Id field
            idField := reflect.ValueOf(item).FieldByName("Id")
            if idField.IsValid() && idField.String() == id {
                (*inMemoryDB)[i] = updatedItem
                SendSuccess(c, updatedItem)
                return
            }
        }
        SendError(c, 404, "Item not found")
    }
}

// Update the Router function
func Router[T any](router *gin.Engine, inMemoryDB *[]T, basePath string) *gin.Engine {
    r := []rune(reflect.TypeOf(*inMemoryDB).Elem().Name())
    r[0] = r[0] + 32
    moduleName := string(r)
    
    if basePath == "" {
        basePath = "/" + moduleName
    }
    
    router.GET(basePath, Get(inMemoryDB))
	router.GET(basePath+"/:id", GetById(inMemoryDB))
    router.POST(basePath, Post(inMemoryDB))
    router.PUT(basePath+"/:id", Put(inMemoryDB))
    router.DELETE(basePath+"/:id", Delete(inMemoryDB))
    
    return router
}

// Helper to validate and handle binding errors
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

// Helper function to standardize success responses
func SendSuccess(c *gin.Context, data interface{}) {
    c.JSON(200, gin.H{
        "success": true,
        "data":    data,
    })
}

func SendError(c *gin.Context, code int, message string) {
    c.JSON(code, gin.H{
        "success": false,
        "message": message,
    })
}

func SendErrorWithDetails(c *gin.Context, code int, message string, details interface{}) {
    c.JSON(code, gin.H{
        "success": false,
        "message": message,
        "details": details,
    })
}

var routerRegistry []func(*gin.Engine)

// Register a router setup function
func RegisterRouter(setup func(*gin.Engine)) {
    routerRegistry = append(routerRegistry, setup)
}

// Apply all registered routers
func SetupAllRouters(router *gin.Engine) {
    for _, setup := range routerRegistry {
        setup(router)
    }
}