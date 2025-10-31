package handlers

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
    "sync"
    "time"
	"strings"

    "github.com/gin-gonic/gin"
)

// LogEntry represents a single API request/response log
type LogEntry struct {
    ID             string                 `json:"id"`
    Timestamp      time.Time              `json:"timestamp"`
    Method         string                 `json:"method"`
    Path           string                 `json:"path"`
    StatusCode     int                    `json:"statusCode"`
    Status         string                 `json:"status"`
    Duration       int64                  `json:"duration"` // in milliseconds
    RequestHeaders map[string]string      `json:"requestHeaders"`
    RequestBody    string                 `json:"requestBody"`
    ResponseHeaders map[string]string     `json:"responseHeaders"`
    ResponseBody   string                 `json:"responseBody"`
    ResponseSize   int                    `json:"responseSize"`
    ClientIP       string                 `json:"clientIP"`
    UserAgent      string                 `json:"userAgent"`
    Error          string                 `json:"error,omitempty"`
}

// APILogger handles API request logging
type APILogger struct {
    logs      []LogEntry
    mu        sync.RWMutex
    maxLogs   int
    enabled   bool
}

var (
    loggerInstance *APILogger
    loggerOnce     sync.Once
)

// GetAPILogger returns the singleton logger instance
func GetAPILogger() *APILogger {
    loggerOnce.Do(func() {
        loggerInstance = &APILogger{
            logs:    make([]LogEntry, 0, 1000),
            maxLogs: 1000, // Keep last 1000 logs in memory
            enabled: true,
        }
    })
    return loggerInstance
}

// AddLog adds a new log entry
func (l *APILogger) AddLog(entry LogEntry) {
    if !l.enabled {
        return
    }

    l.mu.Lock()
    defer l.mu.Unlock()

    // Add timestamp if not set
    if entry.Timestamp.IsZero() {
        entry.Timestamp = time.Now()
    }

    // Add to beginning of slice
    l.logs = append([]LogEntry{entry}, l.logs...)

    // Trim if exceeds max
    if len(l.logs) > l.maxLogs {
        l.logs = l.logs[:l.maxLogs]
    }
}

// GetLogs returns all logs with optional filtering
func (l *APILogger) GetLogs(filter map[string]interface{}) []LogEntry {
    l.mu.RLock()
    defer l.mu.RUnlock()

    if len(filter) == 0 {
        // Return copy of all logs
        result := make([]LogEntry, len(l.logs))
        copy(result, l.logs)
        return result
    }

    // Filter logs
    var filtered []LogEntry
    for _, log := range l.logs {
        if matchesFilter(log, filter) {
            filtered = append(filtered, log)
        }
    }
    return filtered
}

// ClearLogs removes all logs
func (l *APILogger) ClearLogs() {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.logs = make([]LogEntry, 0, 1000)
}

// SetEnabled enables or disables logging
func (l *APILogger) SetEnabled(enabled bool) {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.enabled = enabled
}

// matchesFilter checks if a log entry matches the filter criteria
func matchesFilter(log LogEntry, filter map[string]interface{}) bool {
    if method, ok := filter["method"].(string); ok && method != "" {
        if log.Method != method {
            return false
        }
    }

    if path, ok := filter["path"].(string); ok && path != "" {
        if log.Path != path {
            return false
        }
    }

    if statusCode, ok := filter["statusCode"].(float64); ok {
        if log.StatusCode != int(statusCode) {
            return false
        }
    }

    if minDuration, ok := filter["minDuration"].(float64); ok {
        if log.Duration < int64(minDuration) {
            return false
        }
    }

    if maxDuration, ok := filter["maxDuration"].(float64); ok {
        if log.Duration > int64(maxDuration) {
            return false
        }
    }

    return true
}

// Middleware for logging API requests
func LoggingMiddleware() gin.HandlerFunc {
    logger := GetAPILogger()

    return func(c *gin.Context) {
        if shouldIncludeLogging(c.Request.URL.Path) {
            c.Next()
            return
        }

        startTime := time.Now()

        // Read request body
        var requestBody string
        if c.Request.Body != nil {
            bodyBytes, _ := io.ReadAll(c.Request.Body)
            requestBody = string(bodyBytes)
            // Restore body for next handlers
            c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
        }

        // Capture request headers
        requestHeaders := make(map[string]string)
        for key, values := range c.Request.Header {
            if len(values) > 0 {
                requestHeaders[key] = values[0]
            }
        }

        // Custom response writer to capture response
        blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
        c.Writer = blw

        // Process request
        c.Next()

        // Calculate duration
        duration := time.Since(startTime).Milliseconds()

        // Capture response headers
        responseHeaders := make(map[string]string)
        for key, values := range c.Writer.Header() {
            if len(values) > 0 {
                responseHeaders[key] = values[0]
            }
        }

        // Get response body
        responseBody := blw.body.String()

        // Get error if any
        var errorMsg string
        if len(c.Errors) > 0 {
            errorMsg = c.Errors.String()
        }

        // Create log entry
        entry := LogEntry{
            ID:              generateID(),
            Timestamp:       startTime,
            Method:          c.Request.Method,
            Path:            c.Request.URL.Path,
            StatusCode:      c.Writer.Status(),
            Status:          http.StatusText(c.Writer.Status()),
            Duration:        duration,
            RequestHeaders:  requestHeaders,
            RequestBody:     requestBody,
            ResponseHeaders: responseHeaders,
            ResponseBody:    responseBody,
            ResponseSize:    blw.body.Len(),
            ClientIP:        c.ClientIP(),
            UserAgent:       c.Request.UserAgent(),
            Error:           errorMsg,
        }

        // Add to logger
        logger.AddLog(entry)
    }
}

// bodyLogWriter wraps gin.ResponseWriter to capture response body
type bodyLogWriter struct {
    gin.ResponseWriter
    body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
    w.body.Write(b)
    return w.ResponseWriter.Write(b)
}

func shouldIncludeLogging(path string) bool {
   includePaths := []string{
	   "/api/",
   }

   for _, p := range includePaths {
	   if strings.HasPrefix(path, p) {
		   return false
	   }
   }
   return true
}

func generateID() string {
    return time.Now().Format("20060102150405") + "-" + randomString(6)
}

func randomString(n int) string {
    const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    b := make([]byte, n)
    for i := range b {
        b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
    }
    return string(b)
}

// Handler methods

// List returns all logs with optional filtering
func (l *APILogger) List(c *gin.Context) {
    filter := make(map[string]interface{})

    // Parse query parameters
    if method := c.Query("method"); method != "" {
        filter["method"] = method
    }
    if path := c.Query("path"); path != "" {
        filter["path"] = path
    }
    if statusCode := c.Query("statusCode"); statusCode != "" {
        var code float64
        json.Unmarshal([]byte(statusCode), &code)
        filter["statusCode"] = code
    }

    logs := l.GetLogs(filter)

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": gin.H{
            "logs":  logs,
            "total": len(logs),
        },
    })
}

// GetByID returns a specific log entry
func (l *APILogger) GetByID(c *gin.Context) {
    id := c.Param("id")

    l.mu.RLock()
    defer l.mu.RUnlock()

    for _, log := range l.logs {
        if log.ID == id {
            c.JSON(http.StatusOK, gin.H{
                "success": true,
                "data":    log,
            })
            return
        }
    }

    c.JSON(http.StatusNotFound, gin.H{
        "success": false,
        "error":   "Log entry not found",
    })
}

// Clear removes all logs
func (l *APILogger) Clear(c *gin.Context) {
    l.ClearLogs()

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "All logs cleared",
    })
}

// Toggle enables or disables logging
func (l *APILogger) Toggle(c *gin.Context) {
    var req struct {
        Enabled bool `json:"enabled"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "error":   err.Error(),
        })
        return
    }

    l.SetEnabled(req.Enabled)

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": gin.H{
            "enabled": req.Enabled,
        },
    })
}

// Stats returns statistics about the logs
func (l *APILogger) Stats(c *gin.Context) {
    l.mu.RLock()
    defer l.mu.RUnlock()

    stats := map[string]interface{}{
        "total": len(l.logs),
        "methods": make(map[string]int),
        "statusCodes": make(map[int]int),
        "avgDuration": 0,
    }

    var totalDuration int64
    methods := make(map[string]int)
    statusCodes := make(map[int]int)

    for _, log := range l.logs {
        methods[log.Method]++
        statusCodes[log.StatusCode]++
        totalDuration += log.Duration
    }

    if len(l.logs) > 0 {
        stats["avgDuration"] = totalDuration / int64(len(l.logs))
    }

    stats["methods"] = methods
    stats["statusCodes"] = statusCodes

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    stats,
    })
}

// Export exports logs as JSON
func (l *APILogger) Export(c *gin.Context) {
    logs := l.GetLogs(nil)

    c.Header("Content-Disposition", "attachment; filename=api-logs-"+time.Now().Format("20060102-150405")+".json")
    c.JSON(http.StatusOK, logs)
}