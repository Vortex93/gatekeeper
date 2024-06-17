package gatekeeper

import (
	"log"
	"sync"
	"sync/atomic"
)

// GateKeeper controls access to a resource or section of code among multiple goroutines.
type GateKeeper struct {
	counter  atomic.Int64
	open     atomic.Bool
	mutex    sync.Mutex
	cond    *sync.Cond
}

// NewGateKeeper initializes a new GateKeeper. If `locked` is true, the gate starts in a locked state.
func NewGateKeeper(locked bool) *GateKeeper {
	gk := &GateKeeper{}
	gk.cond = sync.NewCond(&gk.mutex)

	log.Println(gk.open.Load())

	if locked { 
		gk.Lock()
	} else {
		gk.Unlock()
	}

	return gk
}

// IsLocked checks if the gate is in a locked state.
func (gk *GateKeeper) IsLocked() bool {
	return !gk.open.Load()
}

// IsUnlocked checks if the gate is in an open state.
func (gk *GateKeeper) IsUnlocked() bool {
	return gk.open.Load()
}

// Lock sets the gate to a locked state, preventing goroutines from passing until it is unlocked.
func (gk *GateKeeper) Lock() {
	gk.mutex.Lock()
	gk.open.Store(false)
	gk.mutex.Unlock() 
}

// Unlock sets the gate to an open state, allowing all waiting goroutines to proceed.
func (gk *GateKeeper) Unlock() {
	gk.mutex.Lock()
	gk.open.Store(true)
	gk.cond.Broadcast()
	gk.mutex.Unlock()
}

// UnlockOne allows exactly one waiting goroutine to proceed, even if the gate is generally closed.
// It prioritizes one goroutine if multiple are waiting.
func (gk *GateKeeper) UnlockOne() {
	gk.mutex.Lock()
	gk.counter.Add(1)
	gk.cond.Signal()
	gk.mutex.Unlock()
}


// AllowIf lets a goroutine pass through the gate only if a specific condition is true.
// The condition is defined by the predicate function provided as an argument.
// If the gate is open, the predicate is ignored and the goroutine is allowed to proceed.
func (gk *GateKeeper) AllowIf(predicate func() bool) {
	if predicate() {
		return
	} else {
		gk.Wait()
	}
}

// Wait blocks the calling goroutine until the gate is fully opened.
// It is useful when a goroutine needs to wait indefinitely until unrestricted access is allowed.
func (gk *GateKeeper) Wait() {
	gk.mutex.Lock()
	for !gk.open.Load() {
		gk.cond.Wait()

		if gk.counter.Load() > 0 {
			gk.counter.Add(-1)
			break
		}
	}
	gk.mutex.Unlock()
}
