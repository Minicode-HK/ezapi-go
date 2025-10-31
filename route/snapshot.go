package route

import (
    "encoding/json"
    "reflect"
)

// CreateSnapshot creates a snapshot of all registered databases
func CreateSnapshot() map[string]interface{} {
    snapshot := make(map[string]interface{})
    
    for _, module := range moduleRegistry {
        // Get the database slice
        dbValue := reflect.ValueOf(module.Data)
        if dbValue.Kind() == reflect.Ptr {
            dbValue = dbValue.Elem()
        }
        
        // Create a deep copy
        if dbValue.Kind() == reflect.Slice {
            // Serialize to JSON and back for deep copy
            jsonData, err := json.Marshal(dbValue.Interface())
            if err != nil {
                continue
            }
            
            // Store the JSON data
            snapshot[module.TypeName.Name()] = string(jsonData)
        }
    }
    
    return snapshot
}

// RestoreSnapshot restores all databases from a snapshot
func RestoreSnapshot(snapshot map[string]interface{}) error {
    for _, module := range moduleRegistry {
        moduleName := module.TypeName.Name()
        
        // Get snapshot data for this module
        jsonDataStr, ok := snapshot[moduleName].(string)
        if !ok {
            continue
        }
        
        // Get the database pointer
        dbValue := reflect.ValueOf(module.Data)
        if dbValue.Kind() != reflect.Ptr {
            continue
        }
        
        // Get the slice type
        sliceType := dbValue.Elem().Type()
        
        // Create new slice
        newSlice := reflect.New(sliceType).Interface()
        
        // Unmarshal into new slice
        if err := json.Unmarshal([]byte(jsonDataStr), newSlice); err != nil {
            continue
        }
        
        // Set the database to the restored data
        dbValue.Elem().Set(reflect.ValueOf(newSlice).Elem())
    }
    
    return nil
}