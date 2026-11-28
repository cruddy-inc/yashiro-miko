package domain

import (
	"sync/atomic"
	"time"
)

type Reroll struct {
	CurrentSecretKey  string
	PreviousSecretKey string
	Version           string
	UpdatedAt         time.Time
	ExpiresAt         int64
}

type RerollStore struct {
	ptr atomic.Pointer[Reroll]
}

func (r *RerollStore) Store(re *Reroll) {
	r.ptr.Store(re)
}

func (r *RerollStore) Load() *Reroll {
	return r.ptr.Load()
}

func (r *RerollStore) CompareAndSwap(old *Reroll, new *Reroll) bool {
	return r.ptr.CompareAndSwap(old, new)
}

func RerollNew(r *Reroll) *RerollStore {
	re := &RerollStore{}
	if r != nil {
		re.ptr.Store(r)
	}
	return re
}
