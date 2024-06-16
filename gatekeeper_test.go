package gatekeeper

import (
	"sync"
	"testing"
	"time"
)

func TestNewGateKeeper(t *testing.T) {
	gk := NewGateKeeper(true)
	if !gk.IsLocked() {
		t.Error("Expected gate to be locked")
	}

	gk = NewGateKeeper(false)
	if gk.IsLocked() {
		t.Error("Expected gate to be unlocked")
	}
}

func TestIsLocked(t *testing.T) {
	gk := NewGateKeeper(true)
	if !gk.IsLocked() {
		t.Error("Expected gate to be locked")
	}

	gk.Unlock()
	if gk.IsLocked() {
		t.Error("Expected gate to be unlocked")
	}
}

func TestLock(t *testing.T) {
	gk := NewGateKeeper(false)
	gk.Lock()
	if !gk.IsLocked() {
		t.Error("Expected gate to be locked")
	}
}

func TestUnlock(t *testing.T) {
	gk := NewGateKeeper(true)
	gk.Unlock()
	if gk.IsLocked() {
		t.Error("Expected gate to be unlocked")
	}
}

func TestUnlockOne(t *testing.T) {
	gk := NewGateKeeper(true)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)	

		go func(i int) {
			defer wg.Done()
			gk.Wait()
		}(i)
	}

	
	time.Sleep(100 * time.Millisecond)
	gk.UnlockOne()
	gk.UnlockOne()
	gk.UnlockOne()
	time.Sleep(100 * time.Millisecond)
	if !gk.IsLocked() {
		t.Error("Expected gate to be locked after letting one goroutine through")
	}

	time.Sleep(1 * time.Second)
	gk.Unlock()
	wg.Wait()
}

func TestAllowIf(t *testing.T) {
	gk := NewGateKeeper(true)
	condition := false

	go func() {
		time.Sleep(100 * time.Millisecond)
		gk.mutex.Lock()
		condition = true
		gk.mutex.Unlock()
		gk.cond.Broadcast()
	}()

	gk.AllowIf(func() bool {
		return condition
	})

	if !condition {
		t.Error("Expected condition to be true")
	}
}

func TestWait(t *testing.T) {
	gk := NewGateKeeper(true)

	go func() {
		time.Sleep(100 * time.Millisecond)
		gk.Unlock()
	}()

	start := time.Now()
	gk.Wait()
	duration := time.Since(start)

	if duration < 100*time.Millisecond {
		t.Error("Expected Wait to block for at least 100ms")
	}
}
