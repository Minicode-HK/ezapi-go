package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "simple_backend_go/cms/models"
    "simple_backend_go/cms/scheduler"
    "simple_backend_go/cms/snapshot"
)

type SnapshotHandler struct {
    manager   *snapshot.Manager
    scheduler *scheduler.Scheduler
}

func NewSnapshotHandler(snapshotDir string) *SnapshotHandler {
    mgr := snapshot.NewManager(snapshotDir)
    return &SnapshotHandler{
        manager:   mgr,
        scheduler: scheduler.GetScheduler(mgr),
    }
}

// GetModules returns all registered modules
func (h *SnapshotHandler) GetModules(c *gin.Context) {
    modules := snapshot.GetModuleInfo()
    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Data:    modules,
    })
}

// Save creates and saves a new snapshot
func (h *SnapshotHandler) Save(c *gin.Context) {
    var req struct {
        Modules  []string `json:"modules" binding:"required"`
        Filename string   `json:"filename"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Invalid request",
        })
        return
    }

    filename, err := h.manager.CreateAndSave(req.Modules, req.Filename)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Snapshot saved successfully",
        Data: gin.H{
            "filename": filename,
            "modules":  req.Modules,
        },
    })
}

// List returns all available snapshots
func (h *SnapshotHandler) List(c *gin.Context) {
    snapshots, err := h.manager.List()
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Data:    snapshots,
    })
}

// Load restores a snapshot
func (h *SnapshotHandler) Load(c *gin.Context) {
    var req struct {
        Filename string   `json:"filename" binding:"required"`
        Modules  []string `json:"modules"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Invalid request",
        })
        return
    }

    if err := h.manager.LoadAndRestore(req.Filename, req.Modules); err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Snapshot loaded successfully",
    })
}

// Delete removes a snapshot
func (h *SnapshotHandler) Delete(c *gin.Context) {
    filename := c.Param("filename")

    if err := h.manager.Delete(filename); err != nil {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Message: "Snapshot deleted successfully",
    })
}

// GetSchedulerStatus returns scheduler status
func (h *SnapshotHandler) GetSchedulerStatus(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    h.scheduler.GetStatus(),
    })
}

// UpdateSchedulerConfig updates scheduler configuration
func (h *SnapshotHandler) UpdateSchedulerConfig(c *gin.Context) {
    var req struct {
        Enabled  bool     `json:"enabled"`
        Interval string   `json:"interval"`
        Modules  []string `json:"modules"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "error":   err.Error(),
        })
        return
    }

    // Parse and set interval
    if req.Interval != "" {
        duration, err := time.ParseDuration(req.Interval)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   "Invalid interval format",
            })
            return
        }

        if err := h.scheduler.SetInterval(duration); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   err.Error(),
            })
            return
        }
    }

    h.scheduler.SetModules(req.Modules)

    if err := h.scheduler.SetEnabled(req.Enabled); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "error":   err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Scheduler configuration updated",
        "data":    h.scheduler.GetStatus(),
    })
}

// TriggerManualBackup creates a manual backup
func (h *SnapshotHandler) TriggerManualBackup(c *gin.Context) {
    filename, err := h.scheduler.CreateSnapshot("manual")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "error":   err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Manual backup created successfully",
        "data": gin.H{
            "filename": filename,
        },
    })
}