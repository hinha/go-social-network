package utils

import (
	"time"
)

type GuestTokenManager struct {
	Token  string
	timing time.Time
}

func TokenManager() *GuestTokenManager {
	return &GuestTokenManager{}
}

func (t *GuestTokenManager) GetToken() string {
	return t.Token
}

func (t *GuestTokenManager) SetToken(token string) {
	t.Token = token
	t.timing = time.Now()
}

func (t *GuestTokenManager) GetTime() time.Time {
	return t.timing
}

func (t *GuestTokenManager) Reset() {
	t.Token = ""
	t.timing = time.Time{}
}
