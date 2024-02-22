package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/msig"
	"github.com/eoscanada/eos-go/sudo"
	"io"
	"net/http"
	"strings"
)

var (
	DfuseKey = "6fe371219e7a2ea1594dec6c1f86b869" // API key for Dfuse service
	EndPoint = "https://eos.greymass.com"
	// EndPoint = "https://jungle4.cryptolions.io" // Endpoint for EOS blockchain
)

// TokenResponse structure for receiving JSON response
type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// GetDfuseToken retrieves an authentication token for Dfuse API
func GetDfuseToken() (token string, err error) {
	// Prepare request body data in struct format
	data := struct {
		APIKey string `json:"api_key"`
	}{
		APIKey: DfuseKey,
	}
	
	// Encode data to JSON format
	payload, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}
	
	// Send POST request
	url := "https://auth.eosnation.io/v1/auth/issue"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	
	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	
	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}
	
	// Parse JSON response
	var tokenResp TokenResponse
	if err = json.Unmarshal(body, &tokenResp); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}
	
	token = tokenResp.Token
	return
}

// Authorization matches the structure of the authorization field
type Authorization struct {
	Actor      string `json:"actor"`
	Permission string `json:"permission"`
}

// Action matches the structure of elements in the actions array
type Action struct {
	Account       string          `json:"account"`
	Name          string          `json:"name"`
	Authorization []Authorization `json:"authorization"`
	HexData       string          `json:"hex_data"`
}

// Transaction matches the structure of the transaction field
type Transaction struct {
	Expiration         string    `json:"expiration"`
	RefBlockNum        int       `json:"ref_block_num"`
	RefBlockPrefix     int       `json:"ref_block_prefix"`
	MaxNetUsageWords   int       `json:"max_net_usage_words"`
	MaxCPUUsageMs      int       `json:"max_cpu_usage_ms"`
	DelaySec           int       `json:"delay_sec"`
	ContextFreeActions []int     `json:"context_free_actions"` // Only for example, actual type may be complex
	Actions            []*Action `json:"actions"`
}

// Response matches the structure of the entire response
type Response struct {
	TransactionStatus string       `json:"transaction_status"`
	ID                string       `json:"id"`
	Transaction       *Transaction `json:"transaction"`
}

// TransferData represents the data structure for a transfer action
type TransferData struct {
	From     eos.AccountName
	To       eos.AccountName
	Quantity eos.Asset
	Memo     string
}

// GetTransactionByDfuse retrieves transaction details using Dfuse API
func GetTransactionByDfuse(txId string) (actions []*TransferData, err error) {
	token, err := GetDfuseToken()
	if err != nil {
		fmt.Println("Error getting token:", err)
		return
	}
	
	url := "https://eos.dfuse.eosnation.io/v0/transactions/" + txId
	method := "POST"
	
	payload := strings.NewReader(``)
	
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Authorization", "Bearer "+token)
	
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	
	var resp Response
	if err = json.Unmarshal(body, &resp); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}
	
	// Process actions in the transaction
	for _, action := range resp.Transaction.Actions {
		if action.Account != "eosio.token" {
			continue
		}
		
		if action.Name != "transfer" {
			continue
		}
		
		// Decode action data
		data, e := decodeActionData(action.HexData)
		if e != nil {
			continue
		}
		
		actions = append(actions, data)
	}
	
	return
}

// decodeActionData decodes the hex-encoded action data into TransferData struct
func decodeActionData(hexData string) (data *TransferData, err error) {
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
	
	data = &actionData
	return
}

// GetProposal retrieves the proposal details from the eosio.msig contract
func GetProposal(acc, name string) (proposal *msig.ProposalRow, err error) {
	api := eos.New(EndPoint)
	
	resp, err := api.GetTableRows(context.TODO(), eos.GetTableRowsRequest{
		Code:       "eosio.msig",
		Scope:      acc,
		Table:      "proposal",
		LowerBound: name,
		JSON:       true,
	})
	
	if err != nil {
		return
	}
	
	var proposals []msig.ProposalRow
	if err = resp.JSONToStructs(&proposals); err != nil {
		return
	}
	
	for _, p := range proposals {
		
		if p.ProposalName.String() == name {
			
			proposal = &p
			break
		}
	}
	
	if proposal == nil {
		
		err = errors.New("NOT FOUND")
	}
	return
}

// GetL1ProposalActions retrieves actions from Level 1 proposals
func GetL1ProposalActions(acc, name string) (array []*msig.Approve, err error) {
	proposal, err := GetProposal(acc, name)
	if err != nil {
		return
	}
	
	var tx eos.Transaction
	if err = eos.UnmarshalBinary(proposal.PackedTransaction, &tx); err != nil {
		
		fmt.Println(proposal, " -> ", err)
		return
	}
	
	for _, action := range tx.Actions {
		var approvalAction msig.Approve
		if err = eos.UnmarshalBinary(action.HexData, &approvalAction); err != nil {
			
			fmt.Println(proposal, "exec action -> ", err)
			continue
		}
		
		// if approvalAction.Level.Actor == "eosio" {
		
		array = append(array, &approvalAction)
		// }
	}
	
	return
}

// EVMAdminAction represents the structure for EVM admin actions
type EVMAdminAction struct {
	From     eos.HexBytes `json:"from"`
	To       eos.HexBytes `json:"to"`
	Value    eos.HexBytes `json:"value"`
	Data     eos.HexBytes `json:"data"`
	GasLimit uint64       `json:"gas_limit"`
}

// GetL2ProposalActions retrieves actions from Level 2 proposals
func GetL2ProposalActions(acc, name string, useSuper bool) (array []*EVMAdminAction, err error) {
	proposal, err := GetProposal(acc, name)
	if err != nil {
		return
	}
	
	var tx eos.Transaction
	if err = eos.UnmarshalBinary(proposal.PackedTransaction, &tx); err != nil {
		
		fmt.Println(proposal, " -> ", err)
		return
	}
	
	if useSuper {
		
		for _, action := range tx.Actions {
			var execAction sudo.Exec
			
			if err = eos.UnmarshalBinary(action.HexData, &execAction); err != nil {
				
				fmt.Println(proposal, "exec action -> ", err)
				break
			}
			
			for _, rawAction := range execAction.Transaction.Actions {
				var adminAction EVMAdminAction
				if err = eos.UnmarshalBinary(rawAction.HexData, &adminAction); err != nil {
					
					fmt.Println(proposal, "admin action -> ", err)
					break
				}
				
				array = append(array, &adminAction)
			}
		}
		
		return
	}
	
	for _, action := range tx.Actions {
		
		var adminAction EVMAdminAction
		if err = eos.UnmarshalBinary(action.HexData, &adminAction); err != nil {
			
			fmt.Println(proposal, "admin action -> ", err)
			break
		}
		
		array = append(array, &adminAction)
	}
	
	return
}
