package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

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

// SaveCurrencyStream adds Currency values from a stream to currency items
func (c *CurrencyService) SaveCurrencyStream(stream curproto.CurrencyService_SaveCurrencyStreamServer) error {
	curList := new(curproto.CurrencyList)
	for {
		cur, err := stream.Recv()
		if err != nil {
			// if done, close sream and return result
			if err == io.EOF {
				copy(c.curList.Items, curList.Items) // save to list
				return stream.SendAndClose(curList)
			}
			return err
		}
		curList.Items = append(curList.Items, cur)
	}
}

// FindCurrencyStream sends a stream of CurrencyRequest while streaming Currency values from server.
func (c *CurrencyService) FindCurrencyStream(stream curproto.CurrencyService_FindCurrencyStreamServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil // we're done
			}
			return err
		}

		// search and stream result
		for _, cur := range c.curList.Items {
			if cur.Code == req.Code || cur.Number == req.Number {
				if err := stream.Send(cur); err != nil {
					return err
				}
			}
		}
	}
}

func main() {
	port := ":50051"
	//certFile := flag.String("certfile", "", "TLS certificate file")
	//keyFile := flag.String("keyfile", "", "TLS private key")
	//flag.Parse()

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

	// setup tls options
	tlsCreds, err := credentials.NewServerTLSFromFile("./../certs/server.crt", "./../certs/server.key")
	if err != nil {
		log.Fatal(err)
	}

	// setup and register currency service
	curService := newCurrencyService(curList)
	grpcServer := grpc.NewServer(grpc.Creds(tlsCreds))
	curproto.RegisterCurrencyServiceServer(grpcServer, curService)

	// start service's server
	fmt.Println("starting currency rpc service on ", port)
	grpcServer.Serve(lstnr)
}
