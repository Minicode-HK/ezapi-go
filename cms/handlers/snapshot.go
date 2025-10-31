package handlers

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "reflect"
    "time"
    
    "simple_backend_go/route"
    "simple_backend_go/cms/models"
    
    "github.com/gin-gonic/gin"
)

type SnapshotMetadata struct {
    Timestamp time.Time `json:"timestamp"`
    Modules   []string  `json:"modules"`
    Version   string    `json:"version"`
}

type Snapshot struct {
    Metadata SnapshotMetadata       `json:"metadata"`
    Data     map[string]interface{} `json:"data"`
}


type SnapshotHandler struct {
    snapshotDir string
}

func NewSnapshotHandler(snapshotDir string) *SnapshotHandler {
    // Create snapshot directory if not exists
    os.MkdirAll(snapshotDir, 0755)
    
    return &SnapshotHandler{
        snapshotDir: snapshotDir,
    }
}

func (h *SnapshotHandler) GetModules(c *gin.Context) {
    modules := []map[string]interface{}{}
    
    for _, module := range route.GetModuleRegistry() {
        db := module.Data
        count := reflect.ValueOf(db).Elem().Len()
        
        modules = append(modules, map[string]interface{}{
            "name":      module.TypeName.Name(),
            "base_path": module.BasePath,
            "count":     count,
        })
    }
    
    c.JSON(200, models.APIResponse{
        Success: true,
        Data:    modules,
    })
}

func (h *SnapshotHandler) Save(c *gin.Context) {
    var req struct {
        Modules  []string `json:"modules" binding:"required"`
        Filename string   `json:"filename"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, models.APIResponse{
            Success: false,
            Message: "Invalid request",
        })
        return
    }
    
    // Create snapshot
    snapshot := Snapshot{
        Metadata: SnapshotMetadata{
            Timestamp: time.Now(),
            Modules:   req.Modules,
            Version:   "1.0",
        },
        Data: make(map[string]interface{}),
    }
    
    // Collect data from selected modules
    registry := route.GetModuleRegistry()
    for _, moduleName := range req.Modules {
        found := false
        for _, module := range registry {
            if module.TypeName.Name() == moduleName {
                dbValue := reflect.ValueOf(module.Data).Elem()
                snapshot.Data[moduleName] = dbValue.Interface()
                found = true
                break
            }
        }
        
        if !found {
            c.JSON(400, models.APIResponse{
                Success: false,
                Message: fmt.Sprintf("Module '%s' not found", moduleName),
            })
            return
        }
    }
    
    // Generate filename
    filename := req.Filename
    if filename == "" {
        filename = fmt.Sprintf("snapshot_%s.json", time.Now().Format("20060102_150405"))
    }
    if filepath.Ext(filename) != ".json" {
        filename += ".json"
    }
    
    filePath := filepath.Join(h.snapshotDir, filename)
    
    // Save to file
    file, err := os.Create(filePath)
    if err != nil {
        c.JSON(500, models.APIResponse{
            Success: false,
            Message: "Failed to create snapshot file",
        })
        return
    }
    defer file.Close()
    
    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(snapshot); err != nil {
        c.JSON(500, models.APIResponse{
            Success: false,
            Message: "Failed to write snapshot",
        })
        return
    }
    
    c.JSON(200, models.APIResponse{
        Success: true,
        Message: "Snapshot saved successfully",
        Data: gin.H{
            "filename": filename,
            "path":     filePath,
            "modules":  req.Modules,
        },
    })
}

func (h *SnapshotHandler) List(c *gin.Context) {
    files, err := os.ReadDir(h.snapshotDir)
    if err != nil {
        if os.IsNotExist(err) {
            c.JSON(200, models.APIResponse{
                Success: true,
                Data:    []interface{}{},
            })
            return
        }
        c.JSON(500, models.APIResponse{
            Success: false,
            Message: "Failed to read snapshots",
        })
        return
    }
    
    snapshots := []map[string]interface{}{}
    for _, file := range files {
        if filepath.Ext(file.Name()) == ".json" {
            info, _ := file.Info()
            
            // Try to read metadata
            filePath := filepath.Join(h.snapshotDir, file.Name())
            var snapshot Snapshot
            if data, err := os.ReadFile(filePath); err == nil {
                json.Unmarshal(data, &snapshot)
            }
            
            snapshots = append(snapshots, map[string]interface{}{
                "filename":  file.Name(),
                "size":      info.Size(),
                "modified":  info.ModTime(),
                "modules":   snapshot.Metadata.Modules,
                "timestamp": snapshot.Metadata.Timestamp,
            })
        }
    }

    // sort snapshots by modified time descending
    for i := 0; i < len(snapshots)-1; i++ {
        for j := i + 1; j < len(snapshots); j++ {
            timeI := snapshots[i]["modified"].(time.Time)
            timeJ := snapshots[j]["modified"].(time.Time)
            if timeJ.After(timeI) {
                snapshots[i], snapshots[j] = snapshots[j], snapshots[i]
            }
        }
    }
    
    c.JSON(200, models.APIResponse{
        Success: true,
        Data:    snapshots,
    })
}

func (h *SnapshotHandler) Load(c *gin.Context) {
    var req struct {
        Filename string   `json:"filename" binding:"required"`
        Modules  []string `json:"modules"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, models.APIResponse{
            Success: false,
            Message: "Invalid request",
        })
        return
    }
    
    filePath := filepath.Join(h.snapshotDir, req.Filename)
    
    // Read snapshot file
    data, err := os.ReadFile(filePath)
    if err != nil {
        c.JSON(404, models.APIResponse{
            Success: false,
            Message: "Snapshot not found",
        })
        return
    }
    
    var snapshot Snapshot
    if err := json.Unmarshal(data, &snapshot); err != nil {
        c.JSON(400, models.APIResponse{
            Success: false,
            Message: "Invalid snapshot format",
        })
        return
    }
    
    // Determine which modules to load
    modulesToLoad := req.Modules
    if len(modulesToLoad) == 0 {
        modulesToLoad = snapshot.Metadata.Modules
    }
    
    // Load data into databases
    registry := route.GetModuleRegistry()
    loadedModules := []string{}
    
    for _, moduleName := range modulesToLoad {
        if snapshotData, exists := snapshot.Data[moduleName]; exists {
            for _, module := range registry {
                if module.TypeName.Name() == moduleName {
                    snapshotBytes, _ := json.Marshal(snapshotData)
                    
                    dbValue := reflect.ValueOf(module.Data).Elem()
                    newSlice := reflect.New(dbValue.Type()).Interface()
                    
                    if err := json.Unmarshal(snapshotBytes, newSlice); err != nil {
                        c.JSON(500, models.APIResponse{
                            Success: false,
                            Message: fmt.Sprintf("Failed to load module '%s': %v", moduleName, err),
                        })
                        return
                    }
                    
                    dbValue.Set(reflect.ValueOf(newSlice).Elem())
                    loadedModules = append(loadedModules, moduleName)
                    break
                }
            }
        }
    }
    
    c.JSON(200, models.APIResponse{
        Success: true,
        Message: "Snapshot loaded successfully",
        Data: gin.H{
            "loaded": loadedModules,
        },
    })
}

func (h *SnapshotHandler) Delete(c *gin.Context) {
    filename := c.Param("filename")
    filePath := filepath.Join(h.snapshotDir, filename)
    
    if err := os.Remove(filePath); err != nil {
        c.JSON(404, models.APIResponse{
            Success: false,
            Message: "Snapshot not found",
        })
        return
    }
    
    c.JSON(200, models.APIResponse{
        Success: true,
        Message: "Snapshot deleted successfully",
    })
}