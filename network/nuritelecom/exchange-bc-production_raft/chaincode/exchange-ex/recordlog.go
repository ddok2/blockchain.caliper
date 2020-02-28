package main

import (
	"encoding/json"

	"github.com/chaincode/model"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func (t *SmartContract) recordTxRemittanceLimitLog(stub shim.ChaincodeStubInterface, txID model.TXID, tl *model.TxRemittanceLimit,
	resultCode model.TxResultCode) string {

	tx := model.CreateTxRemittanceLimitLog(txID, tl, resultCode)
	txBytes, err := json.Marshal(tx)
	if err != nil {
		return MarshallingError
	}

	key, _ := t.createCompositeKey(tx.GetObjectType(), []string{string(txID)})
	logger.Infof("key : %s, tx : %v", key, tx)
	err = stub.PutState(key, txBytes)
	if err != nil {
		logger.Infof("PutState Error : %s", err.Error())
		return InternalServerError
	}

	return OK
}

func (t *SmartContract) recordTxRemittanceLog(stub shim.ChaincodeStubInterface, txID model.TXID, r *model.Remittance,
	resultCode model.TxResultCode) string {

	tx := model.CreateTxRemittanceLog(txID, r, resultCode)
	txBytes, err := json.Marshal(tx)
	if err != nil {
		return MarshallingError
	}

	key, _ := t.createCompositeKey(tx.GetObjectType(), []string{r.SenderWalletAddress, r.ReceiverWalletAddress, string(tx.TxID)})
	logger.Infof("key : %s, tx : %v", key, tx)
	err = stub.PutState(key, txBytes)
	if err != nil {
		logger.Infof("PutState Error : %s", err.Error())
		return InternalServerError
	}

	return OK
}

func (t *SmartContract) recordTxFreezeMemberAccountLog(stub shim.ChaincodeStubInterface, txID model.TXID, walletAddress string,
	txType string, resultCode model.TxResultCode) string {

	tx := model.CreateTxFreezeMemberAccountLog(txID, walletAddress, txType, resultCode)
	txBytes, err := json.Marshal(tx)
	if err != nil {
		return MarshallingError
	}

	key, _ := t.createCompositeKey(tx.GetObjectType(), []string{walletAddress, string(tx.TxID)})
	logger.Infof("key : %s, tx : %v", key, tx)
	err = stub.PutState(key, txBytes)
	if err != nil {
		logger.Infof("PutState Error : %s", err.Error())
		return InternalServerError
	}

	return OK
}

func (t *SmartContract) recordTxIssueCoinLog(stub shim.ChaincodeStubInterface, txID model.TXID, walletAddress string,
	amount float64, resultCode model.TxResultCode) string {

	tx := model.CreateTxIssueCoinLog(txID, walletAddress, amount, resultCode)
	txBytes, err := json.Marshal(tx)
	if err != nil {
		return MarshallingError
	}

	key, _ := t.createCompositeKey(tx.GetObjectType(), []string{walletAddress, string(tx.TxID)})
	logger.Infof("key : %s, tx : %v", key, tx)
	err = stub.PutState(key, txBytes)
	if err != nil {
		logger.Infof("PutState Error : %s", err.Error())
		return InternalServerError
	}

	return OK
}

func (t *SmartContract) recordTxFeeLog(stub shim.ChaincodeStubInterface, txID model.TXID, walletAddress string,
	fee float64, resultCode model.TxResultCode) string {

	tx := model.CreateTxFeeLog(walletAddress, fee, resultCode)
	txBytes, err := json.Marshal(tx)
	if err != nil {
		return MarshallingError
	}

	key, _ := t.createCompositeKey(tx.GetObjectType(), []string{walletAddress, string(tx.TxID)})
	logger.Infof("key : %s, tx : %v", key, tx)
	err = stub.PutState(key, txBytes)
	if err != nil {
		logger.Infof("PutState Error : %s", err.Error())
		return InternalServerError
	}

	return OK
}
