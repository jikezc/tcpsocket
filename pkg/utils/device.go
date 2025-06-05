package utils

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// GetMACAddresses 获取所有网络接口的MAC地址
func getMACAddresses() (map[string]string, error) {
	// 获取所有网络接口
	interfaces, err := net.Interfaces()

	if err != nil {
		return nil, err
	}
	macs := make(map[string]string)
	for _, iface := range interfaces {
		// 跳过无效地址和回环接口
		if iface.Flags&net.FlagLoopback != 0 ||
			iface.HardwareAddr == nil ||
			len(iface.HardwareAddr) < 6 {
			continue
		}
		addresses, err := iface.Addrs()
		if err != nil {
			continue
		}
		// 格式化未标准的MAC地址格式
		mac := fmt.Sprintf(
			"%02x:%02x:%02x:%02x:%02x:%02x",
			iface.HardwareAddr[0], iface.HardwareAddr[1],
			iface.HardwareAddr[2], iface.HardwareAddr[3],
			iface.HardwareAddr[4], iface.HardwareAddr[5],
		)
		for _, addr := range addresses {
			macs[addr.String()] = mac
		}
	}
	return macs, nil
}

// GetMacByIp 根据IP地址获取MAC地址
func GetMacByIp(ip string) (string, error) {
	macs, err := getMACAddresses()
	fmt.Println(macs)
	if err != nil {
		return "", err
	}
	return macs[ip], nil
}

// GetDeviceUUID 获取设备ID
func GetDeviceUUID() (string, error) {

	// 尝试从DMI信息获取
	if uuid, err := os.ReadFile("/sys/class/dmi/id/product_uuid"); err == nil {
		return strings.TrimSpace(string(uuid)), nil
	}
	// 回退方案：通过dmidecode命令获取
	cmd := exec.Command("dmidecode", "-s", "system-uuid")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("无法获取设备ID: %v", err)
	} else {
		return strings.TrimSpace(string(out)), nil
	}
}

// GetFQDN 获取FQDN(完全限定域名)
func GetFQDN() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	addresses, err := net.LookupIP(hostname)
	for _, addr := range addresses {
		if ipv4 := addr.To4(); ipv4 != nil {
			names, err := net.LookupAddr(ipv4.String())
			if err != nil || len(names) == 0 {
				continue
			}
			return names[0], nil
		}
	}
	return hostname, nil
}

// GetOS  获取操作系统名称
func GetOS() string {
	if runtime.GOOS == "windows" {
		return "Windows"
	} else if runtime.GOOS == "darwin" {
		return "MacOS"
	} else {
		return "Linux"
	}
}

// GetPerformance  获取性能信息
func GetPerformance() (cpuPercent, menPercent float64) {
	memInfo, _ := mem.VirtualMemory()
	menPercent = memInfo.UsedPercent
	fmt.Printf("内存使用率: %v\n", memInfo.UsedPercent)

	cpuInfo, _ := cpu.Percent(1*time.Second, false)
	if len(cpuInfo) > 0 {
		cpuPercent = cpuInfo[0]
	}
	fmt.Printf("CPU使用率: %v\n", cpuPercent)

	return cpuPercent, menPercent
}
