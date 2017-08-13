package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/vladimirvivien/go-grpc/currency_protobuf/curproto"
)

func main() {
	fname := "./curdata.pb"

	// load protobuf file
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("file %s missing, run encode_pb.go first.\n", fname)
		} else {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// unmarshal into curList
	curList := new(curproto.CurrencyList)
	if err := proto.Unmarshal(data, curList); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// display protobuf data
	for i, item := range curList.Items {
		fmt.Printf("%-25s%-20s\n", item.Name, item.Code)
		if i > 50 {
			fmt.Println("...")
			break
		}
	}
	fmt.Println("data decoded successfully!")
}
