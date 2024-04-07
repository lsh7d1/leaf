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

	"github.com/panjf2000/ants/v2"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		panic(fmt.Errorf(`net.Listen("tcp", ":9090") failed, err: %v`, err))
	}

	s := grpc.NewServer()
	pb.RegisterLeafSegmentServiceServer(s, newIdIssuer())

	err = s.Serve(lis)
	if err != nil {
		panic(fmt.Errorf("s.Serve(lis) failed, err: %v", err))
	}
}

const mysqldsn = "root:root1234@tcp(192.168.10.7:13306)/db2?charset=utf8mb4&parseTime=True"

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

	svc := service.NewSegmentService(dao.NewSegmentRepoImpl(dal.ConnectMySQL(mysqldsn).Debug()))
	// svc := service.NewSegmentService(dao.NewSegmentRepoImpl(dal.ConnectSQLite("./test.db")))
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
