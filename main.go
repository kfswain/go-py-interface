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

// TODO: Check mem leaks (I think I need to deref all the pyobjects that are created in go)
// TODO: May be able to remove the C 'glue' code file altogether and just use cgo calls

func main() {
	C.Py_Initialize()
	defer C.Py_Finalize()

	loadPyFunc("testfile", "selectPod")

	var list []*Result
	for i := range 10 {
		start := time.Now()
		list = scorePodsViaPython(podMetrics)
		end := time.Now()
		fmt.Printf("Python took %v\nIteration: %v\n", end.Sub(start), i)
	}
	for _, score := range list {
		fmt.Printf("Pod: %v, Score: %v\n", score.PodName, score.Score)
	}
}

func scorePodsViaPython(metrics []*Metrics) []*Result {
	tuple := C.PyTuple_New(1)
	pyList := podMetricsToPyObject(metrics)
	C.PyTuple_SetItem(tuple, 0, pyList)

	resultList := callPyFuncWithParam("testfile", "selectPod", tuple)

	finalList := []*Result{}
	for i := range len(metrics) {
		tuple := C.PyList_GetItem(resultList, C.Py_ssize_t(i))
		pyPodName := C.PyTuple_GetItem(tuple, 0)
		tempString := C.PyUnicode_AsUTF8(pyPodName)
		podName := C.GoString(tempString)

		pyScore := C.PyTuple_GetItem(tuple, 1)
		score := C.PyLong_AsLong(pyScore)
		finalList = append(finalList, &Result{PodName: podName, Score: int(score)})
	}
	return finalList
}

func callPyFuncWithParam(moduleName, funcName string, tuple *C.PyObject) *C.PyObject {
	fn := loadPyFunc(moduleName, funcName)
	return C.PyObject_CallObject(fn, tuple)
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

func podMetricsToPyObject(metrics []*Metrics) *C.PyObject {
	podLen := len(metrics)

	// Init list with len of pods to prevent superfluous array creation on appends.
	pyList := C.PyList_New(C.Py_ssize_t(podLen))

	// Convert dict keys to PyString
	// Can probably just do this once at start up w/all the keys
	podKey := C.CString("pod_name")
	defer C.free(unsafe.Pointer(podKey))
	podNameKey := C.PyUnicode_FromString(podKey)

	adapterKey := C.CString("adapters")
	defer C.free(unsafe.Pointer(adapterKey))
	pyAdapterKey := C.PyUnicode_FromString(adapterKey)

	kvCacheUtilKey := C.CString("kv_cache_util")
	defer C.free(unsafe.Pointer(kvCacheUtilKey))
	pyKVCacheUtilKey := C.PyUnicode_FromString(kvCacheUtilKey)

	qCountKey := C.CString("queue_count")
	defer C.free(unsafe.Pointer(qCountKey))
	pyQCountKey := C.PyUnicode_FromString(qCountKey)

	for i, pod := range metrics {
		// Initialize py dict
		pyPodMetrics := C.PyDict_New()

		// Convert PodName value to Py
		cString := C.CString(pod.PodName)
		defer C.free(unsafe.Pointer(cString))
		pyPodName := C.PyUnicode_FromString(cString)
		C.PyDict_SetItem(pyPodMetrics, podNameKey, pyPodName)

		// Convert Adapters to PyList
		pyAdapterList := C.PyList_New(C.Py_ssize_t(len(pod.Adapters)))
		for j, adapter := range pod.Adapters {
			cAdapterString := C.CString(adapter)
			defer C.free(unsafe.Pointer(cAdapterString))
			pyAdapter := C.PyUnicode_FromString(cAdapterString)
			C.PyList_SetItem(pyAdapterList, C.Py_ssize_t(j), pyAdapter)
		}
		C.PyDict_SetItem(pyPodMetrics, pyAdapterKey, pyAdapterList)

		// Convert KVCacheUtil to Py
		pyKVCacheUtil := C.PyFloat_FromDouble(C.double(pod.KVCacheUtil))
		C.PyDict_SetItem(pyPodMetrics, pyKVCacheUtilKey, pyKVCacheUtil)

		// Convert Queue Count to Py
		pyQCount := C.PyLong_FromLong(C.long(pod.QueueCount))
		C.PyDict_SetItem(pyPodMetrics, pyQCountKey, pyQCount)

		// Set the podMetric to the appropriate index
		C.PyList_SetItem(pyList, C.Py_ssize_t(i), pyPodMetrics)
	}

	return pyList
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
	KVCacheUtil float64
	QueueCount  int64
}

type Result struct {
	PodName string
	Score   int
}
