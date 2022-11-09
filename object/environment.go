package object

// NewEnclosedEnvironment supports adding an environment to an exisiting
// environment.  This allows closures and proper binding within functions
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// NewEnvironment returns a new environment
func NewEnvironment() *Environment {
	s := &ConcurrentMap[string, Object]{
		kv: make(map[string]Object),
	}
	is := &ConcurrentMap[string, struct{}]{
		kv: make(map[string]struct{}),
	}
	return &Environment{store: s, immutableStore: is}
}

// Environment is a map of strings to `Object`s
type Environment struct {
	store *ConcurrentMap[string, Object]
	// store          map[string]Object
	immutableStore *ConcurrentMap[string, struct{}]
	// immutableStore map[string]struct{}

	outer *Environment
}

// func (e *Environment) String() string {
// 	var out bytes.Buffer
// 	out.WriteString("Environment{\n\tstore:\n")
// 	for s, elem := range e.store {
// 		out.WriteString(fmt.Sprintf("\t\t%q -> %s\n", s, elem.Type()))
// 	}
// 	out.WriteString("\n\timmuatebleStore:\n")
// 	for s := range e.immutableStore {
// 		out.WriteString(fmt.Sprintf("\t\t%qs\n", s))
// 	}
// 	out.WriteString("}")
// 	return out.String()
// }

// Get returns the object from the environment store
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store.Get(name)
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set puts a new object into the environment
func (e *Environment) Set(name string, val Object) Object {
	e.store.Put(name, val)
	return val
}

// ImmutableSet puts the name of the identifier in a map and sets it as true
func (e *Environment) ImmutableSet(name string) {
	e.immutableStore.Put(name, struct{}{})
}

// IsImmutable checks if the give identifier name is in the immutable map
func (e *Environment) IsImmutable(name string) bool {
	_, ok := e.immutableStore.Get(name)
	if !ok && e.outer != nil {
		ok = e.outer.IsImmutable(name)
		return ok
	}
	return ok
}

// RemoveIdentifier removes a key from the environment
// this is used in for loops to remove temporary variables
func (e *Environment) RemoveIdentifier(name string) {
	e.store.Remove(name)
	e.immutableStore.Remove(name)
}
