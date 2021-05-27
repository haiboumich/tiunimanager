package client

import (
	commonPb "github.com/pingcap/tcp/proto/common"
	"github.com/pingcap/tcp/service"

	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/v3"
)

// Make request
/*
	rsp, err := TcpClient.Hello(context.Background(), &pb.HelloRequest{
		Name: "Foo",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(rsp.Greeting)
*/
var CommonClient commonPb.CommonService

func init() {
	appendToInitFpArray(initCommonClient)
}

func initCommonClient(srv micro.Service) {
	CommonClient = commonPb.NewCommonService(service.TCP_COMMON_SERVICE_NAME, srv.Client())
}