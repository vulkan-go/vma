# vma

[![](https://godoc.org/github.com/vulkan-go/vma?status.svg)](http://godoc.org/github.com/vulkan-go/vma)

[VulkanMemoryAllocator](https://github.com/GPUOpen-LibrariesAndSDKs/VulkanMemoryAllocator) bindings
for vulkan-go.

These bindings do not implement functions, structures or enumerations that are deprecated as of
v2.3.0. The following are also not implemented:
- Allocator callbacks (feel free to implement if anybody actually cares)
- Command recording (can be enabled by uncommenting lines from vma.go and vma.cpp. Does not seem to be supported on Linux)

Not all functions have been tested, please report bugs if you use this library.