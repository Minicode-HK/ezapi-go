package core

import (
    "sync"
    "fmt"
    "reflect"

    "github.com/google/uuid"
)

type DBWrapper[T any] struct {
    data      *[]T
    generator IDGenerator
    mu        sync.RWMutex
}

type Identifiable interface {
    GetId() string
    SetId(id string)
}

type IDGenerator interface {
    GenerateNewId() string
}

type UUIDGenerator struct{}

func (g *UUIDGenerator) GenerateNewId() string {
    return uuid.New().String()
}

type IncrementalGenerator struct {
    counter int
}

func (g *IncrementalGenerator) GenerateNewId() string {
    g.counter++
    return fmt.Sprintf("%d", g.counter)
}

func NewDBWrapper[T any](data *[]T, generator IDGenerator) *DBWrapper[T] {
    return &DBWrapper[T]{
        data:      data,
        generator: generator,
        mu:        sync.RWMutex{},
    }
}

func (db *DBWrapper[T]) GetAll() []T {
    db.mu.RLock()
    defer db.mu.RUnlock()
    dataCopy := make([]T, len(*db.data))
    copy(dataCopy, *db.data)
    return dataCopy
}

func (db *DBWrapper[T]) GetCount() int {
    db.mu.RLock()
    defer db.mu.RUnlock()
    return len(*db.data)
}

func (db *DBWrapper[T]) GetById(id string) *T {
    db.mu.RLock()
    defer db.mu.RUnlock()

    for i := range *db.data {
        itemId := getItemId(&(*db.data)[i])  
        if itemId == id {
            return &(*db.data)[i]  
        }
    }
    return nil
}

func (db *DBWrapper[T]) PaginateWith(data *[] T, offset, limit int) []T {

	if offset > len(*data) {
		return []T{}
	}

	end := offset + limit
	if end > len(*data) {
		end = len(*data)
	}

	return (*data)[offset:end]
}


func (db *DBWrapper[T]) Add(item *T) bool {
    db.mu.Lock()
    defer db.mu.Unlock()

    if db.generator != nil {
        id := db.generator.GenerateNewId()
        setItemId(item, id) 
    }

    *db.data = append(*db.data, *item)
    return true
}

func (db *DBWrapper[T]) Update(id string, updatedItem *T) bool {
    db.mu.Lock()
    defer db.mu.Unlock()

    for i, item := range *db.data {
        itemId := getItemId(item)  
        if itemId == id {
            setItemId(updatedItem, id)  
			(*db.data)[i] = *updatedItem
            return true
        }
    }
    return false
}

func (db *DBWrapper[T]) Delete(id string) bool {
    db.mu.Lock()
    defer db.mu.Unlock()

    for i, item := range *db.data {
        itemId := getItemId(&item) 
        if itemId == id {
            *db.data = append((*db.data)[:i], (*db.data)[i+1:]...)
            return true
        }
    }
    return false
}

func getItemId(item any) string {
    if identifiable, ok := item.(Identifiable); ok {
        return identifiable.GetId()
    }
    
    idField := reflect.ValueOf(item).Elem().FieldByName("Id")
    if idField.IsValid() {
        return idField.String()
    }
    return ""
}

func setItemId(item any, id string) {
    if identifiable, ok := item.(Identifiable); ok {
        identifiable.SetId(id)
        return
    }

    val := reflect.ValueOf(item)
    if val.Kind() != reflect.Ptr {
        return
    }
    
    val = val.Elem()
    idField := val.FieldByName("Id")
    if idField.IsValid() && idField.CanSet() {
        idField.SetString(id)
    }
}