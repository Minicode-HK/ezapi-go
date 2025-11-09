package feature

import (
    "fmt"

    "github.com/google/uuid"
)

type IDGenerator interface {
    GenerateNewId() string
}

type UUIDGenerator struct{}

func (g *UUIDGenerator) GenerateNewId() string {
    return uuid.New().String()
}

type IncrementalGenerator struct {
    Counter int
}

func (g *IncrementalGenerator) GenerateNewId() string {
    g.Counter++
    return fmt.Sprintf("%d", g.Counter)
}