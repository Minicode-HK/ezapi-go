package core

import (
	"math"
	"reflect"
	"sync"

	"github.com/Minicode-HK/ezapi-go/core/feature"
)

type DBWrapper[T any] struct {
    data      *[]T
    generator feature.IDGenerator
    mu        sync.RWMutex
}

// DB Wrapper are thread-safe wrappers around in-memory database slices
// used primarily just in CRUDRouter to provide thread-safe simple operations like GetAll, GetById, Add, Update, Delete.
// for other modules, QueryBuilder might are the one you want to use.
// each module should only have one instance of DBWrapper per in-memory slice.
func NewDBWrapper[T any](data *[]T, generator feature.IDGenerator) *DBWrapper[T] {
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

type PaginateRes[T any] struct {
    Data  []T    `json:"data"`  
    Total uint   `json:"total"`  
    
    Page       uint `json:"page"`       
    PageSize   uint `json:"page_size"`   
    TotalPages uint `json:"total_pages"` 
    HasMore    bool `json:"has_more"`    
}

// paginates itself
func (db *DBWrapper[T]) Paginate(page, pageSize int) PaginateRes[T] {
    db.mu.RLock()
    defer db.mu.RUnlock()
    
    total := uint(len(*db.data))
    offset := (page - 1) * pageSize

    if offset > len(*db.data) {
        return PaginateRes[T]{
            Data:       []T{},
            Total:      total,
            Page:       uint(page),
            PageSize:   uint(pageSize),
            TotalPages: uint(math.Ceil(float64(total) / float64(pageSize))),
            HasMore:    false,
        }
    }

    end := offset + pageSize
    if end > len(*db.data) {
        end = len(*db.data)
    }

    return PaginateRes[T]{
        Data:       (*db.data)[offset:end],
        Total:      total,
        Page:       uint(page),
        PageSize:   uint(pageSize),
        TotalPages: uint(math.Ceil(float64(total) / float64(pageSize))),
        HasMore:    end < len(*db.data),
    }
}

// helper function that paginates with given data slice pointer
func (db *DBWrapper[T]) PaginateWith(data *[]T, page, pageSize int) PaginateRes[T] {
    total := uint(len(*data))
    offset := (page - 1) * pageSize
    
    if offset > len(*data) {
        return PaginateRes[T]{
            Data:       []T{},
            Total:      total,
            Page:       uint(page),
            PageSize:   uint(pageSize),
            TotalPages: uint(math.Ceil(float64(total) / float64(pageSize))),
            HasMore:    false,
        }
    }

    end := offset + pageSize
    if end > len(*data) {
        end = len(*data)
    }

    return PaginateRes[T]{
        Data:       (*data)[offset:end],
        Total:      total,
        Page:       uint(page),
        PageSize:   uint(pageSize),
        TotalPages: uint(math.Ceil(float64(total) / float64(pageSize))),
        HasMore:    end < len(*data),
    }
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
    if identifiable, ok := item.(feature.Identifiable); ok {
        return identifiable.GetId()
    }
    
    v := reflect.ValueOf(item)
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
    }

    // try both "Id" and "ID", "id"
    idField := v.FieldByName("Id")
    if !idField.IsValid() {
        idField = v.FieldByName("ID")
    }
    if !idField.IsValid() {
        idField = v.FieldByName("id")
    }
    
    if idField.IsValid() {
        return idField.String()
    }
    
    panic("Item does not have an Id field or does not implement Identifiable interface")
}

func setItemId(item any, id string) {
    if identifiable, ok := item.(feature.Identifiable); ok {
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