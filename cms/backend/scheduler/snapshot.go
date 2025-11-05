package scheduler

import (
    "fmt"
    "log"
    "sync"
    "time"

    "ezapi-go/cms/backend/snapshot"
)

type Scheduler struct {
    enabled        bool
    interval       time.Duration
    modules        []string
    ticker         *time.Ticker
    stopChan       chan bool
    mu             sync.RWMutex
    running        bool
    snapshotMgr    *snapshot.Manager
}

var (
    instance *Scheduler
    once     sync.Once
)

// GetScheduler returns the singleton scheduler instance
func GetScheduler(snapshotMgr *snapshot.Manager) *Scheduler {
    once.Do(func() {
        instance = &Scheduler{
            enabled:     false,
            interval:    24 * time.Hour,
            modules:     []string{},
            stopChan:    make(chan bool),
            snapshotMgr: snapshotMgr,
        }
    })
    return instance
}

// Start begins the scheduled snapshot creation
func (s *Scheduler) Start() error {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.running {
        return fmt.Errorf("scheduler is already running")
    }

    if !s.enabled {
        return fmt.Errorf("scheduler is disabled")
    }

    s.ticker = time.NewTicker(s.interval)
    s.running = true

    go func() {
        log.Printf("Snapshot scheduler started (interval: %v)", s.interval)

        for {
            select {
            case <-s.ticker.C:
                s.createSnapshot()
            case <-s.stopChan:
                log.Println("Snapshot scheduler stopped")
                return
            }
        }
    }()

    return nil
}

// Stop halts the scheduled snapshot creation
func (s *Scheduler) Stop() {
    s.mu.Lock()
    defer s.mu.Unlock()

    if !s.running {
        return
    }

    if s.ticker != nil {
        s.ticker.Stop()
    }

    s.stopChan <- true
    s.running = false
}

// SetInterval updates the snapshot interval
func (s *Scheduler) SetInterval(interval time.Duration) error {
    if interval < time.Minute {
        return fmt.Errorf("interval must be at least 1 minute")
    }

    s.mu.Lock()
    wasRunning := s.running
    s.mu.Unlock()

    if wasRunning {
        s.Stop()
    }

    s.mu.Lock()
    s.interval = interval
    s.mu.Unlock()

    if wasRunning && s.enabled {
        return s.Start()
    }

    return nil
}

// SetModules sets which modules to include in scheduled snapshots
func (s *Scheduler) SetModules(modules []string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.modules = modules
}

// SetEnabled enables or disables the scheduler
func (s *Scheduler) SetEnabled(enabled bool) error {
    s.mu.Lock()
    s.enabled = enabled
    s.mu.Unlock()

    if enabled && !s.running {
        return s.Start()
    } else if !enabled && s.running {
        s.Stop()
    }

    return nil
}

// GetStatus returns the current scheduler status
func (s *Scheduler) GetStatus() map[string]interface{} {
    s.mu.RLock()
    defer s.mu.RUnlock()

    nextRun := "N/A"
    if s.running {
        nextRun = time.Now().Add(s.interval).Format(time.RFC3339)
    }

    return map[string]interface{}{
        "enabled":  s.enabled,
        "running":  s.running,
        "interval": s.interval.String(),
        "modules":  s.modules,
        "nextRun":  nextRun,
    }
}

// CreateSnapshot creates a manual snapshot using current config
func (s *Scheduler) CreateSnapshot(prefix string) (string, error) {
    s.mu.RLock()
    modules := s.modules
    s.mu.RUnlock()

    timestamp := time.Now().Format("20060102_150405")
    filename := fmt.Sprintf("%s_backup_%s.json", prefix, timestamp)

    filename, err := s.snapshotMgr.CreateAndSave(modules, filename)
    if err != nil {
        return "", err
    }

    return filename, nil
}

// createSnapshot creates an automatic scheduled snapshot
func (s *Scheduler) createSnapshot() {
    filename, err := s.CreateSnapshot("auto")
    if err != nil {
        log.Printf("Failed to create scheduled snapshot: %v", err)
        return
    }

    log.Printf("Scheduled snapshot created successfully: %s", filename)
}