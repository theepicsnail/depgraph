package depgraph

import (
	"fmt"
	"sync"
)

// Map and map accessories */
var mapLock = new(sync.RWMutex)
var nodeMap = make(map[string]*dependency)

// Threadsafe reading of map
func readMap(key string) (*dependency, bool) {
	mapLock.RLock()
	defer mapLock.RUnlock()
	val, err := nodeMap[key]
	return val, err
}

// Threadsafe writing of map
func writeMap(key string, val *dependency) {
	mapLock.Lock()
	defer mapLock.Unlock()
	nodeMap[key] = val
}

//
// Callback that's called when a dependency is loaded.
//
// The argument is the slice of interface{}s returned from each dependency that
// was loaded.
//
// The returned interface{} value will be passed in to other Loaders that
// depend on this. It's also the value that's returned from Resolve(name).
//
// If the returned error is not nil, then the loader is considered failed, and
// any dependants will also fail to load.
//
type Loader func() interface{}

//
// Models a dependency. This shouldn't really be exposed.
// TODO: Hide these details. probably move a few things to panics
//
type dependency struct {
	name       string
	loader     Loader
	once       *sync.Once
	resolution interface{}
}

//
// Recursively resolve this nodes dependecies then call this nodes
// loader function with those dependecies passed in.
//
func (node *dependency) resolve() {
	node.once.Do(func() {
		node.resolution = node.loader()
	})
}

//
// Add a node in the dependency graph.
// name - Name of the dependency to be added. This is just an arbitrary string
//        but it must be unique. Dots/underscores/etc.. are all acceptable
// deps - Slice of dep names. These are passed into the loader callback in the
//        provided order.
// loader - Callback to be called when/if this node needs resolved. The value
//          returned by this method is given to other loaders, or Resolve calls
//
// This returns the dependency that was created TODO: Stop this.
// or it returns an error in the case of collisions TODO: this should panic?
//
func Bind(name string, loader Loader) {
	node := &dependency{
		name, loader, new(sync.Once), nil,
	}

	// safe read, make sure it hasn't been added.
	if _, ok := readMap(node.name); ok {
		panic(fmt.Sprintf("dependency '%v' exists.", node.name))
	}
	// This should have a threading issue here? Multiple threads can see
	// that it doesn't exist, then add it. Bind("Foo", a) Bind("Foo", b) can both
	// be here at the same time.

	// add it!
	writeMap(node.name, node)
}

//
// Resolve a dependency and get its value (iterface{}) or if it fails to
// resolve, get the error.
//
func Resolve(name string) interface{} {
	if node, ok := readMap(name); !ok {
		panic(fmt.Sprintf("dependency '%v' not in graph.", name))
	} else {
		node.resolve()
		return node.resolution
	}
}
