// Package qwen Qwen 对话模型测试
package qwen

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"
)

// TestQwenChatSingle 测试 Qwen 对话模型的单轮对话
// go test -v ./internal/pkg/agent/llm/qwen/ -run TestQwenChatSingle
func TestQwenChatSingle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := QwenChat(ctx, "你是谁")
	if err != nil {
		t.Fatalf("QwenChat failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			t.Logf("context done: %v", ctx.Err())
			return
		case reply, ok := <-stream:
			if !ok {
				t.Log("stream closed")
				return
			}
			if reply == "" {
				t.Log("empty reply")
			} else {
				t.Logf("Reply: %.50s...", reply)
			}
		}
	}
}

var questions = []string{
	// 简单问题
	"Go语言的作者是谁",
	// 原有深度问题
	"Go语言的并发模型是怎样的",
	// 简单问题
	"Go语言中声明变量的关键字是什么",
	// 原有深度问题
	"如何用Go写一个高性能HTTP服务器",
	// 简单问题
	"Go中如何定义一个函数",
	// 原有深度问题
	"Go中context的作用是什么",
	// 简单问题
	"Go语言的main函数有返回值吗",
	// 原有深度问题
	"Go的垃圾回收机制是如何工作的",
	// 简单问题
	"Go中声明常量用什么关键字",
	// 原有深度问题
	"如何在Go中处理错误",
	// 简单问题
	"Go中切片和数组的核心区别是什么",
	// 原有深度问题
	"Go的interface有什么特点",
	// 简单问题
	"Go中如何导入第三方包",
	// 原有深度问题
	"怎么用Go实现一个简单的REST API",
	// 简单问题
	"Go中for循环有几种写法",
	// 原有深度问题
	"Go的goroutine和线程有什么区别",
	// 简单问题
	"Go中map是线程安全的吗",
	// 原有深度问题
	"如何对Go程序进行性能分析",
}

// BenchmarkQwenChatConcurrent 测试 Qwen 对话模型的并发对话场景
// # 默认并发 3
// go test -bench=BenchmarkQwenChatConcurrent -benchtime=1x ./internal/pkg/agent/llm/qwen/
//
// # 自定义并发 10
// BENCH_PARALLELISM=10 go test -bench=BenchmarkQwenChatConcurrent -benchtime=1x ./internal/pkg/agent/llm/qwen/
// 
// # 对比不同负载（配合 -count）
//BENCH_PARALLELISM=5 go test -bench=BenchmarkQwenChatConcurrent -count=3 ./...
func BenchmarkQwenChatConcurrent(b *testing.B) {
    parallelismStr := os.Getenv("BENCH_PARALLELISM")
    parallelism := 3 // 默认并发 worker 数
    if parallelismStr != "" {
        if p, err := strconv.Atoi(parallelismStr); err == nil && p > 0 {
            parallelism = p
        }
    }

    b.Logf("Setting parallelism to %d", parallelism)
    b.SetParallelism(parallelism) // ⚠️ 关键：设置每个迭代的并发 worker 数
    b.ResetTimer()

    // 使用 RunParallel 自动管理并发
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            // 每个 worker 独立 context
            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            defer cancel() // 注意：defer 在循环内，每次都会注册

            // 简单轮询问题（避免 rand 锁竞争）
            // 这里不能用 id，但可以用原子计数 or 取 pb 的隐式序号（不暴露）
            // 我们用一个简单 trick：取当前时间纳秒模
            question := questions[(time.Now().UnixNano() % int64(len(questions)))]
			reply, err := QwenChat(ctx, question)
			if err != nil {
				b.Fatalf("QwenChat failed: %v", err)
			}

			for r := range reply {
				b.Logf("Reply: %.50s...", r)
			}

        }
    })
}