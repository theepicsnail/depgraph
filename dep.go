package depgraph

import (
	"errors"
	"fmt"
	"sync"
)

// Map and map accessories */
var mapLock = new(sync.RWMutex)
var nodeMap = make(map[string]*Node)

// Threadsafe reading of map
func readMap(key string) (*Node, bool) {
	mapLock.RLock()
	defer mapLock.RUnlock()
	val, err := nodeMap[key]
	return val, err
}

// Threadsafe writing of map
func writeMap(key string, val *Node) {
	mapLock.Lock()
	defer mapLock.Unlock()
	nodeMap[key] = val
}

//
// Callback that's called when a Node is loaded.
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
type Loader func(...interface{}) (interface{}, error)

// Convenient value for providing no args.
var NoDeps = []string{}

//
// Private model for the resolution of a Node.
// This should only ever have 1 non-nil field.
// These are filled in through errorResolution or valueResolution
//
type resolution struct {
	value interface{}
	err   error
}

func (node *Node) errorResolution(err error) {
	node.resolution.err = err
	node.resolution.value = nil
}
func (node *Node) valueResolution(val interface{}) {
	node.resolution.err = nil
	node.resolution.value = val
}

//
// Models a dependency. This shouldn't really be exposed.
// TODO: Hide these details. probably move a few things to panics
//
type Node struct {
	name       string
	deps       []string
	loader     Loader
	once       *sync.Once
	resolution *resolution
}

//
// Recursively resolve this nodes dependecies then call this nodes
// loader function with those dependecies passed in.
//
func (node *Node) resolve() {
	node.once.Do(func() {
		// Check that all deps exist in the map first
		for idx := range node.deps {
			if _, ok := readMap(node.deps[idx]); !ok {
				node.errorResolution(errors.New(
					fmt.Sprintf("Node (%v) doesn't exist", node.deps[idx])))
				return
			}
		}

		// Ensure each node has been resolved.
		// Also starts resolving any unresolved deps.
		wg := new(sync.WaitGroup)
		for idx := range node.deps {
			dep_node, _ := readMap(node.deps[idx])
			wg.Add(1)
			go func() {

				dep_node.resolve()

				wg.Done()
			}()
		}
		wg.Wait()

		// Check for errors, and build resolved deps slice
		resolved := []interface{}{}
		for idx := range node.deps {
			dep, _ := readMap(node.deps[idx])
			if dep.resolution.err != nil {
				node.errorResolution(errors.New(
					fmt.Sprintf("Failed to resolve %v:\n%v", node.name, dep.resolution.err)))
				return
			}
			resolved = append(resolved, dep.resolution.value)
		}

		value, err := node.loader(resolved...)
		if err != nil {
			node.errorResolution(err)
		} else {
			node.valueResolution(value)
		}
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
// This returns the Node that was created TODO: Stop this.
// or it returns an error in the case of collisions TODO: this should panic?
//
func NewNode(name string, deps []string, loader Loader) (*Node, error) {
	node := &Node{
		name, deps, loader, new(sync.Once), new(resolution),
	}

	// safe read, make sure it hasn't been added.
	if _, ok := readMap(node.name); ok {
		return nil, errors.New(fmt.Sprintf("Node '%v' exists.", node.name))
	}

	// add it!
	writeMap(node.name, node)
	return node, nil
}

//
// Resolve a dependency and get its value (iterface{}) or if it fails to
// resolve, get the error.
//
func Resolve(name string) (interface{}, error) {
	if node, ok := readMap(name); !ok {
		return nil, errors.New(fmt.Sprintf("Node '%v' not in graph.", name))
	} else {
		node.resolve()
		return node.resolution.value, node.resolution.err
	}
}
