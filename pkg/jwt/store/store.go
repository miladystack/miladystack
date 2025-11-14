package store

import "github.com/miladystack/miladystack/pkg/jwt/core"

// Re-export types from core for backward compatibility
type (
	RefreshTokenStorer = core.TokenStore
	RefreshTokenData   = core.RefreshTokenData
)

// Re-export errors from core for backward compatibility
var (
	ErrRefreshTokenNotFound = core.ErrRefreshTokenNotFound
	ErrRefreshTokenExpired  = core.ErrRefreshTokenExpired
)

// Default creates a default memory-based token store
// This is the recommended way to create a store with sensible defaults
func Default() core.TokenStore {
	return NewMemoryStore()
}
