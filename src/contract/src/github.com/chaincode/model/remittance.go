package model

import (
	"errors"
	"fmt"
)

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

type Remittance struct {
	SenderWalletAddress   string            `json:"sender_wallet_address"`
	ReceiverWalletAddress string            `json:"receiver_wallet_address"`
	Amount                float64           `json:"amount"`
	Fee                   float64           `json:"fee"`
	TxFlag                string            `json:"tx_flag"`
	Params                map[string]string `json:"params"`
}

func (r *Remittance) Validate() error {

	txFlag := []string{"1", "2", "3", "4"}

	if r.SenderWalletAddress == "" {
		return errors.New("Missing required SenderWalletAddress")
	}
	if r.ReceiverWalletAddress == "" {
		return errors.New("Missing required ReceiverWalletAddress")
	}
	if r.Amount <= 0 {
		return fmt.Errorf("Invalid Amount %f", r.Amount)
	}
	if r.Fee < 0 {
		return fmt.Errorf("Invalid Fee %f", r.Fee)
	}
	if !Contains(txFlag, r.TxFlag) {
		return fmt.Errorf("Invalid TxFlag %f", r.TxFlag)
	}
	return nil
}
