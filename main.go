package main

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// Define global variables, including the hacker's address, sender, proposal name, and the EVM address to be recovered
var (
	HACKER            = "nrwgthbeupex"
	Sender            = "smhgshaoahnf"                             // send proposal account
	ProposalName      = "faa2fc1f2ce3"                             // L1 ProposalName
	RecoverEVMAddress = "bbbbbbbbbbbbbbbbbbbbbbbb55300ba914daae00" // eosio.evm
)

func main() {
	// Check if the account is valid
	// if res := CheckAccount(); res {
	//
	// 	fmt.Println("Address Valid")
	// } else {
	//
	// 	fmt.Println("Address Invalid")
	// }
	
	// Check if the proposal is valid
	if res := CheckProposal(); res {
		
		fmt.Println("Proposal Valid")
	} else {
		
		fmt.Println("Proposal Invalid")
	}
}

// CheckAccount checks if the addresses in the CSV file are hacker addresses
// and if the actions in the proposals are 'eosio.evm' admincall functions with the hacker address as the sender
func CheckAccount() bool {
	// Read addresses from the file
	array, err := ReadFile()
	if err != nil {
		fmt.Println("Error reading file:", err)
		return false
	}
	
	// Group addresses by transaction ID
	txIds := make(map[string]map[string]*Address)
	for _, address := range array {
		if _, ok := txIds[address.TxId]; !ok {
			txIds[address.TxId] = make(map[string]*Address)
		}
		txIds[address.TxId][address.Address] = address
	}
	
	// Check each transaction
	checkedTxId := make([]string, 0)
	for txId, addresses := range txIds {
		actions, e := GetTransactionByDfuse(txId)
		if e != nil {
			fmt.Println(txId, " get transaction error -> ", e)
			break
		}
		
		for _, action := range actions {
			if action.From.String() != HACKER {
				
				continue
			}
			
			if action.To != "eosio.evm" {
				
				continue
			}
			
			if v, ok := addresses[action.Memo]; ok {
				
				// check balance
				balance := v.Quantity + ".0000 " + v.Token
				
				if action.Quantity.String() == balance {
					
					delete(addresses, action.Memo)
				} else {
					
					fmt.Println(action.Memo, " balance incorrect")
				}
				
			} else {
				
				fmt.Println(action.Memo, " invalid")
				break
			}
		}
		
		if len(addresses) == 0 {
			
			fmt.Println(txId, " -> checked")
			checkedTxId = append(checkedTxId, txId)
		} else {
			
			fmt.Println(txId, " -> not checked")
		}
	}
	
	if len(checkedTxId) == len(txIds) {
		return true
	} else {
		return false
	}
}

// CheckProposal checks if the proposals initiated have actions that are 'eosio.evm' admincall functions
// and the 'from' address in the actions is the hacker's address
func CheckProposal() bool {
	addresses, err := ReadFile()
	if err != nil {
		fmt.Println("Error reading file:", err)
		return false
	}
	
	tempAddresses := make(map[string]*Address)
	for _, address := range addresses {
		
		key := strings.Replace(address.Address, "0x", "", 1)
		tempAddresses[strings.ToLower(key)] = address
	}
	
	array, err := GetL1ProposalActions(Sender, ProposalName)
	if err != nil {
		fmt.Println("Get L1 proposal error -> ", err)
		return false
	}
	
	totalCount := 0
	for _, p := range array {
		var actions []*EVMAdminAction
		
		for {
			actions, err = GetL2ProposalActions(Sender, p.ProposalName.String(), false)
			if err != nil {
				
				fmt.Println("Get L2 proposal error -> ", err)
				err = nil
			} else {
				
				break
			}
			
		}
		
		for _, action := range actions {
			
			addr := tempAddresses[action.From.String()]
			
			if addr == nil {
				
				fmt.Println(action.From.String(), " not found")
				return false
			}
			
			if action.To.String() != RecoverEVMAddress {
				
				fmt.Println(action.To.String(), " receiver address incorrect")
				return false
			}
			
			// Convert the transfer quantity from byte slice to big.Int
			value := new(big.Int).SetBytes(action.Value)
			
			// Base precision: 1x10^17
			precision := new(big.Int)
			precision.SetString("100000000000000000", 10)
			
			// Calculate the real value by dividing by precision
			value.Div(value, precision)
			
			// Convert the balance from `account.csv` to an integer
			balance, e := strconv.ParseInt(addr.Quantity, 10, 64)
			if e != nil {
				
				fmt.Println(addr.Address, "balance error:", e)
				return false
			}
			
			// Adjust the balance to match the precision
			balance *= 10
			
			// Calculate the transfer quantity as an int64
			quantity := value.Int64()
			
			// Reserve 0.1 EOS for EVM gas fees and validate the quantity
			if balance > quantity && balance-quantity > 1 {
				
				fmt.Println(addr.Address, "incorrect transfer quantity")
				return false
			}
			
			totalCount++
			fmt.Println(action.From.String(), " checked -> total ", totalCount)
			delete(tempAddresses, action.From.String())
		}
	}
	
	if len(tempAddresses) == 0 {
		
		return true
	}
	
	return false
}
