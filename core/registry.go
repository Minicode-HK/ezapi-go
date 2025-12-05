package core

import (
	"reflect"

	"github.com/gin-gonic/gin"

	"github.com/Minicode-HK/ezapi-go/core/feature"
)

var routerRegistry []func(*gin.Engine)

// store module information
type ModuleInfo struct {
    BasePath string
    Data    interface{}
}

// TODO: currently only RegisterRouter have this info
var moduleRegistry map[reflect.Type]ModuleInfo = make(map[reflect.Type]ModuleInfo)

func GetAllModuleRegistry() map[reflect.Type]ModuleInfo {
    return moduleRegistry
}

func GetModuleRegistryFor(t reflect.Type) ModuleInfo {
    return moduleRegistry[t]
}


// Register a router setup function
func RegisterRouter[T any](inMemoryDB *[]T, basePath string, generator ... feature.IDGenerator) {

    if len(generator) == 0 {
        generator = append(generator, &feature.UUIDGenerator{})
    }

    if incremental, ok := generator[0].(*feature.IncrementalGenerator); ok {
        incremental.Counter = len(*inMemoryDB)
    }

	moduleRegistry[reflect.TypeOf(*inMemoryDB).Elem()] = ModuleInfo{
        BasePath: basePath,
        Data:    inMemoryDB,
    }

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

    group := router.Group(basePath)

    group.GET("", Get(db))
    group.GET("/:id", GetById(db))
    group.POST("", Post(db))
    group.PUT("/:id", Put(db))
    group.DELETE("/:id", Delete(db))

    return router
}

// Apply all registered routers
func SetupAllRouters(router *gin.Engine) {
    for _, setup := range routerRegistry {
        setup(router)
    }

    feature.ResetableRoute(router)
}