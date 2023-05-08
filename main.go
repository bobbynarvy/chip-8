package main

import (
	"fmt"
)

var currentVm Vm

func main() {
	fmt.Println("Init WASM")
	setup()
	<-make(chan bool)
}
