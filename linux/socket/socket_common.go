// +build !386

package socket

import (
	"fmt"
	"syscall"
	"unsafe"
)

func bind(s int, addr unsafe.Pointer, addrlen _Socklen) (err error) {
	fmt.Printf("syscall.Syscall(syscall.SYS_BIND, uintptr(s)=%v, uintptr(addr)=%v, uintptr(addrlen)=%v)\n", uintptr(s), uintptr(addr), uintptr(addrlen))
	_, _, e1 := syscall.Syscall(syscall.SYS_BIND, uintptr(s), uintptr(addr), uintptr(addrlen))
	if e1 != 0 {
		err = e1
	}
	fmt.Printf("err: %v\n", err)
	return
}

func setsockopt(s int, level int, name int, val unsafe.Pointer, vallen uintptr) (err error) {
	fmt.Printf("syscall.Syscall6(syscall.SYS_SETSOCKOPT, uintptr(s)=%v, uintptr(level)=%v, uintptr(name)=%v, uintptr(val)=%v, uintptr(vallen)=%v, 0)\n",
		uintptr(s), uintptr(level), uintptr(name), uintptr(val), uintptr(vallen))
	_, _, e1 := syscall.Syscall6(syscall.SYS_SETSOCKOPT, uintptr(s), uintptr(level), uintptr(name), uintptr(val), uintptr(vallen), 0)
	if e1 != 0 {
		err = e1
	}
	fmt.Printf("err: %v\n", err)
	return
}
