package main

import (
	"log"

	"permissions/contract"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	PermissionChaincode, err := contractapi.NewChaincode(&contract.SmartContract{})
	if err != nil {
		log.Panicf("Error: %v", err)
	}
	if err := PermissionChaincode.Start(); err != nil {
		log.Panicf("Error: %v", err)
	}
}