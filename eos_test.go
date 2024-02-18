package main

import (
	"encoding/hex"
	"fmt"
	"github.com/eoscanada/eos-go"
	"testing"
)

func Test_Decode(t *testing.T) {
	
	hexData := "d055d5eab4ccf89d0000905b01ea3055106c33000000000004454f53000000002a307861414437616243663430313442343032353666336235623964433637443265376435653332303731"
	// data := eos.NewActionDataFromHexData([]byte(hexData))
	
	rawData, err := hex.DecodeString(hexData)
	if err != nil {
		
		fmt.Println("hex decode error -> ", err)
		return
	}
	
	var actionData TransferData
	if err = eos.NewDecoder(rawData).Decode(&actionData); err != nil {
		
		fmt.Println("decode error -> ", err)
		return
	}
	
	fmt.Println(actionData.Memo)
}

func Test_DFuseGetTransaction(t *testing.T) {
	
	txId := "d1e170f54c4a87926231a1ff90d5f007c565fe923e56888caada7de1fb511a8a"
	array, err := GetTransactionByDfuse(txId)
	if err != nil {
		
		fmt.Println(err)
		return
	}
	
	fmt.Println(len(array))
}

func Test_GetL2Proposal(t *testing.T) {
	
	array, err := GetL2ProposalActions("evmgreatagai", "3324b3545155")
	if err != nil {
		
		fmt.Println(err)
		return
	}
	
	fmt.Println(len(array))
}

func Test_GetL1Proposal(t *testing.T) {
	
	array, err := GetL1ProposalActions("evmgreatagai", "c5fc2e45f4b4")
	if err != nil {
		
		fmt.Println(err)
		return
	}
	
	fmt.Println(len(array))
}
