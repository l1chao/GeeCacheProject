package SyncSection

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSync(t *testing.T) {
	var wg sync.WaitGroup
	urls := []string{"url1", "url2", "url3"}

	for _, url := range urls {
		wg.Add(1) // 增加等待计数
		go func(u string) {
			defer wg.Done() // 任务完成时减少计数
			download(u)
		}(url)
	}

	wg.Wait() // 阻塞，直到所有下载完成
	fmt.Println("所有文件下载完成")
}

func download(url string) {
	print(url + "下载中...\n")
	time.Sleep(time.Duration(2) * time.Second) // time.Duration(2) 是类型转换
	print(url + "下载成功！\n")
}
