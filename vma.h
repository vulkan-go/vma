#ifndef _VMA_GO_H
#define _VMA_GO_H

#define VK_NO_PROTOTYPES
#include "vulkan/vulkan.h"

#define VMA_STATIC_VULKAN_FUNCTIONS 0
#include "vk_mem_alloc.h"

typedef void* (*getInstanceProcAddr)(VkInstance, const char*);

void initVulkanFunctions(void* getProcAddrPtr, VkInstance instance, VmaVulkanFunctions* funcs) {
	getInstanceProcAddr getProcAddr = (getInstanceProcAddr)getProcAddrPtr;

	funcs->vkGetPhysicalDeviceProperties = (PFN_vkGetPhysicalDeviceProperties)((*getProcAddr)(instance, "vkGetPhysicalDeviceProperties"));
	funcs->vkGetPhysicalDeviceMemoryProperties = (PFN_vkGetPhysicalDeviceMemoryProperties)((*getProcAddr)(instance, "vkGetPhysicalDeviceMemoryProperties"));
	funcs->vkAllocateMemory = (PFN_vkAllocateMemory)((*getProcAddr)(instance, "vkAllocateMemory"));
	funcs->vkFreeMemory = (PFN_vkFreeMemory)((*getProcAddr)(instance, "vkFreeMemory"));
	funcs->vkMapMemory = (PFN_vkMapMemory)((*getProcAddr)(instance, "vkMapMemory"));
	funcs->vkUnmapMemory = (PFN_vkUnmapMemory)((*getProcAddr)(instance, "vkUnmapMemory"));
	funcs->vkFlushMappedMemoryRanges = (PFN_vkFlushMappedMemoryRanges)((*getProcAddr)(instance, "vkFlushMappedMemoryRanges"));
	funcs->vkInvalidateMappedMemoryRanges = (PFN_vkInvalidateMappedMemoryRanges)((*getProcAddr)(instance, "vkInvalidateMappedMemoryRanges"));
	funcs->vkBindBufferMemory = (PFN_vkBindBufferMemory)((*getProcAddr)(instance, "vkBindBufferMemory"));
	funcs->vkBindImageMemory = (PFN_vkBindImageMemory)((*getProcAddr)(instance, "vkBindImageMemory"));
	funcs->vkGetBufferMemoryRequirements = (PFN_vkGetBufferMemoryRequirements)((*getProcAddr)(instance, "vkGetBufferMemoryRequirements"));
	funcs->vkGetImageMemoryRequirements = (PFN_vkGetImageMemoryRequirements)((*getProcAddr)(instance, "vkGetImageMemoryRequirements"));
	funcs->vkCreateBuffer = (PFN_vkCreateBuffer)((*getProcAddr)(instance, "vkCreateBuffer"));
	funcs->vkDestroyBuffer = (PFN_vkDestroyBuffer)((*getProcAddr)(instance, "vkDestroyBuffer"));
	funcs->vkCreateImage = (PFN_vkCreateImage)((*getProcAddr)(instance, "vkCreateImage"));
	funcs->vkDestroyImage = (PFN_vkDestroyImage)((*getProcAddr)(instance, "vkDestroyImage"));
	funcs->vkCmdCopyBuffer = (PFN_vkCmdCopyBuffer)((*getProcAddr)(instance, "vkCmdCopyBuffer"));
	funcs->vkGetBufferMemoryRequirements2KHR = (PFN_vkGetBufferMemoryRequirements2)((*getProcAddr)(instance, "vkGetBufferMemoryRequirements2KHR"));
	funcs->vkGetImageMemoryRequirements2KHR = (PFN_vkGetImageMemoryRequirements2KHR)((*getProcAddr)(instance, "vkGetImageMemoryRequirements2KHR"));
}

#endif // include guard