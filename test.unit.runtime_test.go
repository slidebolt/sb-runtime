package runtime

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	contract "github.com/slidebolt/sb-contract"
)

func TestHelloCommand(t *testing.T) {
	// Build a tiny test binary using `go run` isn't practical here,
	// so we test the hello function directly via subprocess pattern.
	if os.Getenv("SB_TEST_SUBPROCESS") == "1" {
		os.Args = []string{"test", "hello"}
		Run(&stubBinary{})
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestHelloCommand")
	cmd.Env = append(os.Environ(), "SB_TEST_SUBPROCESS=1")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("subprocess failed: %v", err)
	}

	var resp contract.HelloResponse
	if err := json.Unmarshal(bytes.TrimSpace(out), &resp); err != nil {
		t.Fatalf("parse hello output: %v", err)
	}

	if resp.ID != "stub" {
		t.Errorf("got id %q, want %q", resp.ID, "stub")
	}
	if resp.ContractVersion != contract.ContractVersion {
		t.Errorf("got version %d, want %d", resp.ContractVersion, contract.ContractVersion)
	}
}

func TestStartAndShutdown(t *testing.T) {
	if os.Getenv("SB_TEST_SUBPROCESS") == "1" {
		os.Args = []string{"test", "start"}
		Run(&stubBinary{})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestStartAndShutdown")
	cmd.Env = append(os.Environ(), "SB_TEST_SUBPROCESS=1")

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	dec := json.NewDecoder(stdout)

	// Should receive ready.
	var ready contract.RuntimeMessage
	if err := dec.Decode(&ready); err != nil {
		t.Fatalf("decode ready: %v", err)
	}
	if ready.Type != contract.RuntimeReady {
		t.Errorf("got type %q, want %q", ready.Type, contract.RuntimeReady)
	}

	// Send shutdown.
	contract.WriteJSON(stdin, contract.ControlMessage{Type: contract.ControlShutdown})
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		t.Fatalf("process exit: %v", err)
	}
}

type stubBinary struct{}

func (s *stubBinary) Hello() contract.HelloResponse {
	return contract.HelloResponse{
		ID:              "stub",
		Kind:            contract.KindService,
		ContractVersion: contract.ContractVersion,
	}
}

func (s *stubBinary) OnStart(deps map[string]json.RawMessage) (json.RawMessage, error) {
	return nil, nil
}
func (s *stubBinary) OnShutdown() error { return nil }
