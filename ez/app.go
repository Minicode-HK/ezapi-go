/*
 *  This file is used for re-exporting core functionalities of the ezapi-go framework.
 *  They are grouped by feature for better organization and usability. (should be)
 */

package ez

import (
	"reflect"
	"time"

	"github.com/Minicode-HK/ezapi-go/core"
	"github.com/Minicode-HK/ezapi-go/core/http"
	"github.com/gin-gonic/gin"

	cms_backend "github.com/Minicode-HK/ezapi-go/cms/backend"
	cms_frontend "github.com/Minicode-HK/ezapi-go/cms/frontend"
	"github.com/Minicode-HK/ezapi-go/core/auth"
	"github.com/Minicode-HK/ezapi-go/core/feature"
)

type App[T any] struct {
	router *gin.Engine
	db     *[]T
}

var (
	sharedRouter *gin.Engine
)

// ///////////////////////////////////// routes setup ///////////////////////////////////////

// New creates a new ez App builder instance
func New[T any](db *[]T) *App[T] {
	if sharedRouter == nil {
		sharedRouter = gin.Default()
	}
	return &App[T]{
		router: sharedRouter,
		db:     db,
	}
}

// Seed the data with initialData.
// Calling /reset endpoint will reset the backend at this state.
func (a *App[T]) Seed(initialData []T) *App[T] {
	*a.db = feature.ResetableDatabase(a.db, initialData)
	return a
}

// CRUD will register the default 4 CRUD routes for the type T
func (a *App[T]) CRUD(path string) *App[T] {
	core.RegisterRouter(a.db, path)
	return a
}

// CustomRoutes allows users to register custom routes using the provided setup function.
func (a *App[T]) CustomRoutes(fn func(*gin.Engine)) *App[T] {
	core.RegisterRouterWith(fn)
	return a
}

// SetupAllRouters apply all the defined routers in ezapi to the provided gin.Engine instance.
// This should be called at main()
func SetupAllRouters(router *gin.Engine) {
	core.SetupAllRouters(router)
}

// ///////////////////////////////////// helper functions ///////////////////////////////////////

// helper function for sending http responses.
var (
	SendSuccess          = http.SendSuccess
	SendError            = http.SendError
	SendErrorWithDetails = http.SendErrorWithDetails
)

////////////////////////////////////// query builder ///////////////////////////////////////

func Query[T any](data *[]T) *feature.QueryBuilder[T] {
	return feature.NewQueryBuilder(data)
}

// ///////////////////////////////////// hook functions ///////////////////////////////////////
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

// //////////////////////////////////// authentication ///////////////////////////////////////
var (
	SetupAuthProvider = auth.SetupAuthProvider

	// JWT
	NewJWTProvider = auth.NewJWTProvider

	SetAuthProvider = auth.SetGlobalAuthProvider
	GetAuthProvider = auth.GetGlobalAuthProvider
)

func SetUserDB(users []auth.User) {
	auth.SetUserDB(users)
}

// //////////////////////////////////// cms ///////////////////////////////////////
var (
	RegisterCMSBackend  = cms_backend.RegisterCMSBackend
	RegisterCMSFrontend = cms_frontend.RegisterCMSFrontend
)

func RegisterCMS(router *gin.Engine) {
	cms_backend.RegisterCMSBackend(router)
	cms_frontend.RegisterCMSFrontend(router)
}

// //////////////////////////////////// helper function ///////////////////////////////////////
func (a *App[T]) AutoTimestamp() *App[T] {

	convertTimeToVariable := func(field reflect.Value, now time.Time) {
		if field.Type().String() == "string" {
			field.SetString(now.Format(time.RFC3339))
		} else if field.Type().String() == "time.Time" {
			field.Set(reflect.ValueOf(now))
		} else if field.Type().String() == "int64" {
			field.SetInt(now.Unix())
		}
	}

	return a.
		BeforeCreate(func(obj *T) error {
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
