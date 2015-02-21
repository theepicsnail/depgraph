package main

import "github.com/theepicsnail/depgraph"

type MapResolver func([]string) map[string]interface{}

func init() {
	depgraph.Bind("mapresolver", func() interface{} {
		var out MapResolver = func(deps []string) map[string]interface{} {
			out := make(map[string]interface{})
			for i := range deps {
				out[deps[i]] = depgraph.Resolve(deps[i])
			}
			return out
		}
		return out
	})
}
