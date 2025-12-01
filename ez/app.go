/*
 *  This file is used for re-exporting core functionalities of the ezapi-go framework.
 *  They are grouped by feature for better organization and usability. (should be)
 */

package ez

import (
	"reflect"
	"time"

	"github.com/gin-gonic/gin"

	"ezapi-go/core"
	"ezapi-go/core/auth"
	"ezapi-go/core/feature"
	"ezapi-go/core/http"
)

type App[T any] struct {
    router *gin.Engine
    db     *[]T
}

/////////////////////////////////////// routes setup ///////////////////////////////////////
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

/////////////////////////////////////// helper functions ///////////////////////////////////////
var (
    SendSuccess          = http.SendSuccess
    SendError           = http.SendError
    SendErrorWithDetails = http.SendErrorWithDetails
)

////////////////////////////////////// query builder ///////////////////////////////////////
func Query[T any](data []T) *feature.QueryBuilder[T] {
    return feature.NewQueryBuilder[T](data)
}

/////////////////////////////////////// hook functions ///////////////////////////////////////
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

////////////////////////////////////// authentication ///////////////////////////////////////
var (
    SetupAuthProvider = auth.SetupAuthProvider

    // JWT 
    NewJWTProvider = auth.NewJWTProvider

    SetAuthProvider = auth.SetGlobalAuthProvider
    GetAuthProvider = auth.GetGlobalAuthProvider
)

////////////////////////////////////// help function ///////////////////////////////////////
func (ez *App[T]) AutoTimestamp() *App[T] {

    convertTimeToVariable := func(field reflect.Value, now time.Time) {
        if field.Type().String() == "string" {
            field.SetString(now.Format(time.RFC3339))
        } else if field.Type().String() == "time.Time" {
            field.Set(reflect.ValueOf(now))
        } else if field.Type().String() == "int64" {
            field.SetInt(now.Unix())
        }
    }

    return ez.
        BeforeCreate(func (obj *T) error {
            // find 'created_at' and 'updated_at' fields and set to current time
            v := reflect.ValueOf(obj)
            if v.Kind() == reflect.Ptr {
                v = v.Elem()
            }
            now := time.Now()
            if createdAtField := v.FieldByName("CreatedAt"); createdAtField.IsValid() && createdAtField.CanSet() {
                // check field type
                convertTimeToVariable(createdAtField, now)
            }
            if updatedAtField := v.FieldByName("UpdatedAt"); updatedAtField.IsValid() && updatedAtField.CanSet() {
                convertTimeToVariable(updatedAtField, now)
            }
            return nil
        }). 
        BeforeUpdate(func(obj *T, requestObj *T) error {
            v := reflect.ValueOf(requestObj)
            if v.Kind() == reflect.Ptr {
                v = v.Elem()
            }
            now := time.Now()
            if updatedAtField := v.FieldByName("UpdatedAt"); updatedAtField.IsValid() && updatedAtField.CanSet() {
                convertTimeToVariable(updatedAtField, now)
            }
            return nil
        })
}