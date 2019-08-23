package main

import (
	"fmt"
	"strconv"

	"github.com/chaincode/model"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	peer "github.com/hyperledger/fabric/protos/peer"
)

func (t *SmartContract) getFeeSum(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		logger.Error("getFeeSum Incorrect arguments. Expecting 1 arguments")
		return shim.Error("Incorrect arguments. Expecting 1 arguments")
	}

	name := args[0]

	deltaResultsIterator, deltaErr := stub.GetStateByPartialCompositeKey("ExchangeFee", []string{name})
	if deltaErr != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve value for %s: %s", name, deltaErr.Error()))
	}
	defer deltaResultsIterator.Close()

	if !deltaResultsIterator.HasNext() {
		return shim.Error(fmt.Sprintf("No variable by the name %s exists", name))
	}

	var amountSum float64
	var i int
	for i = 0; deltaResultsIterator.HasNext(); i++ {

		responseRange, nextErr := deltaResultsIterator.Next()
		if nextErr != nil {
			return shim.Error(nextErr.Error())
		}

		_, keyParts, splitKeyErr := stub.SplitCompositeKey(responseRange.Key)
		if splitKeyErr != nil {
			return shim.Error(splitKeyErr.Error())
		}

		amountStr := keyParts[2]

		amount, convErr := strconv.ParseFloat(amountStr, 64)
		if convErr != nil {
			return shim.Error(convErr.Error())
		}

		amountSum += amount

	}

	return shim.Success([]byte(strconv.FormatFloat(amountSum, 'f', -1, 64)))
}

func (t *SmartContract) pruneFastFeeSum(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		logger.Error("pruneFastFeeSum Incorrect arguments. Expecting 1 arguments")
		return shim.Error("Incorrect arguments. Expecting 1 arguments")
	}

	name := args[0]

	deltaResultsIterator, deltaErr := stub.GetStateByPartialCompositeKey("ExchangeFee", []string{name})
	if deltaErr != nil {
		//return shim.Error(fmt.Sprintf("Could not retrieve value for %s: %s", name, deltaErr.Error()))
		return shim.Success(nil)
	}
	defer deltaResultsIterator.Close()

	if !deltaResultsIterator.HasNext() {
		//return shim.Error(fmt.Sprintf("No variable by the name %s exists", name))
		return shim.Success(nil)
	}

	var amountSum float64
	var amountSumExceptPrune float64
	var i int
	amountSum = 0
	amountSumExceptPrune = 0
	for i = 0; deltaResultsIterator.HasNext(); i++ {

		responseRange, nextErr := deltaResultsIterator.Next()
		if nextErr != nil {
			return shim.Error(nextErr.Error())
		}

		_, keyParts, splitKeyErr := stub.SplitCompositeKey(responseRange.Key)
		if splitKeyErr != nil {
			return shim.Error(splitKeyErr.Error())
		}
		logger.Infof("keyParts[1] : %s", keyParts[1])
		amountStr := keyParts[2]

		amount, convErr := strconv.ParseFloat(amountStr, 64)
		if convErr != nil {
			return shim.Error(convErr.Error())
		}

		deltaRowDelErr := stub.DelState(responseRange.Key)
		if deltaRowDelErr != nil {
			return shim.Error(fmt.Sprintf("Could not delete delta row: %s", deltaRowDelErr.Error()))
		}
		amountSum += amount
		if keyParts[1] != "exchange_platform_prune" {
			amountSumExceptPrune += amount
		}
	}

	if amountSum == 0 {
		return shim.Success(nil)
	}

	if amountSumExceptPrune != 0 {
		err := t.creditFee(stub, name+"_prune", amountSum)
		if err == nil {

			memberAccountBytes, err := t.getMemberAccount(stub, "elmo")
			if err != nil {
				return shim.Error(err.Error())
			}

			if memberAccountBytes == nil {
				return shim.Error("Account not found")
			}

			account := new(model.MemberAccount)
			bytesToStruct(memberAccountBytes, account)

			t.creditCoin(stub, account, amountSumExceptPrune)

			logger.Infof("Successfully pruned variable %s, final value is %f, %d rows pruned", args[0], amountSum, i)
			return shim.Success([]byte(fmt.Sprintf("Successfully pruned variable %s, final value is %f, %d rows pruned", args[0], amountSum, i)))
		}
		logger.Infof("Failed to prune variable: all rows deleted but could not update value to %f, variable no longer exists in ledger", amountSum)
		return shim.Error(fmt.Sprintf("Failed to prune variable: all rows deleted but could not update value to %f, variable no longer exists in ledger", amountSum))
	}
	return shim.Success(nil)
}

func (t *SmartContract) pruneSafeFeeSum(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		logger.Error("pruneSafeFeeSum Incorrect arguments. Expecting 1 arguments")
		return shim.Error("Incorrect arguments. Expecting 1 arguments")
	}

	name := args[0]

	getResp := t.getFeeSum(stub, args)
	if getResp.Status == ERROR {
		return shim.Error(fmt.Sprintf("Could not retrieve the value of %s before pruning, pruning aborted: %s", name, getResp.Message))
	}

	valueStr := string(getResp.Payload)
	val, convErr := strconv.ParseFloat(valueStr, 64)
	if convErr != nil {
		return shim.Error(fmt.Sprintf("Could not convert the value of %s to a number before pruning, pruning aborted: %s", name, convErr.Error()))
	}

	backupPutErr := stub.PutState(fmt.Sprintf("%s_PRUNE_BACKUP", name), []byte(valueStr))
	if backupPutErr != nil {
		return shim.Error(fmt.Sprintf("Could not backup the value of %s before pruning, pruning aborted: %s", name, backupPutErr.Error()))
	}

	deltaResultsIterator, deltaErr := stub.GetStateByPartialCompositeKey("ExchangeFee", []string{name})
	if deltaErr != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve value for %s: %s", name, deltaErr.Error()))
	}
	defer deltaResultsIterator.Close()

	var i int
	for i = 0; deltaResultsIterator.HasNext(); i++ {
		responseRange, nextErr := deltaResultsIterator.Next()
		if nextErr != nil {
			return shim.Error(fmt.Sprintf("Could not retrieve next row for pruning: %s", nextErr.Error()))
		}

		deltaRowDelErr := stub.DelState(responseRange.Key)
		if deltaRowDelErr != nil {
			return shim.Error(fmt.Sprintf("Could not delete delta row: %s", deltaRowDelErr.Error()))
		}
	}
	amount, convErr := strconv.ParseFloat(valueStr, 64)
	if convErr != nil {
		return shim.Error(convErr.Error())
	}

	err := t.creditFee(stub, name+"_prune", amount)
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not insert the final value of the variable after pruning, variable backup is stored in %s_PRUNE_BACKUP: %s", name, err.Error()))
	}

	delErr := stub.DelState(fmt.Sprintf("%s_PRUNE_BACKUP", name))
	if delErr != nil {
		return shim.Error(fmt.Sprintf("Could not delete backup value %s_PRUNE_BACKUP, this does not affect the ledger but should be removed manually", name))
	}

	return shim.Success([]byte(fmt.Sprintf("Successfully pruned variable %s, final value is %f, %d rows pruned", name, val, i)))
}

func (t *SmartContract) deleteFee(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		logger.Error("deleteFee Incorrect arguments. Expecting 1 arguments")
		return shim.Error("Incorrect arguments. Expecting 1 arguments")
	}

	name := args[0]

	deltaResultsIterator, deltaErr := stub.GetStateByPartialCompositeKey("ExchangeFee", []string{name})
	if deltaErr != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve delta rows for %s: %s", name, deltaErr.Error()))
	}
	defer deltaResultsIterator.Close()

	if !deltaResultsIterator.HasNext() {
		return shim.Error(fmt.Sprintf("No variable by the name %s exists", name))
	}

	var i int
	for i = 0; deltaResultsIterator.HasNext(); i++ {
		responseRange, nextErr := deltaResultsIterator.Next()
		if nextErr != nil {
			return shim.Error(fmt.Sprintf("Could not retrieve next delta row: %s", nextErr.Error()))
		}

		deltaRowDelErr := stub.DelState(responseRange.Key)
		if deltaRowDelErr != nil {
			return shim.Error(fmt.Sprintf("Could not delete delta row: %s", deltaRowDelErr.Error()))
		}
	}

	return shim.Success([]byte(fmt.Sprintf("Deleted %s, %d rows removed", name, i)))
}
