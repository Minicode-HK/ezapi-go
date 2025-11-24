package feature

import "reflect"

type BeforeCreateFunc[T any] func(obj *T) error
type AfterCreateFunc[T any] func(obj *T) error

type BeforeUpdateFunc[T any] func(obj *T, requestObj *T) error
type AfterUpdateFunc[T any] func(obj *T) error

type BeforeDeleteFunc[T any] func(obj *T) error
type AfterDeleteFunc[T any] func(obj *T) error


type HookEntry[T any] struct {
	BeforeCreateHooks []BeforeCreateFunc[T]
	AfterCreateHooks  []AfterCreateFunc[T]

	BeforeUpdateHooks []BeforeUpdateFunc[T]
	AfterUpdateHooks  []AfterUpdateFunc[T]

	BeforeDeleteHooks []BeforeDeleteFunc[T]
	AfterDeleteHooks  []AfterDeleteFunc[T]
}

// HookRegistry will be invoked when `route` got invoked. not db_wrapper
type HookRegistry struct {
	hooks map[reflect.Type]HookEntry[any]
}

var routerRegistry = &HookRegistry{
	hooks: make(map[reflect.Type]HookEntry[any]),
}

func AddBeforeCreateHook(t reflect.Type, hook BeforeCreateFunc[any]) {
    entry, exists := routerRegistry.hooks[t]
    if !exists {
        entry = HookEntry[any]{}
    }
    entry.BeforeCreateHooks = append(entry.BeforeCreateHooks, hook)
    routerRegistry.hooks[t] = entry
}

func AddAfterCreateHook(t reflect.Type, hook AfterCreateFunc[any]) {
    entry, exists := routerRegistry.hooks[t]
    if !exists {
        entry = HookEntry[any]{}
    }
    entry.AfterCreateHooks = append(entry.AfterCreateHooks, hook)
    routerRegistry.hooks[t] = entry
}

func AddBeforeUpdateHook(t reflect.Type, hook BeforeUpdateFunc[any]) {
    entry, exists := routerRegistry.hooks[t]
    if !exists {
        entry = HookEntry[any]{}
    }
    entry.BeforeUpdateHooks = append(entry.BeforeUpdateHooks, hook)
    routerRegistry.hooks[t] = entry
}

func AddAfterUpdateHook(t reflect.Type, hook AfterUpdateFunc[any]) {
    entry, exists := routerRegistry.hooks[t]
    if !exists {
        entry = HookEntry[any]{}
    }
    entry.AfterUpdateHooks = append(entry.AfterUpdateHooks, hook)
    routerRegistry.hooks[t] = entry
}

func AddBeforeDeleteHook(t reflect.Type, hook BeforeDeleteFunc[any]) {
    entry, exists := routerRegistry.hooks[t]
    if !exists {
        entry = HookEntry[any]{}
    }
    entry.BeforeDeleteHooks = append(entry.BeforeDeleteHooks, hook)
    routerRegistry.hooks[t] = entry
}

func AddAfterDeleteHook(t reflect.Type, hook AfterDeleteFunc[any]) {
    entry, exists := routerRegistry.hooks[t]
    if !exists {
        entry = HookEntry[any]{}
    }
    entry.AfterDeleteHooks = append(entry.AfterDeleteHooks, hook)
    routerRegistry.hooks[t] = entry
}

func GetHooksForType(t reflect.Type) ( *HookEntry[any] , bool ) { 
	hooks, exists := routerRegistry.hooks[t]
	if !exists {
		return &HookEntry[any]{}, false
	}
	return &hooks, true
}

