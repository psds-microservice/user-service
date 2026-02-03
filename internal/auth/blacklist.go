package auth

import (
	"sync"
	"time"
)

// Blacklist — инвалидированные JWT (logout). In-memory; для production лучше Redis.
type Blacklist struct {
	mu   sync.RWMutex
	byID map[string]time.Time // jti -> expire at
}

func NewBlacklist() *Blacklist {
	return &Blacklist{byID: make(map[string]time.Time)}
}

// Add помечает токен как недействительный до expireAt.
func (b *Blacklist) Add(jti string, expireAt time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.byID[jti] = expireAt
}

// Contains возвращает true если jti в чёрном списке и ещё не истёк.
func (b *Blacklist) Contains(jti string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	exp, ok := b.byID[jti]
	return ok && time.Now().Before(exp)
}
