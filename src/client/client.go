package main

import (
	"github.com/gopherjs/gopherjs/js"
)

func main() {
	js.Global.Set("client", map[string]interface{}{
		"start": start,
	})
}

func start() {
}
