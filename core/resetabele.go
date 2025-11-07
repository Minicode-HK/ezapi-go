package core 

import (
	"reflect"

	"github.com/gin-gonic/gin"
)

var (
	_resetableDBInitialData []any
	_resetableDBs          []any
)

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


func ResetableRoute(router *gin.Engine) {
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