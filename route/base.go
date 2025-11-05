package route

import (
    "reflect"
    "sync"

    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
)

// DBWrapper wraps the in-memory database with a mutex for thread-safe operations
type DBWrapper[T any] struct {
    data *[]T
    mu   sync.RWMutex // RWMutex allows multiple readers or single writer
}

// NewDBWrapper creates a new thread-safe database wrapper
func NewDBWrapper[T any](data *[]T) *DBWrapper[T] {
    return &DBWrapper[T]{
        data: data,
    }
}

// Get returns a handler that retrieves all items from the database
func Get[T any](db *DBWrapper[T]) gin.HandlerFunc {
    return func(c *gin.Context) {
        db.mu.RLock() // Read lock - allows concurrent reads
        defer db.mu.RUnlock()
        
        // Create a copy to avoid returning pointer to internal slice
        dataCopy := make([]T, len(*db.data))
        copy(dataCopy, *db.data)
        
        SendSuccess(c, dataCopy)
    }
}

// GetById returns a handler that retrieves a single item by ID
func GetById[T any](db *DBWrapper[T]) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")

        db.mu.RLock() // Read lock
        defer db.mu.RUnlock()

        for _, item := range *db.data {
            idField := reflect.ValueOf(item).FieldByName("Id")
            if idField.IsValid() && idField.String() == id {
                SendSuccess(c, item)
                return
            }
        }
        SendError(c, 404, "Item not found")
    }
}

// Post returns a handler that creates a new item
func Post[T any](db *DBWrapper[T]) gin.HandlerFunc {
    return func(c *gin.Context) {
        var newItem T

        if !ValidateBind(c, &newItem) {
            return
        }

        db.mu.Lock() // Write lock - exclusive access
        defer db.mu.Unlock()
        
        *db.data = append(*db.data, newItem)
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

        db.mu.Lock() // Write lock
        defer db.mu.Unlock()

        for i, item := range *db.data {
            idField := reflect.ValueOf(item).FieldByName("Id")
            if idField.IsValid() && idField.String() == id {
                (*db.data)[i] = updatedItem
                SendSuccess(c, updatedItem)
                return
            }
        }
        SendError(c, 404, "Item not found")
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

// Router sets up the standard CRUD routes for a given database
func Router[T any](router *gin.Engine, inMemoryDB *[]T, basePath string) *gin.Engine {
    // Wrap the database with thread-safe wrapper
    db := NewDBWrapper(inMemoryDB)
    
    // Generate module name from type if basePath is empty
    r := []rune(reflect.TypeOf(*inMemoryDB).Elem().Name())
    r[0] = r[0] + 32
    moduleName := string(r)

    if basePath == "" {
        basePath = "/" + moduleName
    }

    router.GET(basePath, Get(db))
    router.GET(basePath+"/:id", GetById(db))
    router.POST(basePath, Post(db))
    router.PUT(basePath+"/:id", Put(db))
    router.DELETE(basePath+"/:id", Delete(db))

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


var _resetableDBInitialData = []any{}
var _resetableDBs = []any{}

// Usage in init(): var ProductDB = ResetableDatabase(&ProductDB, []Product{...})
func ResetableDatabase[T any](dbPtr *[]T, initialData []T) []T {
    db := make([]T, len(initialData))
    copy(db, initialData)
    
    initialCopy := make([]T, len(initialData))
    copy(initialCopy, initialData)
    
    _resetableDBInitialData = append(_resetableDBInitialData, initialCopy)
    _resetableDBs = append(_resetableDBs, dbPtr)
    
    return db
}


var routerRegistry []func(*gin.Engine)

// store module information
type ModuleInfo struct {
	TypeName reflect.Type
    BasePath string
    Data    interface{}
}
// TODO: currently only RegisterRouter have this info
var moduleRegistry []ModuleInfo

func GetModuleRegistry() []ModuleInfo {
    return moduleRegistry
}

// Register a router setup function
func RegisterRouter[T any](inMemoryDB *[]T, basePath string) {
	routerRegistry = append(routerRegistry, func(router *gin.Engine) {
		Router(router, inMemoryDB, basePath)
	})
    moduleRegistry = append(moduleRegistry, ModuleInfo{
        TypeName: reflect.TypeOf(*inMemoryDB).Elem(),
        BasePath: basePath,
        Data:    inMemoryDB,
    })
}

func RegisterRouterWith(setup func(*gin.Engine)) {
    routerRegistry = append(routerRegistry, setup)
}

// Apply all registered routers
func SetupAllRouters(router *gin.Engine) {
    for _, setup := range routerRegistry {
        setup(router)
    }

    // add /api/reset endpoint to reset all resetableDBs
    router.POST("/api/reset", func(c *gin.Context) {
        for i := range _resetableDBs {
            initialData := reflect.ValueOf(_resetableDBInitialData[i])
            dbPtr := reflect.ValueOf(_resetableDBs[i])
            
            dbSlice := dbPtr.Elem()
            
            newSlice := reflect.MakeSlice(dbSlice.Type(), initialData.Len(), initialData.Len())
            reflect.Copy(newSlice, initialData)

            dbSlice.Set(newSlice)
        }

        SendSuccess(c, gin.H{"message": "All databases have been reset"})
    })
}