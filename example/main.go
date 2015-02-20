package main

import "fmt"
import "github.com/theepicsnail/depgraph"

func main() {
	fmt.Println("main")

	node("foo", "bar", "baz", "quz")
	node("bar", "baz")
	node("baz", "quuz")
	node("quz")

	node("bar", "foo")

	b, err := depgraph.Resolve("foo")
	fmt.Println("Returned value:", b)
	fmt.Println("Returned error:", err)
}

func node(name string, deps ...string) (*depgraph.Node, error) {
	return depgraph.NewNode(name, deps, func(d ...interface{}) (interface{}, error) {
		fmt.Println("Loaded:", name, deps)
		return name, nil
	})
}
