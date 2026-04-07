package gatekeeper

import "sync"

// GateKeeper controls whether goroutines may pass.
type GateKeeper struct {
	open    bool
	permits int
	waiters int
	mutex   sync.Mutex
	cond    *sync.Cond
}

func NewGateKeeper(locked bool) *GateKeeper {
	gk := &GateKeeper{
		open: !locked,
	}
	gk.cond = sync.NewCond(&gk.mutex)
	return gk
}

func (gk *GateKeeper) IsLocked() bool {
	gk.mutex.Lock()
	defer gk.mutex.Unlock()
	return !gk.open
}

func (gk *GateKeeper) IsUnlocked() bool {
	gk.mutex.Lock()
	defer gk.mutex.Unlock()
	return gk.open
}

func (gk *GateKeeper) Lock() {
	gk.mutex.Lock()
	gk.open = false
	gk.mutex.Unlock()
}

func (gk *GateKeeper) Unlock() {
	gk.mutex.Lock()
	gk.open = true
	gk.cond.Broadcast()
	gk.mutex.Unlock()
}

func (gk *GateKeeper) UnlockOne() {
	gk.mutex.Lock()
	if gk.waiters > 0 {
		gk.permits++
		gk.cond.Signal()
	}
	gk.mutex.Unlock()
}

func (gk *GateKeeper) Wait() {
	gk.mutex.Lock()
	gk.waiters++
	defer func() {
		gk.waiters--
		gk.mutex.Unlock()
	}()

	for !gk.open && gk.permits == 0 {
		gk.cond.Wait()
	}

	if !gk.open && gk.permits > 0 {
		gk.permits--
	}
}

func (gk *GateKeeper) SkipIf(predicate func() bool) {
	if predicate() {
		return
	}
	gk.Wait()
}

func (gk *GateKeeper) Reset() {
	gk.mutex.Lock()
	gk.open = false
	gk.permits = 0
	gk.mutex.Unlock()
}
