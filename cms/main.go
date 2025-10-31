package cms

import "time"

type Config struct {
    Users           map[string]string
    SessionDuration time.Duration
    SnapshotDir     string
    TemplateDir     string
}

func DefaultConfig() *Config {
    return &Config{
        Users: map[string]string{
            "superadmin": "superadmin",
            "admin":      "admin",
        },
        SessionDuration: 24 * time.Hour,
        SnapshotDir:     "./snapshots",
        TemplateDir:     "./cms/templates",
    }
}