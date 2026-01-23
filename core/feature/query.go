package feature

import (
	"math"
	"reflect"
	"sync"
)


type QueryBuilder[T any] struct {
	originalSlice *[]T // a pointer to the original slice

	workSet []*T
	mu sync.RWMutex
}


func NewQueryBuilder[T any](data *[]T) *QueryBuilder[T] {
    pointers := make([]*T, len(*data))
    for i := range *data {
        pointers[i] = &(*data)[i]
    }

    return &QueryBuilder[T]{
        originalSlice: data,  
        workSet: pointers,
        mu:      sync.RWMutex{},
    }
}

func (qb *QueryBuilder[T]) Get() []*T {
	qb.mu.RLock()
	defer qb.mu.RUnlock()
	return qb.workSet
}

func (qb *QueryBuilder[T]) Delete() []*T  {
    qb.mu.Lock()
    defer qb.mu.Unlock()

    workSetMap := make(map[uintptr]bool)
    for _, item := range qb.workSet {
        workSetMap[reflect.ValueOf(item).Pointer()] = true
    }
    
    var deleted []*T
    var remaining []T  
    
    for i := range *qb.originalSlice {
		item := &(*qb.originalSlice)[i]

        if workSetMap[reflect.ValueOf(item).Pointer()] {
            deleted = append(deleted, item)
        } else {
            remaining = append(remaining, *item)  
        }
    }
    
    *qb.originalSlice = remaining 
    qb.workSet = []*T{}
    
    return deleted
}

func (qb *QueryBuilder[T]) First() *T {
	qb.mu.RLock()
	defer qb.mu.RUnlock()

	if len(qb.workSet) == 0 {
		return nil
	}
	return qb.workSet[0]
}

func (qb *QueryBuilder[T]) Filter(filterFunc func(*T) bool) *QueryBuilder[T] {
	qb.mu.Lock()
	defer qb.mu.Unlock()

	var filtered []*T
	for i := range qb.workSet {
		if filterFunc(qb.workSet[i]) {
			filtered = append(filtered, qb.workSet[i])
		}
	}
	qb.workSet = filtered
	return qb
}

func (qb *QueryBuilder[T]) Limit(n int) *QueryBuilder[T] {
	qb.mu.Lock()
	defer qb.mu.Unlock()

	if n < len(qb.workSet) {
		qb.workSet = qb.workSet[:n]
	}
	return qb
}

func (qb *QueryBuilder[T]) Offset(n int) *QueryBuilder[T] {
	qb.mu.Lock()
	defer qb.mu.Unlock()

	if n < len(qb.workSet) {
		qb.workSet = qb.workSet[n:]
	} else {
		qb.workSet = []*T{}
	}
	return qb
}

func (qb *QueryBuilder[T]) WhereWith(predicate func(*T) bool) *QueryBuilder[T] {
	qb.mu.Lock()
	defer qb.mu.Unlock()
	
	var filtered []*T
	for i := range qb.workSet {
		if predicate(qb.workSet[i]) {
			filtered = append(filtered, qb.workSet[i])
		}
	}
	qb.workSet = filtered
	return qb
}

func (qb *QueryBuilder[T]) Where(fieldName string, operator string, value any) *QueryBuilder[T] {
	qb.mu.Lock()
	defer qb.mu.Unlock()

	var filtered []*T
	for i := range qb.workSet {
		item := qb.workSet[i]
		for j := 0; j < reflect.TypeOf(*item).NumField(); j++ {
			field := reflect.TypeOf(*item).Field(j)
			if field.Name == fieldName {
				fieldValue := reflect.ValueOf(*item).FieldByName(fieldName).Interface()
				match := false
				switch operator {
				case "=":
					match = fieldValue == value
				case "==":
					match = fieldValue == value
				case "!=":
					match = fieldValue != value
				case ">":
					match = reflect.ValueOf(fieldValue).Float() > reflect.ValueOf(value).Float()
				case "<":
					match = reflect.ValueOf(fieldValue).Float() < reflect.ValueOf(value).Float()
				case ">=":
					match = reflect.ValueOf(fieldValue).Float() >= reflect.ValueOf(value).Float()
				case "<=":
					match = reflect.ValueOf(fieldValue).Float() <= reflect.ValueOf(value).Float()
				}
				if match {
					filtered = append(filtered, item)
				}
				break
			}
		}
	}
	qb.workSet = filtered
	return qb
}

func (qb *QueryBuilder[T]) Count() int {
	qb.mu.RLock()
	defer qb.mu.RUnlock()
	return len(qb.workSet)
}

func compareValues(a, b reflect.Value) int {
	switch a.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intA := a.Int()
		intB := b.Int()
		if intA < intB {
			return -1
		} else if intA > intB {
			return 1
		} else {
			return 0
		}
	case reflect.Float32, reflect.Float64:
		floatA := a.Float()
		floatB := b.Float()
		if math.Abs(floatA-floatB) < 1e-9 {
			return 0
		} else if floatA < floatB {
			return -1
		} else {
			return 1
		}
	case reflect.String:
		strA := a.String()
		strB := b.String()
		if strA < strB {
			return -1
		} else if strA > strB {
			return 1
		} else {
			return 0
		}
	default:
		return 0
	}
}		

func (qb *QueryBuilder[T]) OrderBy(fieldName string, ascending ...bool) *QueryBuilder[T] {
	qb.mu.Lock()
	defer qb.mu.Unlock()

	asc := true
	if len(ascending) > 0 {
		asc = ascending[0]
	}

	less := func(i, j int) bool {
		itemI := qb.workSet[i]
		itemJ := qb.workSet[j]
		fieldValueI := reflect.ValueOf(*itemI).FieldByName(fieldName)
		fieldValueJ := reflect.ValueOf(*itemJ).FieldByName(fieldName)

		if asc {
			return compareValues(fieldValueI, fieldValueJ) < 0
		} else {
			return compareValues(fieldValueI, fieldValueJ) > 0
		}
	}

	n := len(qb.workSet)
	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			if !less(j, j+1) {
				qb.workSet[j], qb.workSet[j+1] = qb.workSet[j+1], qb.workSet[j]
			}
		}
	}

	return qb
}

func (qb *QueryBuilder[T]) Select(fields ...string) *QueryBuilder[map[string]any] {
	qb.mu.Lock()
	defer qb.mu.Unlock()

	var selected []map[string]any

	for i := range qb.workSet {
		item := qb.workSet[i]
		record := make(map[string]any)
		val := reflect.ValueOf(*item)
		for _, fieldName := range fields {
			fieldVal := val.FieldByName(fieldName)
			if fieldVal.IsValid() {
				structField, ok := reflect.TypeOf(*item).FieldByName(fieldName)
				if ok {
					tag := structField.Tag.Get("json")
					if tag != "" {
						record[tag] = fieldVal.Interface()
					} else {
						record[fieldName] = fieldVal.Interface()
					}
				} else {
					record[fieldName] = fieldVal.Interface()
				}
			}
		}
		selected = append(selected, record)
	}
	
	return &QueryBuilder[map[string]any]{
		workSet: convertToPointers(selected),
		mu:      sync.RWMutex{},
	}
}

// Helper function for Select
func convertToPointers[T any](slice []T) []*T {
	result := make([]*T, len(slice))
	for i := range slice {
		result[i] = &slice[i]
	}
	return result
}