package main

import (
	"context"
	"fmt"
	"net"
	"runtime"

	"leaf/dal"
	"leaf/dao"
	pb "leaf/pb/api/leaf/v1"
	"leaf/service"

	"github.com/hashicorp/consul/api"
	"github.com/panjf2000/ants/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	serviceName = "segment"
	port        = 9090
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(fmt.Errorf(`net.Listen("tcp", ":%d") failed, err: %v`, port, err))
	}

	s := grpc.NewServer()
	pb.RegisterLeafSegmentServiceServer(s, newIdIssuer())

	healthpb.RegisterHealthServer(s, health.NewServer()) // consul 发来健康检查的RPC请求，这个负责返回OK

	cc, err := api.NewClient(api.DefaultConfig()) // 127.0.0.1:8500
	if err != nil {
		panic(fmt.Errorf("api.NewClient failed, err: %v", err))
	}

	ipInfo, err := getOutIP()
	if err != nil {
		panic(fmt.Errorf("getOutIP failed, err: %v", err))
	}

	// 配置健康检查策略
	check := &api.AgentServiceCheck{
		GRPC:                           fmt.Sprintf("%s:%d", ipInfo.String(), port),
		Timeout:                        "30s",
		Interval:                       "10s",
		DeregisterCriticalServiceAfter: "1m30s",
	}

	srv := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%s-%d", serviceName, ipInfo.String(), port), // 服务唯一ID
		Name:    serviceName,
		Tags:    []string{"segment", "lsh7d1"},
		Address: ipInfo.String(),
		Port:    port,
		Check:   check,
	}

	// 注册服务到Consul
	cc.Agent().ServiceRegister(srv)

	err = s.Serve(lis)
	if err != nil {
		panic(fmt.Errorf("s.Serve(lis) failed, err: %v", err))
	}
}

func getOutIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	localaddr := conn.LocalAddr().(*net.UDPAddr)
	return localaddr.IP, nil
}

const mysqldsn = "root:root1234@tcp(192.168.10.7:13306)/db2?charset=utf8mb4&parseTime=True"

var _ = mysqldsn

type idIssuer struct {
	pb.UnimplementedLeafSegmentServiceServer
	svc *service.SegmentService
	p   *ants.Pool
}

func newIdIssuer() pb.LeafSegmentServiceServer {
	p, err := ants.NewPool(runtime.NumCPU())
	if err != nil {
		panic(fmt.Errorf("ants.NewPool(runtime.NumCPU()) failed, err: %v", err))
	}

	// svc := service.NewSegmentService(dao.NewSegmentRepoImpl(dal.ConnectMySQL(mysqldsn).Debug()))
	svc := service.NewSegmentService(dao.NewSegmentRepoImpl(dal.ConnectSQLite("./test.db")))
	i := &idIssuer{
		svc: svc,
		p:   p,
	}
	return i
}

func (issuer *idIssuer) GetSegmentId(ctx context.Context, req *pb.GenIdsRequest) (resp *pb.GenIdsResponse, err error) {
	resp = &pb.GenIdsResponse{Ids: make([]int64, 0, req.GetCount())}
	// fmt.Printf("req.BizTag: %v, req.Count: %v\n", req.BizTag, req.Count)
	for id, i := int64(-1), int32(0); i < req.Count; i++ {
		// issuer.p.Submit(func() {
		id, err = issuer.svc.GetID(ctx, req.BizTag)
		// })
		if err != nil {
			return nil, err
		}
		resp.Ids = append(resp.Ids, id)
	}
	return
}
