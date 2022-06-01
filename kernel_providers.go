//go:build windows
// +build windows
package etw

/*
	#cgo LDFLAGS: -ltdh

	#include "etw.h"
*/
import "C"
import "golang.org/x/sys/windows"

var (
	/* 45d8cccd-539f-4b72-a8b7-5c683142609a */
	KERNEL_ALPC_GUID = windows.GUID{
		0x45d8cccd,
		0x539f,
		0x4b72,
		[8]byte{0xa8, 0xb7, 0x5c, 0x68, 0x31, 0x42, 0x60, 0x9a},
	}

	/* 13976d09-a327-438c-950b-7f03192815c7 */
	KERNEL_DEBUG_GUID = windows.GUID{
		0x13976d09,
		0xa327,
		0x438c,
		[8]byte{0x95, 0x0b, 0x7f, 0x03, 0x19, 0x28, 0x15, 0xc7},
	}

	/* 3d6fa8d4-fe05-11d0-9dda-00c04fd7ba7c */
	KERNEL_DISK_IO_GUID = windows.GUID{
		0x3d6fa8d4,
		0xfe05,
		0x11d0,
		[8]byte{0x9d, 0xda, 0x00, 0xc0, 0x4f, 0xd7, 0xba, 0x7c},
	}

	/* 01853a65-418f-4f36-aefc-dc0f1d2fd235 */
	KERNEL_EVENT_TRACE_CONFIG_GUID = windows.GUID{
		0x01853a65,
		0x418f,
		0x4f36,
		[8]byte{0xae, 0xfc, 0xdc, 0x0f, 0x1d, 0x2f, 0xd2, 0x35},
	}

	/* 90cbdc39-4a3e-11d1-84f4-0000f80464e3 */
	KERNEL_FILE_IO_GUID = windows.GUID{
		0x90cbdc39,
		0x4a3e,
		0x11d1,
		[8]byte{0x84, 0xf4, 0x00, 0x00, 0xf8, 0x04, 0x64, 0xe3},
	}

	/* 2cb15d1d-5fc1-11d2-abe1-00a0c911f518 */
	KERNEL_IMAGE_LOAD_GUID = windows.GUID{
		0x2cb15d1d,
		0x5fc1,
		0x11d2,
		[8]byte{0xab, 0xe1, 0x00, 0xa0, 0xc9, 0x11, 0xf5, 0x18},
	}

	/* 3d6fa8d3-fe05-11d0-9dda-00c04fd7ba7c */
	KERNEL_PAGE_FAULT_GUID = windows.GUID{
		0x3d6fa8d3,
		0xfe05,
		0x11d0,
		[8]byte{0x9d, 0xda, 0x00, 0xc0, 0x4f, 0xd7, 0xba, 0x7c},
	}

	/* ce1dbfb4-137e-4da6-87b0-3f59aa102cbc */
	KERNEL_PERF_INFO_GUID = windows.GUID{
		0xce1dbfb4,
		0x137e,
		0x4da6,
		[8]byte{0x87, 0xb0, 0x3f, 0x59, 0xaa, 0x10, 0x2c, 0xbc},
	}

	/* 3d6fa8d0-fe05-11d0-9dda-00c04fd7ba7c */
	KERNEL_PROCESS_GUID = windows.GUID{
		0x3d6fa8d0,
		0xfe05,
		0x11d0,
		[8]byte{0x9d, 0xda, 0x00, 0xc0, 0x4f, 0xd7, 0xba, 0x7c},
	}

	/* AE53722E-C863-11d2-8659-00C04FA321A1 */
	KERNEL_REGISTRY_GUID = windows.GUID{
		0xae53722e,
		0xc863,
		0x11d2,
		[8]byte{0x86, 0x59, 0x0, 0xc0, 0x4f, 0xa3, 0x21, 0xa1},
	}

	/* d837ca92-12b9-44a5-ad6a-3a65b3578aa8 */
	KERNEL_SPLIT_IO_GUID = windows.GUID{
		0xd837ca92,
		0x12b9,
		0x44a5,
		[8]byte{0xad, 0x6a, 0x3a, 0x65, 0xb3, 0x57, 0x8a, 0xa8},
	}

	/* 9a280ac0-c8e0-11d1-84e2-00c04fb998a2 */
	KERNEL_TCP_IP_GUID = windows.GUID{
		0x9a280ac0,
		0xc8e0,
		0x11d1,
		[8]byte{0x84, 0xe2, 0x00, 0xc0, 0x4f, 0xb9, 0x98, 0xa2},
	}

	/* 3d6fa8d1-fe05-11d0-9dda-00c04fd7ba7c */
	KERNEL_THREAD_GUID = windows.GUID{
		0x3d6fa8d1,
		0xfe05,
		0x11d0,
		[8]byte{0x9d, 0xda, 0x00, 0xc0, 0x4f, 0xd7, 0xba, 0x7c},
	}

	/* bf3a50c5-a9c9-4988-a005-2df0b7c80f80 */
	KERNEL_UDP_IP_GUID = windows.GUID{
		0xbf3a50c5,
		0xa9c9,
		0x4988,
		[8]byte{0xa0, 0x05, 0x2d, 0xf0, 0xb7, 0xc8, 0x0f, 0x80},
	}

	/* 9e814aad-3204-11d2-9a82-006008a86939 */
	KERNEL_SYSTEM_TRACE_GUID = windows.GUID{
		0x9e814aad,
		0x3204,
		0x11d2,
		[8]byte{0x9a, 0x82, 0x00, 0x60, 0x08, 0xa8, 0x69, 0x39},
	}

	/* 89497f50-effe-4440-8cf2-ce6b1cdcaca7 */
	KERNEL_OB_TRACE_GUID = windows.GUID{
		0x89497f50,
		0xeffe,
		0x4440,
		[8]byte{0x8c, 0xf2, 0xce, 0x6b, 0x1c, 0xdc, 0xac, 0xa7},
	}

	/* 0268a8b6-74fd-4302-9dd0-6e8f1795c0cf */
	KERNEL_POOL_TRACE_GUID = windows.GUID{
		0x0268a8b6,
		0x74fd,
		0x4302,
		[8]byte{0x9d, 0xd0, 0x6e, 0x8f, 0x17, 0x95, 0xc0, 0xcf},
	}

	/* 68fdd900-4a3e-11d1-84f4-0000f80464e3 */
	KERNEL_EVENT_TRACE_GUID = windows.GUID{
		0x68fdd900,
		0x4a3e,
		0x11d1,
		[8]byte{0x84, 0xf4, 0x00, 0x00, 0xf8, 0x04, 0x64, 0xe3},
	}

	/* 6a399ae0-4bc6-4de9-870b-3657f8947e7e */
	KERNEL_LOST_EVENT_GUID = windows.GUID{
		0x6a399ae0,
		0x4bc6,
		0x4de9,
		[8]byte{0x87, 0x0b, 0x36, 0x57, 0xf8, 0x94, 0x7e, 0x7e},
	}

	/* 9aec974b-5b8e-4118-9b92-3186d8002ce5 */
	KERNEL_UMS_EVENT_GUID = windows.GUID{
		0x9aec974b,
		0x5b8e,
		0x4118,
		[8]byte{0x9b, 0x92, 0x31, 0x86, 0xd8, 0x00, 0x2c, 0xe5},
	}

	/* def2fe46-7bd6-4b80-bd94-f57fe20d0ce3 */
	KERNEL_STACK_WALK_GUID = windows.GUID{
		0xdef2fe46,
		0x7bd6,
		0x4b80,
		[8]byte{0xbd, 0x94, 0xf5, 0x7f, 0xe2, 0x0d, 0x0c, 0xe3},
	}

	/* e43445e0-0903-48c3-b878-ff0fccebdd04 */
	KERNEL_POWER_GUID = windows.GUID{
		0xe43445e0,
		0x0903,
		0x48c3,
		[8]byte{0xb8, 0x78, 0xff, 0x0f, 0xcc, 0xeb, 0xdd, 0x04},
	}

	/* f8f10121-b617-4a56-868b-9df1b27fe32c */
	KERNEL_MMCSS_TRACE_GUID = windows.GUID{
		0xf8f10121,
		0xb617,
		0x4a56,
		[8]byte{0x86, 0x8b, 0x9d, 0xf1, 0xb2, 0x7f, 0xe3, 0x2c},
	}

	/* 3b9c9951-3480-4220-9377-9c8e5184f5cd */
	KERNEL_RUNDOWN_GUID = windows.GUID{
		0x3b9c9951,
		0x3480,
		0x4220,
		[8]byte{0x93, 0x77, 0x9c, 0x8e, 0x51, 0x84, 0xf5, 0xcd},
	}

	/**
	 * <summary>A provider that enables ALPC events.</summary>
	 */
	KERNEL_ALPC_PROVIDER = NewKernelProvider(KERNEL_ALPC_GUID, C.EVENT_TRACE_FLAG_ALPC)

	/**
	 * <summary>A provider that enables context switch events.</summary>
	 */
	KERNEL_CONTEXT_SWITCH_PROVIDER = NewKernelProvider(KERNEL_THREAD_GUID, C.EVENT_TRACE_FLAG_CSWITCH)

	/**
	 * <summary>A provider that enables debug print events.</summary>
	 */
	KERNEL_DEBUG_PRINT_PROVIDER = NewKernelProvider(KERNEL_DEBUG_GUID, C.EVENT_TRACE_FLAG_DBGPRINT)

	/**
	 * <summary>A provider that enables file I/O name events.</summary>
	 */
	KERNEL_DISK_FILE_IO_PROVIDER = NewKernelProvider(KERNEL_FILE_IO_GUID, C.EVENT_TRACE_FLAG_DISK_FILE_IO)

	/**
	 * <summary>A provider that enables disk I/O completion events.</summary>
	 */
	KERNEL_DISK_IO_PROVIDER = NewKernelProvider(KERNEL_DISK_IO_GUID, C.EVENT_TRACE_FLAG_DISK_IO)

	/**
	 * <summary>A provider that enables disk I/O start events.</summary>
	 */
	KERNEL_DISK_INIT_IO_PROVIDER = NewKernelProvider(KERNEL_DISK_IO_GUID, C.EVENT_TRACE_FLAG_DISK_IO_INIT)

	/**
	 * <summary>A provider that enables file I/O completion events.</summary>
	 */
	KERNEL_FILE_IO_PROVIDER = NewKernelProvider(KERNEL_FILE_IO_GUID, C.EVENT_TRACE_FLAG_FILE_IO)

	/**
	 * <summary>A provider that enables file I/O start events.</summary>
	 */
	KERNEL_FILE_INIT_IO_PROVIDER = NewKernelProvider(KERNEL_FILE_IO_GUID, C.EVENT_TRACE_FLAG_FILE_IO_INIT)

	/**
	 * <summary>A provider that enables thread dispatch events.</summary>
	 */
	KERNEL_THREAD_DISPATCH_PROVIDER = NewKernelProvider(KERNEL_THREAD_GUID, C.EVENT_TRACE_FLAG_DISPATCHER)

	/**
	 * <summary>A provider that enables device deferred procedure call events.</summary>
	 */
	KERNEL_DPC_PROVIDER = NewKernelProvider(KERNEL_PERF_INFO_GUID, C.EVENT_TRACE_FLAG_DPC)

	/**
	 * <summary>A provider that enables driver events.</summary>
	 */
	KERNEL_DRIVER_PROVIDER = NewKernelProvider(KERNEL_DISK_IO_GUID, C.EVENT_TRACE_FLAG_DRIVER)

	/**
	 * <summary>A provider that enables image load events.</summary>
	 */
	KERNEL_IMAGE_LOAD_PROVIDER = NewKernelProvider(KERNEL_IMAGE_LOAD_GUID, C.EVENT_TRACE_FLAG_IMAGE_LOAD)

	/**
	 * <summary>A provider that enables interrupt events.</summary>
	 */
	KERNEL_INTERRUPT_PROVIDER = NewKernelProvider(KERNEL_PERF_INFO_GUID, C.EVENT_TRACE_FLAG_INTERRUPT)

	/**
	 * <summary>A provider that enables memory hard fault events.</summary>
	 */
	KERNEL_MEMORY_HARD_FAULT_PROVIDER = NewKernelProvider(KERNEL_PAGE_FAULT_GUID, C.EVENT_TRACE_FLAG_MEMORY_HARD_FAULTS)

	/**
	 * <summary>A provider that enables memory page fault events.</summary>
	 */
	KERNEL_MEMORY_PAGE_FAULT_PROVIDER = NewKernelProvider(KERNEL_PAGE_FAULT_GUID, C.EVENT_TRACE_FLAG_MEMORY_PAGE_FAULTS)

	/**
	 * <summary>A provider that enables network tcp/ip events.</summary>
	 */
	KERNEL_NETWORK_TCPIP_PROVIDER = NewKernelProvider(KERNEL_TCP_IP_GUID, C.EVENT_TRACE_FLAG_NETWORK_TCPIP)

	/**
	 * <summary>A provider that enables process events.</summary>
	 */
	KERNEL_PROCESS_PROVIDER = NewKernelProvider(KERNEL_PROCESS_GUID, C.EVENT_TRACE_FLAG_PROCESS)

	/**
	 * <summary>A provider that enables process counter events.</summary>
	 */
	KERNEL_PROCESS_COUNTER_PROVIDER = NewKernelProvider(KERNEL_PROCESS_GUID, C.EVENT_TRACE_FLAG_PROCESS_COUNTERS)

	/**
	 * <summary>A provider that enables profiling events.</summary>
	 */
	KERNEL_PROFILE_PROVIDER = NewKernelProvider(KERNEL_PERF_INFO_GUID, C.EVENT_TRACE_FLAG_PROFILE)

	/**
	 * <summary>A provider that enables registry events.</summary>
	 */
	KERNEL_REGISTRY_PROVIDER = NewKernelProvider(KERNEL_REGISTRY_GUID, C.EVENT_TRACE_FLAG_REGISTRY)

	/**
	 * <summary>A provider that enables split I/O events.</summary>
	 */
	KERNEL_SPLIT_IO_PROVIDER = NewKernelProvider(KERNEL_SPLIT_IO_GUID, C.EVENT_TRACE_FLAG_SPLIT_IO)

	/**
	 * <summary>A provider that enables system call events.</summary>
	 */
	KERNEL_SYSTEM_CALL_PROVIDER = NewKernelProvider(KERNEL_PERF_INFO_GUID, C.EVENT_TRACE_FLAG_SYSTEMCALL)

	/**
	 * <summary>A provider that enables thread start and stop events.</summary>
	 */
	KERNEL_THREAD_PROVIDER = NewKernelProvider(KERNEL_THREAD_GUID, C.EVENT_TRACE_FLAG_THREAD)

	/**
	 * <summary>A provider that enables file map and unmap (excluding images) events.</summary>
	 */
	KERNEL_VAMAP_PROVIDER = NewKernelProvider(KERNEL_FILE_IO_GUID, C.EVENT_TRACE_FLAG_VAMAP)

	/**
	 * <summary>A provider that enables VirtualAlloc and VirtualFree events.</summary>
	 */
	KERNEL_VIRTUAL_ALLOC_PROVIDER = NewKernelProvider(KERNEL_PAGE_FAULT_GUID, C.EVENT_TRACE_FLAG_VIRTUAL_ALLOC)
)
