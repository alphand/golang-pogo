package auth

import (
	"errors"
	"fmt"
)

// Provider - common interface to manage auth token from different third
// party authenticators
type Provider interface {
	Login() (authToken string, err error)
	GetProviderString() string
	GetAccessToken() string
}

// UnknownProvider - null provider by default if real provider not avail
type UnknownProvider struct {
}

// Login - login method for unknonwn provider
func (u *UnknownProvider) Login() (string, error) {
	return "", errors.New("Cannot login using unknown provider")
}

// GetProviderString - return the name of this provider
func (u *UnknownProvider) GetProviderString() string {
	return "unknown"
}

// GetAccessToken - return empty due to no access token
func (u *UnknownProvider) GetAccessToken() string {
	return ""
}

// NewProvider - generate provider based on the type
func NewProvider(provider, username, password string) (Provider, error) {
	switch provider {
	default:
		return &UnknownProvider{}, fmt.Errorf("Provider \"%s\" is not supported", provider)
	}
}
