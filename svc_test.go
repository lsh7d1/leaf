package main

import (
	"context"
	"fmt"
	"leaf/dal"
	"leaf/dao"
	"leaf/service"
	"sync"
	"testing"

	"github.com/panjf2000/ants/v2"
)

const mysqldsn = "root:root1234@tcp(127.0.0.1:13306)/db2?charset=utf8mb4&parseTime=True"

var _ = mysqldsn

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
