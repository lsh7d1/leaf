package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"leaf/dal"
	"leaf/dao"
	v1 "leaf/pb/api/leaf/v1"
	"leaf/service"

	_ "github.com/mbobakov/grpc-consul-resolver"
	"github.com/panjf2000/ants/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 当发号器step逼近1e6时
// 本机通过rpc调用，qps大约在1e6上下
// goos: linux
// goarch: amd64
// pkg: leaf
// cpu: AMD Ryzen 7 5800H with Radeon Graphics
// BenchmarkMainInClient-16            8838            515490 ns/op            5168 B/op        103 allocs/op
// PASS
// ok      leaf    5.633s

func BenchmarkMainInClient(b *testing.B) {
	var addr = "127.0.0.1:9090"
	// 连接到server，禁用安全传输
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := v1.NewLeafSegmentServiceClient(conn)

	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, err := c.GetSegmentId(ctx, &v1.GenIdsRequest{BizTag: "bbbb", Count: 1})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
	}
}

func TestMainInClient(t *testing.T) {
	var addr = "127.0.0.1:9090"
	// 连接到server，禁用安全传输
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := v1.NewLeafSegmentServiceClient(conn)

	// 执行RPC调用并打印收到的响应数据
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	start := time.Now()
	_, err = c.GetSegmentId(ctx, &v1.GenIdsRequest{BizTag: "bbbb", Count: 1000000})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	// t.Logf("Greeting: %#v", r.GetIds())
	t.Logf("time used: %v", time.Since(start))
}
func TestServiceDiscovery(t *testing.T) {
	conn, err := grpc.NewClient(
		// consul服务
		"consul://localhost:8500/segment?healthy=true",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("grpc.NewClient failed, err: %v", err)
	}
	defer conn.Close()

	c := v1.NewLeafSegmentServiceClient(conn)
	ctx := context.TODO()
	resp, err := c.GetSegmentId(ctx, &v1.GenIdsRequest{BizTag: "bbbb", Count: 1})
	if err != nil {
		fmt.Printf("c.GetSegmentId failed, err:%v\n", err)
		return
	}
	fmt.Printf("resp:%v\n", resp.GetIds())
}
func TestSegmentService(t *testing.T) {
	svc := service.NewSegmentService(dao.NewSegmentRepoImpl(dal.ConnectSQLite("./test.db")))
	//svc.ShowBuffers()
	wg := sync.WaitGroup{}
	task := func() {
		value, err := svc.GetID(context.TODO(), "bbbb")
		_ = value
		if err != nil {
			t.Errorf("svc.GetID failed, err: %v", err)
		}
		fmt.Printf("%v ", value)
		wg.Done()
	}

	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		ants.Submit(task)
	}
	wg.Wait()
}

// BenchmarkServiceInPool
// BenchmarkServiceInPool-16        2307625               498.2 ns/op
func BenchmarkServiceInPool(b *testing.B) {
	svc := service.NewSegmentService(dao.NewSegmentRepoImpl(dal.ConnectSQLite("./test.db")))

	wg := sync.WaitGroup{}
	p, _ := ants.NewPool(16)
	defer p.Release()

	task := func() {
		_, _ = svc.GetID(context.TODO(), "bbbb")
		wg.Done()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		_ = p.Submit(task)
	}
	wg.Wait()
}

// BenchmarkServiceInChan
// BenchmarkServiceInChan-16        2570415               429.5 ns/op
func BenchmarkServiceInChan(b *testing.B) {
	svc := service.NewSegmentService(dao.NewSegmentRepoImpl(dal.ConnectSQLite("./test.db")))
	const numWorkers = 16
	tasks := make(chan func())
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		go func() {
			for task := range tasks {
				task()
				wg.Done()
			}
		}()
	}

	task := func() {
		_, _ = svc.GetID(context.TODO(), "bbbb")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		tasks <- task
	}

	wg.Wait()
	close(tasks)
}
