package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

type Address struct {
	TxId     string `json:"tx_id"`
	Address  string `json:"address"`
	Quantity string `json:"quantity"`
	Token    string `json:"token"`
}

func ReadFile() (array []*Address, err error) {
	
	//  Open CSV File
	file, err := os.Open("account.csv")
	if err != nil {
		
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	
	r := csv.NewReader(file)
	
	if _, err = r.Read(); err != nil {
		
		fmt.Println("Error reading header:", err)
		return
	}
	
	var addresses []*Address
	
	// Read File
	for {
		var record []string
		record, err = r.Read()
		if err == io.EOF {
			
			err = nil
			break // End
		}
		
		if err != nil {
			fmt.Println("Error reading record:", err)
			return
		}
		
		//  Make Record
		address := &Address{
			TxId:     record[0],
			Address:  record[1],
			Quantity: record[2],
			Token:    record[3],
		}
		
		addresses = append(addresses, address)
	}
	
	array = addresses
	return
}
