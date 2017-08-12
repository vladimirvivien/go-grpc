package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

type Currency struct {
	Code    string `json:"currency_code"`
	Name    string `json:"currency_name"`
	Number  int32  `json:"currency_number"`
	Country string `json:"currency_country"`
}

func main() {
	items, err := createJsonFromCsv("./curdata.csv")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// print on screen
	for i, item := range items {
		fmt.Printf("%-25s%-20s\n", item.Name, item.Code)
		if i > 50 {
			fmt.Println("...")
			break
		}
	}
	// encode as protobuf data
	data, err := json.Marshal(items)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// save to file
	if err := ioutil.WriteFile("./curdata.json", data, 0644); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("data saved protobuf")
}

// createJsonFromCsv creates []Currency from CSV file
func createJsonFromCsv(path string) ([]Currency, error) {
	table := make([]Currency, 0)
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
		c := Currency{
			Country: row[0],
			Name:    row[1],
			Code:    row[2],
			Number:  num,
		}
		table = append(table, c)
	}
	return table, nil
}
