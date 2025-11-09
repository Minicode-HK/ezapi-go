package ez

import (
    "github.com/gin-gonic/gin"
    "ezapi-go/core"
    "ezapi-go/core/feature"
    "ezapi-go/core/http"
)

type App[T any] struct {
    router *gin.Engine
    db     *[]T
}

func New[T any](db *[]T) *App[T] {
    return &App[T]{
        router: gin.Default(),
        db:     db,
    }
}

func (a *App[T]) ResetableDB(initialData []T) *App[T] {
    *a.db = feature.ResetableDatabase(a.db, initialData)
	return a
}

func (a *App[T]) CRUDRoutes(path string) *App[T] {
    core.RegisterRouter(a.db, path)
	return a
}

func (a *App[T]) CustomRoutes(fn func(*gin.Engine)) *App[T] {
    core.RegisterRouterWith(fn)
	return a
}

var (
    SendSuccess          = http.SendSuccess
    SendError           = http.SendError
    SendErrorWithDetails = http.SendErrorWithDetails
)