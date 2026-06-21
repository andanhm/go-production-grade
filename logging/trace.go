package logging

import "runtime"

// Get File name and Line number
func getFileAndLine(callerLevel int) (file string, line int, ok bool) {
	_, file, line, ok = runtime.Caller(callerLevel)
	return file, line, ok
}

// GetMethod returns the calling method
func getMethod(callerLevel int) string {
	// we get the callers as interprets - but we just need 1
	fPcs := make([]uintptr, 1)
	// skip 3 levels to get to the caller of whoever called Caller()
	n := runtime.Callers(callerLevel, fPcs)
	if n == 0 {
		return "n/a"
	}
	// get the info of the actual function that's in the pointer
	methodObj := runtime.FuncForPC(fPcs[0] - 1)
	if methodObj == nil {
		return "n/a"
	}
	return methodObj.Name()
}
