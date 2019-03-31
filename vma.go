// Package vma implements a Go wrapper for AMD's VulkanMemoryAllocator library.
package vma

/*
#cgo CFLAGS: -I.
#include "vma.h"
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

func NewAllocationCreateInfo(flags uint32, usage uint32,
	requiredFlags, preferredFlags vk.MemoryPropertyFlags,
	memoryTypeBits uint32, pool Pool, userData unsafe.Pointer) AllocationCreateInfo {

	return AllocationCreateInfo(C.VmaAllocationCreateInfo{
		flags:          C.VmaAllocationCreateFlags(flags),
		usage:          C.VmaMemoryUsage(usage),
		requiredFlags:  C.VkMemoryPropertyFlags(requiredFlags),
		preferredFlags: C.VkMemoryPropertyFlags(preferredFlags),
		memoryTypeBits: C.uint32_t(memoryTypeBits),
		pool:           C.VmaPool(pool),
		pUserData:      userData,
	})
}

// AllocationInfo parameters of VmaAllocation objects, that can be retrieved using function GetAllocationInfo()
type AllocationInfo C.VmaAllocationInfo

func (a *AllocationInfo) MemoryType() uint32 {
	return uint32(C.VmaAllocationInfo(*a).memoryType)
}

func (a *AllocationInfo) DeviceMemory() vk.DeviceMemory {
	return vk.DeviceMemory(unsafe.Pointer(C.VmaAllocationInfo(*a).deviceMemory))
}

func (a *AllocationInfo) Offset() vk.DeviceSize {
	return vk.DeviceSize(uint64(C.VmaAllocationInfo(*a).offset))
}

func (a *AllocationInfo) Size() vk.DeviceSize {
	return vk.DeviceSize(uint64(C.VmaAllocationInfo(*a).size))
}

func (a *AllocationInfo) MappedData() unsafe.Pointer {
	return C.VmaAllocationInfo(*a).pMappedData
}

func (a *AllocationInfo) UserData() unsafe.Pointer {
	return C.VmaAllocationInfo(*a).pUserData
}

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
		return nil, errors.New("NewAllocator: vulkanProcAddr is nil")
	}

	allocator := new(Allocator)
	var create C.VmaAllocatorCreateInfo

	create.flags = C.uint32_t(flags)
	create.physicalDevice = C.VkPhysicalDevice(unsafe.Pointer(physicalDevice))
	create.device = C.VkDevice(unsafe.Pointer(device))
	create.preferredLargeHeapBlockSize = C.VkDeviceSize(preferredLargeHeapBlockSize)
	create.pAllocationCallbacks = (*C.VkAllocationCallbacks)(unsafe.Pointer(allocationCallbacks))
	create.pDeviceMemoryCallbacks = (*C.VmaDeviceMemoryCallbacks)(deviceMemoryCallbacks)
	create.frameInUseCount = C.uint32_t(frameInUseCount)

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

// Destroy destroys allocator object.
func (a *Allocator) Destroy() {
	C.vmaDestroyAllocator(a.cAlloc)
}

// DefragmentationContext represents Opaque object that represents started defragmentation process
type DefragmentationContext C.VmaDefragmentationContext

// DefragmentationInfo2 is the parameters for defragmentation. Note that the earlier, deprecated version of
// this struct is not included in these bindings.
type DefragmentationInfo2 C.VmaDefragmentationInfo2

// NewDefragmentationInfo2 creates a new DefragmentationInfo2. allocsChanged should either be nil or a slice of equal
// length to allocations.
func NewDefragmentationInfo2(flags uint32, allocations []Allocation, allocsChanged []vk.Bool32, pools []Pool,
	maxCPUBytesToMove vk.DeviceSize, maxCPUAllocationsToMove uint32,
	maxGPUBytesToMove vk.DeviceSize, maxGPUAllocationsToMove uint32,
	commandBuffer vk.CommandBuffer) DefragmentationInfo2 {

	if len(allocsChanged) != 0 && len(allocsChanged) != len(allocations) {
		panic("NewDefragmentationInfo2: allocsChanged not empty and not same length as allocations")
	}

	var d C.VmaDefragmentationInfo2
	d.flags = C.uint32_t(flags)
	d.allocationCount = C.uint32_t(len(allocations))
	d.pAllocations = (*C.VmaAllocation)(&allocations[0])
	if len(allocsChanged) != 0 {
		d.pAllocationsChanged = (*C.VkBool32)(&allocsChanged[0])
	}
	if pools != nil {
		d.poolCount = C.uint32_t(len(pools))
		d.pPools = (*C.VmaPool)(&pools[0])
	}
	d.maxCpuBytesToMove = C.VkDeviceSize(maxCPUBytesToMove)
	d.maxCpuAllocationsToMove = C.uint32_t(maxCPUAllocationsToMove)
	d.maxGpuBytesToMove = C.VkDeviceSize(maxGPUBytesToMove)
	d.maxGpuAllocationsToMove = C.uint32_t(maxGPUAllocationsToMove)
	d.commandBuffer = C.VkCommandBuffer(unsafe.Pointer(commandBuffer))
	return DefragmentationInfo2(d)
}

// DefragmentationStats is the statistics returned by function vmaDefragment()
type DefragmentationStats C.VmaDefragmentationStats

// DeviceMemoryCallbacks is the set of callbacks that the library will call for vkAllocateMemory and vkFreeMemory
type DeviceMemoryCallbacks C.VmaDeviceMemoryCallbacks

// Pool represents custom memory pool
type Pool C.VmaPool

// PoolCreateInfo describes parameter of created VmaPool
type PoolCreateInfo C.VmaPoolCreateInfo

// NewPoolCreateInfo creates a new PoolCreateInfo.
func NewPoolCreateInfo(memoryTypeIndex, flags uint32, blockSize vk.DeviceSize,
	minBlockCount, maxBlockCount uintptr, frameInUseCount uint32) PoolCreateInfo {

	var p C.VmaPoolCreateInfo
	p.memoryTypeIndex = C.uint32_t(memoryTypeIndex)
	p.flags = C.VmaPoolCreateFlags(flags)
	p.blockSize = C.VkDeviceSize(blockSize)
	p.minBlockCount = C.size_t(minBlockCount)
	p.maxBlockCount = C.size_t(maxBlockCount)
	p.frameInUseCount = C.uint32_t(frameInUseCount)
	return PoolCreateInfo(p)
}

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

// TODO: Have this return vulkan-go/vulkan wrapper structs
func (a *Allocator) GetPhysicalDeviceProperties() []vk.PhysicalDeviceProperties

// TODO: Have this return vulkan-go/vulkan wrapper structs
func (a *Allocator) GetMemoryProperties() []vk.PhysicalDeviceMemoryProperties

func (a *Allocator) GetMemoryTypeProperties(memoryTypeIndex uint32) vk.MemoryPropertyFlags {
	var flags vk.MemoryPropertyFlags
	C.vmaGetMemoryTypeProperties(a.cAlloc, C.uint32_t(memoryTypeIndex),
		(*C.VkMemoryPropertyFlags)(&flags))
	return flags
}

func (a *Allocator) SetCurrentFrameIndex(frameIndex uint32) {
	C.vmaSetCurrentFrameIndex(a.cAlloc, C.uint32_t(frameIndex))
}

func (a *Allocator) CalculateStats() Stats {
	var stats Stats
	C.vmaCalculateStats(a.cAlloc, (*C.VmaStats)(&stats))
	return stats
}

// BuildStatsString builds and returns statistics as string in JSON format. This function can be called safely without
// needing to manually free the built string.
func (a *Allocator) BuildStatsString(detailedMap bool) string {
	var charPtr *C.char
	var detailed C.VkBool32 = C.VK_FALSE
	if detailedMap {
		detailed = C.VK_TRUE
	}
	C.vmaBuildStatsString(a.cAlloc, &charPtr, detailed)
	buildStats := C.GoString(charPtr)
	C.vmaFreeStatsString(a.cAlloc, charPtr)
	return buildStats
}

func (a *Allocator) FindMemoryTypeIndex(memoryTypeBits uint32, allocationCreateInfo *AllocationCreateInfo) (uint32, error) {
	var memTypeIndex C.uint32_t
	ret := vk.Result(C.vmaFindMemoryTypeIndex(a.cAlloc, C.uint32_t(memoryTypeBits),
		(*C.VmaAllocationCreateInfo)(allocationCreateInfo), &memTypeIndex))
	if ret != vk.Success {
		return 0, vk.Error(ret)
	}
	return uint32(memTypeIndex), nil
}

func (a *Allocator) FindMemoryTypeIndexForBufferInfo(bufferCreateInfo *vk.BufferCreateInfo, allocationCreateInfo *AllocationCreateInfo) (uint32, error) {
	var memTypeIndex C.uint32_t
	ret := vk.Result(C.vmaFindMemoryTypeIndexForBufferInfo(a.cAlloc,
		(*C.VkBufferCreateInfo)(unsafe.Pointer(bufferCreateInfo.Ref())),
		(*C.VmaAllocationCreateInfo)(allocationCreateInfo), &memTypeIndex))
	if ret != vk.Success {
		return 0, vk.Error(ret)
	}
	return uint32(memTypeIndex), nil
}

func (a *Allocator) FindMemoryTypeIndexForImageInfo(imageCreateInfo *vk.ImageCreateInfo, allocationCreateInfo *AllocationCreateInfo) (uint32, error) {
	var memTypeIndex C.uint32_t
	ret := vk.Result(C.vmaFindMemoryTypeIndexForImageInfo(a.cAlloc,
		(*C.VkImageCreateInfo)(unsafe.Pointer(imageCreateInfo.Ref())),
		(*C.VmaAllocationCreateInfo)(allocationCreateInfo), &memTypeIndex))
	if ret != vk.Success {
		return 0, vk.Error(ret)
	}
	return uint32(memTypeIndex), nil
}

func (a *Allocator) CreatePool(createInfo *PoolCreateInfo) (Pool, error) {
	var pool Pool
	ret := vk.Result(C.vmaCreatePool(a.cAlloc, (*C.VmaPoolCreateInfo)(createInfo),
		(*C.VmaPool)(&pool)))
	if ret != vk.Success {
		return pool, vk.Error(ret)
	}
	return pool, nil
}

func (a *Allocator) DestroyPool(pool Pool) {
	C.vmaDestroyPool(a.cAlloc, C.VmaPool(pool))
}

func (a *Allocator) GetPoolStats(pool Pool) PoolStats {
	var poolStats PoolStats
	C.vmaGetPoolStats(a.cAlloc, C.VmaPool(pool), (*C.VmaPoolStats)(&poolStats))
	return poolStats
}

func (a *Allocator) MakePoolAllocationsLost(pool Pool) int {
	var count C.size_t
	C.vmaMakePoolAllocationsLost(a.cAlloc, C.VmaPool(pool), &count)
	return int(count)
}

func (a *Allocator) CheckPoolCorruption(pool Pool) vk.Result {
	return vk.Result(C.vmaCheckPoolCorruption(a.cAlloc, C.VmaPool(pool)))
}

func (a *Allocator) AllocateMemory(memoryRequirements *vk.MemoryRequirements, createInfo *AllocationCreateInfo, returnInfo bool) (Allocation, AllocationInfo, error) {
	var alloc Allocation
	var allocInfo AllocationInfo
	var allocInfoPtr *AllocationInfo
	if returnInfo {
		allocInfoPtr = &allocInfo
	}
	ret := vk.Result(C.vmaAllocateMemory(a.cAlloc, (*C.VkMemoryRequirements)(unsafe.Pointer(memoryRequirements.Ref())),
		(*C.VmaAllocationCreateInfo)(createInfo), (*C.VmaAllocation)(&alloc), (*C.VmaAllocationInfo)(allocInfoPtr)))
	if ret != vk.Success {
		return alloc, allocInfo, vk.Error(ret)
	}
	return alloc, allocInfo, nil
}

func (a *Allocator) AllocateMemoryPages(memoryRequirements *vk.MemoryRequirements, createInfo *AllocationCreateInfo, allocationCount int, returnInfo bool) ([]Allocation, []AllocationInfo, error) {
	if allocationCount <= 0 {
		return nil, nil, nil
	}
	allocs := make([]Allocation, allocationCount)
	var allocInfos []AllocationInfo
	var allocInfosPtr *AllocationInfo
	if returnInfo {
		allocInfos = make([]AllocationInfo, allocationCount)
		allocInfosPtr = &allocInfos[0]
	}
	ret := vk.Result(C.vmaAllocateMemoryPages(a.cAlloc, (*C.VkMemoryRequirements)(unsafe.Pointer(memoryRequirements.Ref())),
		(*C.VmaAllocationCreateInfo)(createInfo), C.size_t(allocationCount),
		(*C.VmaAllocation)(&allocs[0]), (*C.VmaAllocationInfo)(allocInfosPtr)))
	if ret != vk.Success {
		return nil, nil, vk.Error(ret)
	}
	return allocs, allocInfos, nil
}

func (a *Allocator) AllocateMemoryForBuffer(buffer vk.Buffer, createInfo *AllocationCreateInfo, returnInfo bool) (Allocation, AllocationInfo, error) {
	var alloc Allocation
	var allocInfo AllocationInfo
	var allocInfoPtr *AllocationInfo
	if returnInfo {
		allocInfoPtr = &allocInfo
	}
	ret := vk.Result(C.vmaAllocateMemoryForBuffer(a.cAlloc, C.VkBuffer(unsafe.Pointer(buffer)), (*C.VmaAllocationCreateInfo)(createInfo),
		(*C.VmaAllocation)(&alloc), (*C.VmaAllocationInfo)(allocInfoPtr)))
	if ret != vk.Success {
		return alloc, allocInfo, vk.Error(ret)
	}
	return alloc, allocInfo, nil
}

func (a *Allocator) AllocateMemoryForImage(image vk.Image, createInfo *AllocationCreateInfo, returnInfo bool) (Allocation, AllocationInfo, error) {
	var alloc Allocation
	var allocInfo AllocationInfo
	var allocInfoPtr *AllocationInfo
	if returnInfo {
		allocInfoPtr = &allocInfo
	}
	ret := vk.Result(C.vmaAllocateMemoryForImage(a.cAlloc, C.VkImage(unsafe.Pointer(image)), (*C.VmaAllocationCreateInfo)(createInfo),
		(*C.VmaAllocation)(&alloc), (*C.VmaAllocationInfo)(allocInfoPtr)))
	if ret != vk.Success {
		return alloc, allocInfo, vk.Error(ret)
	}
	return alloc, allocInfo, nil
}

func (a *Allocator) FreeMemory(allocation Allocation) {
	C.vmaFreeMemory(a.cAlloc, C.VmaAllocation(allocation))
}

func (a *Allocator) FreeMemoryPages(allocations []Allocation) {
	if len(allocations) == 0 {
		return
	}
	C.vmaFreeMemoryPages(a.cAlloc, C.size_t(len(allocations)), (*C.VmaAllocation)(&allocations[0]))
}

func (a *Allocator) ResizeAllocation(allocation Allocation, newSize vk.DeviceSize) error {
	ret := vk.Result(C.vmaResizeAllocation(a.cAlloc, C.VmaAllocation(allocation), C.VkDeviceSize(newSize)))
	if ret != vk.Success {
		return vk.Error(ret)
	}
	return nil
}

func (a *Allocator) GetAllocationInfo(allocation Allocation) AllocationInfo {
	var allocInfo AllocationInfo
	C.vmaGetAllocationInfo(a.cAlloc, C.VmaAllocation(allocation), (*C.VmaAllocationInfo)(&allocInfo))
	return allocInfo
}

func (a *Allocator) TouchAllocation(allocation Allocation) bool {
	return C.vmaTouchAllocation(a.cAlloc, C.VmaAllocation(allocation)) == C.VK_TRUE
}

func (a *Allocator) SetAllocationUserData(allocation Allocation, userData unsafe.Pointer) {
	C.vmaSetAllocationUserData(a.cAlloc, C.VmaAllocation(allocation), userData)
}

func (a *Allocator) CreateLostAllocation() Allocation {
	var alloc Allocation
	C.vmaCreateLostAllocation(a.cAlloc, (*C.VmaAllocation)(&alloc))
	return alloc
}

func (a *Allocator) MapMemory(allocation Allocation) (uintptr, error) {
	var location unsafe.Pointer
	ret := vk.Result(C.vmaMapMemory(a.cAlloc, C.VmaAllocation(allocation), (&location)))
	if ret != vk.Success {
		return 0, vk.Error(ret)
	}
	return uintptr(location), nil
}

func (a *Allocator) UnmapMemory(allocation Allocation) {
	C.vmaUnmapMemory(a.cAlloc, C.VmaAllocation(allocation))
}

func (a *Allocator) FlushAllocation(allocation Allocation, offset vk.DeviceSize, size vk.DeviceSize) {
	C.vmaFlushAllocation(a.cAlloc, C.VmaAllocation(allocation), C.VkDeviceSize(offset), C.VkDeviceSize(size))
}

func (a *Allocator) InvalidateAllocation(allocation Allocation, offset vk.DeviceSize, size vk.DeviceSize) {
	C.vmaInvalidateAllocation(a.cAlloc, C.VmaAllocation(allocation), C.VkDeviceSize(offset), C.VkDeviceSize(size))
}

func (a *Allocator) CheckCorruption(memoryTypeBits uint32) vk.Result {
	return vk.Result(C.vmaCheckCorruption(a.cAlloc, C.uint32_t(memoryTypeBits)))
}

func (a *Allocator) DefragmentationBegin(info *DefragmentationInfo2) (DefragmentationStats, DefragmentationContext, error) {
	var stats DefragmentationStats
	var context DefragmentationContext
	ret := vk.Result(C.vmaDefragmentationBegin(a.cAlloc, (*C.VmaDefragmentationInfo2)(info),
		(*C.VmaDefragmentationStats)(&stats), (*C.VmaDefragmentationContext)(&context)))
	if ret != vk.Success {
		return stats, context, vk.Error(ret)
	}
	return stats, context, nil
}

func (a *Allocator) DefragmentationEnd(context DefragmentationContext) error {
	ret := vk.Result(C.vmaDefragmentationEnd(a.cAlloc, C.VmaDefragmentationContext(context)))
	if ret != vk.Success {
		return vk.Error(ret)
	}
	return nil
}

// func (a *Allocator) Defragment(VmaAllocation *pAllocations, size_t allocationCount, VkBool32 *pAllocationsChanged, const VmaDefragmentationInfo *pDefragmentationInfo, VmaDefragmentationStats *pDefragmentationStats)
// Deprecated. Compacts memory by moving allocations. More...

func (a *Allocator) BindBufferMemory(allocation Allocation, buffer vk.Buffer) error {
	ret := vk.Result(C.vmaBindBufferMemory(a.cAlloc, C.VmaAllocation(allocation), C.VkBuffer(unsafe.Pointer(buffer))))
	if ret != vk.Success {
		return vk.Error(ret)
	}
	return nil
}

func (a *Allocator) BindImageMemory(allocation Allocation, image vk.Image) error {
	ret := vk.Result(C.vmaBindImageMemory(a.cAlloc, C.VmaAllocation(allocation), C.VkImage(unsafe.Pointer(image))))
	if ret != vk.Success {
		return vk.Error(ret)
	}
	return nil
}

func (a *Allocator) CreateBuffer(bufferCreateInfo *vk.BufferCreateInfo, allocationCreateInfo *AllocationCreateInfo, returnInfo bool) (vk.Buffer, Allocation, AllocationInfo, error) {
	var alloc Allocation
	var buffer vk.Buffer
	var allocationInfo AllocationInfo
	var allocInfoPtr *AllocationInfo
	if returnInfo {
		allocInfoPtr = &allocationInfo
	}
	ret := vk.Result(C.vmaCreateBuffer(a.cAlloc,
		(*C.VkBufferCreateInfo)(unsafe.Pointer(bufferCreateInfo.Ref())),
		(*C.VmaAllocationCreateInfo)(allocationCreateInfo),
		(*C.VkBuffer)(unsafe.Pointer(&buffer)), (*C.VmaAllocation)(&alloc),
		(*C.VmaAllocationInfo)(allocInfoPtr)))
	if ret != vk.Success {
		return buffer, alloc, allocationInfo, vk.Error(ret)
	}
	return buffer, alloc, allocationInfo, nil
}

func (a *Allocator) DestroyBuffer(buffer vk.Buffer, allocation Allocation) {
	C.vmaDestroyBuffer(a.cAlloc, C.VkBuffer(unsafe.Pointer(buffer)),
		C.VmaAllocation(allocation))
}

func (a *Allocator) CreateImage(imageCreateInfo *vk.ImageCreateInfo, allocationCreateInfo *AllocationCreateInfo, returnInfo bool) (vk.Image, Allocation, AllocationInfo, error) {
	var alloc Allocation
	var image vk.Image
	var allocationInfo AllocationInfo
	var allocInfoPtr *AllocationInfo
	if returnInfo {
		allocInfoPtr = &allocationInfo
	}
	ret := vk.Result(C.vmaCreateImage(a.cAlloc,
		(*C.VkImageCreateInfo)(unsafe.Pointer(imageCreateInfo.Ref())),
		(*C.VmaAllocationCreateInfo)(allocationCreateInfo),
		(*C.VkImage)(unsafe.Pointer(&image)), (*C.VmaAllocation)(&alloc),
		(*C.VmaAllocationInfo)(allocInfoPtr)))
	if ret != vk.Success {
		return image, alloc, allocationInfo, vk.Error(ret)
	}
	return image, alloc, allocationInfo, nil
}

func (a *Allocator) DestroyImage(image vk.Image, allocation Allocation) {
	C.vmaDestroyImage(a.cAlloc, C.VkImage(unsafe.Pointer(image)),
		C.VmaAllocation(allocation))
}
