package main

// Much of this is inspired by: https://github.com/ardanlabs/python-go/
import (
	// #cgo pkg-config: python3-embed
	// #include <Python.h>
	// #include "glue.h"
	"C"
)
import (
	"fmt"
	"time"
	"unsafe"
)

func main() {
	C.Py_Initialize()
	defer C.Py_Finalize()

	for i := range 10 {
		start := time.Now()
		loadPyFuncAndGetNumber("testfile", "number")
		end := time.Now()
		fmt.Printf("Python took %v \n; iteration: %v \n", end.Sub(start), i)
	}

	fmt.Println(getPodNameViaCConversion())
	for i := range 10 {
		start := time.Now()
		getPodNameViaPyFunc()
		end := time.Now()
		fmt.Printf("Python with a param took %v \n; iteration: %v \n", end.Sub(start), i)
	}
}

func loadPyFuncAndGetNumber(moduleName, funcName string) {

	fn := loadPyFunc(moduleName, funcName)

	result := C.PyObject_CallObject(fn, nil)

	cLong := C.PyLong_AsLong(result)
	fmt.Println(cLong)
}

func loadPyFunc(moduleName, funcName string) *C.PyObject {

	cMod := C.CString(moduleName)
	cFunc := C.CString(funcName)

	// Free memory allocated by C.CString
	defer func() {
		C.free(unsafe.Pointer(cMod))
		C.free(unsafe.Pointer(cFunc))
	}()

	fn := C.load_func(cMod, cFunc)
	if fn == nil {
		fmt.Println("Something exploded")
		return nil
	}

	return fn
}

func callPyFuncWithParam() {
	fn := loadPyFunc("testfile", "paramTester")
	tuple := C.PyTuple_New(1)
	cLong := C.long(43)
	C.PyTuple_SetItem(tuple, 0, C.PyLong_FromLong(cLong))

	result := C.PyObject_CallObject(fn, tuple)

	cLong = C.PyLong_AsLong(result)
	fmt.Println(cLong)
}

func getPodNameViaPyFunc() string {
	fn := loadPyFunc("testfile", "stringParam")
	tuple := C.PyTuple_New(1)
	cString := C.CString("I said:")
	pyString := C.PyUnicode_FromString(cString)
	defer C.free(unsafe.Pointer(cString))
	// defer C.Py_DECREF(pyString)
	C.PyTuple_SetItem(tuple, 0, pyString)
	result := C.PyObject_CallObject(fn, tuple)

	tempString := C.PyUnicode_AsUTF8(result)
	fmt.Println(C.GoString(tempString))
	return "test"
}

func podMetricsToC(podMetric *Metrics) *C.struct_Metrics {
	cMetrics := &C.struct_Metrics{}

	cMetrics.PodName = C.CString(podMetrics[0].PodName)
	defer C.free(unsafe.Pointer(cMetrics.PodName))

	cMetrics.KVCacheUtil = C.float(podMetrics[0].KVCacheUtil)
	cMetrics.QueueCount = C.int(podMetrics[0].QueueCount)

	return cMetrics
}

// This is just to prove out how to convert a go struct to C.
// Unfortunately we cant just pass a pointer, as Go strings and C *char are not compatible types,
// and so must be explicitly converted using C.CString
func getPodNameViaCConversion() string {
	cMetrics := podMetricsToC(podMetrics[0])
	podName := C.capture_struct(cMetrics)
	return C.GoString(podName)
}

var (
	podMetrics []*Metrics = []*Metrics{
		{
			PodName: "pod1",
			Adapters: []string{
				"cs-agent-2025-03-14-150746",
				"cs-agent-2025-03-07-025432",
				"point-of-sale-reccomender-2024-10-31-042069",
				"code-completion-2025-03-16-130724",
			},
			KVCacheUtil: 56.6834,
			QueueCount:  3,
		},
		{
			PodName: "pod2",
			Adapters: []string{
				"skynet-2025-03-07-025432",
				"point-of-sale-reccomender-2024-10-31-042069",
				"legal-loophole-finder-2025-03-16-130724",
				"tax-code-exploiter-2025-03-16-130724",
			},
			KVCacheUtil: 56.6834,
			QueueCount:  0,
		},
		{
			PodName: "pod3",
			Adapters: []string{
				"cs-agent-2025-03-14-150746",
				"cs-agent-2025-03-07-025432",
				"point-of-sale-reccomender-2024-10-31-042069",
				"code-completion-2025-03-16-130724",
			},
			KVCacheUtil: 87.6834,
			QueueCount:  6,
		},
	}
)

type Metrics struct {
	PodName     string
	Adapters    []string
	KVCacheUtil float32
	QueueCount  int
}
