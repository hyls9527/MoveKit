package migrate

import sys "github.com/hyls9527/MoveKit/internal/syscall"

// Sys is the package-level syscall abstraction layer used by the migrate
// package. In production, it is initialized to the real Win32 API implementation.
// In tests, it can be replaced with mock implementations for any or all of its
// 7 interface fields, enabling pure-Go testing without touching the real
// Windows registry, filesystem, or process APIs.
var Sys = sys.NewRealLayer()
