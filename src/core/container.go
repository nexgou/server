package core

import (
	"fmt"
	"reflect"
)

// Container is a lightweight IoC container that resolves dependencies by type.
// It stores constructor functions and instantiates singletons on first use.
type Container struct {
	constructors map[reflect.Type]any
	singletons   map[reflect.Type]reflect.Value
}

// NewContainer creates a new empty Container.
func NewContainer() *Container {
	return &Container{
		constructors: make(map[reflect.Type]any),
		singletons:   make(map[reflect.Type]reflect.Value),
	}
}

// Register stores a constructor function indexed by its first return type.
// The constructor must be a function that returns at least one value (typically a pointer).
func (c *Container) Register(ctor any) {
	ctorType := reflect.TypeOf(ctor)
	if ctorType == nil || ctorType.Kind() != reflect.Func {
		panic(fmt.Sprintf("nexgou/container: provider must be a function, got %T", ctor))
	}
	if ctorType.NumOut() == 0 {
		panic("nexgou/container: provider function must return at least one value")
	}
	returnType := ctorType.Out(0)
	c.constructors[returnType] = ctor
}

// Resolve instantiates the dependency for the given reflect.Type by calling its
// registered constructor with recursively resolved arguments.
// Results are cached as singletons — each type is instantiated at most once.
func (c *Container) Resolve(t reflect.Type) (reflect.Value, error) {
	if val, ok := c.singletons[t]; ok {
		return val, nil
	}

	ctor, ok := c.constructors[t]
	if !ok {
		return reflect.Value{}, fmt.Errorf("nexgou/container: no provider registered for type %s", t)
	}

	ctorType := reflect.TypeOf(ctor)
	args := make([]reflect.Value, ctorType.NumIn())

	for i := 0; i < ctorType.NumIn(); i++ {
		argType := ctorType.In(i)
		arg, err := c.Resolve(argType)
		if err != nil {
			return reflect.Value{}, fmt.Errorf(
				"nexgou/container: cannot resolve argument %d (%s) for %s: %w",
				i, argType, t, err,
			)
		}
		args[i] = arg
	}

	results := reflect.ValueOf(ctor).Call(args)
	val := results[0]

	// Support constructors that return (T, error).
	if len(results) > 1 {
		errVal := results[1]
		if errVal.Kind() == reflect.Interface && !errVal.IsNil() {
			return reflect.Value{}, errVal.Interface().(error)
		}
	}

	c.singletons[t] = val
	return val, nil
}
