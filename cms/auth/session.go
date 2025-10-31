package auth

import (
    "crypto/rand"
    "encoding/base64"
    "sync"
    "time"
)

type Session struct {
    Token     string
    Username  string
    ExpiresAt time.Time
}

type SessionManager struct {
    sessions map[string]*Session
    mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
    sm := &SessionManager{
        sessions: make(map[string]*Session),
    }
    
    // Cleanup expired sessions every hour
    go sm.cleanupLoop()
    
    return sm
}

func (sm *SessionManager) cleanupLoop() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        sm.cleanup()
    }
}

func (sm *SessionManager) cleanup() {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    now := time.Now()
    for token, session := range sm.sessions {
        if session.ExpiresAt.Before(now) {
            delete(sm.sessions, token)
        }
    }
}

func (sm *SessionManager) Create(username string, duration time.Duration) (string, error) {
    token, err := generateToken()
    if err != nil {
        return "", err
    }
    
    session := &Session{
        Token:     token,
        Username:  username,
        ExpiresAt: time.Now().Add(duration),
    }
    
    sm.mu.Lock()
    sm.sessions[token] = session
    sm.mu.Unlock()
    
    return token, nil
}

func (sm *SessionManager) Get(token string) (*Session, bool) {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    
    session, exists := sm.sessions[token]
    if !exists {
        return nil, false
    }
    
    if session.ExpiresAt.Before(time.Now()) {
        return nil, false
    }
    
    return session, true
}

func (sm *SessionManager) Delete(token string) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    delete(sm.sessions, token)
}

func generateToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}