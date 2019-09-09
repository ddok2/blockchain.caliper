package model

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"
)

const TxRemittanceObjectType = "TxRemittance"
const TxFeeObjectType = "TxFee"
const TxRemittanceLimitObjectType = "TxRemittanceLimit"
const TxFreezeMemberAccountObjectType = "TxFreezeMemberAccount"
const TxIssueCoinObjectType = "TxIssueCoin"

type TXID string

type TxResultCode string
type TxType string
type TxFlag string

type TxFee struct {
	WalletAddress string            `json:"wallet_address"`
	Fee           float64           `json:"fee"`
	CreatedTime   int64             `json:"created_time"`
	Type          TxType            `json:"tx_status"`
	Params        map[string]string `json:"params"`
}

const (
	TxSuccess           TxResultCode = "tx_success"
	E01                 TxResultCode = "incorrect_arguments"
	E02                 TxResultCode = "insufficient_coin"
	InSufficientCash    TxResultCode = "insufficient_cash"
	MemberAccountfrozen TxResultCode = "member_account_frozen"
	UnexpectedFailure   TxResultCode = "unexpected_failure"
)

const (
	TxSellCoin     string = "sell_coin"
	TxBuyCoin      string = "buy_coin"
	TxTransferCoin string = "transfer_coin"
	TxTransferCash string = "transfer_cash"
	//TxFee          		string = "fee"
	TxCanceled            string = "canceled"
	TxFreezeMember        string = "freeze_member"
	TxRecoverFrozenMember string = "recover_frozen_member"
)

type TransactionHdr struct {
	Entity
	TxID        TXID         `json:"tx_id"`
	ResultCode  TxResultCode `json:"result_code"`
	CreatedTime int64        `json:"created_time"`
}

type TxRemmittanceLog struct {
	TransactionHdr
	Remittance
}

type TxFeeLog struct {
	TransactionHdr
	TxFee
}

type TxRemittanceLimitLog struct {
	TransactionHdr
	TxRemittanceLimit
}

type TxFreezeMemberAccountLog struct {
	TransactionHdr
	WalletAddress string `json:"wallet_address"`
	Type          string `json:"tx_type"`
}

type TxIssueCoinLog struct {
	TransactionHdr
	WalletAddress string  `json:"wallet_address"`
	Amount        float64 `json:"amount"`
}

func CreateTxRemittanceLimitLog(txID TXID, t *TxRemittanceLimit, resultCode TxResultCode) *TxRemittanceLimitLog {
	tx := &TxRemittanceLimitLog{TransactionHdr: TransactionHdr{Entity: Entity{TxRemittanceLimitObjectType},
		TxID: txID, ResultCode: resultCode, CreatedTime: time.Now().Unix()}}

	tx.TxRemittanceLimit = *t
	// tx.TxRemittanceLimit = TxRemittanceLimit{
	// 	Level: t.Level,
	// 	OneTimeRemittanceLimit: t.OneTimeRemittanceLimit,
	// 	OneTimeWithdrawLimit:   t.OneTimeWithdrawLimit,
	// 	OneDayRemittanceLimit:  t.OneDayRemittanceLimit,
	// 	OneDayWithdrawLimit:    t.OneDayWithdrawLimit,
	// }

	return tx
}

func CreateTxRemittanceLog(txID TXID, r *Remittance, resultCode TxResultCode) *TxRemmittanceLog {

	tx := &TxRemmittanceLog{TransactionHdr: TransactionHdr{Entity: Entity{TxRemittanceObjectType},
		TxID: txID, ResultCode: resultCode, CreatedTime: time.Now().Unix()}}

	tx.Remittance = *r

	// txBytes, _ := json.Marshal(tx)
	// tx.TxID = TXID(fmt.Sprintf("%x", GenTxID(txBytes)))

	return tx
}

func CreateTxFreezeMemberAccountLog(txID TXID, walletAddress string, txType string, resultCode TxResultCode) *TxFreezeMemberAccountLog {
	tx := &TxFreezeMemberAccountLog{TransactionHdr: TransactionHdr{Entity: Entity{TxFreezeMemberAccountObjectType},
		TxID: txID, ResultCode: resultCode, CreatedTime: time.Now().Unix()}}

	tx.WalletAddress = walletAddress
	tx.Type = txType

	return tx
}

func CreateTxIssueCoinLog(txID TXID, walletAddress string, amount float64, resultCode TxResultCode) *TxIssueCoinLog {
	tx := &TxIssueCoinLog{TransactionHdr: TransactionHdr{Entity: Entity{TxIssueCoinObjectType},
		TxID: txID, ResultCode: resultCode, CreatedTime: time.Now().Unix()}}

	tx.WalletAddress = walletAddress
	tx.Amount = amount

	return tx
}

func CreateTxFeeLog(walletAddress string, fee float64, resultCode TxResultCode) *TxFeeLog {

	tx := &TxFeeLog{TransactionHdr: TransactionHdr{Entity: Entity{TxFeeObjectType},
		ResultCode: resultCode, CreatedTime: time.Now().Unix()}}

	tx.TxFee = TxFee{
		WalletAddress: walletAddress,
		Fee:           fee,
	}

	txBytes, _ := json.Marshal(tx)
	tx.TransactionHdr.TxID = TXID(fmt.Sprintf("%x", GenTxID(txBytes)))

	return tx
}

func GenTxID(data []byte) []byte {
	md5 := md5.New()
	md5.Write(data)
	return md5.Sum(nil)
}
