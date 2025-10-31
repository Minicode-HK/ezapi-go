package handlers

import (
    "encoding/json"
    "net/http"
    "reflect"

    "github.com/gin-gonic/gin"
    
    "simple_backend_go/route"
)

type FieldDefinition struct {
    Name     string `json:"name"`
    Type     string `json:"type"`
    JSONTag  string `json:"json_tag"`
    Required bool   `json:"required"`
    Rules    string `json:"rules,omitempty"`
}

type SchemaDefinition struct {
    Name     string            `json:"name"`
    Fields   []FieldDefinition `json:"fields"`
    BasePath string            `json:"base_path,omitempty"`
}

type SchemaHandler struct{}

func NewSchemaHandler() *SchemaHandler {
    return &SchemaHandler{}
}

func (h *SchemaHandler) GetSchemas(w http.ResponseWriter, r *http.Request) {
    schemas := []SchemaDefinition{}
    
    for _, module := range route.GetModuleRegistry() {
        schema := h.reflectSchema(module)
        schemas = append(schemas, schema)
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(schemas)
}

func (h *SchemaHandler) GetSchema(c *gin.Context) {
    name := c.Param("name")
    var foundSchema *SchemaDefinition
    
    for _, module := range route.GetModuleRegistry() {
        schema := h.reflectSchema(module)
        if schema.Name == name {
            foundSchema = &schema
            break
        }
    }
    
    if foundSchema == nil {
        c.JSON(http.StatusNotFound, gin.H{
            "success": false,
            "error":   "Schema not found",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":   foundSchema,
    })
}

func (h *SchemaHandler) reflectSchema(module route.ModuleInfo) SchemaDefinition {
    t := module.TypeName
    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }
    
    schema := SchemaDefinition{
        Name:     t.Name(),
        Fields:   []FieldDefinition{},
        BasePath: module.BasePath,
    }
    
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        jsonTag := field.Tag.Get("json")
        required := field.Tag.Get("validate") != ""
        
        schema.Fields = append(schema.Fields, FieldDefinition{
            Name:     field.Name,
            Type:     field.Type.String(),
            JSONTag:  jsonTag,
            Required: required,
            Rules:    field.Tag.Get("validate"),
        })
    }
    
    return schema
}