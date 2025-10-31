package cms

import (
    "crypto/rand"
    "encoding/base64"
    "sync"
    "time"
    
    "github.com/gin-gonic/gin"
)

type Session struct {
    Token     string
    Username  string
    ExpiresAt time.Time
}

type AuthManager struct {
    sessions map[string]*Session
    users    map[string]string
    mu       sync.RWMutex
}

var authManager *AuthManager

func init() {
    authManager = &AuthManager{
        sessions: make(map[string]*Session),
        users: Users,
    }
    
    // Cleanup expired sessions every hour
    go authManager.cleanupSessions()
}

func (am *AuthManager) cleanupSessions() {
    ticker := time.NewTicker(1 * time.Hour)
    for range ticker.C {
        am.mu.Lock()
        now := time.Now()
        for token, session := range am.sessions {
            if session.ExpiresAt.Before(now) {
                delete(am.sessions, token)
            }
        }
        am.mu.Unlock()
    }
}

func (am *AuthManager) generateToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}

func (am *AuthManager) Login(username, password string) (string, error) {
    am.mu.RLock()
    storedPassword, exists := am.users[username]
    am.mu.RUnlock()
    
    if !exists || storedPassword != password {
        return "", gin.Error{Err: nil, Type: gin.ErrorTypePublic, Meta: "Invalid credentials"}
    }
    
    token, err := am.generateToken()
    if err != nil {
        return "", err
    }
    
    session := &Session{
        Token:     token,
        Username:  username,
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }
    
    am.mu.Lock()
    am.sessions[token] = session
    am.mu.Unlock()
    
    return token, nil
}

func (am *AuthManager) ValidateToken(token string) (*Session, bool) {
    am.mu.RLock()
    defer am.mu.RUnlock()
    
    session, exists := am.sessions[token]
    if !exists {
        return nil, false
    }
    
    if session.ExpiresAt.Before(time.Now()) {
        return nil, false
    }
    
    return session, true
}

func (am *AuthManager) Logout(token string) {
    am.mu.Lock()
    defer am.mu.Unlock()
    delete(am.sessions, token)
}

// Middleware to protect routes
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Try to get token from cookie first
        token, err := c.Cookie("cms_token")
        if err != nil || token == "" {
            // Fallback to Authorization header
            token = c.GetHeader("Authorization")
        }
        
        if token == "" {
            c.JSON(401, gin.H{"success": false, "message": "Unauthorized - No token provided"})
            c.Abort()
            return
        }
        
        session, valid := authManager.ValidateToken(token)
        if !valid {
            c.JSON(401, gin.H{"success": false, "message": "Invalid or expired token"})
            c.Abort()
            return
        }
        
        c.Set("username", session.Username)
        c.Next()
    }
}