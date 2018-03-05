package datafortest

import (
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)

	// TestDataRootPath is Root path of datafortest package
	TestDataRootPath = filepath.Dir(b)
)
