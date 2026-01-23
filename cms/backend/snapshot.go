package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/Minicode-HK/ezapi-go/core"
	"github.com/Minicode-HK/ezapi-go/core/feature"
	core_http "github.com/Minicode-HK/ezapi-go/core/http"
	"github.com/gin-gonic/gin"
)

const snapshotDir = "./snapshots"

type SnapshotResponse struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	Modules []string `json:"modules"`

	CreatedAt   string  `json:"created_at"`
    UpdatedAt   string  `json:"updated_at"`
}

var snapshotResponseDB []SnapshotResponse


func readFSSnapshots() []SnapshotResponse {
	var snapshots []SnapshotResponse
	
	// if dir not exist, create it
	if _, err := os.Stat(snapshotDir); os.IsNotExist(err) {
		os.MkdirAll(snapshotDir, 0755)
		return snapshots
	}

	entries, err := os.ReadDir(snapshotDir)
	if err != nil {
		return snapshots
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			// only load .json
			if !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			snapshots = append(snapshots, SnapshotResponse{
				Id: entry.Name(),
				Name: entry.Name(),
				Size: info.Size(),
				CreatedAt: info.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
	}

	return snapshots
}

type DateIdGenerator struct {
}

func (g *DateIdGenerator) GenerateNewId() string {
	return fmt.Sprintf("%d",  time.Now().UnixNano() / int64(time.Millisecond))
}

func createSnapshotFile(snapshotId string, name string) (*SnapshotResponse, error) {
    if err := os.MkdirAll(snapshotDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create snapshot directory: %w", err)
    }

    snapshotData := make(map[string]interface{})
    modules := []string{}

    registry := core.GetAllModuleRegistry()
    for typeObj, info := range registry {
        moduleName := strings.TrimPrefix(info.BasePath, "/")
        if moduleName == "" {
            moduleName = typeObj.Name()
        }

        val := reflect.ValueOf(info.Data)
        if val.Kind() == reflect.Ptr {
            val = val.Elem()
        }
        
        snapshotData[moduleName] = val.Interface()
        modules = append(modules, moduleName)
    }

    jsonData, err := json.MarshalIndent(snapshotData, "", "  ")
    if err != nil {
        return nil, fmt.Errorf("failed to marshal snapshot data: %w", err)
    }

    fileName := snapshotId + ".json"
    filePath := snapshotDir + "/" + fileName
    if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
        return nil, fmt.Errorf("failed to write snapshot file: %w", err)
    }

    info, err := os.Stat(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to stat snapshot file: %w", err)
    }

    now := time.Now().Format("2006-01-02 15:04:05")
    return &SnapshotResponse{
        Id:        snapshotId,
        Name:      name,
        Size:      info.Size(),
        Modules:   modules,
        CreatedAt: now,
        UpdatedAt: now,
    }, nil
}

// restoreSnapshotFile restores data from a snapshot file
func restoreSnapshotFile(snapshotId string) error {
    filePath := snapshotDir + "/" + snapshotId 

    data, err := os.ReadFile(filePath)
    if err != nil {
        return fmt.Errorf("failed to read snapshot file: %w", err)
    }

    var snapshotData map[string]json.RawMessage
    if err := json.Unmarshal(data, &snapshotData); err != nil {
        return fmt.Errorf("failed to parse snapshot data: %w", err)
    }

    registry := core.GetAllModuleRegistry()
    
    dbMap := make(map[string]interface{})
    for typeObj, info := range registry {
        name := strings.TrimPrefix(info.BasePath, "/")
        if name == "" {
            name = typeObj.Name()
        }
        dbMap[name] = info.Data
    }

    for moduleName, rawData := range snapshotData {
        dbPtr, exists := dbMap[moduleName]
        if !exists {
            continue // Skip databases that no longer exist
        }

        // Get the type of the database slice
        val := reflect.ValueOf(dbPtr)
        if val.Kind() != reflect.Ptr {
            continue
        }

        // Create a new slice of the same type
        sliceType := val.Elem().Type()
        newSlice := reflect.New(sliceType)

        // Unmarshal into the new slice
        if err := json.Unmarshal(rawData, newSlice.Interface()); err != nil {
            return fmt.Errorf("failed to restore %s: %w", moduleName, err)
        }

        // Set the database to the new data
        val.Elem().Set(newSlice.Elem())
    }

    return nil
}


func SetupSnapshots(router *gin.Engine) {
    snapshotResponseDB = readFSSnapshots()

    // List snapshots
    router.GET("/cms/api/snapshots", func(c *gin.Context) {
        snapshotResponseDB = readFSSnapshots()
        core_http.SendSuccess(c, snapshotResponseDB)
    })

    // Get single snapshot
    router.GET("/cms/api/snapshots/:id/content", func(c *gin.Context) {
        id := c.Param("id")
        // Basic security check
        if strings.Contains(id, "..") || strings.Contains(id, "/") || strings.Contains(id, "\\") {
            core_http.SendError(c, 400, "Invalid snapshot ID")
            return
        }

        filePath := snapshotDir + "/" + id
        if _, err := os.Stat(filePath); os.IsNotExist(err) {
            core_http.SendError(c, 404, "Snapshot file not found")
            return
        }

        // Serve the JSON file directly
        c.File(filePath)
    })

    // Create snapshot
    router.POST("/cms/api/snapshots", func(c *gin.Context) {
        var req struct {
            Name string `json:"name"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            req.Name = time.Now().Format("2006-01-02_15-04-05")
        }
        if req.Name == "" {
            req.Name = time.Now().Format("2006-01-02_15-04-05")
        }

        gen := &DateIdGenerator{}
        id := gen.GenerateNewId()

        snapshot, err := createSnapshotFile(id, req.Name)
        if err != nil {
            core_http.SendError(c, 500, err.Error())
            return
        }

        snapshotResponseDB = append(snapshotResponseDB, *snapshot)
        core_http.SendSuccess(c, snapshot)
    })

    // Delete snapshot
    router.DELETE("/cms/api/snapshots/:id", func(c *gin.Context) {
        id := c.Param("id")
        snap := feature.NewQueryBuilder(&snapshotResponseDB).Where("Id", "=", id).First()
        if snap == nil {
            core_http.SendError(c, 404, "Snapshot not found")
            return
        }
        os.Remove(snapshotDir + "/" + id )
        core_http.SendSuccess(c, gin.H{"message": "Snapshot deleted"})
    })

    // Restore snapshot
    router.POST("/cms/api/snapshots/:id/restore", func(c *gin.Context) {
        id := c.Param("id")
        if err := restoreSnapshotFile(id); err != nil {
            core_http.SendError(c, 500, err.Error())
            return
        }
        core_http.SendSuccess(c, gin.H{"message": "Snapshot restored successfully"})
    })
}