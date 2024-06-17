package gatekeeper

import (
	"sync"
	"testing"
	"time"
)

func TestNewGateKeeper(t *testing.T) {
	gk := NewGateKeeper(true)
	if gk.IsUnlocked() {
		t.Error("Expected gate to be locked")
	}

	gk = NewGateKeeper(false)
	if gk.IsLocked() {
		t.Error("Expected gate to be unlocked")
	}
}

func TestIsLocked(t *testing.T) {
	gk := NewGateKeeper(true)
	if gk.IsUnlocked() {
		t.Error("Expected gate to be locked")
	}

	gk.Unlock()
	if gk.IsLocked() {
		t.Error("Expected gate to be unlocked")
	}

	gk.Lock()
	if gk.IsUnlocked() {
		t.Error("Expected gate to be locked")
	}
}

func TestUnlock(t *testing.T) {
	var t0 time.Time
	var td time.Duration

	gk := NewGateKeeper(true) // Start in locked state

	go func() {
		t0 = time.Now()
		time.Sleep(100 * time.Millisecond)
		gk.Unlock()
	}()

	gk.Wait()
	td = time.Since(t0)

	if td < 100*time.Millisecond {
		t.Error("Expected Unlock to block for at least 100ms")
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

	allow_0 := false
	allow_1 := false

	go func() {
		gk.AllowIf(func() bool {
			allow_0 = true
			return true
		})

		if !allow_0 {
			t.Error("Expected allow_0 to be false")
		}
	}()

	go func() {
		gk.AllowIf(func() bool {
			allow_1 = false
			return false
		})

		if allow_1 {
			t.Error("Expected allow_1 to be false")
		}
	}()

	time.Sleep(100 * time.Millisecond)
	gk.Unlock()
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
