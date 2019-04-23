package vma

/*
#cgo CFLAGS: -I.

#define VK_NO_PROTOTYPES
#include "vulkan/vulkan.h"

#define VMA_STATIC_VULKAN_FUNCTIONS 0
#include "vk_mem_alloc.h"
*/
import "C"

import (
	"reflect"
	"errors"
	"io"
	"unsafe"
)

// This file contains functions and types that provide a more Go-like access to VMA features.

var (
	// ErrorAlreadyUnmapped is returned when trying to write to an unmapped buffer.
	ErrorAlreadyUnmapped = errors.New("memory already unmapped")

	// ErrorBufferIsFull is returned when trying to write to a full buffer.
	ErrorBufferIsFull = errors.New("attempting to write into a full buffer")
)

// ReadWriter implements a Go-like interface for accessing mapped memory regions. It should not be
// copied after creating.
//
// Close must be called to unmap the memory after reading or writing to the buffer.
type ReadWriter struct {
	buffer     uintptr // The memory location of the underlying buffer
	size       uint64  // The total capacity of the underlying buffer in bytes
	offset     uint64  // Offset from the beginning of the buffer in bytes
	allocator  C.VmaAllocator
	allocation C.VmaAllocation
}

var _ io.ReadWriter = (*ReadWriter)(nil)
var _ io.Closer = (*ReadWriter)(nil)

// NewReadWriter returns a new Writer for the allocation. User must keep track of the allocation's size.
func (a *Allocator) NewReadWriter(allocation Allocation, size uint64) (ReadWriter, error) {
	buf, err := a.MapMemory(allocation)
	if err != nil {
		return ReadWriter{}, err
	}

	return ReadWriter{
		buffer: buf,
		size: size,
		offset: 0,
		allocator: C.VmaAllocator(a.cAlloc),
		allocation: C.VmaAllocation(allocation),
	}, nil
}

// NewReadWriterSimple is identical to NewWriter, but asks the allocator for the size. Using this
// function is less efficient than using NewWriter.
func (a *Allocator) NewReadWriterSimple(allocation Allocation) (ReadWriter, error) {
	info := a.GetAllocationInfo(allocation)
	return a.NewReadWriter(allocation, uint64(info.Size()))
}

// Write can be used to write regular byte slices. It is safe against buffer overflows (assuming you
// give the correct size when creating it) and will return an ErrorBufferIsFull error when
// attempting to write a larger slice than the buffer can hold. It will still fill the buffer with
// what it can hold.
func (rw *ReadWriter) Write(p []byte) (n int, err error) {
	if rw.buffer == 0 {
		return 0, ErrorAlreadyUnmapped
	}

	if len(p) > int(rw.size-rw.offset) {
		n = int(rw.size - rw.offset)
		err = ErrorBufferIsFull
	} else {
		n = len(p)
	}

	// Cast the underlying C buffer into a slice of bytes
	dst := (*[memCap]byte)(unsafe.Pointer(rw.buffer+uintptr(rw.offset)))[:n:n]

	copy(dst, p)

	return
}

// WriteAny writes an arbitrary type into the buffer. The function returns the number of bytes
// written. v must be a slice or a pointer. This function uses runtime reflection, so it is slower
// than the Write method. WriteAny should not be used for elements containing pointers, maps or slices.
func (rw *ReadWriter) WriteAny(v interface{}) (n int, err error) {
	var elemSize int
	var elemNum int

	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	elemSize = int(t.Elem().Size())
	if t.Kind() == reflect.Slice {
		elemNum = val.Len()
	} else {
		elemNum = 1
	}

	if elemNum*elemSize > int(rw.size - rw.offset) {
		n = (int(rw.size - rw.offset) / elemSize) * elemSize
		err = ErrorBufferIsFull
	} else {
		n = elemNum * elemSize
	}

	// Cast the underlying C buffer into a slice of bytes
	src := (*[memCap]byte)(unsafe.Pointer(val.Pointer()))[:n:n]
	dst := (*[memCap]byte)(unsafe.Pointer(rw.buffer+uintptr(rw.offset)))[:n:n]

	copy(dst, src)

	return
}

// WriteFromPointer lets you write directly from an unsafe.Pointer. It is the most efficient but
// also the least safe option.
func (rw *ReadWriter) WriteFromPointer(srcPtr unsafe.Pointer, size uint64) (n int, err error) {
	if size > rw.size - rw.offset {
		err = ErrorBufferIsFull
		n = int(rw.size - rw.offset)
	} else {
		n = int(size)
	}

	src := (*[memCap]byte)(srcPtr)[:n:n]
	dst := (*[memCap]byte)(unsafe.Pointer(rw.buffer+uintptr(rw.offset)))[:n:n]

	copy(dst, src)

	return
}

// Close unmaps the memory and should be called for all non-persistent buffers after filling the buffer.
// Attempting to write to or close an already closed Writer will yield ErrorAlreadyUnmapped.
func (rw *ReadWriter) Close() error {
	if rw.buffer == 0 {
		return ErrorAlreadyUnmapped
	}

	C.vmaUnmapMemory(rw.allocator, rw.allocation)
	rw.buffer = 0
	return nil
}

// Reset resets the number of elements written, so that the next write or read will start
// from the beginning.
func (rw *ReadWriter) Reset() {
	rw.offset = 0
}

func (rw *ReadWriter) Read(p []byte) (n int, err error) {
	if rw.buffer == 0 {
		return 0, ErrorAlreadyUnmapped
	}

	if uint64(len(p)) >= rw.size-rw.offset {
		n = int(rw.size - rw.offset)
		err = io.EOF
	} else {
		n = len(p)
	}

	// Cast the underlying C buffer into a slice of bytes
	src := (*[memCap]byte)(unsafe.Pointer(rw.buffer+uintptr(rw.offset)))[:n:n]

	copy(p, src)

	return
}

// ReadAny reads to an arbitrary type. The function returns the number of bytes read. v must be a slice
// or a pointer. This function uses runtime reflection, so it is slower than the Read method.
// ReadAny should not be used for elements containing pointers, maps or slices.
func (rw *ReadWriter) ReadAny(v interface{}) (n int, err error) {
	var elemSize int
	var elemNum int

	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	elemSize = int(t.Elem().Size())
	if t.Kind() == reflect.Slice {
		elemNum = val.Len()
	} else {
		elemNum = 1
	}

	if elemNum*elemSize >= int(rw.size - rw.offset) {
		n = (int(rw.size - rw.offset) / elemSize) * elemSize
		err = io.EOF
	} else {
		n = elemNum * elemSize
	}

	// Cast the underlying C buffer into a slice of bytes
	src := (*[memCap]byte)(unsafe.Pointer(rw.buffer+uintptr(rw.offset)))[:n:n]
	dst := (*[memCap]byte)(unsafe.Pointer(val.Pointer()))[:n:n]

	copy(dst, src)

	return
}

// ReadToPointer lets you read directly into an unsafe.Pointer. It is the most efficient but
// also the least safe option.
func (rw *ReadWriter) ReadToPointer(dstPtr unsafe.Pointer, size uint64) (n int, err error) {
	if size >= rw.size - rw.offset {
		n = int(rw.size - rw.offset)
		err = io.EOF
	} else {
		n = int(size)
	}

	src := (*[memCap]byte)(unsafe.Pointer(rw.buffer+uintptr(rw.offset)))[:n:n]
	dst := (*[memCap]byte)(dstPtr)[:n:n]

	copy(dst, src)

	return
}