package util

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"net"
	"os"
	"sync"
	"time"
)

// UUID 类型定义
type UUID [16]byte

// UUID 版本常量
const (
	VersionNil = 0x00
	VersionV1  = 0x10 // basic time the UUID
	VersionV3  = 0x30 // basic name the MD5 hash
	VersionV4  = 0x40 // round UUID
	VersionV5  = 0x50 // basic name the SHA1 hash
	VersionV6  = 0x60 // basic date sort the UUID
	VersionV7  = 0x70 // basic Unix timestamp UUID
)

// UUID 变体常量
const (
	VariantNCS       = 0x80 // NCS向后兼容
	VariantRFC4122   = 0x40 // RFC 4122标准
	VariantMicrosoft = 0x20 // Microsoft向后兼容
	VariantFuture    = 0x00 // 未来保留
)

// Generator UUID生成器结构体
type Generator struct {
	mu              sync.Mutex
	clockSequence   uint16
	nodeID          []byte
	lastTimestamp   int64
	randomGenerator *randReader
}

// randReader 线程安全的随机数生成器
type randReader struct {
	mu sync.Mutex
}

func (r *randReader) Read(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return rand.Read(p)
}

// NewGenerator 创建新的UUID生成器
func NewGenerator() (*Generator, error) {
	g := &Generator{
		clockSequence:   generateClockSequence(),
		nodeID:          getNodeID(),
		lastTimestamp:   0,
		randomGenerator: &randReader{},
	}
	return g, nil
}

// generateClockSequence 生成时钟序列
func generateClockSequence() uint16 {
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		return uint16(time.Now().UnixNano() & 0x3FFF)
	}
	return binary.BigEndian.Uint16(b) & 0x3FFF
}

// getNodeID 获取节点ID（MAC地址或随机生成）
func getNodeID() []byte {
	// 尝试获取MAC地址
	if interfaces, err := net.Interfaces(); err == nil {
		for _, iface := range interfaces {
			if len(iface.HardwareAddr) >= 6 {
				nodeID := make([]byte, 6)
				copy(nodeID, iface.HardwareAddr)
				// 设置多播位
				nodeID[0] |= 0x01
				return nodeID
			}
		}
	}

	// 如果无法获取MAC地址，使用随机值
	nodeID := make([]byte, 6)
	if _, err := rand.Read(nodeID); err != nil {
		// 最后的备选方案
		hostname, _ := os.Hostname()
		h := sha1.Sum([]byte(hostname))
		copy(nodeID, h[:6])
		nodeID[0] |= 0x01
	}
	return nodeID
}

// getCurrentTimestamp 获取当前时间戳（100纳秒精度）
func getCurrentTimestamp() int64 {
	return time.Now().UnixNano() / 100
}

// UUIDVersion 获取UUID版本
func (u UUID) Version() byte {
	return u[6] >> 4
}

// UUIDVariant 获取UUID变体
func (u UUID) Variant() byte {
	switch {
	case u[8]>>7 == 0x01:
		return VariantRFC4122
	case u[8]>>6 == 0x02:
		return VariantMicrosoft
	case u[8]>>5 == 0x06:
		return VariantFuture
	default:
		return VariantNCS
	}
}

// String 实现Stringer接口
func (u UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		u[0:4],
		u[4:6],
		u[6:8],
		u[8:10],
		u[10:16])
}

// Hex 返回十六进制字符串（无分隔符）
func (u UUID) Hex() string {
	return hex.EncodeToString(u[:])
}

// Bytes 返回字节切片
func (u UUID) Bytes() []byte {
	return u[:]
}

// IsNil 检查是否为零值UUID
func (u UUID) IsNil() bool {
	for _, b := range u {
		if b != 0 {
			return false
		}
	}
	return true
}

// Equal 比较两个UUID是否相等
func (u UUID) Equal(other UUID) bool {
	return u == other
}

// ============ UUID生成方法 ============

// UUIDV1 生成基于时间的UUID (RFC 4122)
func (g *Generator) UUIDV1() UUID {
	g.mu.Lock()
	defer g.mu.Unlock()

	var uuid UUID
	timestamp := getCurrentTimestamp()

	// 处理时钟序列回退
	if timestamp <= g.lastTimestamp {
		g.clockSequence = (g.clockSequence + 1) & 0x3FFF
	}
	g.lastTimestamp = timestamp

	// 设置时间戳 (60位)
	// 从1582-10-15 00:00:00到现在的100纳秒数
	timestamp += 0x01B21DD213814000

	// 时间戳低32位
	binary.BigEndian.PutUint32(uuid[0:4], uint32(timestamp&0xFFFFFFFF))
	// 时间戳中间16位
	binary.BigEndian.PutUint16(uuid[4:6], uint16((timestamp>>32)&0xFFFF))
	// 时间戳高16位
	binary.BigEndian.PutUint16(uuid[6:8], uint16((timestamp>>48)&0x0FFF))

	// 设置版本 (V1)
	uuid[6] = (uuid[6] & 0x0F) | VersionV1

	// 设置时钟序列
	binary.BigEndian.PutUint16(uuid[8:10], g.clockSequence)
	uuid[8] = (uuid[8] & 0x3F) | VariantRFC4122

	// 设置节点ID
	copy(uuid[10:16], g.nodeID)

	return uuid
}

// UUIDV3 生成基于名称的MD5哈希UUID
func UUIDV3(namespace UUID, name string) UUID {
	return generateHashUUID(namespace, name, sha1.New(), VersionV3)
}

// UUIDV4 生成随机UUID
func (g *Generator) UUIDV4() UUID {
	var uuid UUID
	if _, err := g.randomGenerator.Read(uuid[:]); err != nil {
		// 如果随机数生成失败，使用time作为备选
		timestamp := time.Now().UnixNano()
		for i := 0; i < 16; i++ {
			uuid[i] = byte(timestamp >> (i * 8))
		}
	}

	// 设置版本 (V4)
	uuid[6] = (uuid[6] & 0x0F) | VersionV4
	// 设置变体 (RFC4122)
	uuid[8] = (uuid[8] & 0x3F) | VariantRFC4122

	return uuid
}

// UUIDV5 生成基于名称的SHA1哈希UUID
func UUIDV5(namespace UUID, name string) UUID {
	return generateHashUUID(namespace, name, sha1.New(), VersionV5)
}

// UUIDV6 生成基于时间的排序UUID (支持排序)
func (g *Generator) UUIDV6() UUID {
	g.mu.Lock()
	defer g.mu.Unlock()

	var uuid UUID
	timestamp := getCurrentTimestamp()

	if timestamp <= g.lastTimestamp {
		g.clockSequence = (g.clockSequence + 1) & 0x3FFF
	}
	g.lastTimestamp = timestamp

	// UUIDV6: 时间戳在前 (便于排序)
	binary.BigEndian.PutUint64(uuid[0:8], uint64(timestamp))

	// 设置版本 (V6)
	uuid[6] = (uuid[6] & 0x0F) | VersionV6

	// 设置时钟序列
	binary.BigEndian.PutUint16(uuid[8:10], g.clockSequence)
	uuid[8] = (uuid[8] & 0x3F) | VariantRFC4122

	// 设置节点ID
	copy(uuid[10:16], g.nodeID)

	return uuid
}

// UUIDV7 生成基于Unix时间戳的UUID (按时间排序)
func (g *Generator) UUIDV7() UUID {
	var uuid UUID

	// Unix时间戳 (毫秒)
	timestampMs := uint64(time.Now().UnixMilli())

	// 填充时间戳 (前48位)
	binary.BigEndian.PutUint64(uuid[0:8], timestampMs<<16)

	// 填充随机数据
	randData := make([]byte, 8)
	if _, err := g.randomGenerator.Read(randData); err != nil {
		// 备选方案
		nano := time.Now().UnixNano()
		binary.BigEndian.PutUint64(randData, uint64(nano))
	}

	// 填充随机部分
	copy(uuid[8:16], randData[:8])

	// 设置版本 (V7)
	uuid[6] = (uuid[6] & 0x0F) | VersionV7
	// 设置变体 (RFC4122)
	uuid[8] = (uuid[8] & 0x3F) | VariantRFC4122

	return uuid
}

// generateHashUUID 生成基于哈希的UUID
func generateHashUUID(namespace UUID, name string, h hash.Hash, version byte) UUID {
	var uuid UUID

	// 重置哈希
	h.Reset()
	// 写入命名空间
	h.Write(namespace[:])
	// 写入名称
	h.Write([]byte(name))

	// 计算哈希
	hashBytes := h.Sum(nil)

	// 复制到UUID
	copy(uuid[:], hashBytes[:16])

	// 设置版本
	uuid[6] = (uuid[6] & 0x0F) | version
	// 设置变体
	uuid[8] = (uuid[8] & 0x3F) | VariantRFC4122

	return uuid
}

// ============ 便捷方法 ============

var defaultGenerator, _ = NewGenerator()

// UUIDV1 快捷方法
func UUIDV1() UUID {
	uuid := defaultGenerator.UUIDV1()
	return uuid
}

// UUIDV4 快捷方法
func UUIDV4() UUID {
	return defaultGenerator.UUIDV4()
}

// UUIDV6 快捷方法
func UUIDV6() UUID {
	return defaultGenerator.UUIDV6()
}

// UUIDV7 快捷方法
func UUIDV7() UUID {
	return defaultGenerator.UUIDV7()
}

// ============ 常用命名空间 ============

var (
	NamespaceDNS  = UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	NamespaceURL  = UUID{0x6b, 0xa7, 0xb8, 0x11, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	NamespaceOID  = UUID{0x6b, 0xa7, 0xb8, 0x12, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	NamespaceX500 = UUID{0x6b, 0xa7, 0xb8, 0x14, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
)
