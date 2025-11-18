package feature

import (
	"math"
	"reflect"
	"sync"
)


type QueryBuilder[T any] struct {
	workSet []T
	mu sync.RWMutex
}


func NewQueryBuilder[T any](data []T) *QueryBuilder[T] {
	return &QueryBuilder[T]{
		workSet: data,
		mu:      sync.RWMutex{},
	}
}

func (qb *QueryBuilder[T]) Get() []T {
	qb.mu.RLock()
	defer qb.mu.RUnlock()
	return qb.workSet
}

func (qb *QueryBuilder[T]) Filter(filterFunc func(*T) bool) *QueryBuilder[T] {
	qb.mu.Lock()
	defer qb.mu.Unlock()

	var filtered []T
	for i := range qb.workSet {
		if filterFunc(&qb.workSet[i]) {
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
		qb.workSet = []T{}
	}
	return qb
}

func (qb *QueryBuilder[T]) WhereWith(predicate func(*T) bool) *QueryBuilder[T] {
	qb.mu.Lock()
	defer qb.mu.Unlock()
	
	var filtered []T
	for i := range qb.workSet {
		if predicate(&qb.workSet[i]) {
			filtered = append(filtered, qb.workSet[i])
		}
	}
	qb.workSet = filtered
	return qb
}

func (qb *QueryBuilder[T]) Where(fieldName string, operator string, value any) *QueryBuilder[T] {
	qb.mu.Lock()
	defer qb.mu.Unlock()

	var filtered []T
	for i := range qb.workSet {
		item := &qb.workSet[i]
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
					filtered = append(filtered, *item)
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
		itemI := &qb.workSet[i]
		itemJ := &qb.workSet[j]
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
