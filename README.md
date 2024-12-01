# GateKeeper
[![Go Reference](https://pkg.go.dev/badge/github.com/Vortex93/gatekeeper.svg)](https://pkg.go.dev/github.com/Vortex93/gatekeeper)

The `GateKeeper` package provides a concurrency control mechanism for Go applications. It combines features of mutexes, condition variables, and selective unlocking to offer fine-grained control over resource access among goroutines.

## Installation

To use the `GateKeeper` package in your Go project, simply run the following command:
```bash
go get github.com/Vortex93/gatekeeper
```

Then, import it in your Go files:

```go
import "github.com/Vortex93/gatekeeper"
```

## Usage

Below are some simple examples illustrating how to use the `GateKeeper` functions.

### Example: Before Using GateKeeper

Typically, managing access to a shared resource without `GateKeeper` might look like this:

```go
var (
	mutex   sync.Mutex
	cond    = sync.NewCond(&mutex)
	ready   bool
)

func Producer() {
	for {
		mutex.Lock()
		// Simulate some work
		time.Sleep(time.Second)
		fmt.Println("Producer: Work done, ready to consume")
		ready = true
		cond.Signal()
		mutex.Unlock()
		time.Sleep(2 * time.Second) // Delay next production
	}
}

func Consumer() {
	for {
		mutex.Lock()
		for !ready { // Wait until ready is true
			cond.Wait()
		}
		fmt.Println("Consumer: Consuming the resource")
		ready = false // Reset ready after consuming
		mutex.Unlock()
	}
}

func main() {
    go Producer()
    go Consumer()

    // Let the main goroutine sleep to let producer and consumer run
    time.Sleep(10 * time.Second)
}
```

### Example: After Using GateKeeper

Here's how you might handle the same scenario using `GateKeeper` to provide more nuanced control:

```go
var (
	gk = gatekeeper.NewGateKeeper(true) // Initialize the GateKeeper in a locked state
)

func Producer() {
	for {
		// Simulate some work
		time.Sleep(time.Second)
		fmt.Println("Producer: Work done, ready to consume")
		
		// Open the gate for the consumer
		gk.Unlock()
		
		// Delay next production
		time.Sleep(2 * time.Second)
	}
}

func Consumer() {
	for {
		// Wait until the gate is open
		gk.Wait()

		// Consume the resource
		fmt.Println("Consumer: Consuming the resource")
		
		// Close the gate after consuming to wait for next production
		gk.Lock()
	}
}

func main() {
    go Producer()
    go Consumer()

    // Let the main goroutine sleep to let producer and consumer run
    time.Sleep(10 * time.Second)
}
```

### More Examples

#### Using `Lock` and `Unlock`

```go
gk := gatekeeper.NewGateKeeper(true)
gk.Lock()
fmt.Println("Gate is now locked.")
gk.Unlock()
fmt.Println("Gate is now open.")
```

#### Using `UnlockOne`

```go
// Allows only one goroutine to proceed even if the gate is locked.
gk := gatekeeper.NewGateKeeper(true)
go func() {
    gk.UnlockOne() // Allow only one goroutine to pass.
}()
```

#### Using `AllowIf`

```go
gk := gatekeeper.NewGateKeeper(true)
go func() {
    gk.AllowIf(func() bool {
        return true // Allow only goroutine whose condition is met.
    })
    fmt.Println("Condition met, gate opened for this goroutine.")
}()
```

#### Using `Wait`

```go
gk := gatekeeper.NewGateKeeper(true)
go func() {
    gk.Unlock() // This will allow the `Wait` below to complete.
}()
gk.Wait()
fmt.Println("Gate fully open, all goroutines may proceed.")
```

#### Using `Reset`

```go
gk := gatekeeper.NewGateKeeper(true)
gk.Reset() // Resets the gate to its initial state
fmt.Println("Gate has been reset.")
```

## Functions

### `NewGateKeeper(locked bool) *GateKeeper`

Initializes a new `GateKeeper`. If `locked` is true, the gate starts in a locked state.

### `IsLocked() bool`

Checks if the gate is in a locked state.

### `IsUnlocked() bool`

Checks if the gate is in an open state.

### `Lock()`

Sets the gate to a locked state, preventing goroutines from passing until it is unlocked.

### `Unlock()`

Sets the gate to an open state, allowing all waiting goroutines to proceed.

### `UnlockOne()`

Allows exactly one waiting goroutine to proceed, even if the gate is generally closed.

### `AllowIf(predicate func() bool)`

Lets a goroutine pass through the gate only if a specific condition is true. If the gate is open, the predicate is ignored and the goroutine is allowed to proceed.

### `Wait()`

Blocks the calling goroutine until the gate is fully opened.

### `Reset()`

Resets the gate to its initial state, closing it and resetting the counter.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.

## License

This project is licensed under the MIT License.
