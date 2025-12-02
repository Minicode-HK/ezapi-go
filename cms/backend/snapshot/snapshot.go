package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/Minicode-HK/ezapi-go/core"
)

// Metadata contains snapshot information
type Metadata struct {
    Timestamp time.Time `json:"timestamp"`
    Modules   []string  `json:"modules"`
    Version   string    `json:"version"`
}

// Snapshot represents a complete data snapshot
type Snapshot struct {
    Metadata Metadata               `json:"metadata"`
    Data     map[string]interface{} `json:"data"`
}

// Manager handles snapshot operations
type Manager struct {
    snapshotDir string
}

// NewManager creates a new snapshot manager
func NewManager(snapshotDir string) *Manager {
    os.MkdirAll(snapshotDir, 0755)
    return &Manager{
        snapshotDir: snapshotDir,
    }
}

// Create creates a snapshot of specified modules (empty means all)
func (m *Manager) Create(modules []string) (*Snapshot, error) {
    snapshot := &Snapshot{
        Metadata: Metadata{
            Timestamp: time.Now(),
            Modules:   modules,
            Version:   "1.0",
        },
        Data: make(map[string]interface{}),
    }

    registry := core.GetModuleRegistry()
    
    // If no modules specified, snapshot all
    if len(modules) == 0 {
        for _, module := range registry {
            modules = append(modules, module.TypeName.Name())
        }
        snapshot.Metadata.Modules = modules
    }

    // Collect data from specified modules
    for _, moduleName := range modules {
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
            return nil, fmt.Errorf("module '%s' not found", moduleName)
        }
    }

    return snapshot, nil
}

// Save saves a snapshot to a file
func (m *Manager) Save(snapshot *Snapshot, filename string) error {
    // Generate filename if not provided
    if filename == "" {
        filename = fmt.Sprintf("snapshot_%s.json", time.Now().Format("20060102_150405"))
    }
    if filepath.Ext(filename) != ".json" {
        filename += ".json"
    }

    filePath := filepath.Join(m.snapshotDir, filename)

    // Marshal and write
    jsonData, err := json.MarshalIndent(snapshot, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal snapshot: %w", err)
    }

    if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
        return fmt.Errorf("failed to write snapshot file: %w", err)
    }

    return nil
}

// CreateAndSave creates and saves a snapshot in one operation
func (m *Manager) CreateAndSave(modules []string, filename string) (string, error) {
    snapshot, err := m.Create(modules)
    if err != nil {
        return "", err
    }

    if err := m.Save(snapshot, filename); err != nil {
        return "", err
    }

    if filename == "" {
        filename = fmt.Sprintf("snapshot_%s.json", snapshot.Metadata.Timestamp.Format("20060102_150405"))
    }
    if filepath.Ext(filename) != ".json" {
        filename += ".json"
    }

    return filename, nil
}

// Load loads a snapshot from a file
func (m *Manager) Load(filename string) (*Snapshot, error) {
    filePath := filepath.Join(m.snapshotDir, filename)

    data, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read snapshot: %w", err)
    }

    var snapshot Snapshot
    if err := json.Unmarshal(data, &snapshot); err != nil {
        return nil, fmt.Errorf("invalid snapshot format: %w", err)
    }

    return &snapshot, nil
}

// Restore restores data from a snapshot to specified modules
func (m *Manager) Restore(snapshot *Snapshot, modules []string) error {
    // If no modules specified, restore all from snapshot
    if len(modules) == 0 {
        modules = snapshot.Metadata.Modules
    }

    registry := core.GetModuleRegistry()

    for _, moduleName := range modules {
        snapshotData, exists := snapshot.Data[moduleName]
        if !exists {
            return fmt.Errorf("module '%s' not found in snapshot", moduleName)
        }

        found := false
        for _, module := range registry {
            if module.TypeName.Name() == moduleName {
                // Marshal and unmarshal for type safety
                snapshotBytes, err := json.Marshal(snapshotData)
                if err != nil {
                    return fmt.Errorf("failed to marshal module '%s': %w", moduleName, err)
                }

                dbValue := reflect.ValueOf(module.Data).Elem()
                newSlice := reflect.New(dbValue.Type()).Interface()

                if err := json.Unmarshal(snapshotBytes, newSlice); err != nil {
                    return fmt.Errorf("failed to unmarshal module '%s': %w", moduleName, err)
                }

                dbValue.Set(reflect.ValueOf(newSlice).Elem())
                found = true
                break
            }
        }

        if !found {
            return fmt.Errorf("module '%s' not registered", moduleName)
        }
    }

    return nil
}

// LoadAndRestore loads and restores a snapshot in one operation
func (m *Manager) LoadAndRestore(filename string, modules []string) error {
    snapshot, err := m.Load(filename)
    if err != nil {
        return err
    }

    return m.Restore(snapshot, modules)
}

// List returns all available snapshots
func (m *Manager) List() ([]map[string]interface{}, error) {
    files, err := os.ReadDir(m.snapshotDir)
    if err != nil {
        if os.IsNotExist(err) {
            return []map[string]interface{}{}, nil
        }
        return nil, fmt.Errorf("failed to read snapshots: %w", err)
    }

    snapshots := []map[string]interface{}{}
    for _, file := range files {
        if filepath.Ext(file.Name()) != ".json" {
            continue
        }

        info, _ := file.Info()

        // Try to read metadata
        var snapshot Snapshot
        if s, err := m.Load(file.Name()); err == nil {
            snapshot = *s
        }

        snapshots = append(snapshots, map[string]interface{}{
            "filename":  file.Name(),
            "size":      info.Size(),
            "modified":  info.ModTime(),
            "modules":   snapshot.Metadata.Modules,
            "timestamp": snapshot.Metadata.Timestamp,
        })
    }

    // Sort by modified time descending
    for i := 0; i < len(snapshots)-1; i++ {
        for j := i + 1; j < len(snapshots); j++ {
            timeI := snapshots[i]["modified"].(time.Time)
            timeJ := snapshots[j]["modified"].(time.Time)
            if timeJ.After(timeI) {
                snapshots[i], snapshots[j] = snapshots[j], snapshots[i]
            }
        }
    }

    return snapshots, nil
}

// Delete deletes a snapshot file
func (m *Manager) Delete(filename string) error {
    filePath := filepath.Join(m.snapshotDir, filename)
    if err := os.Remove(filePath); err != nil {
        return fmt.Errorf("failed to delete snapshot: %w", err)
    }
    return nil
}

// GetModuleInfo returns information about all registered modules
func GetModuleInfo() []map[string]interface{} {
    modules := []map[string]interface{}{}

    for _, module := range core.GetModuleRegistry() {
        db := module.Data
        count := reflect.ValueOf(db).Elem().Len()

        modules = append(modules, map[string]interface{}{
            "name":      module.TypeName.Name(),
            "base_path": module.BasePath,
            "count":     count,
        })
    }

    return modules
}