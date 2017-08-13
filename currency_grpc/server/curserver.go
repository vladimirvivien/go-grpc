package main

import (
	"errors"
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

// GetCurrency searches and return Currency by code or name
func (c *CurrencyService) GetCurrency(ctx context.Context, req *curproto.CurrencyRequest) (*curproto.Currency, error) {
	for _, cur := range c.curList.Items {
		if cur.Code == req.Code {
			return cur, nil
		}
	}
	return nil, errors.New("currency not found")
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
