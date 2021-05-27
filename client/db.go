package client

import (
	"github.com/asim/go-micro/v3"
	dbPb "github.com/pingcap/tcp/proto/db"
	"github.com/pingcap/tcp/service"

	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
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
var DbClient dbPb.DbService

func init() {
	appendToInitFpArray(initDbClient)
}

func initDbClient(srv micro.Service) {
	DbClient = dbPb.NewDbService(service.TCP_DB_SERVICE_NAME, srv.Client())
}