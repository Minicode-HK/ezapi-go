package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "reflect"
    
    "github.com/gin-gonic/gin"
    
    "ezapi-go/cms/backend/mock"
    "ezapi-go/core"
)

type MockDataHandler struct {
    generator *mockdata.Generator
}

func NewMockDataHandler() *MockDataHandler {
    return &MockDataHandler{
        generator: mockdata.NewGenerator(),
    }
}

type GenerateRequest struct {
    ModuleName string `json:"moduleName" binding:"required"`
    Count      int    `json:"count" binding:"required,min=1,max=1000"`
}

// Helper function to find module by name
func findModuleByName(moduleName string) *core.ModuleInfo {
    modules := core.GetModuleRegistry()
    for _, m := range modules {
        if m.TypeName.Name() == moduleName {
            return &m
        }
    }
    return nil
}

func (h *MockDataHandler) Generate(c *gin.Context) {
    var req GenerateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "error":   "Invalid request: " + err.Error(),
        })
        return
    }

    // Get module info
    module := findModuleByName(req.ModuleName)
    if module == nil {
        c.JSON(http.StatusNotFound, gin.H{
            "success": false,
            "error":   "Module not found: " + req.ModuleName,
        })
        return
    }

    // Get the database pointer and element type
    dbValue := reflect.ValueOf(module.Data)
    if dbValue.Kind() != reflect.Ptr {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "error":   "Module DB is not a pointer",
        })
        return
    }

    dbSlice := dbValue.Elem()
    if dbSlice.Kind() != reflect.Slice {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "error":   "Module DB is not a slice",
        })
        return
    }

    elemType := dbSlice.Type().Elem()

    // Generate records
    generatedRecords := make([]interface{}, 0, req.Count)
    
    for i := 0; i < req.Count; i++ {
        // Create new instance
        newItem := reflect.New(elemType).Elem()

        // Fill fields
        for j := 0; j < elemType.NumField(); j++ {
            field := elemType.Field(j)
            fieldValue := newItem.Field(j)

            if !fieldValue.CanSet() {
                continue
            }

            // Generate value based on field name and type
            mockValue := h.generator.GenerateValue(field.Name, field.Type)

            // Set value
            val := reflect.ValueOf(mockValue)
            if val.Type().ConvertibleTo(fieldValue.Type()) {
                fieldValue.Set(val.Convert(fieldValue.Type()))
            }
        }

        // Append to database
        dbSlice.Set(reflect.Append(dbSlice, newItem))
        generatedRecords = append(generatedRecords, newItem.Interface())
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": fmt.Sprintf("Generated %d records", req.Count),
        "data": gin.H{
            "count": req.Count,
            "total": dbSlice.Len(),
        },
    })
}

func (h *MockDataHandler) Preview(c *gin.Context) {
    moduleName := c.Param("module")

    // Get module info
    module := findModuleByName(moduleName)
    if module == nil {
        c.JSON(http.StatusNotFound, gin.H{
            "success": false,
            "error":   "Module not found: " + moduleName,
        })
        return
    }

    // Get the database pointer and element type
    dbValue := reflect.ValueOf(module.Data)
    if dbValue.Kind() != reflect.Ptr {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "error":   "Module DB is not a pointer",
        })
        return
    }

    dbSlice := dbValue.Elem()
    if dbSlice.Kind() != reflect.Slice {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "error":   "Module DB is not a slice",
        })
        return
    }

    elemType := dbSlice.Type().Elem()

    // Generate one sample record
    sample := reflect.New(elemType).Elem()

    // Fill fields
    for i := 0; i < elemType.NumField(); i++ {
        field := elemType.Field(i)
        fieldValue := sample.Field(i)

        if !fieldValue.CanSet() {
            continue
        }

        mockValue := h.generator.GenerateValue(field.Name, field.Type)
        val := reflect.ValueOf(mockValue)
        if val.Type().ConvertibleTo(fieldValue.Type()) {
            fieldValue.Set(val.Convert(fieldValue.Type()))
        }
    }

    // Convert to JSON
    jsonData, err := json.Marshal(sample.Interface())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "error":   "Failed to marshal sample: " + err.Error(),
        })
        return
    }

    var result map[string]interface{}
    if err := json.Unmarshal(jsonData, &result); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "error":   "Failed to unmarshal sample: " + err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    result,
    })
}