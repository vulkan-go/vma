package vma

/*
#cgo CFLAGS: -I.
#include "vk_mem_alloc.h"
*/
import "C"

// type MemoryUsage C.VmaMemoryUsage

const (
	// VmaAllocatorCreateFlagBits

	AllocatorCreateExternallySynchronized  = C.VMA_ALLOCATOR_CREATE_EXTERNALLY_SYNCHRONIZED_BIT
	AllocatorCreateDedicatedAllocation     = C.VMA_ALLOCATOR_CREATE_KHR_DEDICATED_ALLOCATION_BIT
	AllocatorCreateBindMemory2             = C.VMA_ALLOCATOR_CREATE_KHR_BIND_MEMORY2_BIT
	AllocatorCreateMemoryBudget            = C.VMA_ALLOCATOR_CREATE_EXT_MEMORY_BUDGET_BIT
	AllocatorCreateAMDDeviceCoherentMemory = C.VMA_ALLOCATOR_CREATE_AMD_DEVICE_COHERENT_MEMORY_BIT
	AllocatorCreateBufferDeviceAddress     = C.VMA_ALLOCATOR_CREATE_BUFFER_DEVICE_ADDRESS_BIT
	AllocatorCreateFlagBitsMaxEnum         = C.VMA_ALLOCATOR_CREATE_FLAG_BITS_MAX_ENUM

	// VmaRecordFlagBits

	RecordFlushAfterCall  = C.VMA_RECORD_FLUSH_AFTER_CALL_BIT
	RecordFlagBitsMaxEnum = C.VMA_RECORD_FLAG_BITS_MAX_ENUM

	// VmaMemoryUsage

	MemoryUsageUnknown            = C.VMA_MEMORY_USAGE_UNKNOWN
	MemoryUsageGPUOnly            = C.VMA_MEMORY_USAGE_GPU_ONLY
	MemoryUsageCPUOnly            = C.VMA_MEMORY_USAGE_CPU_ONLY
	MemoryUsageCPUToGPU           = C.VMA_MEMORY_USAGE_CPU_TO_GPU
	MemoryUsageGPUToCPU           = C.VMA_MEMORY_USAGE_GPU_TO_CPU
	MemoryUsageCPUCopy            = C.VMA_MEMORY_USAGE_CPU_COPY
	MemoryUsageGPULazilyAllocated = C.VMA_MEMORY_USAGE_GPU_LAZILY_ALLOCATED
	MemoryUsageUsageMaxEnum       = C.VMA_MEMORY_USAGE_MAX_ENUM

	// VmaAllocationCreateFlagBits

	AllocationCreateDedicatedMemory          = C.VMA_ALLOCATION_CREATE_DEDICATED_MEMORY_BIT
	AllocationCreateNeverAllocate            = C.VMA_ALLOCATION_CREATE_NEVER_ALLOCATE_BIT
	AllocationCreateMapped                   = C.VMA_ALLOCATION_CREATE_MAPPED_BIT
	AllocationCreateCanBecomeLost            = C.VMA_ALLOCATION_CREATE_CAN_BECOME_LOST_BIT
	AllocationCreateCanMakeOtherLost         = C.VMA_ALLOCATION_CREATE_CAN_MAKE_OTHER_LOST_BIT
	AllocationCreateUserDataCopyString       = C.VMA_ALLOCATION_CREATE_USER_DATA_COPY_STRING_BIT
	AllocationCreateUpperAddress             = C.VMA_ALLOCATION_CREATE_UPPER_ADDRESS_BIT
	AllocationCreateDontBind                 = C.VMA_ALLOCATION_CREATE_DONT_BIND_BIT
	AllocationCreateStrategyBestFit          = C.VMA_ALLOCATION_CREATE_STRATEGY_BEST_FIT_BIT
	AllocationCreateStrategyWorstFit         = C.VMA_ALLOCATION_CREATE_STRATEGY_WORST_FIT_BIT
	AllocationCreateStrategyFirstFit         = C.VMA_ALLOCATION_CREATE_STRATEGY_FIRST_FIT_BIT
	AllocationCreateStrategyMinMemory        = C.VMA_ALLOCATION_CREATE_STRATEGY_MIN_MEMORY_BIT
	AllocationCreateStrategyMinTime          = C.VMA_ALLOCATION_CREATE_STRATEGY_MIN_TIME_BIT
	AllocationCreateStrategyMinFragmentation = C.VMA_ALLOCATION_CREATE_STRATEGY_MIN_FRAGMENTATION_BIT
	AllocationCreateStrategyMask             = C.VMA_ALLOCATION_CREATE_STRATEGY_MASK
	AllocationCreateFlagBitsMaxEnum          = C.VMA_ALLOCATION_CREATE_FLAG_BITS_MAX_ENUM

	// VmaPoolCreateFlagBits

	PoolCreateIgnoreBufferImageGranularity = C.VMA_POOL_CREATE_IGNORE_BUFFER_IMAGE_GRANULARITY_BIT
	PoolCreateLinearAlgorithm              = C.VMA_POOL_CREATE_LINEAR_ALGORITHM_BIT
	PoolCreateBuddyAlgorithm               = C.VMA_POOL_CREATE_BUDDY_ALGORITHM_BIT
	PoolCreateAlgorithmMask                = C.VMA_POOL_CREATE_ALGORITHM_MASK
	PoolCreateFlagBitsMaxEnum              = C.VMA_POOL_CREATE_FLAG_BITS_MAX_ENUM

	// VmaDefragmentationFlagBits

	DefragmentationFlagIncremental = C.VMA_DEFRAGMENTATION_FLAG_INCREMENTAL
	DefragmentationFlagBitsMaxEnum = C.VMA_DEFRAGMENTATION_FLAG_BITS_MAX_ENUM
)
