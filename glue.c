#include "glue.h"

// Load function, same as "import module_name.func_name as obj" in Python
// Returns the function object or NULL if not found
PyObject *load_func(const char *module_name, char *func_name) {
    // Import the module
    PyObject *py_mod_name = PyUnicode_FromString(module_name);
    if (py_mod_name == NULL) {
      return NULL;
    }
  
    PyObject *module = PyImport_Import(py_mod_name);
    Py_DECREF(py_mod_name);
    if (module == NULL) {
      return NULL;
    }
  
    // Get function, same as "getattr(module, func_name)" in Python
    PyObject *func = PyObject_GetAttrString(module, func_name);
    Py_DECREF(module);
    return func;
  }



void py_decref(PyObject *obj) { Py_DECREF(obj); }
