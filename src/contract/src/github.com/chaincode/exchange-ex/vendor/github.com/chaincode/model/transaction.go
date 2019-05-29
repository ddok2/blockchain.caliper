package model

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"
)

const TransactionObjectType = "Transaction"
const TransactionFeeObjectType = "TransactionFee"

type TXID string

type TxFailureCode string
type TxStatus string

type TxDetails struct {
	SenderWalletAddress   string            `json:"sender_wallet_address"`
	ReceiverWalletAddress string            `json:"receiver_wallet_address"`
	Amount                float64           `json:"amount"`
	Fee                   float64           `json:"fee"`
	CreatedTime           int64             `json:"created_time"`
	TxStatus              TxStatus          `json:"tx_status"`
	Params                map[string]string `json:"params"`
}

type TxDetailsFee struct {
	WalletAddress string            `json:"sender_wallet_address"`
	Fee           float64           `json:"fee"`
	CreatedTime   int64             `json:"created_time"`
	TxStatus      TxStatus          `json:"tx_status"`
	Params        map[string]string `json:"params"`
}

const (
	TxSuccess           TxFailureCode = "tx_success"
	E01                 TxFailureCode = "incorrect_arguments"
	E02                 TxFailureCode = "insufficient_coin"
	InSufficientCash    TxFailureCode = "insufficient_cash"
	MemberAccountfrozen TxFailureCode = "member_account_frozen"
	UnexpectedFailure   TxFailureCode = "unexpected_failure"
)

const (
	TxSellCoin     TxStatus = "sell_coin"
	TxBuyCoin      TxStatus = "buy_coin"
	TxTransferCoin TxStatus = "transfer_coin"
	TxTransferCash TxStatus = "transfer_cash"
	TxFee          TxStatus = "fee"
	TxCanceled     TxStatus = "canceled"
)

type Transaction struct {
	Entity
	TxId TXID `json:"tx_id"`
	TxDetails
	FailureCode TxFailureCode `json:"failure_code"`
}

type TransactionFee struct {
	Entity
	TxId TXID `json:"tx_id"`
	TxDetailsFee
	FailureCode TxFailureCode `json:"failure_code"`
}

func CreateTransaction(r *Remittance, txStatus TxStatus, failureCode TxFailureCode) (*Transaction, error) {

	tx := &Transaction{Entity: Entity{TransactionObjectType}, FailureCode: failureCode}
	tx.TxDetails = TxDetails{

		SenderWalletAddress:   r.SenderWalletAddress,
		ReceiverWalletAddress: r.ReceiverWalletAddress,
		Amount:                r.Amount,
		Fee:                   r.Fee,
		CreatedTime:           time.Now().Unix(),
		TxStatus:              txStatus,
		Params:                r.Params,
	}

	txBytes, _ := json.Marshal(tx)
	tx.TxId = TXID(fmt.Sprintf("%x", GenTxId(txBytes)))

	return tx, nil
}

func CreateTransactionFee(walletAddress string, fee float64, txStatus TxStatus, failureCode TxFailureCode) (*TransactionFee, error) {

	tx := &TransactionFee{Entity: Entity{TransactionFeeObjectType}, FailureCode: failureCode}
	tx.TxDetailsFee = TxDetailsFee{

		WalletAddress: walletAddress,
		Fee:           fee,
		CreatedTime:   time.Now().Unix(),
		TxStatus:      txStatus,
	}

	txBytes, _ := json.Marshal(tx)
	tx.TxId = TXID(fmt.Sprintf("%x", GenTxId(txBytes)))

	return tx, nil
}

func GenTxId(data []byte) []byte {
	md5 := md5.New()
	md5.Write(data)
	return md5.Sum(nil)
}
