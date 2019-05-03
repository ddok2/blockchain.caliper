package model

import (
	"errors"
	"time"
)

const AccountObjecttype = "Account"

type TransactionLimit struct {
	Level                string  `json:"level"`
	OneTimeTransferLimit float64 `json:"one_time_transfer_limit"`
	OneTimeWithdrawLimit float64 `json:"one_time_withdraw_limit"`
	OneDayTransferLimit  float64 `json:"one_day_transfer_limit"`
	OneDayWithdrawLimit  float64 `json:"one_day_withdraw_limit"`
}

type TransactionLimitMap struct {
	TxLimit map[string]TransactionLimit
}

func NewTransactionLimit() *TransactionLimitMap {
	return &TransactionLimitMap{make(map[string]TransactionLimit)}
}

func (t *TransactionLimitMap) Add(level string, txLimit TransactionLimit) {
	t.TxLimit[level] = txLimit
}

type Wallet struct {
	WalletAddress string  `json:"wallet_address"`
	CoinBalance   float64 `json:"coin_balance"`
	CashBalance   float64 `json:"cash_balance"`
}

func (w *Wallet) Validate() error {
	if w.WalletAddress == "" {
		return errors.New("Missing required WalletAddress")
	}
	if w.CoinBalance < 0 {
		return errors.New("Invalid CoinBalance")
	}
	if w.CashBalance < 0 {
		return errors.New("Invalid CashBalance")
	}

	return nil
}

func (w *Wallet) DebitCoin(amount float64) {

	w.CoinBalance -= amount
}

func (w *Wallet) CreditCoin(amount float64) {
	w.CoinBalance += amount
}

func (w *Wallet) DebitCash(amount float64) {
	w.CashBalance -= amount
}

func (w *Wallet) CreditCash(amount float64) {
	w.CashBalance += amount
}

type TxDate struct {
	Year  int        `json:"year"`
	Month time.Month `json:"month" `
	Day   int        `json:day`
}

type MemberAccount struct {
	Entity
	MemberId                   string  `json:"member_id"`
	VSCode                     string  `json:"vs_code"`
	CountryCode                string  `json:"country_code"`
	CurrencyCode               string  `json:"currency_code"`
	MemberRole                 string  `json:"member_role"`
	MemberLevel                string  `json:"member_level"`
	CustomOneTimeTransferLimit float64 `json:"custom_one_time_transfer_limit"`
	CustomOneTimeWithdrawLimit float64 `json:"custom_one_time_withdraw_limit"`
	CustomOneDayTransferLimit  float64 `json:"custom_one_day_transfer_limit"`
	CustomOneDayWithdrawLimit  float64 `json:"custom_one_day_withdraw_limit"`
	OneDayTransferSum          float64 `json:"one_day_transfer_sum"`
	OneDayTransferDate         TxDate  `json:"one_day_transfer_date"`
	OneDayWithdrawSum          float64 `json:"one_day_withdraw_sum"`
	OneDayWithdrawDate         TxDate  `json:"one_day_withdraw_date"`
	MemberWallet               Wallet  `json:"member_wallet"`
	Frozen                     bool    `json:"frozen"`
	CreatedDate                string  `json:"create_date"`
	Deleted                    bool    `json:"deleted"`
	Description                string  `json:"description"`
}

type MemberAccountList struct {
	MemberAccounts []*MemberAccount `json:"member_accounts"`
}

func (ma *MemberAccount) Validate() error {
	if ma.MemberId == "" {
		return errors.New("Missing required MemberId")
	}
	if ma.VSCode == "" {
		return errors.New("Missing required VSCode")
	}
	if ma.CountryCode == "" {
		return errors.New("Missing required CountryCode")
	}
	if ma.CurrencyCode == "" {
		return errors.New("Missing required CurrencyCode")
	}
	if ma.MemberRole == "" {
		return errors.New("Missing required MemberRole")
	}
	if ma.CreatedDate == "" {
		return errors.New("Missing required CreatedDate")
	}

	if err := ma.MemberWallet.Validate(); err != nil {
		return err
	}

	return nil
}

func (ma *MemberAccount) DebitCoin(amount float64) {
	ma.MemberWallet.DebitCoin(amount)
}

func (ma *MemberAccount) CreditCoin(amount float64) {
	ma.MemberWallet.CreditCoin(amount)
}

func (ma *MemberAccount) DebitCash(amount float64) {
	ma.MemberWallet.DebitCash(amount)
}

func (ma *MemberAccount) CreditCash(amount float64) {
	ma.MemberWallet.CreditCash(amount)
}
