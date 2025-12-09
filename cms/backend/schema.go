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
            case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
                reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
                reflect.Float32, reflect.Float64:
                uiType = "number"
            case reflect.Bool:
                uiType = "checkbox"
            case reflect.Slice, reflect.Array:
                // TODO: maybe display a table
                uiType = "text"
            case reflect.Struct, reflect.Map, reflect.Interface:
                // TODO: recursive schema generation?
                if field.Type.String() == "time.Time" {
                    uiType = "datetime-local"
                } else {
                    uiType = "textarea"
                }
            case reflect.String:
                if field.Type.Name() == "string" && strings.Contains(strings.ToLower(field.Name), "email") {
                    uiType = "email"
                }
            }

            // or based on `cms:"{type}"` tag
            cmsTag := field.Tag.Get("cms")
            if cmsTag != "" {
                uiType = cmsTag
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