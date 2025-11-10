package sysinfo

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type SysInfo struct {
	CPUPercent      float64 `json:"cpu_percent"`       //
	DiskUsedPercent float64 `json:"disk_used_percent"` //
	MemoryPercent   float64 `json:"memory_percent"`    //
	// TotalMemory     uint64  `json:"total_memory_bytes"`
	// UsedMemory      uint64  `json:"used_memorybytes"`
	// DiskTotal       uint64  `json:"disk_total_bytes"`
	// DiskUsed        uint64  `json:"disk_used_bytes"`
	// Timestamp       int64   `json:"timestamp_unix"`
}

func GetSystemInfo(drivePath string) (SysInfo, error) {
	var info SysInfo

	//CPU
	cpuPercentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return info, err
	}

	if len(cpuPercentages) > 0 {
		info.CPUPercent = cpuPercentages[0]
	}

	//Memória
	vm, err := mem.VirtualMemory()

	if err != nil {
		return info, err
	}

	// info.TotalMemory = vm.Total
	// info.UsedMemory = vm.Used
	info.MemoryPercent = vm.UsedPercent

	//Disco
	disk, err := disk.Usage(drivePath)
	if err != nil {
		return info, err
	}

	// info.DiskTotal = disk.Total
	// info.DiskUsed = disk.Used
	info.DiskUsedPercent = disk.UsedPercent

	// info.Timestamp = time.Now().Unix()

	return info, nil
}
