package main

import (
	"syscall"
	"unsafe"
)

func mq_open(path []byte) uintptr {
	mq, _, _ := syscall.Syscall6(
		uintptr(MqOpenSyscall),
		uintptr(unsafe.Pointer(&path[0])),
		uintptr(0x02),
		uintptr(0x1c7),
		uintptr(0),
		uintptr(0),
		uintptr(0),
	)

	return mq
}

func mq_timedsend(fd uintptr, data []byte) uintptr {
	mqLock, _, _ := syscall.Syscall6(
		uintptr(MqTimedSendSyscall),
		uintptr(fd),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(uintptr(len(data))),
		uintptr(0),
		uintptr(0),
		uintptr(0),
	)

	return mqLock
}

func mq_close(fd uintptr) {
	fdi := int(fd)
	syscall.Close(fdi)
}
