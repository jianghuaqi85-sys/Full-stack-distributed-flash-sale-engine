package mq

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"order-system/internal/pkg/idgen"
)

// idGenAdapter ID 生成器适配器
type idGenAdapter struct{}

func (a *idGenAdapter) OrderNo() string {
	return idgen.OrderNo()
}

var adapter *idGenAdapter

// getIDGen 获取 ID 生成器适配器
func getIDGen() *idGenAdapter {
	if adapter != nil {
		return adapter
	}
	// 尝试初始化
	if err := idgen.Init(); err != nil {
		return nil
	}
	adapter = &idGenAdapter{}
	return adapter
}

// generateFallbackOrderNo 降级订单号生成方案
func generateFallbackOrderNo() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("TK%s%s", time.Now().Format("20060102150405"), hex.EncodeToString(b))
}
