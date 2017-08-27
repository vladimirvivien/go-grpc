package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/vladimirvivien/go-grpc/proto-encoding/curproto"
)

const fileName = "data.js"

func main() {
	items, err := createJsonFromCsv("../curdata.csv")
	if err != nil {
		log.Fatalln(err)
	}
	// print on screen
	for i, item := range items {
		fmt.Printf("%-25s%-20s\n", item.Name, item.Code)
		if i > 10 {
			fmt.Println("...")
			break
		}
	}
	// encode as protobuf data
	data, err := json.Marshal(items)
	if err != nil {
		log.Fatalln(err)
	}

	// save to file
	if err := ioutil.WriteFile(fileName, data, 0644); err != nil {
		log.Fatalln(err)
	}
	log.Println("data file saved as", fileName)
}

// createJsonFromCsv creates []Currency from CSV file
func createJsonFromCsv(path string) ([]curproto.Currency, error) {
	table := make([]curproto.Currency, 0)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		var num int32
		if i, err := strconv.Atoi(row[3]); err == nil {
			num = int32(i)
		}
		c := curproto.Currency{
			Country: row[0],
			Name:    row[1],
			Code:    row[2],
			Number:  num,
		}
		table = append(table, c)
	}
	return table, nil
}
