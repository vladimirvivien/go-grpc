package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc/codes"

	pb "github.com/vladimirvivien/go-grpc/protobuf"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	server     = "127.0.0.1"
	serverPort = "50051"
)

// printUSD demonstrates simple binary call from client
func printUSD(client pb.CurrencyServiceClient) {
	curReq := &pb.CurrencyRequest{}
	curList, err := client.GetCurrencyList(context.Background(), curReq)
	if err != nil {
		fmt.Println("error in printUSD:", err)
		return
	}

	fmt.Println("\nUSD Countries")
	fmt.Println("-------------")
	for _, cur := range curList.Items {
		fmt.Printf("%-50s%-10s\n", cur.GetCountry(), cur.GetCode())
	}
}

// printEUR demonstrates server stream call from client
func printEUR(client pb.CurrencyServiceClient) {
	curReq := &pb.CurrencyRequest{}

	// client.GetCurrencyStream() is not calling server.GetCurrencyStream()
	// directly. This is a setup call to the gRPC framework to setup stream.
	stream, err := client.GetCurrencyStream(context.Background(), curReq)

	// errors, at this point, are most likely communication errors
	if err != nil {
		log.Fatal("error in printEUR:", err)
	}

	fmt.Println("\nEUR Countries")
	fmt.Println("-------------")

	for {
		// stream.Rcvd() is where server.GetCurrencyStream() is
		// actually called, handle your error here.
		cur, err := stream.Recv()

		// errors at this point could be framework-created
		// or could be business logic errors from server.GetCurrencyStream()
		// each type can be handled by looking at the the err type and status codes
		if err != nil {
			if err == io.EOF {
				break // we're done
			}

			// determine err based on status code
			// to take appropriate action.
			if stat, ok := status.FromError(err); ok {
				switch stat.Code() {
				case codes.InvalidArgument:
					fmt.Println("error in printEUR:", err)
					return
				default:
					// other err type, do something with it
					fmt.Println(err)
					continue
				}
			}
		}
		fmt.Printf("%-50s%-10s\n", cur.GetCountry(), cur.GetCode())
	}
}

// addCurrencies demonstrates client to server stream
func addCurrencies(client pb.CurrencyServiceClient) {
	currencies := []*pb.Currency{
		&pb.Currency{Country: "HAITI", Name: "Gourde", Code: "HTG", Number: 332},
		&pb.Currency{Country: "MARTINIQUE", Name: "Euro", Code: "EUR", Number: 978},
		&pb.Currency{Country: "UNKNOWN", Number: -1}, // bad data
		&pb.Currency{Country: "CUBA", Name: "Cuban Peso", Code: "CUP", Number: 192},
		&pb.Currency{Country: "JAMAICA", Name: "Jamaican Dollar", Code: "JMD", Number: 388},
	}

	// setup server stream (remember, not calling server.SaveCurrencyStream yet)
	stream, err := client.SaveCurrencyStream(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// stream Currency values to the server.
	// The streamed data is not yet being saved.
	for _, cur := range currencies {
		if err := stream.Send(cur); err != nil {
			fmt.Println(err)
			return
		}
	}

	// CloseAndSave actually ends up calling server.SaveCurrencyStream
	// and any data/validation error returned can only be handled at this point.
	curList, err := stream.CloseAndRecv()

	// if err != nil (curList = nil), we can handle err here.
	// this example shows how to handle errors with structured
	// details in them.  This error extract the Currency value
	// that caused the error.
	if err != nil {
		if stat, ok := status.FromError(err); ok {
			switch stat.Code() {
			case codes.InvalidArgument:
				// see if detail object sent, if so assert to Currency
				if len(stat.Details()) > 0 {
					detail := stat.Details()[0]
					if cur, ok := detail.(*pb.Currency); ok {
						fmt.Printf("error in addCurrencies: %v [Detail: currency=%v]\n", err, cur)
					}
					return
				}
			default:
				// handle other errors here
				fmt.Println(err)
				return
			}
		}
	}

	fmt.Println("\nSaved currencies")
	fmt.Println("-----------------")
	for _, cur := range curList.Items {
		fmt.Printf("%-50s%-10s\n", cur.GetCountry(), cur.GetCode())
	}
}

// findCurrencies demonstrates bi-directional stream: one direction streams
// requests to the server while receiving replies from the server.
func findCurrencies(client pb.CurrencyServiceClient) {
	reqs := []*pb.CurrencyRequest{
		&pb.CurrencyRequest{Code: "CDF"},
		&pb.CurrencyRequest{Code: "AZN"},
		&pb.CurrencyRequest{Number: 392},
		&pb.CurrencyRequest{}, // bad request
		&pb.CurrencyRequest{Number: 949},
	}

	// setup stream
	stream, err := client.FindCurrencyStream(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// goroutine to stream outbound requests to server
	go func() {
		for _, req := range reqs {
			if err := stream.Send(req); err != nil {
				log.Fatal(err)
			}
		}
		if err := stream.CloseSend(); err != nil {
			log.Fatal(err)
		}
	}()

	// handle incoming Currency reponses from stream
	fmt.Println("\nFound Currencies")
	fmt.Println("-----------------")
	for {
		cur, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			if stat, ok := status.FromError(err); ok {
				switch stat.Code() {
				case codes.InvalidArgument:
					fmt.Println("error in findCurrencies:", err)
					return
				default:
					// other err type, do something with it
					fmt.Println(err)
					continue
				}
			}
		}
		fmt.Printf("%-50s%-10s\n", cur.GetCountry(), cur.GetCode())
	}
}

func main() {
	serverAddr := net.JoinHostPort(server, serverPort)

	// setup insecure connection
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	client := pb.NewCurrencyServiceClient(conn)

	printUSD(client)

	printEUR(client)

	addCurrencies(client)

	findCurrencies(client)
}
