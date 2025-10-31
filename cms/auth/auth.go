package auth

import (
    "errors"
    "sync"
    "time"
)

var (
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrUserNotFound       = errors.New("user not found")
)

type AuthService struct {
    users          map[string]string // username -> password (should use bcrypt in production)
    sessionManager *SessionManager
    mu             sync.RWMutex
}

func NewAuthService(users map[string]string) *AuthService {
    return &AuthService{
        users:          users,
        sessionManager: NewSessionManager(),
    }
}

func (as *AuthService) Login(username, password string) (string, error) {
    as.mu.RLock()
    storedPassword, exists := as.users[username]
    as.mu.RUnlock()
    
    if !exists {
        return "", ErrUserNotFound
    }
    
    if storedPassword != password {
        return "", ErrInvalidCredentials
    }
    
    // Create session (24 hour expiry)
    token, err := as.sessionManager.Create(username, 24*time.Hour)
    if err != nil {
        return "", err
    }
    
    return token, nil
}

func (as *AuthService) ValidateToken(token string) (*Session, bool) {
    return as.sessionManager.Get(token)
}

func (as *AuthService) Logout(token string) {
    as.sessionManager.Delete(token)
}

func (as *AuthService) GetUsername(token string) (string, bool) {
    session, valid := as.sessionManager.Get(token)
    if !valid {
        return "", false
    }
    return session.Username, true
}