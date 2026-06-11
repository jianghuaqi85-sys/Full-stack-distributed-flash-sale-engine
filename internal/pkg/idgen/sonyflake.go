package idgen

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/sony/sonyflake"
)

var sf *sonyflake.Sonyflake

// Init 初始化 Sonyflake ID 生成器
// startTime 是纪元起点，设置为 2024-01-01 00:00:00 UTC
// MachineID 优先级：环境变量 MACHINE_ID > K8s Pod 序号 > 本地 IP
func Init() error {
	return InitWithMachineID(nil)
}

// InitWithMachineID 使用自定义 MachineID 函数初始化
// 如果 machineIDFn 为 nil，则使用默认策略
func InitWithMachineID(machineIDFn func() (uint16, error)) error {
	var st sonyflake.Settings
	st.StartTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	if machineIDFn != nil {
		st.MachineID = machineIDFn
	} else {
		st.MachineID = defaultMachineID
	}

	sf = sonyflake.NewSonyflake(st)
	if sf == nil {
		return fmt.Errorf("failed to create sonyflake instance")
	}

	// 验证可用性
	_, err := sf.NextID()
	if err != nil {
		return fmt.Errorf("sonyflake not working: %w", err)
	}

	return nil
}

// defaultMachineID 默认 MachineID 策略
// 优先级：MACHINE_ID 环境变量 > K8s Pod 序号 > 本地 IP
func defaultMachineID() (uint16, error) {
	// 1. 尝试从环境变量获取（适用于手动分配）
	if idStr := os.Getenv("MACHINE_ID"); idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 16)
		if err == nil {
			return uint16(id), nil
		}
	}

	// 2. 尝试从 K8s Pod 名称提取序号（适用于 Kubernetes）
	// Pod 名称格式：statefulset-name-0, statefulset-name-1, ...
	if podName := os.Getenv("HOSTNAME"); podName != "" {
		if id := extractPodOrdinal(podName); id >= 0 {
			return uint16(id), nil
		}
	}

	// 3. 使用本地 IP 最后两个字节（默认策略）
	return ipMachineID()
}

// extractPodOrdinal 从 K8s Pod 名称中提取序号
// 例如：seckill-0 -> 0, seckill-1 -> 1
func extractPodOrdinal(podName string) int {
	for i := len(podName) - 1; i >= 0; i-- {
		if podName[i] == '-' {
			ordinal, err := strconv.Atoi(podName[i+1:])
			if err == nil {
				return ordinal
			}
			return -1
		}
	}
	return -1
}

// ipMachineID 使用本地 IP 最后两个字节作为 MachineID
func ipMachineID() (uint16, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return 0, fmt.Errorf("failed to get interface addresses: %w", err)
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ip4 := ipNet.IP.To4(); ip4 != nil {
				return uint16(ip4[2])<<8 | uint16(ip4[3]), nil
			}
		}
	}

	return 0, fmt.Errorf("no suitable network interface found")
}

// HealthCheck 验证 Sonyflake 生成器是否正常工作
func HealthCheck() error {
	if sf == nil {
		return fmt.Errorf("sonyflake not initialized")
	}
	_, err := sf.NextID()
	return err
}

// IsInitialized 检查 Sonyflake 是否已初始化
func IsInitialized() bool {
	return sf != nil
}

// NextID 生成下一个唯一 ID (uint64)
func NextID() (uint64, error) {
	if sf == nil {
		return 0, fmt.Errorf("sonyflake not initialized, call Init() first")
	}
	return sf.NextID()
}

// MustNextID 生成下一个唯一 ID，失败时 panic
func MustNextID() uint64 {
	id, err := NextID()
	if err != nil {
		panic(fmt.Sprintf("failed to generate ID: %v", err))
	}
	return id
}

// OrderNo 生成订单号
// 格式: TK + Sonyflake ID 的十六进制表示
func OrderNo() string {
	id := MustNextID()
	return fmt.Sprintf("TK%016x", id)
}
