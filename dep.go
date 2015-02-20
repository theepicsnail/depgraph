package depgraph

import (
	"errors"
	"fmt"
)
import "sync"

/*
	Loading a depdency. The args passed in are the resolved values from
	its deps. in the order specified.

	Return either an interface{}, or an error
*/
type Loader func(...interface{}) (interface{}, error)

var NoDeps = []string{}

type Resolution struct {
	value interface{}
	err   error
}

type Node struct {
	name       string
	deps       []string
	loader     Loader
	once       *sync.Once
	resolution *Resolution
}

func (node *Node) errorResolution(err error) {
	node.resolution.err = err
	node.resolution.value = nil
}
func (node *Node) valueResolution(val interface{}) {
	node.resolution.err = nil
	node.resolution.value = val
}

var mapLock = new(sync.RWMutex)
var nodeMap = make(map[string]*Node)

func readMap(key string) (*Node, bool) {
	mapLock.RLock()
	defer mapLock.RUnlock()
	val, err := nodeMap[key]
	return val, err
}
func writeMap(key string, val *Node) {
	mapLock.Lock()
	defer mapLock.Unlock()
	nodeMap[key] = val
}

func (node *Node) resolve() {
	node.once.Do(func() {
		fmt.Println("resolving ", node.name)
		// Check that all deps exist in the map first
		for idx := range node.deps {
			if _, ok := readMap(node.deps[idx]); !ok {
				node.errorResolution(errors.New(
					fmt.Sprintf("Node (%v) doesn't exist", node.deps[idx])))
				return
			}
		}

		// Ensure each dep has been resolved
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

func NewNode(name string, deps []string, loader Loader) (*Node, error) {
	node := &Node{
		name, deps, loader, new(sync.Once), new(Resolution),
	}

	return addNode(node)
}

func addNode(node *Node) (*Node, error) {
	fmt.Println("Add node", node)

	// safe read, make sure it hasn't been added.
	if _, ok := readMap(node.name); ok {
		return nil, errors.New(fmt.Sprintf("Node '%v' exists.", node.name))
	}

	// add it!
	writeMap(node.name, node)
	return node, nil
}

func Resolve(name string) (interface{}, error) {
	if node, ok := readMap(name); !ok {
		return nil, errors.New(fmt.Sprintf("Node '%v' not in graph.", name))
	} else {
		node.resolve()
		return node.resolution.value, node.resolution.err
	}
}
