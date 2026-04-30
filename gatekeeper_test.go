package gatekeeper

import (
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
	gk := NewGateKeeper(true) // Start in locked state

	go func() {
		time.Sleep(100 * time.Millisecond)
		gk.Unlock()
	}()

	start := time.Now()
	gk.Wait()

	if time.Since(start) < 100*time.Millisecond {
		t.Error("Expected Unlock to block for at least 100ms")
	}
}

func TestUnlockOne(t *testing.T) {
	gk := NewGateKeeper(true)
	started := make(chan struct{}, 3)
	passed := make(chan struct{}, 3)

	for i := 0; i < 3; i++ {
		go func() {
			started <- struct{}{}
			gk.Wait()
			passed <- struct{}{}
		}()
	}

	for i := 0; i < 3; i++ {
		<-started
	}

	time.Sleep(100 * time.Millisecond)

	for i := 1; i <= 3; i++ {
		gk.UnlockOne()

		select {
		case <-passed:
		case <-time.After(250 * time.Millisecond):
			t.Fatalf("Expected UnlockOne #%d to release one goroutine", i)
		}

		select {
		case <-passed:
			t.Fatalf("Expected UnlockOne #%d to release only one goroutine", i)
		case <-time.After(100 * time.Millisecond):
		}

		if !gk.IsLocked() {
			t.Fatal("Expected gate to remain locked after UnlockOne")
		}
	}
}

func TestAllowIf(t *testing.T) {
	t.Run("predicate true", func(t *testing.T) {
		gk := NewGateKeeper(true)
		allow0 := false

		gk.AllowIf(func() bool {
			allow0 = true
			return true
		})

		if !allow0 {
			t.Fatal("Expected allow_0 to be true")
		}
	})

	t.Run("predicate false waits", func(t *testing.T) {
		gk := NewGateKeeper(true)
		allow1 := false
		done := make(chan struct{})

		go func() {
			gk.AllowIf(func() bool {
				return allow1
			})
			close(done)
		}()

		select {
		case <-done:
			t.Fatal("Expected AllowIf to wait while predicate is false and gate is locked")
		case <-time.After(100 * time.Millisecond):
		}

		gk.Unlock()

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Expected AllowIf to return after Unlock")
		}

		if allow1 {
			t.Fatal("Expected allow_1 to be false")
		}
	})
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

func TestWaitConsumesStoredPermit(t *testing.T) {
	gk := NewGateKeeper(true)
	gk.UnlockOne()

	done := make(chan struct{})
	go func() {
		gk.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected Wait to consume a stored UnlockOne permit")
	}

	if !gk.IsLocked() {
		t.Fatal("Expected gate to remain locked after consuming a stored permit")
	}
}

func TestTryWait(t *testing.T) {
	gk := NewGateKeeper(true)

	if gk.TryWait() {
		t.Fatal("Expected TryWait to fail while the gate is locked")
	}

	gk.UnlockOne()
	if !gk.TryWait() {
		t.Fatal("Expected TryWait to consume a stored UnlockOne permit")
	}

	if gk.TryWait() {
		t.Fatal("Expected TryWait to fail after the stored permit is consumed")
	}

	gk.Unlock()
	if !gk.TryWait() {
		t.Fatal("Expected TryWait to succeed while the gate is open")
	}

	gk.Lock()
	if gk.TryWait() {
		t.Fatal("Expected full Unlock to clear stored single-use permits")
	}
}

func TestResetClearsStoredPermit(t *testing.T) {
	gk := NewGateKeeper(true)
	gk.UnlockOne()
	gk.Reset()

	if gk.TryWait() {
		t.Fatal("Expected Reset to clear stored single-use permits")
	}
}
