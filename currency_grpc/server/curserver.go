package main

import (
	"fmt"
	"net"
	"os"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/vladimirvivien/go-grpc/currency_grpc/curproto"
	"github.com/vladimirvivien/go-grpc/currency_grpc/servutil"
)

type CurrencyService struct {
	curList *curproto.CurrencyList
}

func newCurrencyService(curList *curproto.CurrencyList) *CurrencyService {
	return &CurrencyService{curList: curList}
}

// GetCurrencyList searches (by Code or Number) and return CurrencyList
func (c *CurrencyService) GetCurrencyList(ctx context.Context, req *curproto.CurrencyRequest) (*curproto.CurrencyList, error) {
	var items []*curproto.Currency
	for _, cur := range c.curList.Items {
		if cur.Number == req.Number || cur.Code == req.Code {
			items = append(items, cur)
		}
	}
	return &curproto.CurrencyList{Items: items}, nil
}

// GetCurrencyStream returns matching Currencies as a server stream
func (c *CurrencyService) GetCurrencyStream(req *curproto.CurrencyRequest, stream curproto.CurrencyService_GetCurrencyStreamServer) error {
	for _, cur := range c.curList.Items {
		if cur.Number == req.Number || cur.Code == req.Code {
			if err := stream.Send(cur); err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	port := ":50051"

	// load data into protobuf structures
	curList, err := servutil.LoadPbFromCsv("./../curdata.csv")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	lstnr, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// setup and register currency service
	curService := newCurrencyService(curList)
	grpcServer := grpc.NewServer()
	curproto.RegisterCurrencyServiceServer(grpcServer, curService)

	// start service's server
	fmt.Println("starting currency rpc service on ", port)
	grpcServer.Serve(lstnr)
}
