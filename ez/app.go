package ez

import (
	"ezapi-go/core"
	"ezapi-go/core/feature"
	"ezapi-go/core/http"
	"reflect"

	"github.com/gin-gonic/gin"
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

func Query[T any](data []T) *feature.QueryBuilder[T] {
    return feature.NewQueryBuilder[T](data)
}


func (a *App[T]) BeforeCreate(hook feature.BeforeCreateFunc[T]) *App[T] {
    t := reflect.TypeOf((*T)(nil)).Elem()
    feature.AddBeforeCreateHook(t, func(obj *any) error {
        return hook((*obj).(*T))
    })
    return a
}

func (a *App[T]) AfterCreate(hook feature.AfterCreateFunc[T]) *App[T] {
    t := reflect.TypeOf((*T)(nil)).Elem()
    feature.AddAfterCreateHook(t, func(obj *any) error {
        return hook((*obj).(*T))
    })
    return a
}

func (a *App[T]) BeforeUpdate(hook feature.BeforeUpdateFunc[T]) *App[T] {
    t := reflect.TypeOf((*T)(nil)).Elem()
    feature.AddBeforeUpdateHook(t, func(obj *any, requestObj *any) error {
        return hook((*obj).(*T), (*requestObj).(*T))
    })
    return a
}

func (a *App[T]) AfterUpdate(hook feature.AfterUpdateFunc[T]) *App[T] { 
    t := reflect.TypeOf((*T)(nil)).Elem()
    feature.AddAfterUpdateHook(t, func(obj *any) error {
        return hook((*obj).(*T))
    })
    return a
}

func (a *App[T]) BeforeDelete(hook feature.BeforeDeleteFunc[T]) *App[T] {
    t := reflect.TypeOf((*T)(nil)).Elem()
    feature.AddBeforeDeleteHook(t, func(obj *any) error {
        return hook((*obj).(*T))
    })
    return a
}

func (a *App[T]) AfterDelete(hook feature.AfterDeleteFunc[T]) *App[T] {
    t := reflect.TypeOf((*T)(nil)).Elem()
    feature.AddAfterDeleteHook(t, func(obj *any) error {
        return hook((*obj).(*T))
    })
    return a
 }
