package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	// 打印CPU核心数
	fmt.Printf("CPU核心数（gomaxproces）：%d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("Goroutine数量：%d\n", runtime.NumGoroutine())
	// 启动100个Goroutine
	for i := 0; i < 100; i++ {
		// 启动一个Goroutine
		go func(id int) {
			// time.Millisecond 是时间单位，100 * time.Millisecond 是 100 毫秒
			time.Sleep(100 * time.Millisecond)
		}(i)
	}
	// 打印启动后Goroutine数量
	fmt.Printf("启动后Goroutine数量: %d\n", runtime.NumGoroutine())
	// 等待200毫秒
	time.Sleep(200 * time.Millisecond)
	// 打印完成后Goroutine数量
	fmt.Printf("完成后Goroutine数量: %d\n", runtime.NumGoroutine())
}
