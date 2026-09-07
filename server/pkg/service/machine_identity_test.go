package service

import (
	"context"
	"testing"

	"toolplane/pkg/model"
	"toolplane/pkg/trace"
)

func newMachineIdentityStack() *MachinesService {
	toolService := NewToolService(trace.NopTracer(), nil)
	return NewMachinesService(context.Background(), toolService, trace.NopTracer(), nil)
}

func TestRegisterMachineMintsOnceOnlyCredential(t *testing.T) {
	machines := newMachineIdentityStack()

	machine, err := machines.RegisterMachine("session-mid", "", "1.0", "go", "127.0.0.1", nil, "")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if machine.Token == "" {
		t.Fatal("fresh registration must return the minted credential once")
	}
	if machine.TokenHash == "" {
		t.Fatal("registration must store the credential hash")
	}
	if !machine.MatchesMachineToken(machine.Token) {
		t.Fatal("stored hash must match the minted credential")
	}

	// The cached copy must not retain the plaintext.
	cached, err := machines.GetMachineByID("session-mid", machine.ID)
	if err != nil {
		t.Fatalf("get machine: %v", err)
	}
	if cached.Token != "" {
		t.Fatal("cached machine retained the plaintext credential")
	}
}

func TestRegisterMachineReRegistrationRequiresCredential(t *testing.T) {
	machines := newMachineIdentityStack()

	first, err := machines.RegisterMachine("session-mid", "machine-1", "1.0", "go", "127.0.0.1", nil, "")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	token := first.Token

	// A different caller presenting no credential cannot take over the ID.
	if _, err := machines.RegisterMachine("session-mid", "machine-1", "1.0", "go", "127.0.0.1", nil, ""); err == nil {
		t.Fatal("re-registration without the machine credential must fail")
	}

	// A wrong credential is rejected too.
	if _, err := machines.RegisterMachine("session-mid", "machine-1", "1.0", "go", "127.0.0.1", nil, "forged"); err == nil {
		t.Fatal("re-registration with a forged credential must fail")
	}

	// The legitimate credential re-registers.
	reregistered, err := machines.RegisterMachine("session-mid", "machine-1", "1.1", "go", "127.0.0.1", nil, token)
	if err != nil {
		t.Fatalf("re-registration with the credential: %v", err)
	}
	if reregistered.Token != "" {
		t.Fatal("re-registration must not re-reveal the credential")
	}
	if reregistered.SDKVersion != "1.1" {
		t.Fatalf("re-registration did not update the machine: %s", reregistered.SDKVersion)
	}
}

func TestAuthorizeMachineTokenBindsLegacyAndValidates(t *testing.T) {
	machines := newMachineIdentityStack()

	// Unknown machines pass (registration will mint).
	if err := machines.AuthorizeMachineToken("session-mid", "machine-x", ""); err != nil {
		t.Fatalf("unknown machine: %v", err)
	}

	// Pre-token machine row: first presented credential binds.
	legacy := &model.Machine{ID: "machine-legacy", SessionID: "session-mid"}
	machines.machines["session-mid"] = map[string]*model.Machine{"machine-legacy": legacy}
	if err := machines.AuthorizeMachineToken("session-mid", "machine-legacy", "first-token"); err != nil {
		t.Fatalf("legacy bind: %v", err)
	}
	if !legacy.MatchesMachineToken("first-token") {
		t.Fatal("legacy machine did not bind the first presented credential")
	}
	// After binding, the gate validates.
	if err := machines.AuthorizeMachineToken("session-mid", "machine-legacy", "other-token"); err == nil {
		t.Fatal("mismatched credential must be rejected after binding")
	}
	if err := machines.AuthorizeMachineToken("session-mid", "machine-legacy", "first-token"); err != nil {
		t.Fatalf("bound credential: %v", err)
	}
}
