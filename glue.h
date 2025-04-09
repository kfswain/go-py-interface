#ifndef GLUE_H
#define GLUE_H

#include <Python.h>

struct Metrics {
    char* PodName;
    // char** Adapters;
    int AdaptersCount;
    float KVCacheUtil;
    int QueueCount;
};

PyObject *load_func(const char *module_name, char *func_name);
char * capture_struct(struct Metrics *podMetrics);
void py_decref(PyObject *obj);

#endif // GLUE_H