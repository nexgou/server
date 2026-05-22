package core_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/core"
)

// ── basic resolution ──────────────────────────────────────────────────────────

type Service struct{ Name string }
type Repo struct{}

func NewRepo() *Repo       { return &Repo{} }
func NewService() *Service { return &Service{Name: "svc"} }

func TestContainer_RegisterAndResolve(t *testing.T) {
	c := core.NewContainer()
	c.Register(NewService)

	val, err := c.Resolve(reflectTypeOf[*Service]())
	if err != nil {
		t.Fatalf("Resolve: unexpected error: %v", err)
	}
	svc, ok := val.Interface().(*Service)
	if !ok || svc.Name != "svc" {
		t.Errorf("Resolve: got %v", val.Interface())
	}
}

func TestContainer_Singleton(t *testing.T) {
	c := core.NewContainer()
	c.Register(NewService)

	v1, _ := c.Resolve(reflectTypeOf[*Service]())
	v2, _ := c.Resolve(reflectTypeOf[*Service]())

	if v1.Pointer() != v2.Pointer() {
		t.Error("Resolve: should return the same singleton instance")
	}
}

func TestContainer_DependencyChain(t *testing.T) {
	type DB struct{}
	type Svc struct{ DB *DB }

	c := core.NewContainer()
	c.Register(func() *DB { return &DB{} })
	c.Register(func(db *DB) *Svc { return &Svc{DB: db} })

	val, err := c.Resolve(reflectTypeOf[*Svc]())
	if err != nil {
		t.Fatalf("chain: unexpected error: %v", err)
	}
	svc := val.Interface().(*Svc)
	if svc.DB == nil {
		t.Error("chain: DB dependency not injected")
	}
}

func TestContainer_MissingProvider(t *testing.T) {
	c := core.NewContainer()
	_, err := c.Resolve(reflectTypeOf[*Service]())
	if err == nil {
		t.Fatal("expected error for missing provider")
	}
}

func TestContainer_MissingTransitiveDep(t *testing.T) {
	type DB struct{}
	type Svc struct{ DB *DB }

	c := core.NewContainer()
	// Register Svc but NOT DB
	c.Register(func(db *DB) *Svc { return &Svc{DB: db} })

	_, err := c.Resolve(reflectTypeOf[*Svc]())
	if err == nil {
		t.Fatal("expected error for missing transitive dep")
	}
}

func TestContainer_ConstructorReturnsError(t *testing.T) {
	type Failing struct{}

	c := core.NewContainer()
	c.Register(func() (*Failing, error) {
		return nil, errors.New("init failed")
	})

	_, err := c.Resolve(reflectTypeOf[*Failing]())
	if err == nil {
		t.Fatal("expected constructor error to propagate")
	}
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

func TestContainer_ConstructorReturnsNoError(t *testing.T) {
	type Good struct{}

	c := core.NewContainer()
	c.Register(func() (*Good, error) {
		return &Good{}, nil
	})

	val, err := c.Resolve(reflectTypeOf[*Good]())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.IsNil() {
		t.Error("value should not be nil")
	}
}

func TestContainer_PanicsOnNonFunction(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for non-function provider")
		}
	}()
	c := core.NewContainer()
	c.Register("not a function")
}

func TestContainer_PanicsOnNoReturn(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for no-return function")
		}
	}()
	c := core.NewContainer()
	c.Register(func() {})
}

// ── module ────────────────────────────────────────────────────────────────────

func TestNewModule_Options(t *testing.T) {
	m := core.NewModule(common.ModuleOptions{
		Providers: []any{NewService},
		Exports:   []any{NewService},
	})
	opts := m.Options()
	if opts == nil {
		t.Fatal("Options: returned nil")
	}
	if len(opts.Providers) != 1 {
		t.Errorf("Providers: got %d, want 1", len(opts.Providers))
	}
}

func TestNewModule_Empty(t *testing.T) {
	m := core.NewModule(common.ModuleOptions{})
	opts := m.Options()
	if opts == nil {
		t.Fatal("Options: returned nil")
	}
}

// ── helper ────────────────────────────────────────────────────────────────────

func reflectTypeOf[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}
