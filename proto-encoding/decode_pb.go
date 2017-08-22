package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/vladimirvivien/go-grpc/proto-encoding/curproto"
)

const fileName = "data.pb"

func main() {
	// load protobuf file
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("file %s missing, run encode_pb.go first.\n", fileName)
		} else {
			log.Fatalln(err)
		}
	}

	// unmarshal into curList
	curList := new(curproto.CurrencyList)
	if err := proto.Unmarshal(data, curList); err != nil {
		log.Fatalln(err)
	}

	// display protobuf data
	for i, item := range curList.Items {
		fmt.Printf("%-25s%-20s\n", item.Name, item.Code)
		if i > 10 {
			fmt.Println("...")
			break
		}
	}
	fmt.Println("data decoded successfully!")
}
