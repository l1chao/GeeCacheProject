package main

// import (
// 	"context"
// 	"fmt"
// )

// // 监听服务实例变化（比喻：电话簿更新时自动通知）
// func Watch(serviceName string, updateCh chan<- []string) {
// 	cli, err := clientv3.New(clientv3.Config{
// 		Endpoints: []string{"localhost:2379"},
// 	})
// 	if err != nil {
// 		close(updateCh)
// 		return
// 	}
// 	prefix := fmt.Sprintf("/services/%s/", serviceName)
// 	watcher := clientv3.NewWatcher(cli)
// 	// 首次获取全量数据
// 	initial, _ := Discover(serviceName)
// 	updateCh <- initial
// 	// 监听后续变化
// 	watchChan := watcher.Watch(context.Background(), prefix, clientv3.WithPrefix())
// 	go func() {
// 		for resp := range watchChan {
// 			// 当有节点增删时，重新获取全量数据（简单实现）
// 			// 生产环境建议增量更新
// 			current, _ := Discover(serviceName)
// 			updateCh <- current
// 		}
// 		defer cli.Close()
// 	}()
// }

// // 使用示例
// func main() {
// 	ch := make(chan []string, 10)
// 	go Watch("web-server", ch)
// 	// 实时打印变化
// 	for servers := range ch {
// 		fmt.Println("服务列表更新:", servers)
// 	}
// 	// 输出示例：
// 	// 服务列表更新: [192.168.1.10:8080]
// 	// 服务列表更新: [192.168.1.10:8080 192.168.1.11:8080]（新节点加入）
// 	// 服务列表更新: [192.168.1.11:8080]（旧节点下线）
// }
