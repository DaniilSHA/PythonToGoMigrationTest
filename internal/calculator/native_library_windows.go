package calculator

import "syscall"

func openNativeLibrary(path string) (uintptr, error) {
	id, err := syscall.LoadLibrary(path)
	return uintptr(id), err
}

func findNativeFunction(id uintptr, name string) (uintptr, error) {
	return syscall.GetProcAddress(syscall.Handle(id), name)
}

func closeNativeLibrary(id uintptr) error {
	return syscall.FreeLibrary(syscall.Handle(id))
}
