package backend

import (
	"reflect"
	"strings"

	"github.com/Minicode-HK/ezapi-go/core"
)

type FieldDef struct {
    Name     string `json:"name"`	  // Struct field name
    Key      string `json:"key"`      // JSON key
    Type     string `json:"type"`     // string, number, bool, etc.
    Required bool   `json:"required"` // based on binding tag
}

type ResourceDef struct {
    Name     string     `json:"name"`
    Endpoint string     `json:"endpoint"`
    Fields   []FieldDef `json:"fields"`
}

func GenerateSchema() []ResourceDef {
    registry := core.GetAllModuleRegistry()
    var resources []ResourceDef

    for typ, info := range registry {
        // Get the struct type (handle pointer vs struct)
        structType := typ
        if structType.Kind() == reflect.Ptr {
            structType = structType.Elem()
        }

        var fields []FieldDef
        for i := 0; i < structType.NumField(); i++ {
            field := structType.Field(i)
            jsonTag := field.Tag.Get("json")
            bindingTag := field.Tag.Get("binding")

            if jsonTag == "-" {
                continue
            }

            // Parse JSON key (e.g., "name,omitempty")
            jsonKey := strings.Split(jsonTag, ",")[0]
            if jsonKey == "" {
                jsonKey = field.Name
            }

            // Map Go types to html input types
            uiType := "text"
            switch field.Type.Kind() {
            case reflect.Int, reflect.Int64, reflect.Float64:
                uiType = "number"
            case reflect.Bool:
                uiType = "checkbox"
            }
            
            fields = append(fields, FieldDef{
                Name:     field.Name,
                Key:      jsonKey,
                Type:     uiType,
                Required: strings.Contains(bindingTag, "required"),
            })
        }

        resources = append(resources, ResourceDef{
            Name:     structType.Name(),
            Endpoint: info.BasePath,
            Fields:   fields,
        })
    }

    return resources
}