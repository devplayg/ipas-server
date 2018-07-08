package sysmonitor

import (
	"encoding/json"
	"github.com/devplayg/ipas-server"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"
	"time"
)

func UpdateResource(e *ipasserver.Engine, partition string) error {
	if err := updateResource(e, partition); err != nil {
		log.Error(err)
	}

	return nil
}

func updateResource(e *ipasserver.Engine, partition string) error {
	// CPU
	cpu, cpuInfo, err := getCPUInfo()
	if err != nil {
		return err
	}

	// Memory
	memInfo, err := getMemoryInfo()
	if err != nil {
		return err
	}

	// Disk
	diskInfo, err := getDiskInfo()
	if err != nil {
		return err
	}

	query := `
     insert into ast_server(
         name,
         category1,
         category2,
         hostname,
         cpu_usage,
         mem_total,
         mem_used,
         disk_total,
         disk_used,
         cpu_comment,
         mem_comment,
         disk_comment,
		 updated
     )
     values ('localhost', 1, 1, '127.0.0.1', ?, ?, ?, ?, ?, ?, ?, ?, now())
     on duplicate key update
         cpu_usage = values(cpu_usage),
         mem_total = values(mem_total),
         mem_used = values(mem_used),
         disk_total = values(disk_total),
         disk_used = values(disk_used),
         cpu_comment = values(cpu_comment),
         mem_comment = values(mem_comment),
         disk_comment = values(disk_comment),
		updated = values(updated)
		
     ;
	`
	cpuComment, _ := json.Marshal(cpuInfo)
	memComment, _ := json.Marshal(memInfo)
	diskComment, _ := json.Marshal(diskInfo)

	// CPU
	var cpuUsage float64
	if len(cpu) > 0 {
		cpuUsage = cpu[0]
	}

	// Memory
	var memTotal uint64
	var memUsed uint64
	if memInfo != nil {
		memTotal = memInfo.Total
		memUsed = memInfo.Used
	}

	// Disk
	var disk *disk.UsageStat
	for _, d := range diskInfo {
		if d.Path == partition {
			disk = d
		}
	}
	var diskTotal uint64
	var diskUsed uint64
	if disk != nil {
		diskTotal = disk.Total
		diskUsed = disk.Used
	}
	_, err = e.DB.Exec(query, cpuUsage, memTotal, memUsed, diskTotal, diskUsed, cpuComment, memComment, diskComment)
	return err
}

func getCPUInfo() ([]float64, []cpu.InfoStat, error) {
	percent, err := cpu.Percent(1000*time.Millisecond, false)
	if err != nil {
		return percent, nil, err
	}

	info, err := cpu.Info()
	if err != nil {
		return percent, nil, err
	}

	return percent, info, nil
}

func getMemoryInfo() (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemory()
}

func getDiskInfo() ([]*disk.UsageStat, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	var usage []*disk.UsageStat
	for _, part := range parts {
		u, err := disk.Usage(part.Mountpoint)
		if err != nil {
			return nil, err
		}
		usage = append(usage, u)
	}
	return usage, err
}

type SystemMonitor struct {
	engine *ipasserver.Engine
	disk string
}
