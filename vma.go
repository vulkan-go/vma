// Package vma implements a Go wrapper for AMD's VulkanMemoryAllocator library.
package vma

/*
#cgo CFLAGS: -I.
#include "vma.h"

int fn() {
	return 1;
}
*/
import "C"

import (
	"errors"
	"unsafe"

	vk "github.com/vulkan-go/vulkan"
)

type result C.VkResult

// Allocation represents single memory allocation
type Allocation C.VmaAllocation

type AllocationCreateInfo C.VmaAllocationCreateInfo

// AllocationInfo parameters of VmaAllocation objects, that can be retrieved using function vmaGetAllocationInfo()
type AllocationInfo C.VmaAllocationInfo

// Allocator represents main object of this library initialized
type Allocator struct {
	cAlloc      C.VmaAllocator // The underlying VmaAllocator struct
	vkFunctions C.VmaVulkanFunctions
}

// NewAllocator creates a new memory allocator. TODO: is it safe to copy?
func NewAllocator(
	vulkanProcAddr unsafe.Pointer,
	physicalDevice vk.PhysicalDevice,
	device vk.Device,
	instance vk.Instance,
	flags uint32,
	preferredLargeHeapBlockSize vk.DeviceSize,
	allocationCallbacks *vk.AllocationCallbacks,
	deviceMemoryCallbacks *DeviceMemoryCallbacks,
	frameInUseCount uint32,
	heapSizeLimit []vk.DeviceSize,
	recordSettings *RecordSettings,
) (*Allocator, error) {

	if vulkanProcAddr == nil {
		return nil, errors.New("NewAllocator: vulkanProcAddr nil")
	}

	allocator := new(Allocator)
	var create C.VmaAllocatorCreateInfo

	create.flags = C.uint(flags)
	create.physicalDevice = C.VkPhysicalDevice(unsafe.Pointer(physicalDevice))
	create.device = C.VkDevice(unsafe.Pointer(device))
	create.preferredLargeHeapBlockSize = C.VkDeviceSize(preferredLargeHeapBlockSize)
	create.pAllocationCallbacks = (*C.VkAllocationCallbacks)(unsafe.Pointer(allocationCallbacks))
	create.pDeviceMemoryCallbacks = (*C.VmaDeviceMemoryCallbacks)(deviceMemoryCallbacks)
	create.frameInUseCount = C.uint(frameInUseCount)

	if len(heapSizeLimit) == 0 {
		create.pHeapSizeLimit = nil
	} else {
		create.pHeapSizeLimit = (*C.VkDeviceSize)(&heapSizeLimit[0])
	}

	C.initVulkanFunctions(vulkanProcAddr,
		C.VkInstance(unsafe.Pointer(instance)), &allocator.vkFunctions)

	ret := vk.Result(C.vmaCreateAllocator(&create, &allocator.cAlloc))
	if ret != vk.Success {
		return nil, vk.Error(ret)
	}
	return allocator, nil
}

// DefragmentationContext represents Opaque object that represents started defragmentation process
type DefragmentationContext C.VmaDefragmentationContext

// DefragmentationInfo is the optional configuration parameters to be passed to function vmaDefragment()
// This struct is deprecated, you should use DefragmentationInfo2 instead
type DefragmentationInfo C.VmaDefragmentationInfo

// DefragmentationInfo2 is the parameters for defragmentation
type DefragmentationInfo2 C.VmaDefragmentationInfo2

// DefragmentationStats is the statistics returned by function vmaDefragment()
type DefragmentationStats C.VmaDefragmentationStats

// DeviceMemoryCallbacks is the set of callbacks that the library will call for vkAllocateMemory and vkFreeMemory
type DeviceMemoryCallbacks C.VmaDeviceMemoryCallbacks

// Pool represents custom memory pool
type Pool C.VmaPool

// PoolCreateInfo describes parameter of created VmaPool
type PoolCreateInfo C.VmaPoolCreateInfo

// PoolStats describes parameter of existing VmaPool
type PoolStats C.VmaPoolStats

// RecordSettings contains the parameters for recording calls to VMA functions. To be used in VmaAllocatorCreateInfo::pRecordSettings
type RecordSettings C.VmaRecordSettings

// StatInfo is the calculated statistics of memory usage in entire allocator
type StatInfo C.VmaStatInfo

// Stats is the general statistics from current state of Allocator
type Stats C.VmaStats

// Pointers to some Vulkan functions - a subset used by the library
type vulkanFunctions C.VmaVulkanFunctions
