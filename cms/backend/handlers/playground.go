package handlers

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net"
    "net/http"
    "net/url"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"

    // "ezapi-go/core"
)

type PlaygroundHandler struct {
    rateLimiter *rate.Limiter
    config      *PlaygroundConfig
}

type PlaygroundConfig struct {
    // Security settings
    AllowedHosts        []string      // Whitelist of allowed hosts (empty = allow only current host)
    BlockPrivateIPs     bool          // Block requests to private IP ranges
    MaxBodySize         int64         // Maximum request body size in bytes
    Timeout             time.Duration // Request timeout
    MaxRequestsPerMin   int           // Rate limit per IP
    AllowOnlyLocalhost  bool          // Only allow requests to localhost/127.0.0.1
    RequireHTTPS        bool          // Force HTTPS for external requests
    
    // Feature flags
    EnableMockMode      bool          // Allow mock transactions
    EnableExternalHosts bool          // Allow requests to external hosts
}

func NewPlaygroundHandler(config *PlaygroundConfig) *PlaygroundHandler {
    if config == nil {
        config = &PlaygroundConfig{
            AllowOnlyLocalhost:  true,  // Default: only localhost
            BlockPrivateIPs:     true,
            MaxBodySize:         1 * 1024 * 1024, // 1MB
            Timeout:             10 * time.Second,
            MaxRequestsPerMin:   30,
            EnableMockMode:      true,
            EnableExternalHosts: false,
            RequireHTTPS:        false,
        }
    }

    return &PlaygroundHandler{
        rateLimiter: rate.NewLimiter(rate.Limit(config.MaxRequestsPerMin)/60, config.MaxRequestsPerMin),
        config:      config,
    }
}

type PlaygroundRequest struct {
    Method   string            `json:"method" binding:"required"`
    URL      string            `json:"url" binding:"required"`
    Headers  map[string]string `json:"headers"`
    Body     string            `json:"body"`
    MockMode bool              `json:"mockMode"`
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
    // Rate limiting
    if !h.rateLimiter.Allow() {
        c.JSON(http.StatusTooManyRequests, gin.H{
            "success": false,
            "error":   "Rate limit exceeded. Please slow down.",
        })
        return
    }

    var req PlaygroundRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "error":   "Invalid request: " + err.Error(),
        })
        return
    }

    // Validate request
    if err := h.validateRequest(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "error":   err.Error(),
        })
        return
    }

    // Check mock mode availability
    if req.MockMode && !h.config.EnableMockMode {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "error":   "Mock mode is disabled",
        })
        return
    }

    // Create snapshot before request if mock mode is enabled and it's a modifying method
    // var snapshot map[string]interface{}
    shouldRollback := false

    // if req.MockMode && isModifyingMethod(req.Method) {
    //     snapshot = core.CreateSnapshot()
    //     shouldRollback = true
    // }

    // Start timing
    startTime := time.Now()

    // Create HTTP request with size limit
    var bodyReader io.Reader
    if req.Body != "" {
        if int64(len(req.Body)) > h.config.MaxBodySize {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   fmt.Sprintf("Request body too large (max: %d bytes)", h.config.MaxBodySize),
            })
            return
        }
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

    // Set safe headers only (filter out dangerous headers)
    for key, value := range req.Headers {
        if h.isSafeHeader(key) {
            httpReq.Header.Set(key, value)
        }
    }

    // If no Content-Type and body exists, default to JSON
    if req.Body != "" && httpReq.Header.Get("Content-Type") == "" {
        httpReq.Header.Set("Content-Type", "application/json")
    }

    // Add User-Agent to identify requests
    httpReq.Header.Set("User-Agent", "EZapi-Playground/1.0")

    // Execute request with timeout and size limit
    client := &http.Client{
        Timeout: h.config.Timeout,
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            // Limit redirects to 3
            if len(via) >= 3 {
                return fmt.Errorf("too many redirects")
            }
            // Validate redirect target
            if err := h.validateURL(req.URL.String()); err != nil {
                return err
            }
            return nil
        },
    }

    resp, err := client.Do(httpReq)
    if err != nil {
        // Rollback before returning error
        // if shouldRollback && snapshot != nil {
        //     core.RestoreSnapshot(snapshot)
        // }

        c.JSON(http.StatusOK, gin.H{
            "success": false,
            "error":   "Request failed: " + err.Error(),
        })
        return
    }
    defer resp.Body.Close()

    // Read response body with size limit
    limitedReader := io.LimitReader(resp.Body, h.config.MaxBodySize)
    bodyBytes, err := io.ReadAll(limitedReader)
    if err != nil {
        // Rollback before returning error
        // if shouldRollback && snapshot != nil {
        //     core.RestoreSnapshot(snapshot)
        // }

        c.JSON(http.StatusOK, gin.H{
            "success": false,
            "error":   "Failed to read response: " + err.Error(),
        })
        return
    }

    // Rollback if mock mode was enabled
    // if shouldRollback && snapshot != nil {
    //     core.RestoreSnapshot(snapshot)
    // }

    // Calculate time taken
    duration := time.Since(startTime).Milliseconds()

    // Parse response body
    var bodyData interface{}
    if len(bodyBytes) > 0 {
        // Try to parse as JSON
        if err := json.Unmarshal(bodyBytes, &bodyData); err != nil {
            // If not JSON, return as string (truncate if too large)
            bodyStr := string(bodyBytes)
            if len(bodyStr) > 10000 {
                bodyStr = bodyStr[:10000] + "... (truncated)"
            }
            bodyData = bodyStr
        }
    }

    // Extract response headers (safe ones only)
    headers := make(map[string]string)
    for key, values := range resp.Header {
        if h.isSafeHeader(key) && len(values) > 0 {
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

// validateRequest validates the playground request
func (h *PlaygroundHandler) validateRequest(req *PlaygroundRequest) error {
    // Validate HTTP method
    validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
    isValid := false
    for _, method := range validMethods {
        if req.Method == method {
            isValid = true
            break
        }
    }
    if !isValid {
        return fmt.Errorf("invalid HTTP method: %s", req.Method)
    }

    // Validate URL
    if err := h.validateURL(req.URL); err != nil {
        return err
    }

    // Validate body size
    if int64(len(req.Body)) > h.config.MaxBodySize {
        return fmt.Errorf("request body too large (max: %d bytes)", h.config.MaxBodySize)
    }

    return nil
}

// validateURL validates and sanitizes the URL
func (h *PlaygroundHandler) validateURL(urlStr string) error {
    parsedURL, err := url.Parse(urlStr)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }

    // Check scheme
    if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
        return fmt.Errorf("only http and https schemes are allowed")
    }

    // Require HTTPS for external hosts
    if h.config.RequireHTTPS && !h.isLocalhost(parsedURL.Hostname()) && parsedURL.Scheme != "https" {
        return fmt.Errorf("HTTPS required for external hosts")
    }

    // Check if external hosts are allowed
    if !h.config.EnableExternalHosts && !h.isLocalhost(parsedURL.Hostname()) {
        return fmt.Errorf("external hosts are not allowed")
    }

    // If only localhost is allowed
    if h.config.AllowOnlyLocalhost && !h.isLocalhost(parsedURL.Hostname()) {
        return fmt.Errorf("only localhost requests are allowed")
    }

    // Check whitelist
    if len(h.config.AllowedHosts) > 0 {
        allowed := false
        for _, host := range h.config.AllowedHosts {
            if strings.EqualFold(parsedURL.Hostname(), host) {
                allowed = true
                break
            }
        }
        if !allowed {
            return fmt.Errorf("host '%s' is not in the allowed list", parsedURL.Hostname())
        }
    }

    // Resolve hostname to IP and check if it's private
    if h.config.BlockPrivateIPs {
        ips, err := net.LookupIP(parsedURL.Hostname())
        if err != nil {
            return fmt.Errorf("failed to resolve hostname: %w", err)
        }

        for _, ip := range ips {
            if h.isPrivateIP(ip) && !h.isLocalhost(parsedURL.Hostname()) {
                return fmt.Errorf("requests to private IP addresses are blocked")
            }
        }
    }

    // Block common cloud metadata endpoints
    blockedHosts := []string{
        "169.254.169.254",          // AWS/Azure metadata
        "metadata.google.internal", // GCP metadata
        "kubernetes.default.svc",   // Kubernetes API
    }
    for _, blocked := range blockedHosts {
        if strings.Contains(strings.ToLower(parsedURL.Hostname()), blocked) {
            return fmt.Errorf("access to cloud metadata endpoints is blocked")
        }
    }

    return nil
}

// isLocalhost checks if hostname is localhost
func (h *PlaygroundHandler) isLocalhost(hostname string) bool {
    localhost := []string{"localhost", "127.0.0.1", "::1", "0.0.0.0"}
    for _, local := range localhost {
        if strings.EqualFold(hostname, local) {
            return true
        }
    }
    return false
}

// isPrivateIP checks if IP is in private range
func (h *PlaygroundHandler) isPrivateIP(ip net.IP) bool {
    privateRanges := []string{
        "10.0.0.0/8",
        "172.16.0.0/12",
        "192.168.0.0/16",
        "169.254.0.0/16", // Link-local
        "127.0.0.0/8",    // Loopback
        "fc00::/7",       // IPv6 private
        "fe80::/10",      // IPv6 link-local
    }

    for _, cidr := range privateRanges {
        _, subnet, _ := net.ParseCIDR(cidr)
        if subnet != nil && subnet.Contains(ip) {
            return true
        }
    }

    return false
}

// isSafeHeader checks if header is safe to use
func (h *PlaygroundHandler) isSafeHeader(header string) bool {
    // Blocked headers
    blocked := []string{
        "cookie",
        "authorization",
        "x-forwarded-for",
        "x-real-ip",
        "proxy-authorization",
    }

    headerLower := strings.ToLower(header)
    for _, b := range blocked {
        if headerLower == b {
            return false
        }
    }

    return true
}

// isModifyingMethod checks if HTTP method modifies data
func isModifyingMethod(method string) bool {
    return method == "POST" || method == "PUT" || method == "DELETE" || method == "PATCH"
}

// GetConfig returns current playground configuration
func (h *PlaygroundHandler) GetConfig(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": gin.H{
            "allowOnlyLocalhost":  h.config.AllowOnlyLocalhost,
            "blockPrivateIPs":     h.config.BlockPrivateIPs,
            "maxBodySize":         h.config.MaxBodySize,
            "timeout":             h.config.Timeout.Seconds(),
            "maxRequestsPerMin":   h.config.MaxRequestsPerMin,
            "enableMockMode":      h.config.EnableMockMode,
            "enableExternalHosts": h.config.EnableExternalHosts,
            "requireHTTPS":        h.config.RequireHTTPS,
            "allowedHosts":        h.config.AllowedHosts,
        },
    })
}

// UpdateConfig updates playground configuration
func (h *PlaygroundHandler) UpdateConfig(c *gin.Context) {
    var req struct {
        AllowOnlyLocalhost  *bool    `json:"allowOnlyLocalhost"`
        BlockPrivateIPs     *bool    `json:"blockPrivateIPs"`
        MaxBodySize         *int64   `json:"maxBodySize"`
        Timeout             *float64 `json:"timeout"`
        MaxRequestsPerMin   *int     `json:"maxRequestsPerMin"`
        EnableMockMode      *bool    `json:"enableMockMode"`
        EnableExternalHosts *bool    `json:"enableExternalHosts"`
        RequireHTTPS        *bool    `json:"requireHTTPS"`
        AllowedHosts        []string `json:"allowedHosts"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "error":   err.Error(),
        })
        return
    }

    // Update config (only non-nil values)
    if req.AllowOnlyLocalhost != nil {
        h.config.AllowOnlyLocalhost = *req.AllowOnlyLocalhost
    }
    if req.BlockPrivateIPs != nil {
        h.config.BlockPrivateIPs = *req.BlockPrivateIPs
    }
    if req.MaxBodySize != nil {
        if *req.MaxBodySize < 1024 || *req.MaxBodySize > 10*1024*1024 {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   "maxBodySize must be between 1KB and 10MB",
            })
            return
        }
        h.config.MaxBodySize = *req.MaxBodySize
    }
    if req.Timeout != nil {
        if *req.Timeout < 1 || *req.Timeout > 60 {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   "timeout must be between 1 and 60 seconds",
            })
            return
        }
        h.config.Timeout = time.Duration(*req.Timeout) * time.Second
    }
    if req.MaxRequestsPerMin != nil {
        if *req.MaxRequestsPerMin < 1 || *req.MaxRequestsPerMin > 1000 {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error":   "maxRequestsPerMin must be between 1 and 1000",
            })
            return
        }
        h.config.MaxRequestsPerMin = *req.MaxRequestsPerMin
        // Update rate limiter
        h.rateLimiter = rate.NewLimiter(rate.Limit(*req.MaxRequestsPerMin)/60, *req.MaxRequestsPerMin)
    }
    if req.EnableMockMode != nil {
        h.config.EnableMockMode = *req.EnableMockMode
    }
    if req.EnableExternalHosts != nil {
        h.config.EnableExternalHosts = *req.EnableExternalHosts
    }
    if req.RequireHTTPS != nil {
        h.config.RequireHTTPS = *req.RequireHTTPS
    }
    if req.AllowedHosts != nil {
        h.config.AllowedHosts = req.AllowedHosts
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Configuration updated successfully",
        "data": gin.H{
            "allowOnlyLocalhost":  h.config.AllowOnlyLocalhost,
            "blockPrivateIPs":     h.config.BlockPrivateIPs,
            "maxBodySize":         h.config.MaxBodySize,
            "timeout":             h.config.Timeout.Seconds(),
            "maxRequestsPerMin":   h.config.MaxRequestsPerMin,
            "enableMockMode":      h.config.EnableMockMode,
            "enableExternalHosts": h.config.EnableExternalHosts,
            "requireHTTPS":        h.config.RequireHTTPS,
            "allowedHosts":        h.config.AllowedHosts,
        },
    })
}