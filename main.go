package main

// Much of this is inspired by: https://github.com/ardanlabs/python-go/
import (
	// #cgo pkg-config: python3-embed
	// #include <Python.h>
	// #include "glue.h"
	"C"
)
import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"
	"unsafe"

	"math/rand"

	"github.com/google/uuid"
)

// TODO: Check mem leaks (I think I need to deref all the pyobjects that are created in go)
// TODO: May be able to remove the C 'glue' code file altogether and just use cgo calls

func main() {
	C.Py_Initialize()
	defer C.Py_Finalize()

	// We init here to load the py file preemptively. It doesnt matter what func it is,
	// just loads the module. This saves startup time.
	loadPyFunc("testfile", "selectPod")

	// Initialize dict key values
	podKey := C.CString("pod_name")
	defer C.free(unsafe.Pointer(podKey))
	podNameKey = C.PyUnicode_FromString(podKey)

	adapterKey := C.CString("adapters")
	defer C.free(unsafe.Pointer(adapterKey))
	pyAdapterKey = C.PyUnicode_FromString(adapterKey)

	kvCacheUtilKey := C.CString("kv_cache_util")
	defer C.free(unsafe.Pointer(kvCacheUtilKey))
	pyKVCacheUtilKey = C.PyUnicode_FromString(kvCacheUtilKey)

	qCountKey := C.CString("queue_count")
	defer C.free(unsafe.Pointer(qCountKey))
	pyQCountKey = C.PyUnicode_FromString(qCountKey)

	// generate 1000 pods with 30 adapters
	// (adapter names are uuids so thats 36characters * 30 which is just slightly over 1kb)
	scaleTestMetricList := generatePodMetrics(1000, 30)

	var list []*Result
	//for i := range 10 {
	start := time.Now()
	list = scorePodsViaPython(scaleTestMetricList[0:30])

	end := time.Now()
	fmt.Printf("Python took %v\n", end.Sub(start))

	// start := time.Now()
	// list = scorePodsWJsonBytesAsParam(scaleTestMetricList[0:30])

	// end := time.Now()
	// fmt.Printf("Python byte method took %v\n", end.Sub(start))

	//}
	for _, score := range list[0:10] {
		fmt.Printf("Pod: %v, Score: %v\n", score.PodName, score.Score)
	}
}

func scorePodsViaPython(metrics []*Metrics) []*Result {
	tuple := C.PyTuple_New(1)
	pyList, cleanUp := podMetricsToPyObject(metrics)
	cleanUp = append(cleanUp, func() { C.py_decref(pyList) })
	cleanUp = append(cleanUp, func() { C.py_decref(tuple) })
	C.PyTuple_SetItem(tuple, 0, pyList)

	resultList := callPyFuncWithParam("testfile", "selectPod", tuple)
	cleanUp = append(cleanUp, func() { C.py_decref(resultList) })

	finalList := []*Result{}
	for i := range len(metrics) {
		resultTuple := C.PyList_GetItem(resultList, C.Py_ssize_t(i))
		cleanUp = append(cleanUp, func() { C.py_decref(resultTuple) })
		pyPodName := C.PyTuple_GetItem(resultTuple, 0)
		tempString := C.PyUnicode_AsUTF8(pyPodName)
		podName := C.GoString(tempString)

		pyScore := C.PyTuple_GetItem(resultTuple, 1)
		score := C.PyLong_AsLong(pyScore)
		finalList = append(finalList, &Result{PodName: podName, Score: int(score)})
	}
	for _, f := range cleanUp {
		f()
	}
	return finalList
}

func scorePodsWJsonBytesAsParam(metrics []*Metrics) []*Result {
	pyBytes := goBytesToPyBytes(encodeArrAsJson(metrics))
	tuple := C.PyTuple_New(1)
	C.PyTuple_SetItem(tuple, 0, pyBytes)

	resultList := callPyFuncWithParam("testfile", "decodeJsonBytes", tuple)
	finalList := []*Result{}
	for i := range len(metrics) {
		resultTuple := C.PyList_GetItem(resultList, C.Py_ssize_t(i))
		pyPodName := C.PyTuple_GetItem(resultTuple, 0)
		tempString := C.PyUnicode_AsUTF8(pyPodName)
		podName := C.GoString(tempString)

		pyScore := C.PyTuple_GetItem(resultTuple, 1)
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

func podMetricsToPyObject(metrics []*Metrics) (*C.PyObject, []func()) {
	podLen := len(metrics)
	cleanUp := []func(){}
	// Init list with len of pods to prevent superfluous array creation on appends.
	pyList := C.PyList_New(C.Py_ssize_t(podLen))
	cleanUp = append(cleanUp, func() { C.py_decref(pyList) })

	// Convert dict keys to PyString
	// Can probably just do this once at start up w/all the keys

	for i, pod := range metrics {
		// Initialize py dict
		pyPodMetrics := C.PyDict_New()
		cleanUp = append(cleanUp, func() { C.py_decref(pyPodMetrics) })

		// Convert PodName value to Py
		cString := C.CString(pod.PodName)
		defer C.free(unsafe.Pointer(cString))
		pyPodName := C.PyUnicode_FromString(cString)
		cleanUp = append(cleanUp, func() { C.py_decref(pyPodName) })
		C.PyDict_SetItem(pyPodMetrics, podNameKey, pyPodName)

		// Convert Adapters to PyList
		pyAdapterList := C.PyList_New(C.Py_ssize_t(len(pod.Adapters)))
		cleanUp = append(cleanUp, func() { C.py_decref(pyAdapterList) })
		for j, adapter := range pod.Adapters {
			cAdapterString := C.CString(adapter)
			defer C.free(unsafe.Pointer(cAdapterString))
			pyAdapter := C.PyUnicode_FromString(cAdapterString)
			cleanUp = append(cleanUp, func() { C.py_decref(pyAdapter) })
			C.PyList_SetItem(pyAdapterList, C.Py_ssize_t(j), pyAdapter)
		}
		C.PyDict_SetItem(pyPodMetrics, pyAdapterKey, pyAdapterList)

		// Convert KVCacheUtil to Py
		pyKVCacheUtil := C.PyFloat_FromDouble(C.double(pod.KVCacheUtil))
		cleanUp = append(cleanUp, func() { C.py_decref(pyKVCacheUtil) })
		C.PyDict_SetItem(pyPodMetrics, pyKVCacheUtilKey, pyKVCacheUtil)

		// Convert Queue Count to Py
		pyQCount := C.PyLong_FromLong(C.long(pod.QueueCount))
		cleanUp = append(cleanUp, func() { C.py_decref(pyQCount) })
		C.PyDict_SetItem(pyPodMetrics, pyQCountKey, pyQCount)

		// Set the podMetric to the appropriate index
		C.PyList_SetItem(pyList, C.Py_ssize_t(i), pyPodMetrics)
	}

	return pyList, cleanUp
}

func generatePodMetrics(podCount int, adapterCount int) []*Metrics {
	allMetrics := []*Metrics{}
	for i := range podCount {
		metric := &Metrics{}
		adapters := []string{}
		for _ = range adapterCount {
			adapters = append(adapters, uuid.New().String())
		}
		metric.Adapters = adapters
		metric.KVCacheUtil = rand.Float64()
		metric.QueueCount = rand.Int63n(10)
		metric.PodName = fmt.Sprintf("Pod%v", i)
		allMetrics = append(allMetrics, metric)
	}
	return allMetrics
}

func encodeArrAsJson(metrics []*Metrics) []byte {
	bytes, err := json.Marshal(metrics)
	if err != nil {
		fmt.Printf("something bad happened: %v", err)
	}
	return bytes
}

func goBytesToPyBytes(json []byte) *C.PyObject {
	byteArr := C.CBytes(json)
	defer C.free(unsafe.Pointer(byteArr))

	pyBytes := C.PyByteArray_FromStringAndSize((*C.char)(byteArr), C.Py_ssize_t(len(json)))
	runtime.KeepAlive(json)
	runtime.KeepAlive(byteArr)
	return pyBytes
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
	podNameKey, pyAdapterKey, pyKVCacheUtilKey, pyQCountKey *C.PyObject
)

type Metrics struct {
	PodName     string   `json:"pod_name"`
	Adapters    []string `json:"adapters"`
	KVCacheUtil float64  `json:"kv_cache_util"`
	QueueCount  int64    `json:"queue_count"`
}

type Result struct {
	PodName string
	Score   int
}
