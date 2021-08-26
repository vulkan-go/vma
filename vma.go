// Package vma implements a Go wrapper for AMD's VulkanMemoryAllocator library.
package vma

/*
#cgo CFLAGS: -I.
#include "vma.h"
#include "stdlib.h"

VmaVulkanFunctions vkFuncs;
*/
import "C"

import (
	"errors"
	"sync"
	"sync/atomic"
	"unsafe"

	vk "github.com/vulkan-go/vulkan"
)

var (
	fetchedVkFuncs int32 // Atomic flag which is set once Vulkan functions are fetched
	fetchVkFuncsMu sync.Mutex
)

type result C.VkResult

// Allocation represents single memory allocation
type Allocation C.VmaAllocation

type AllocationCreateInfo struct {
	Flags          uint32
	Usage          uint32
	RequiredFlags  uint32
	PreferredFlags uint32
	MemoryTypeBits uint32
	Pool           Pool
	UserData       unsafe.Pointer
}

func (a *AllocationCreateInfo) makeCStruct() C.VmaAllocationCreateInfo {
	return C.VmaAllocationCreateInfo{
		flags:          C.VmaAllocationCreateFlags(a.Flags),
		usage:          C.VmaMemoryUsage(a.Usage),
		requiredFlags:  C.VkMemoryPropertyFlags(a.RequiredFlags),
		preferredFlags: C.VkMemoryPropertyFlags(a.PreferredFlags),
		memoryTypeBits: C.uint32_t(a.MemoryTypeBits),
		pool:           C.VmaPool(a.Pool),
		pUserData:      a.UserData,
	}
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
	cAlloc C.VmaAllocator // The underlying VmaAllocator struct
}

// Function recording is not enabled. See related comments below in NewAllocator().
// type RecordSettings struct {
// 	Flags    uint32
// 	FilePath string
// }

// func (r *RecordSettings) makeCStruct() C.VmaRecordSettings {
// 	return C.VmaRecordSettings{
// 		flags:     C.VmaRecordFlags(r.Flags),
// 		pFilePath: C.CString(r.FilePath), // REMEMBER TO C.free() IN NewAllocator!
// 	}
// }

type AllocatorCreateInfo struct {
	VulkanProcAddr              unsafe.Pointer
	PhysicalDevice              vk.PhysicalDevice
	Device                      vk.Device
	Instance                    vk.Instance
	Flags                       uint32
	PreferredLargeHeapBlockSize vk.DeviceSize
	// Allocation callbacks are currently not supported. If anybody actually cares, feel free to implement
	// AllocationCallbacks         *vk.AllocationCallbacks
	// DeviceMemoryCallbacks       *DeviceMemoryCallbacks
	FrameInUseCount uint32
	HeapSizeLimit   []vk.DeviceSize
	// RecordSettings   *RecordSettings // Function recording is not used.
	VulkanAPIVersion uint32 // Vulkan version to use, use vk.MakeVersion. Leave as 0 for default.
}

func (a *AllocatorCreateInfo) makeCStruct() C.VmaAllocatorCreateInfo {
	var heapLimit *C.VkDeviceSize
	if len(a.HeapSizeLimit) > 0 {
		heapLimit = (*C.VkDeviceSize)(&a.HeapSizeLimit[0])
	}
	return C.VmaAllocatorCreateInfo{
		physicalDevice:              C.VkPhysicalDevice(unsafe.Pointer(a.PhysicalDevice)),
		device:                      C.VkDevice(unsafe.Pointer(a.Device)),
		instance:                    C.VkInstance(unsafe.Pointer(a.Instance)),
		flags:                       C.uint32_t(a.Flags),
		preferredLargeHeapBlockSize: C.VkDeviceSize(a.PreferredLargeHeapBlockSize),
		// pAllocationCallbacks:         C.VkAllocationCallbacks(a.AllocationCallbacks),
		// pDeviceMemoryCallbacks:       C.VmaDeviceMemoryCallbacks(a.DeviceMemoryCallbacks),
		frameInUseCount: C.uint32_t(a.FrameInUseCount),
		// pRecordSettings:              C.VmaRecordSettings(a.RecordSettings), // This one is done in NewAllocator
		pHeapSizeLimit:   heapLimit,
		vulkanApiVersion: C.uint32_t(a.VulkanAPIVersion),
	}
}

// NewAllocator creates a new memory allocator. TODO: is it safe to copy?
func NewAllocator(c *AllocatorCreateInfo) (*Allocator, error) {

	if c.VulkanProcAddr == nil {
		return nil, errors.New("NewAllocator: VulkanProcAddr is nil")
	}

	allocator := &Allocator{}
	create := c.makeCStruct()

	// Function recording is not enabled. It requires VMA_RECORDING_ENABLED to
	// be defined on compile time. You can enable it in your vendored copy of
	// this library by uncommenting the following lines and related structs above
	// as well as defining the aforementioned macro in vma.cpp. Note that it is
	// only supported on Windows.
	/*
		var record *C.VmaRecordSettings = (*C.VmaRecordSettings)(C.malloc(C.size_t(unsafe.Sizeof(C.VmaRecordSettings{}))))
		if c.RecordSettings != nil {
			*record = c.RecordSettings.makeCStruct()
			defer C.free(unsafe.Pointer(record.pFilePath))
			create.pRecordSettings = record
		 defer C.free(unsafe.Pointer(record))
		}
	*/

	if len(c.HeapSizeLimit) == 0 {
		create.pHeapSizeLimit = nil
	} else {
		create.pHeapSizeLimit = (*C.VkDeviceSize)(&c.HeapSizeLimit[0])
	}

	initVkFuncs(c.VulkanProcAddr, c.Instance)
	create.pVulkanFunctions = &C.vkFuncs

	ret := vk.Result(C.vmaCreateAllocator(&create, &allocator.cAlloc))
	if ret != vk.Success {
		return nil, vk.Error(ret)
	}
	// allocator.cAlloc = a
	return allocator, nil
}

func initVkFuncs(procAddr unsafe.Pointer, instance vk.Instance) {
	// Although calling this function at the same time on multiple threads is
	// extremely unlikely, better to be safe than sorry.
	if atomic.LoadInt32(&fetchedVkFuncs) == 1 {
		return
	}
	fetchVkFuncsMu.Lock()
	defer fetchVkFuncsMu.Unlock()
	if atomic.LoadInt32(&fetchedVkFuncs) == 1 {
		return
	}
	C.initVulkanFunctions(procAddr,
		C.VkInstance(unsafe.Pointer(instance)), &C.vkFuncs)
	atomic.StoreInt32(&fetchedVkFuncs, 1)
}

// Destroy destroys allocator object.
func (a *Allocator) Destroy() {
	C.vmaDestroyAllocator(a.cAlloc)
}

// DefragmentationContext represents Opaque object that represents started defragmentation process
type DefragmentationContext C.VmaDefragmentationContext

// DefragmentationInfo2 is the parameters for defragmentation. Note that the earlier, deprecated version of
// this struct is not included in these bindings.
type DefragmentationInfo2 struct {
	Flags                   uint32
	Allocations             []Allocation
	AllocationsChanged      []vk.Bool32
	Pools                   []Pool
	MaxCPUBytesToMove       vk.DeviceSize
	MaxCPUAllocationsToMove uint32
	MaxGPUBytesToMove       vk.DeviceSize
	MaxGPUAllocationsToMove uint32
	CommandBuffer           vk.CommandBuffer
}

func (d *DefragmentationInfo2) makeCStruct() C.VmaDefragmentationInfo2 {
	defragInfo := C.VmaDefragmentationInfo2{
		flags:                   C.VmaDefragmentationFlags(d.Flags),
		allocationCount:         C.uint32_t(len(d.Allocations)),
		poolCount:               C.uint32_t(len(d.Pools)),
		maxCpuBytesToMove:       C.VkDeviceSize(d.MaxCPUBytesToMove),
		maxCpuAllocationsToMove: C.uint32_t(d.MaxCPUAllocationsToMove),
		maxGpuBytesToMove:       C.VkDeviceSize(d.MaxGPUBytesToMove),
		maxGpuAllocationsToMove: C.uint32_t(d.MaxGPUAllocationsToMove),
		commandBuffer:           C.VkCommandBuffer(unsafe.Pointer(d.CommandBuffer)),
	}
	if len(d.Allocations) > 0 {
		defragInfo.pAllocations = (*C.VmaAllocation)(&d.Allocations[0])
		if len(d.AllocationsChanged) == len(d.Allocations) {
			defragInfo.pAllocationsChanged = (*C.VkBool32)(&d.AllocationsChanged[0])
		}
	}
	if len(d.Pools) > 0 {
		defragInfo.pPools = (*C.VmaPool)(&d.Pools[0])
	}
	return defragInfo
}

// DefragmentationStats is the statistics returned by function vmaDefragment()
type DefragmentationStats struct {
	BytesMoved              vk.DeviceSize
	BytesFreed              vk.DeviceSize
	AllocationsMoved        uint32
	DeviceMemoryBlocksFreed uint32
}

// DeviceMemoryCallbacks is the set of callbacks that the library will call for vkAllocateMemory and vkFreeMemory
// type DeviceMemoryCallbacks C.VmaDeviceMemoryCallbacks

// Pool represents custom memory pool
type Pool C.VmaPool

// PoolCreateInfo describes parameter of created VmaPool
type PoolCreateInfo struct {
	MemoryTypeIndex uint32
	Flags           uint32
	BlockSize       vk.DeviceSize
	MinBlockCount   uint
	MaxBlockCount   uint
	FrameInUseCount uint32
}

func (p *PoolCreateInfo) makeCStruct() C.VmaPoolCreateInfo {
	return C.VmaPoolCreateInfo{
		memoryTypeIndex: C.uint32_t(p.MemoryTypeIndex),
		flags:           C.VmaPoolCreateFlags(p.Flags),
		blockSize:       C.VkDeviceSize(p.BlockSize),
		minBlockCount:   C.size_t(p.MinBlockCount),
		maxBlockCount:   C.size_t(p.MaxBlockCount),
		frameInUseCount: C.uint32_t(p.FrameInUseCount),
	}
}

// PoolStats describes parameter of existing VmaPool
type PoolStats C.VmaPoolStats

// RecordSettings contains the parameters for recording calls to VMA functions. To be used in VmaAllocatorCreateInfo::pRecordSettings
// type RecordSettings C.VmaRecordSettings

// StatInfo is the calculated statistics of memory usage in entire allocator
type StatInfo C.VmaStatInfo

// Stats is the general statistics from current state of Allocator
type Stats C.VmaStats

// Pointers to some Vulkan functions - a subset used by the library
type vulkanFunctions C.VmaVulkanFunctions

func (a *Allocator) GetPhysicalDeviceProperties() *vk.PhysicalDeviceProperties {
	var propPtr *C.VkPhysicalDeviceProperties
	C.vmaGetPhysicalDeviceProperties(a.cAlloc, &propPtr)
	return vk.NewPhysicalDevicePropertiesRef(unsafe.Pointer(propPtr))
}

func (a *Allocator) GetMemoryProperties() *vk.PhysicalDeviceMemoryProperties {
	var propPtr *C.VkPhysicalDeviceMemoryProperties
	C.vmaGetMemoryProperties(a.cAlloc, &propPtr)
	return vk.NewPhysicalDeviceMemoryPropertiesRef(unsafe.Pointer(propPtr))
}

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
	allocCreateInfo := allocationCreateInfo.makeCStruct()
	ret := vk.Result(C.vmaFindMemoryTypeIndex(a.cAlloc, C.uint32_t(memoryTypeBits),
		&allocCreateInfo, &memTypeIndex))
	if ret != vk.Success {
		return 0, vk.Error(ret)
	}
	return uint32(memTypeIndex), nil
}

func (a *Allocator) FindMemoryTypeIndexForBufferInfo(bufferCreateInfo *vk.BufferCreateInfo, allocationCreateInfo *AllocationCreateInfo) (uint32, error) {
	var memTypeIndex C.uint32_t
	allocCreateInfo := allocationCreateInfo.makeCStruct()
	cpBufferCreate, _ := bufferCreateInfo.PassRef()
	ret := vk.Result(C.vmaFindMemoryTypeIndexForBufferInfo(a.cAlloc,
		(*C.VkBufferCreateInfo)(unsafe.Pointer(cpBufferCreate)),
		&allocCreateInfo, &memTypeIndex))
	if ret != vk.Success {
		return 0, vk.Error(ret)
	}
	return uint32(memTypeIndex), nil
}

func (a *Allocator) FindMemoryTypeIndexForImageInfo(imageCreateInfo *vk.ImageCreateInfo, allocationCreateInfo *AllocationCreateInfo) (uint32, error) {
	var memTypeIndex C.uint32_t
	allocCreateInfo := allocationCreateInfo.makeCStruct()
	cpImageCreate, _ := imageCreateInfo.PassRef()
	ret := vk.Result(C.vmaFindMemoryTypeIndexForImageInfo(a.cAlloc,
		(*C.VkImageCreateInfo)(unsafe.Pointer(cpImageCreate)),
		&allocCreateInfo, &memTypeIndex))
	if ret != vk.Success {
		return 0, vk.Error(ret)
	}
	return uint32(memTypeIndex), nil
}

func (a *Allocator) CreatePool(createInfo *PoolCreateInfo) (Pool, error) {
	var pool Pool
	poolCreateInfo := createInfo.makeCStruct()
	ret := vk.Result(C.vmaCreatePool(a.cAlloc, &poolCreateInfo,
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
	allocCreateInfo := createInfo.makeCStruct()
	if returnInfo {
		allocInfoPtr = &allocInfo
	}
	ret := vk.Result(C.vmaAllocateMemory(a.cAlloc, (*C.VkMemoryRequirements)(unsafe.Pointer(memoryRequirements.Ref())),
		&allocCreateInfo, (*C.VmaAllocation)(&alloc), (*C.VmaAllocationInfo)(allocInfoPtr)))
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
	allocCreateInfo := createInfo.makeCStruct()
	if returnInfo {
		allocInfos = make([]AllocationInfo, allocationCount)
		allocInfosPtr = &allocInfos[0]
	}
	cpMemoryRequirements, _ := memoryRequirements.PassRef()
	ret := vk.Result(C.vmaAllocateMemoryPages(a.cAlloc, (*C.VkMemoryRequirements)(unsafe.Pointer(cpMemoryRequirements)),
		&allocCreateInfo, C.size_t(allocationCount),
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
	allocCreateInfo := createInfo.makeCStruct()
	if returnInfo {
		allocInfoPtr = &allocInfo
	}
	ret := vk.Result(C.vmaAllocateMemoryForBuffer(a.cAlloc, C.VkBuffer(unsafe.Pointer(buffer)), &allocCreateInfo,
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
	allocCreateInfo := createInfo.makeCStruct()
	if returnInfo {
		allocInfoPtr = &allocInfo
	}
	ret := vk.Result(C.vmaAllocateMemoryForImage(a.cAlloc, C.VkImage(unsafe.Pointer(image)), &allocCreateInfo,
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

// DEPRECATED.
// func (a *Allocator) ResizeAllocation(allocation Allocation, newSize vk.DeviceSize) error {
// 	ret := vk.Result(C.vmaResizeAllocation(a.cAlloc, C.VmaAllocation(allocation), C.VkDeviceSize(newSize)))
// 	if ret != vk.Success {
// 		return vk.Error(ret)
// 	}
// 	return nil
// }

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
	var cStats C.VmaDefragmentationStats
	var context DefragmentationContext
	defragInfo := info.makeCStruct()
	ret := vk.Result(C.vmaDefragmentationBegin(a.cAlloc, &defragInfo,
		(*C.VmaDefragmentationStats)(&cStats), (*C.VmaDefragmentationContext)(&context)))
	stats := DefragmentationStats{
		BytesMoved:              vk.DeviceSize(cStats.bytesMoved),
		BytesFreed:              vk.DeviceSize(cStats.bytesFreed),
		AllocationsMoved:        uint32(cStats.allocationsMoved),
		DeviceMemoryBlocksFreed: uint32(cStats.deviceMemoryBlocksFreed),
	}
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

func (a *Allocator) BindBufferMemory(allocation Allocation, buffer vk.Buffer) error {
	ret := vk.Result(C.vmaBindBufferMemory(a.cAlloc, C.VmaAllocation(allocation), C.VkBuffer(unsafe.Pointer(buffer))))
	if ret != vk.Success {
		return vk.Error(ret)
	}
	return nil
}

func (a *Allocator) BindBufferMemory2(allocation Allocation, offset vk.DeviceSize, buffer vk.Buffer, pNext unsafe.Pointer) error {
	ret := vk.Result(C.vmaBindBufferMemory2(a.cAlloc, C.VmaAllocation(allocation), C.VkDeviceSize(offset), C.VkBuffer(unsafe.Pointer(buffer)), pNext))
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

func (a *Allocator) BindImageMemory2(allocation Allocation, offset vk.DeviceSize, image vk.Image, pNext unsafe.Pointer) error {
	ret := vk.Result(C.vmaBindImageMemory2(a.cAlloc, C.VmaAllocation(allocation), C.VkDeviceSize(offset), C.VkImage(unsafe.Pointer(image)), pNext))
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
	allocCreateInfo := allocationCreateInfo.makeCStruct()
	if returnInfo {
		allocInfoPtr = &allocationInfo
	}
	cpBufferCreate, _ := bufferCreateInfo.PassRef()
	ret := vk.Result(C.vmaCreateBuffer(a.cAlloc,
		(*C.VkBufferCreateInfo)(unsafe.Pointer(cpBufferCreate)),
		&allocCreateInfo,
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
	allocCreateInfo := allocationCreateInfo.makeCStruct()
	if returnInfo {
		allocInfoPtr = &allocationInfo
	}
	cpImageCreate, _ := imageCreateInfo.PassRef()
	ret := vk.Result(C.vmaCreateImage(a.cAlloc,
		(*C.VkImageCreateInfo)(unsafe.Pointer(cpImageCreate)),
		&allocCreateInfo,
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

type Budget struct {
	BlockBytes      uint64
	AllocationBytes uint64
	Usage           uint64
	Budget          uint64
}

func (a *Allocator) GetBudget() Budget {
	var budget C.VmaBudget
	C.vmaGetBudget(a.cAlloc, &budget)
	return Budget{
		BlockBytes:      uint64(budget.blockBytes),
		AllocationBytes: uint64(budget.allocationBytes),
		Usage:           uint64(budget.usage),
		Budget:          uint64(budget.budget),
	}
}

func (a *Allocator) SetPoolName(pool Pool, name string) {
	cstr := C.CString(name)
	C.vmaSetPoolName(a.cAlloc, C.VmaPool(pool), cstr)
	C.free(unsafe.Pointer(cstr))
}

func (a *Allocator) GetPoolName(pool Pool) string {
	var charPtr *C.char
	C.vmaGetPoolName(a.cAlloc, C.VmaPool(pool), &charPtr)
	return C.GoString(charPtr)
}
