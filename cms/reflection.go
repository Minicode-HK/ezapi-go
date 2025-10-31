package cms

import (
    "encoding/json"
    "net/http"
    "reflect"

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
    Name   string            `json:"name"`
    Fields []FieldDefinition `json:"fields"`
    BasePath string           `json:"base_path,omitempty"`
}

func ReflectSchema(model interface{}) SchemaDefinition {
    t := reflect.TypeOf(model)
    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }

    // find the base path from module registry
    var basePath string
    for _, module := range route.GetModuleRegistry() {
        if module.TypeName == t {
            basePath = module.BasePath
            break
        }
    }

    schema := SchemaDefinition{
        Name:   t.Name(),
        Fields: []FieldDefinition{},
        BasePath: basePath,
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

func GetSchemaHandler(w http.ResponseWriter, r *http.Request) {
    schemas := []SchemaDefinition{}
    

    if len(schemas) == 0 {
        // If no modules registered, reflect all from route registry
        for _, module := range route.GetModuleRegistry() {
            schemas = append(schemas, ReflectSchema(reflect.New(module.TypeName).Interface()))
        }
    } else {
        for _, module := range Modules {
            schemas = append(schemas, ReflectSchema(module))
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(schemas)
}