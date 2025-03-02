package main

import (
	// #cgo pkg-config: python3
	// #include "Python.h"
	"C"
	"fmt"
)

func main() {
	C.Py_Initialize()
	fmt.Println(C.GoString(C.Py_GetVersion()))
	C.Py_Finalize()
}
