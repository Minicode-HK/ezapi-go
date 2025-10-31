package handlers

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    
    "simple_backend_go/route"
)

type PlaygroundHandler struct{}

func NewPlaygroundHandler() *PlaygroundHandler {
    return &PlaygroundHandler{}
}

type PlaygroundRequest struct {
    Method  string            `json:"method" binding:"required"`
    URL     string            `json:"url" binding:"required"`
    Headers map[string]string `json:"headers"`
    Body    string            `json:"body"`
    MockMode bool             `json:"mockMode"`
}

type PlaygroundResponse struct {
    StatusCode int               `json:"statusCode"`
    Status     string            `json:"status"`
    Headers    map[string]string `json:"headers"`
    Body       interface{}       `json:"body"`
    Time       int64             `json:"time"` // milliseconds
    Size       int               `json:"size"` // bytes
    MockMode   bool              `json:"mockMode"`
}

func (h *PlaygroundHandler) ExecuteRequest(c *gin.Context) {
    var req PlaygroundRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "error":   "Invalid request: " + err.Error(),
        })
        return
    }

    // Create snapshot before request if mock mode is enabled and it's a modifying method
    var snapshot map[string]interface{}
    shouldRollback := false
    
    if req.MockMode && isModifyingMethod(req.Method) {
        snapshot = route.CreateSnapshot()
        shouldRollback = true
    }

    // Start timing
    startTime := time.Now()

    // Create HTTP request
    var bodyReader io.Reader
    if req.Body != "" {
        bodyReader = bytes.NewBufferString(req.Body)
    }

    httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "error":   "Failed to create request: " + err.Error(),
        })
        return
    }

    // Set headers
    for key, value := range req.Headers {
        httpReq.Header.Set(key, value)
    }

    // If no Content-Type and body exists, default to JSON
    if req.Body != "" && httpReq.Header.Get("Content-Type") == "" {
        httpReq.Header.Set("Content-Type", "application/json")
    }

    // Execute request
    client := &http.Client{
        Timeout: 30 * time.Second,
    }
    
    resp, err := client.Do(httpReq)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{
            "success": false,
            "error":   "Request failed: " + err.Error(),
        })
        return
    }
    defer resp.Body.Close()

    // Read response body
    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{
            "success": false,
            "error":   "Failed to read response: " + err.Error(),
        })
        return
    }

    // Rollback if mock mode was enabled
    if shouldRollback && snapshot != nil {
        route.RestoreSnapshot(snapshot)
    }

    // Calculate time taken
    duration := time.Since(startTime).Milliseconds()

    // Parse response body
    var bodyData interface{}
    if len(bodyBytes) > 0 {
        // Try to parse as JSON
        if err := json.Unmarshal(bodyBytes, &bodyData); err != nil {
            // If not JSON, return as string
            bodyData = string(bodyBytes)
        }
    }

    // Extract response headers
    headers := make(map[string]string)
    for key, values := range resp.Header {
        if len(values) > 0 {
            headers[key] = values[0]
        }
    }

    // Build response
    playgroundResp := PlaygroundResponse{
        StatusCode: resp.StatusCode,
        Status:     resp.Status,
        Headers:    headers,
        Body:       bodyData,
        Time:       duration,
        Size:       len(bodyBytes),
        MockMode:   req.MockMode && shouldRollback,
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    playgroundResp,
    })
}


// Helper function to check if HTTP method modifies data
func isModifyingMethod(method string) bool {
    return method == "POST" || method == "PUT" || method == "DELETE" || method == "PATCH"
}