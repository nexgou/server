package pipe_test

import (
	"testing"

	"github.com/nexgou/server/src/pipe"
)

// ── ParseIntPipe ──────────────────────────────────────────────────────────────

func TestParseIntPipe_Valid(t *testing.T) {
	p := &pipe.ParseIntPipe{}
	v, err := p.Transform("42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.(int) != 42 {
		t.Errorf("got %v, want 42", v)
	}
}

func TestParseIntPipe_Negative(t *testing.T) {
	p := &pipe.ParseIntPipe{}
	v, err := p.Transform("-7")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.(int) != -7 {
		t.Errorf("got %v, want -7", v)
	}
}

func TestParseIntPipe_Invalid(t *testing.T) {
	p := &pipe.ParseIntPipe{}
	_, err := p.Transform("abc")
	if err == nil {
		t.Fatal("expected error for non-integer input")
	}
}

// ── ParseUUIDPipe ─────────────────────────────────────────────────────────────

func TestParseUUIDPipe_Valid(t *testing.T) {
	p := &pipe.ParseUUIDPipe{}
	uuid := "550e8400-e29b-41d4-a716-446655440000"
	v, err := p.Transform(uuid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.(string) != uuid {
		t.Errorf("got %v, want %v", v, uuid)
	}
}

func TestParseUUIDPipe_Invalid(t *testing.T) {
	p := &pipe.ParseUUIDPipe{}
	_, err := p.Transform("not-a-uuid")
	if err == nil {
		t.Fatal("expected error for invalid UUID")
	}
}

// ── DefaultValuePipe ──────────────────────────────────────────────────────────

func TestDefaultValuePipe_Empty(t *testing.T) {
	p := &pipe.DefaultValuePipe{Default: "guest"}
	v, err := p.Transform("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.(string) != "guest" {
		t.Errorf("got %v, want guest", v)
	}
}

func TestDefaultValuePipe_NonEmpty(t *testing.T) {
	p := &pipe.DefaultValuePipe{Default: "guest"}
	v, err := p.Transform("alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.(string) != "alice" {
		t.Errorf("got %v, want alice", v)
	}
}
