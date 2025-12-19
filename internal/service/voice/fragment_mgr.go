// Package voice  分片管理器
package voice

import (
	"bytes"
	"sync"
	"time"
)

type FragmentManager struct {
	buffers sync.Map // map[string]*fragmentBuffer
}

type fragmentBuffer struct {
	buf      *bytes.Buffer
	lastSeen time.Time
}

func NewFragmentManager() *FragmentManager {
	mgr := &FragmentManager{}
	// 启动清理 goroutine
	go mgr.cleanupLoop()
	return mgr
}

// Append 追加分片数据到缓冲区
// 参数：
//   - clientID: 客户端唯一标识符，用于区分不同的客户端分片
//   - data: 包含音频数据的字节切片
// 返回值：
//   - *bytes.Buffer: 包含所有已追加数据的缓冲区
func (mgr *FragmentManager) Append(clientID string, data []byte) *bytes.Buffer {
	now := time.Now()
	bufAny, _ := mgr.buffers.LoadOrStore(clientID, &fragmentBuffer{
		buf:      &bytes.Buffer{},
		lastSeen: now,
	})
	fb := bufAny.(*fragmentBuffer)
	fb.buf.Write(data)
	fb.lastSeen = now
	return fb.buf
}

// GetAndDelete 获取并删除指定客户端的分片缓冲区
// 参数：
//   - clientID: 客户端唯一标识符，用于区分不同的客户端分片
// 返回值：
//   - *bytes.Buffer: 包含所有已追加数据的缓冲区
func (mgr *FragmentManager) GetAndDelete(clientID string) *bytes.Buffer {
	if bufAny, ok := mgr.buffers.LoadAndDelete(clientID); ok {
		return bufAny.(*fragmentBuffer).buf
	}
	return nil
}

// cleanupLoop 清理过期分片缓冲区的 goroutine
// 该 goroutine 每 60 秒检查一次所有分片缓冲区，删除超过 2 分钟未更新的缓冲区
func (mgr *FragmentManager) cleanupLoop() {
	ticker := time.NewTicker(60 * time.Second)
	for range ticker.C {
		now := time.Now()
		mgr.buffers.Range(func(key, value any) bool {
			fb := value.(*fragmentBuffer)
			if now.Sub(fb.lastSeen) > 2*time.Minute { // 超时 2 分钟
				mgr.buffers.Delete(key)
			}
			return true
		})
	}
}