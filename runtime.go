// Package runtime provides the standard harness for SlideBolt binaries.
// A binary implements the Binary interface and calls Run().
package runtime

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	contract "github.com/slidebolt/sb-contract"
)

// Binary is the interface every SlideBolt binary implements.
type Binary interface {
	// Hello returns the manifest for this binary. Must have no side effects.
	Hello() contract.HelloResponse
	// OnStart is called after all dependency payloads have been delivered.
	// deps is keyed by dependency ID with raw JSON payloads.
	// Return a payload to advertise to dependents (or nil).
	OnStart(deps map[string]json.RawMessage) (json.RawMessage, error)
	// OnShutdown is called when the manager requests a graceful stop.
	OnShutdown() error
}

// Run is the entry point for a SlideBolt binary. It dispatches based on
// os.Args and handles the stdin/stdout contract with the manager.
func Run(b Binary) {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: <binary> <hello|start>")
		os.Exit(1)
	}

	cmd := os.Args[1]

	// Set up structured logging for all commands except hello (which must
	// only write the manifest to stdout).
	if cmd != "hello" {
		SetupLogger(b.Hello().ID)
	}

	switch cmd {
	case "hello":
		hello(b)
	case "start":
		start(b)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func hello(b Binary) {
	resp := b.Hello()
	if err := contract.WriteJSON(os.Stdout, resp); err != nil {
		log.Fatal(err)
	}
}

func start(b Binary) {
	manifest := b.Hello()
	scanner := bufio.NewScanner(os.Stdin)

	// Collect dependency payloads before starting.
	deps := make(map[string]json.RawMessage)
	expected := len(manifest.DependsOn)

	for len(deps) < expected {
		if !scanner.Scan() {
			contract.WriteJSON(os.Stdout, contract.RuntimeMessage{
				Type:    contract.RuntimeError,
				Message: "stdin closed before all dependencies received",
			})
			os.Exit(1)
		}

		var msg contract.ControlMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}

		if msg.Type == contract.ControlDependency {
			deps[msg.ID] = msg.Payload
		}
	}

	// Start the binary with its dependencies.
	payload, err := b.OnStart(deps)
	if err != nil {
		contract.WriteJSON(os.Stdout, contract.RuntimeMessage{
			Type:    contract.RuntimeError,
			Message: fmt.Sprintf("start failed: %v", err),
		})
		os.Exit(1)
	}

	// Signal ready with optional payload.
	contract.WriteJSON(os.Stdout, contract.RuntimeMessage{
		Type:    contract.RuntimeReady,
		Payload: payload,
	})

	// Listen for control messages.
	for scanner.Scan() {
		var msg contract.ControlMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			contract.WriteJSON(os.Stdout, contract.RuntimeMessage{
				Type:    contract.RuntimeError,
				Message: fmt.Sprintf("bad control message: %v", err),
			})
			continue
		}

		switch msg.Type {
		case contract.ControlShutdown:
			if err := b.OnShutdown(); err != nil {
				contract.WriteJSON(os.Stdout, contract.RuntimeMessage{
					Type:    contract.RuntimeError,
					Message: fmt.Sprintf("shutdown error: %v", err),
				})
			}
			os.Exit(0)
		default:
			contract.WriteJSON(os.Stdout, contract.RuntimeMessage{
				Type:    contract.RuntimeLog,
				Level:   "warn",
				Message: fmt.Sprintf("unknown control type: %s", msg.Type),
			})
		}
	}
}
