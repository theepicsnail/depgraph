package main

import "fmt"
import "github.com/theepicsnail/depgraph"

func main() {
	fmt.Println("main")

	node("foo", "bar", "baz", "quz")
	node("bar", "baz")
	node("baz", "quuz")
	node("quz")
	node("quuz")
	b := depgraph.Resolve("foo")
	fmt.Println("Returned value:", b)
}

func node(name string, deps ...string) {
	depgraph.Bind(name, func() interface{} {
		fmt.Println(name, depgraph.Resolve("mapresolver").(MapResolver)(deps))
		return name
	})
}
