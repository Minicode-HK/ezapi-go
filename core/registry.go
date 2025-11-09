package core 

import (
	"reflect"

	"github.com/gin-gonic/gin"

    "ezapi-go/core/feature"
)

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
func RegisterRouter[T any](inMemoryDB *[]T, basePath string, generator ... feature.IDGenerator) {

    if len(generator) == 0 {
        generator = append(generator, &feature.UUIDGenerator{})
    }

    if incremental, ok := generator[0].(*feature.IncrementalGenerator); ok {
        incremental.Counter = len(*inMemoryDB)
    }

	moduleRegistry = append(moduleRegistry, ModuleInfo{
        TypeName: reflect.TypeOf(*inMemoryDB).Elem(),
        BasePath: basePath,
        Data:    inMemoryDB,
    })

	routerRegistry = append(routerRegistry, func(router *gin.Engine) {
		CRUDRouter(router, inMemoryDB, basePath, generator[0])
	})
    
}

func RegisterRouterWith(setup func(*gin.Engine)) {
    routerRegistry = append(routerRegistry, setup)
}

// Router sets up the standard CRUD routes for a given database
func CRUDRouter[T any](router *gin.Engine, inMemoryDB *[]T, basePath string, generator feature.IDGenerator) *gin.Engine {
    // Wrap the database with thread-safe wrapper
    db := NewDBWrapper(inMemoryDB, generator)
    
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

// Apply all registered routers
func SetupAllRouters(router *gin.Engine) {
    for _, setup := range routerRegistry {
        setup(router)
    }

    feature.ResetableRoute(router)
}