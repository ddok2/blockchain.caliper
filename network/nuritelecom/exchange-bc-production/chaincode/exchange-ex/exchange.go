/*
 * Copyright 2019. Nuri Telecom. All Rights Reserved.
 *
 * - exchange.go
 * - author: Sungyub NA <mailto: syna@nuritelecom.com>
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chaincode/model"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	peer "github.com/hyperledger/fabric/protos/peer"
	"os"
	"strconv"
)

var (
	logger     = shim.NewLogger("exchange")
	handlerMap = NewHandlerMap()
)

const (
	ERROR = 500
)

const (
	TxStartIdx   = 1
	TxIdOffset   = 1
	TxArgsOffset = 2

	RegisterMemberID   = "1"
	TransferCoinID     = "2"
	GetBalanceID       = "3"
	RemittanceID       = "4"
	RegisterMemberArgs = 9
	TransferCoinArgs   = 9
	GetBalanceArgs     = 2
	RemittanceArgs     = 3
)

const (
	OK                                 = "200"
	Created                            = "201"
	BadRequest                         = "400"
	MethodNotAllowed                   = "405"
	AlreadyRegistered                  = "420"
	GetStateError                      = "421"
	InvalidArgument                    = "422"
	SenderAccountNotFound              = "430"
	SenderAccountFrozen                = "431"
	SenderAccountInsufficientBalance   = "432"
	ReceiverAccountNotFound            = "433"
	ReceiverAccountFrozen              = "434"
	ReceiverAccountInsufficientBalance = "435"
	AccountNotFound                    = "436"
	OneTimeRemittanceLimitError        = "440"
	OneDayRemittanceLimitError         = "441"
	CustomOneTimeRemittanceLimitError  = "442"
	CustomOneDayRemittanceLimitError   = "443"
	WalletBalanceLimitError            = "444"
	AlreadySetMemberLevel              = "450"
	AlreadyFrozenMember                = "451"
	DuplicateTxID                      = "452"
	IncorrectSenderBalance             = "460"
	IncorrectReceiverBalance           = "461"
	IncorrectBalance                   = "462"
	InternalServerError                = "500"
	MarshallingError                   = "501" // Error marshalling
)

type SmartContract struct{}

func main() {

	initLogging()
	logger.Info("Starting exchange chaincode")
	sc := new(SmartContract)
	sc.registerHandlers()

	err := shim.Start(sc)
	if err != nil {
		logger.Errorf("Error starting exchange chaincode : %s", err)
	}

}

func initLogging() {
	logger.SetLevel(shim.LogInfo)
	logLevel, _ := shim.LogLevel(os.Getenv("SHIM_LOGGING_LEVEL"))
	shim.SetLoggingLevel(logLevel)
}

func (t *SmartContract) registerHandlers() {

	handlerMap.Add("batchProcess", t.batchProcess)
	handlerMap.Add("registerMember", t.registerMember)
	handlerMap.Add("setTransactionLimit", t.setTransactionLimit)
	handlerMap.Add("getMemberAccountList", t.getMemberAccountList)
	handlerMap.Add("setMemberLevel", t.setMemberLevel)
	handlerMap.Add("freezeMemberAccount", t.freezeMemberAccount)
	handlerMap.Add("recoverFrozenMemberAccount", t.recoverFrozenMemberAccount)
	handlerMap.Add("issueCoin", t.issueCoin)
	handlerMap.Add("processOrder", t.processOrder)
	handlerMap.Add("transferCoin", t.transferCoin)
	handlerMap.Add("transferCash", t.transferCash)
	handlerMap.Add("chargeCoin", t.chargeCoin)
	handlerMap.Add("depositCash", t.depositCash)
	handlerMap.Add("withdrawCash", t.withdrawCash)
	handlerMap.Add("getFeeSum", t.getFeeSum)
	handlerMap.Add("pruneFastFeeSum", t.pruneFastFeeSum)
	handlerMap.Add("pruneSafeFeeSum", t.pruneSafeFeeSum)
	handlerMap.Add("deleteFee", t.deleteFee)
}

func (t *SmartContract) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (t *SmartContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	function, args := stub.GetFunctionAndParameters()
	logger.Infof("Invoke function : %s", function)
	return handlerMap.Handle(stub, function, args)
}

func (t *SmartContract) batchProcess(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	logger.Infof("batchProcess Args: txCount(%s)", args[0])

	txArrayListCount, _ := strconv.Atoi(args[0])
	txArgsCount := 0
	idxArrayList := 0
	idx := TxStartIdx
	i := 0

	type Result struct {
		UUID string `json:"uuid"`
		Code string `json:"code"`
	}

	type ResultList struct {
		Results []*Result `json:"results"`
	}

	resultList := ResultList{}

	for idxArrayList < txArrayListCount {
		startIdx := idx + TxArgsOffset
		id := idx + TxIdOffset
		uuididx := idx
		switch args[id] {
		case RegisterMemberID:
			txArgsCount = RegisterMemberArgs
		case TransferCoinID:
			txArgsCount = TransferCoinArgs
		case GetBalanceID:
			txArgsCount = GetBalanceArgs
		case RemittanceID:
			idx++
			startIdx++
			remittanceCount, _ := strconv.Atoi(args[id+1])
			txArgsCount = remittanceCount * RemittanceArgs
		}

		idx += TxArgsOffset + txArgsCount
		i = 0
		txArgs := make([]string, txArgsCount)
		for startIdx < idx {
			// logger.Info(i, ",", startIdx)
			txArgs[i] = args[startIdx]
			startIdx++
			i++
		}

		// logger.Infof("args[%d]: %s", id, args[id])

		result := new(Result)

		switch args[id] {
		case RegisterMemberID:
			result.Code = t.batchRegisterMember(stub, txArgs)
		case TransferCoinID:
			result.Code = t.batchTransferCoin(stub, txArgs)
		case GetBalanceID:
			result.Code = t.batchGetBalance(stub, txArgs)
		case RemittanceID:
			result.Code = t.batchRemittanceCoin(stub, txArgs)
		}

		result.UUID = args[uuididx]
		resultList.Results = append(resultList.Results, result)

		idxArrayList++

	}

	eventPayload := "BatchProcessEvent"
	if err := stub.SetEvent(args[1], []byte(eventPayload)); err != nil {
		return shim.Error(err.Error())
	}

	jsonList, _ := json.Marshal(resultList)

	return shim.Success(jsonList)
}

func (t *SmartContract) batchRegisterMember(stub shim.ChaincodeStubInterface, args []string) string {

	if len(args) != RegisterMemberArgs {
		logger.Error("batchRegisterMember Incorrect arguments. Expecting %s arguments", RegisterMemberArgs)
		return BadRequest
	}

	logger.Infof(`batchRegisterMember Args: TxID(%s), MemberId(%s), VSCode(%s), CountryCode(%s), 
	CurrencyCode(%s), MemberRole(%s), Wallet Address(%s), txTime(%s), MemberLevel(%s)\n`,
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8])

	memberBytes, err := stub.GetState(args[6])
	if err != nil {
		logger.Infof("GetState Error: %s", err.Error())
		return GetStateError
	}

	if memberBytes != nil {
		logger.Infof("AlreadyRegistered: %s", string(memberBytes))
		return AlreadyRegistered
	}

	account := new(model.MemberAccount)

	account.TxID = args[0]
	account.MemberID = args[1]
	account.VSCode = args[2]
	account.CountryCode = args[3]
	account.CurrencyCode = args[4]
	account.MemberRole = args[5]
	account.CreatedTime = args[7]
	account.MemberWallet.WalletAddress = args[6]
	account.MemberWallet.CoinBalance = 0
	account.MemberWallet.CashBalance = 0
	account.MemberLevel = args[8]
	account.CustomOneTimeRemittanceLimit = 0
	account.CustomOneTimeWithdrawLimit = 0
	account.CustomOneDayRemittanceLimit = 0
	account.CustomOneDayWithdrawLimit = 0
	account.OneDayRemittanceSum = 0
	// account.OneDayTransferDate
	account.OneDayWithdrawSum = 0
	// account.OneDayWithdrawDate

	account.MemberWallet.CoinLimit = account.GetWalletLimit()

	if err := account.Validate(); err != nil {
		return InvalidArgument
	}

	memberAsBytes, _ := json.Marshal(account)

	// @@ need to define exchange memberRole identifier
	// if account.MemberRole == "exchange" {
	// 	uid := fmt.Sprintf("%x", model.GenTxID(memberAsBytes))
	// 	key, err := stub.CreateCompositeKey("", []string{account.MemberWallet.WalletAddress, uid})
	// 	if err != nil {
	// 		return shim.Error(err.Error())
	// 	}

	// 	err = stub.PutState(key, memberAsBytes)
	// 	if err != nil {
	// 		return shim.Error("Failed to set member account")
	// 	}
	// } else {
	if err := stub.PutState(account.MemberWallet.WalletAddress, memberAsBytes); err != nil {
		logger.Infof("PutState Error: %s", err.Error())
		return InternalServerError
	}

	return Created
}

func (t *SmartContract) batchGetBalance(stub shim.ChaincodeStubInterface, args []string) string {
	if len(args) != 2 {
		logger.Error("batchGetBalance Incorrect arguments. Expecting 2 arguments")
		return BadRequest
	}

	logger.Infof(`batchGetBalance Args: walletAddress(%s), balance(%s)\n`, args[0], args[1])

	memberBytes, err := stub.GetState(args[0])
	if err != nil {
		logger.Infof("GetState Error: %s", err.Error())
		return GetStateError
	}

	if memberBytes == nil {
		return AccountNotFound
	}

	account := new(model.MemberAccount)
	bytesToStruct(memberBytes, account)

	balance, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return InternalServerError
	}

	if !account.MemberWallet.ValidateBalance(balance) {
		return IncorrectBalance
	}

	// if account.MemberWallet.CoinBalance != balance {
	// 	return IncorrectBalance
	// }

	// balance := strconv.FormatFloat(account.MemberWallet.CoinBalance, 'f', -1, 64)
	return args[1]
}

func (t *SmartContract) registerMember(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	ret := Created

	if len(args) != RegisterMemberArgs {
		logger.Error("registerMember Incorrect arguments. Expecting %s arguments", RegisterMemberArgs)
		ret = BadRequest
		return t.doResult(ret)
	}

	logger.Infof(`registerMember Args: TxID(%s), MemberId(%s), VSCode(%s), CountryCode(%s), 
	CurrencyCode(%s), MemberRole(%s), Wallet Address(%s), txTime(%s), MemberLevel(%s)\n`,
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8])

	memberBytes, err := stub.GetState(args[6])
	if err != nil {
		// return shim.Error(err.Error())
		logger.Infof("GetState Error: %s", err.Error())
		ret = GetStateError
		return t.doResult(ret)
	}

	if memberBytes != nil {
		// return shim.Error("Already registered")
		logger.Infof("AlreadyRegistered: %s", string(memberBytes))
		ret = AlreadyRegistered
		return t.doResult(ret)
	}

	account := new(model.MemberAccount)

	account.TxID = args[0]
	account.MemberID = args[1]
	account.VSCode = args[2]
	account.CountryCode = args[3]
	account.CurrencyCode = args[4]
	account.MemberRole = args[5]
	account.CreatedTime = args[7]
	account.MemberWallet.WalletAddress = args[6]
	account.MemberWallet.CoinBalance = 0
	account.MemberWallet.CashBalance = 0
	account.MemberLevel = args[8]
	account.CustomOneTimeRemittanceLimit = 0
	account.CustomOneTimeWithdrawLimit = 0
	account.CustomOneDayRemittanceLimit = 0
	account.CustomOneDayWithdrawLimit = 0
	account.OneDayRemittanceSum = 0
	// account.OneDayTransferDate
	account.OneDayWithdrawSum = 0
	// account.OneDayWithdrawDate

	account.MemberWallet.CoinLimit = account.GetWalletLimit()

	if err := account.Validate(); err != nil {
		// return shim.Error(err.Error())
		ret = InvalidArgument
		return t.doResult(ret)
	}

	memberAsBytes, _ := json.Marshal(account)

	// @@ need to define exchange memberRole identifier
	// if account.MemberRole == "exchange" {
	// 	uid := fmt.Sprintf("%x", model.GenTxID(memberAsBytes))
	// 	key, err := stub.CreateCompositeKey("", []string{account.MemberWallet.WalletAddress, uid})
	// 	if err != nil {
	// 		return shim.Error(err.Error())
	// 	}

	// 	err = stub.PutState(key, memberAsBytes)
	// 	if err != nil {
	// 		return shim.Error("Failed to set member account")
	// 	}
	// } else {

	err = stub.PutState(account.MemberWallet.WalletAddress, memberAsBytes)
	if err != nil {
		// return shim.Error("Failed to set member account")
		logger.Infof("PutState Error: %s", err.Error())
		ret = InternalServerError
		return t.doResult(ret)
	}

	// eventPayload := "Test Event"
	// if err := stub.SetEvent(args[0], []byte(eventPayload)); err != nil {
	// 	//return shim.Error(err.Error())
	// 	ret = "530"
	// 	return t.doResult(ret)
	// }

	return t.doResult(ret)
}

func (t *SmartContract) doResult(ret string) peer.Response {

	// eventPayload := "Test Event"
	// if err := stub.SetEvent("SingleInvoke", []byte(eventPayload)); err != nil {
	// 	//return shim.Error(err.Error())
	// 	ret = "530"
	// }

	// type Result struct {
	// 	Code string `json:"code"`
	// }

	type ResultList struct {
		Results string `json:"results"`
	}

	resultList := ResultList{}
	resultList.Results = ret

	jsonList, _ := json.Marshal(resultList)

	return shim.Success(jsonList)
}

func (t *SmartContract) setTransactionLimit(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	ret := OK

	if len(args) != 6 {
		logger.Error("setTransactionLimit Incorrect arguments. Expecting 6 arguments")
		ret = BadRequest
		return t.doResult(ret)
	}

	logger.Infof(`setTransactionLimit Args: TxID(%s), Level(%s), OneTimeRemittanceLimit(%s), OneTimeWithdrawLimit(%s),
	OneDayRemittanceLimit(%s), OneDayWithdrawLimit(%s)\n`, args[0], args[1], args[2], args[3], args[4], args[5])

	key, _ := t.createCompositeKey(model.TxRemittanceLimitObjectType, []string{args[0]})
	remittanceLimitLogBytes, err := stub.GetState(key)
	if err != nil {
		logger.Infof("GetState Error: %s", err.Error())
		ret = GetStateError
		return t.doResult(ret)
	}

	if remittanceLimitLogBytes != nil {
		ret = DuplicateTxID
		return t.doResult(ret)
	}

	keyStr := "TxLimit:" + args[1]
	remittanceLimitBytes, err := stub.GetState(keyStr)
	if err != nil {
		logger.Infof("GetState Error: %s", err.Error())
		ret = GetStateError
		return t.doResult(ret)
	}

	remittanceLimit := new(model.TxRemittanceLimit)

	if remittanceLimitBytes == nil {

		remittanceLimit.Level = "TxLimit:" + args[1]
		remittanceLimit.OneTimeRemittanceLimit, _ = strconv.ParseFloat(args[2], 64)
		remittanceLimit.OneTimeWithdrawLimit, _ = strconv.ParseFloat(args[3], 64)
		remittanceLimit.OneDayRemittanceLimit, _ = strconv.ParseFloat(args[4], 64)
		remittanceLimit.OneDayWithdrawLimit, _ = strconv.ParseFloat(args[5], 64)
	} else {

		bytesToStruct(remittanceLimitBytes, remittanceLimit)

		oneTimeRemittanceLimit, _ := strconv.ParseFloat(args[2], 64)
		oneTimeWithdrawLimit, _ := strconv.ParseFloat(args[3], 64)
		oneDayRemittanceLimit, _ := strconv.ParseFloat(args[4], 64)
		oneDayWithdrawLimit, _ := strconv.ParseFloat(args[5], 64)

		if oneTimeRemittanceLimit != 0 {
			remittanceLimit.OneTimeRemittanceLimit = oneTimeRemittanceLimit
		}
		if oneTimeWithdrawLimit != 0 {
			remittanceLimit.OneTimeWithdrawLimit = oneTimeWithdrawLimit
		}
		if oneDayRemittanceLimit != 0 {
			remittanceLimit.OneDayRemittanceLimit = oneDayRemittanceLimit
		}
		if oneDayWithdrawLimit != 0 {
			remittanceLimit.OneDayWithdrawLimit = oneDayWithdrawLimit
		}
	}

	remittanceLimitAsBytes, _ := json.Marshal(remittanceLimit)
	err = stub.PutState(remittanceLimit.Level, remittanceLimitAsBytes)
	if err != nil {
		logger.Infof("PutState Error: %s", err.Error())
		ret = InternalServerError
		return t.doResult(ret)
	}

	ret = t.recordTxRemittanceLimitLog(stub, model.TXID(args[0]), remittanceLimit, OK)

	return t.doResult(ret)
}

func (t *SmartContract) issueCoin(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	ret := OK

	if len(args) != 5 {
		logger.Error("issueCoin Incorrect arguments. Expecting 5 arguments")
		ret = BadRequest
		return t.doResult(ret)
	}
	logger.Infof(`issueCoin Args: TxID(%s), WalletAddress(%s), Amount(%s), Balance(%s), txTime(%s)\n`,
		args[0], args[1], args[2], args[3], args[4])

	walletAddress := args[1]
	amount, _ := strconv.ParseFloat(args[2], 64)

	key, _ := t.createCompositeKey(model.TxIssueCoinObjectType, []string{walletAddress, args[0]})
	issueCoinLogLogBytes, err := stub.GetState(key)
	if err != nil {
		logger.Infof("GetState Error: %s", err.Error())
		ret = GetStateError
		return t.doResult(ret)
	}

	if issueCoinLogLogBytes != nil {
		ret = DuplicateTxID
		return t.doResult(ret)
	}

	if amount <= 0 {
		ret = InvalidArgument
		return t.doResult(ret)
	}

	memberAccountBytes, err := t.getMemberAccount(stub, walletAddress)
	if err != nil {
		logger.Infof("GetState Error: %s", err.Error())
		ret = GetStateError
		return t.doResult(ret)
	}

	if memberAccountBytes == nil {
		ret = AccountNotFound
		return t.doResult(ret)
	}

	account := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, account)

	balance, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		ret = InternalServerError
		return t.doResult(ret)
	}

	if !account.MemberWallet.ValidateBalance(balance) {
		ret = IncorrectBalance
		return t.doResult(ret)
	}

	if err := t.creditCoin(stub, account, amount); err != nil {
		if err.Error() == "Maxed Out Wallet" {
			ret = WalletBalanceLimitError
		} else {
			ret = InternalServerError
		}
		return t.doResult(ret)
	}

	ret = t.recordTxIssueCoinLog(stub, model.TXID(args[0]), walletAddress, amount, OK)

	return t.doResult(ret)
}

func (t *SmartContract) processOrder(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 6 {
		logger.Error("processOrder Incorrect arguments. Expecting 6 arguments")
		return shim.Error("Incorrect arguments. Expecting 6 arguments")
	}

	sr := new(model.Remittance)
	sr.SenderWalletAddress = args[0]
	sr.ReceiverWalletAddress = args[1]
	sr.Amount, _ = strconv.ParseFloat(args[2], 64)
	sr.Fee, _ = strconv.ParseFloat(args[3], 64)

	if err := sr.Validate(); err != nil {
		return shim.Error(err.Error())
	}

	sellerMemberAccountBytes, err := t.getMemberAccount(stub, sr.SenderWalletAddress)

	if err != nil {
		return shim.Error(err.Error())
	}

	if sellerMemberAccountBytes == nil {
		return shim.Error("Seller Account not found")
	}

	sellerAccount := new(model.MemberAccount)
	bytesToStruct(sellerMemberAccountBytes, sellerAccount)

	if sellerAccount.Frozen {
		return shim.Error("Cannot execute a transaction into frozen member")
	}

	if sellerAccount.MemberWallet.CoinBalance < sr.Amount {
		return shim.Error("Insufficient Coin Balance")
	}

	if sellerAccount.MemberWallet.CashBalance < sr.Fee {
		return shim.Error("Insufficient Cash Balance")
	}

	if args[0] == args[1] {
		sellerAccount.DebitCash(sr.Fee)
		t.creditFee(stub, sellerAccount.MemberWallet.WalletAddress, sr.Fee)
		// t.recordTransaction(stub, sr, model.TxSellCoin, model.TxSuccess)

		sellerAccount.DebitCash(sr.Fee)
		t.creditFee(stub, sellerAccount.MemberWallet.WalletAddress, sr.Fee)
		// t.recordTransaction(stub, sr, model.TxSellCoin, model.TxSuccess)

		t.updateState(stub, sellerAccount)
	} else {
		br := new(model.Remittance)
		br.SenderWalletAddress = args[1]
		br.ReceiverWalletAddress = args[0]
		br.Amount, _ = strconv.ParseFloat(args[4], 64)
		br.Fee, _ = strconv.ParseFloat(args[5], 64)

		if err := br.Validate(); err != nil {
			return shim.Error(err.Error())
		}

		buyerMemberAccountBytes, err := t.getMemberAccount(stub, br.SenderWalletAddress)

		if err != nil {
			return shim.Error(err.Error())
		}

		if buyerMemberAccountBytes == nil {
			return shim.Error("Buyer Account not found")
		}
		buyerAccount := new(model.MemberAccount)
		bytesToStruct(buyerMemberAccountBytes, buyerAccount)

		if buyerAccount.Frozen {
			return shim.Error("Cannot execute a transaction into frozen member")
		}

		if buyerAccount.MemberWallet.CashBalance < br.Amount+br.Fee {
			return shim.Error("Insufficient Cash Balance")
		}

		sellerAccount.DebitCoin(sr.Amount)
		sellerAccount.DebitCash(sr.Fee)
		t.creditFee(stub, sellerAccount.MemberWallet.WalletAddress, sr.Fee)
		ve := buyerAccount.CreditCoin(sr.Amount)
		if ve != nil {
			return shim.Error(ve.Error())
		}

		// t.recordTransaction(stub, sr, model.TxSellCoin, model.TxSuccess)

		buyerAccount.DebitCash(br.Amount)
		buyerAccount.DebitCash(br.Fee)
		t.creditFee(stub, buyerAccount.MemberWallet.WalletAddress, br.Fee)

		sellerAccount.CreditCash(br.Amount)

		// t.recordTransaction(stub, br, model.TxBuyCoin, model.TxSuccess)

		t.updateState(stub, sellerAccount)
		t.updateState(stub, buyerAccount)
	}

	return shim.Success(nil)
}

func (t *SmartContract) batchRemittanceCoin(stub shim.ChaincodeStubInterface, args []string) string {
	// walletAddress, expectedBalance, amount
	logger.Info(args)
	argLen := len(args)
	if argLen % 3 != 0 {
		logger.Error("[batchRemittanceCoin] Incorrect arguments. Expecting more than 2 args")
		return BadRequest
	}

	type element struct {
		walletAddress string
		expected float64
		amount float64
	}

	// elementCount := argLen / 3

	/*2020-02-07 12:17:45.693 UTC [exchange] Info -> INFO 001 Starting exchange chaincode
	2020-02-07 12:17:45.747 UTC [exchange] Infof -> INFO 002 Invoke function : batchProcess
	2020-02-07 12:17:45.747 UTC [exchange] Infof -> INFO 003 batchProcess Args: txCount(1)
	2020-02-07 12:17:45.747 UTC [exchange] Info -> INFO 004 [elmo 0 -20 user2 0 10 user1 0 10]
	2020-02-07 12:17:45.747 UTC [exchange] Info -> INFO 005 [{ 0 0} { 0 0} { 0 0}]*/


	elements := make([]element, 0)

	for i := 0; i < argLen; i += 3 {
		end := i + 3

		if end > argLen {
			end = argLen
		}

		wallet := args[i]
		expected, _ := strconv.ParseFloat(args[i + 1], 64)
		amount, _ := strconv.ParseFloat(args[i + 2], 64)

		newElement := element{
			walletAddress: wallet,
			expected:      expected,
			amount:        amount,
		}

		elements = append(elements, newElement)
	}


	for _, e := range elements {
		memberAccount, err := t.getMemberAccount(stub, e.walletAddress)
		if err != nil {
			logger.Errorf("[batchRemittanceCoin] GetState Error: %s", err.Error())
			return GetStateError
		}
		if memberAccount == nil {
			return ReceiverAccountNotFound
		}
		account := new(model.MemberAccount)
		err = bytesToStruct(memberAccount, &account)
		if err != nil {
			logger.Error(`[batchRemittanceCoin] - bytesToStruct:`, err.Error())
		}

		if !account.MemberWallet.ValidateBalance(e.expected) {
			return IncorrectSenderBalance
		}

		if account.Frozen {
			return ReceiverAccountFrozen
		}

		account.MemberWallet.CreditCoin(e.amount)

		// if err := t.creditCoin(stub, account, e.amount); err != nil {
		// 	if err.Error() == "Maxed Out Wallet" {
		// 		return WalletBalanceLimitError
		// 	} else {
		// 		return InternalServerError
		// 	}
		// }
		accountToByte, _ := json.Marshal(account)
		err = stub.PutState(account.MemberWallet.WalletAddress, accountToByte)
		logger.Info("[batchRemittanceCoin] - PutState:", account)
		if err != nil {
			return InternalServerError
		}

	}

	// elements := make([]element, elementCount)
	// for i, e := range elements {
	// 	e.walletAddress = args[i]
	// 	e.expected, _ = strconv.ParseFloat(args[i+1], 64)
	// 	e.amount, _ = strconv.ParseFloat(args[i+1], 64)
	// }
	//
	// logger.Info(elements)

	// memberAccountBytes, err := t.getMemberAccount(stub, args[0])
	// if err != nil {
	// 	logger.Errorf("[batchRemittanceCoin] GetState Error: %s", err.Error())
	// 	return GetStateError
	// }
	//
	// if memberAccountBytes == nil {
	// 	return ReceiverAccountNotFound
	// }
	//
	// account := new(model.MemberAccount)
	// e := bytesToStruct(memberAccountBytes, account)
	// if e != nil {
	// 	logger.Error(`[batchRemittanceCoin] - bytesToStruct:`, err.Error())
	// }
	//
	// balance, err := strconv.ParseFloat(args[1], 64)
	// if err != nil {
	// 	return InternalServerError
	// }
	//
	// if !account.MemberWallet.ValidateBalance(balance) {
	// 	return IncorrectSenderBalance
	// }
	//
	// if account.Frozen {
	// 	return ReceiverAccountFrozen
	// }
	//
	// amount, err := strconv.ParseFloat(args[2], 64)
	// if err := t.creditCoin(stub, account, amount); err != nil {
	// 	if err.Error() == "Maxed Out Wallet" {
	// 		return WalletBalanceLimitError
	// 	} else {
	// 		return InternalServerError
	// 	}
	// }
	//
	// err = stub.PutState(account.MemberWallet.WalletAddress, memberAccountBytes)
	// if err != nil {
	// 	return InternalServerError
	// }

	return OK
}

func (t *SmartContract) batchTransferCoin(stub shim.ChaincodeStubInterface, args []string) string {

	if len(args) != 9 {
		logger.Error("transferCoin Incorrect arguments. Expecting 9 arguments")
		return BadRequest
	}

	logger.Infof("transferCoin Args: txID(%s), senderWalletAddress(%s), senderBalance(%s), receiverWalletAddress(%s), receiverBalance(%s), coinAmount(%s), fee(%s), txFlag(%s), txTime(%s)\n",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8])

	r := new(model.Remittance)
	r.SenderWalletAddress = args[1]
	r.ReceiverWalletAddress = args[3]
	r.Amount, _ = strconv.ParseFloat(args[5], 64)
	r.Fee, _ = strconv.ParseFloat(args[6], 64)
	r.TxFlag = args[7]
	r.TxType = model.TxTransferCoin

	if err := r.Validate(); err != nil {
		return InvalidArgument
	}

	key, _ := t.createCompositeKey(model.TxRemittanceObjectType, []string{args[1], args[3], args[0]})
	remittanceLogBytes, err := stub.GetState(key)
	if err != nil {
		logger.Infof("GetState Error: %s", err.Error())
		return GetStateError
	}

	if remittanceLogBytes != nil {
		return DuplicateTxID
	}

	memberAccountBytes, err := t.getMemberAccount(stub, r.SenderWalletAddress)
	if err != nil {
		logger.Infof("GetState Error: %s", err.Error())
		return GetStateError
	}

	if memberAccountBytes == nil {
		return SenderAccountNotFound
	}

	senderAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, senderAccount)

	senderBalance, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return InternalServerError
	}

	if !senderAccount.MemberWallet.ValidateBalance(senderBalance) {
		return IncorrectSenderBalance
	}

	if senderAccount.Frozen {
		return SenderAccountFrozen
	}

	if senderAccount.MemberWallet.CoinBalance < (r.Amount + r.Fee) {
		return SenderAccountInsufficientBalance
	}

	// deprecated
	// if senderAccount.MemberWallet.CashBalance < r.Fee {
	// 	return shim.Error("Insufficient Cash Balance")
	// }

	// if senderAccount.MemberRole != "exchange" {
	// 	if senderAccount.CustomOneTimeRemittanceLimit != 0 {
	// 		if senderAccount.CustomOneTimeRemittanceLimit < r.Amount {
	// 			return CustomOneTimeRemittanceLimitError
	// 		}
	// 	} else {
	// 		txLimitBytes, _ := t.getRemittanceLimit(stub, senderAccount.MemberLevel)
	// 		txLimit := new(model.TxRemittanceLimit)
	// 		bytesToStruct(txLimitBytes, txLimit)

	// 		if txLimit.OneTimeRemittanceLimit < r.Amount {
	// 			return OneTimeRemittanceLimitError
	// 		}
	// 	}

	// 	tm := time.Now()
	// 	txDate := model.TxDate{tm.Year(), tm.Month(), tm.Day()}
	// 	if txDate != senderAccount.OneDayRemittanceDate {
	// 		senderAccount.OneDayRemittanceDate = txDate
	// 		senderAccount.OneDayRemittanceSum = 0
	// 	}

	// 	if senderAccount.CustomOneDayRemittanceLimit != 0 {
	// 		if senderAccount.CustomOneDayRemittanceLimit < senderAccount.OneDayRemittanceSum+r.Amount {
	// 			return CustomOneDayRemittanceLimitError
	// 		}
	// 	} else {
	// 		txLimitBytes, _ := t.getRemittanceLimit(stub, senderAccount.MemberLevel)
	// 		txLimit := new(model.TxRemittanceLimit)
	// 		bytesToStruct(txLimitBytes, txLimit)

	// 		if txLimit.OneDayRemittanceLimit < senderAccount.OneDayRemittanceSum+r.Amount {
	// 			return OneDayRemittanceLimitError
	// 		}
	// 	}
	// }

	// senderAccount.OneDayRemittanceSum += r.Amount

	memberAccountBytes, _ = json.Marshal(senderAccount)

	err = stub.PutState(senderAccount.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		return InternalServerError
	}

	memberAccountBytes, err = t.getMemberAccount(stub, r.ReceiverWalletAddress)

	if err != nil {
		logger.Infof("GetState Error: %s", err.Error())
		return GetStateError
	}

	if memberAccountBytes == nil {
		return ReceiverAccountNotFound
	}
	receiverAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, receiverAccount)

	receiverBalance, err := strconv.ParseFloat(args[4], 64)
	if err != nil {
		return InternalServerError
	}

	if !receiverAccount.MemberWallet.ValidateBalance(receiverBalance) {
		return IncorrectReceiverBalance
	}

	if receiverAccount.Frozen {
		return ReceiverAccountFrozen
	}

	if err := t.debitCoin(stub, senderAccount, r.Amount+r.Fee); err != nil {
		return InternalServerError
	}
	// t.debitCash(stub, senderAccount, r.Fee) //deprecated
	if r.Fee > 0 {
		if err := t.creditFee(stub, senderAccount.MemberWallet.WalletAddress, r.Fee); err != nil {
			return InternalServerError
		}
	}
	if err := t.creditCoin(stub, receiverAccount, r.Amount); err != nil {
		if err.Error() == "Maxed Out Wallet" {
			return WalletBalanceLimitError
		} else {
			return InternalServerError
		}
	}

	ret := t.recordTxRemittanceLog(stub, model.TXID(args[0]), r, OK)

	return ret
}

func (t *SmartContract) transferCoin(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 7 {
		logger.Error("transferCoin Incorrect arguments. Expecting 7 arguments")
		return shim.Error("Incorrect arguments. Expecting 7 arguments")
	}

	logger.Infof("transferCoin Args: txID(%s), senderWalletAddress(%s), receiverWalletAddress(%s), coinAmount(%s), fee(%s), txFlag(%s), txTime(%s)\n",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6])

	r := new(model.Remittance)
	r.SenderWalletAddress = args[1]
	r.ReceiverWalletAddress = args[2]
	r.Amount, _ = strconv.ParseFloat(args[3], 64)
	r.Fee, _ = strconv.ParseFloat(args[4], 64)
	r.TxFlag = args[5]
	r.TxType = model.TxTransferCoin

	if err := r.Validate(); err != nil {
		return shim.Error(err.Error())
	}

	key, _ := t.createCompositeKey(model.TxRemittanceObjectType, []string{args[1], args[2], args[0]})
	remittanceLogBytes, err := stub.GetState(key)
	if err != nil {
		logger.Infof("GetState Error: %s", err.Error())
		return shim.Error(err.Error())
	}

	if remittanceLogBytes != nil {
		return shim.Error("Dupllicate TxID")
	}

	memberAccountBytes, err := t.getMemberAccount(stub, r.SenderWalletAddress)

	if err != nil {
		return shim.Error(err.Error())
	}

	if memberAccountBytes == nil {
		return shim.Error(err.Error())
	}

	senderAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, senderAccount)

	if senderAccount.Frozen {
		return shim.Error("Cannot execute a transaction into frozen member")
	}

	if senderAccount.MemberWallet.CoinBalance < (r.Amount + r.Fee) {
		return shim.Error("Insufficient Coin Balance")
	}

	// deprecated
	// if senderAccount.MemberWallet.CashBalance < r.Fee {
	// 	return shim.Error("Insufficient Cash Balance")
	// }
	// if senderAccount.MemberRole != "exchange" {
	// 	if senderAccount.CustomOneTimeRemittanceLimit != 0 {
	// 		if senderAccount.CustomOneTimeRemittanceLimit < r.Amount {
	// 			return shim.Error("One Time Remittance Limit Over")
	// 		}
	// 	} else {
	// 		txLimitBytes, _ := t.getRemittanceLimit(stub, senderAccount.MemberLevel)
	// 		txLimit := new(model.TxRemittanceLimit)
	// 		bytesToStruct(txLimitBytes, txLimit)

	// 		if txLimit.OneTimeRemittanceLimit < r.Amount {
	// 			return shim.Error("One Time Remittance Limit Over")
	// 		}
	// 	}

	// 	tm := time.Now()
	// 	txDate := model.TxDate{tm.Year(), tm.Month(), tm.Day()}
	// 	if txDate != senderAccount.OneDayRemittanceDate {
	// 		senderAccount.OneDayRemittanceDate = txDate
	// 		senderAccount.OneDayRemittanceSum = 0
	// 	}

	// 	if senderAccount.CustomOneDayRemittanceLimit != 0 {
	// 		if senderAccount.CustomOneDayRemittanceLimit < senderAccount.OneDayRemittanceSum+r.Amount {
	// 			return shim.Error("One Day Remittance Limit Over")
	// 		}
	// 	} else {
	// 		txLimitBytes, _ := t.getRemittanceLimit(stub, senderAccount.MemberLevel)
	// 		txLimit := new(model.TxRemittanceLimit)
	// 		bytesToStruct(txLimitBytes, txLimit)

	// 		if txLimit.OneDayRemittanceLimit < senderAccount.OneDayRemittanceSum+r.Amount {
	// 			return shim.Error("One Day Remittance Limit Over")
	// 		}
	// 	}
	// }

	// senderAccount.OneDayRemittanceSum += r.Amount

	memberAccountBytes, _ = json.Marshal(senderAccount)

	err = stub.PutState(senderAccount.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	memberAccountBytes, err = t.getMemberAccount(stub, r.ReceiverWalletAddress)

	if err != nil {
		return shim.Error(err.Error())
	}

	if memberAccountBytes == nil {
		return shim.Error("Receiver Account not found")
	}
	receiverAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, receiverAccount)

	if receiverAccount.Frozen {
		return shim.Error("Cannot execute a transaction into frozen member")
	}

	t.debitCoin(stub, senderAccount, r.Amount+r.Fee)
	// t.debitCash(stub, senderAccount, r.Fee) //deprecated
	if r.Fee > 0 {
		t.creditFee(stub, senderAccount.MemberWallet.WalletAddress, r.Fee)
	}
	if err := t.creditCoin(stub, receiverAccount, r.Amount); err != nil {
		return shim.Error(err.Error())
	}

	// t.recordTxRemittanceLog(stub, model.TXID(args[0]), r, OK)

	eventPayload := "Test Event"
	if err := stub.SetEvent(args[0], []byte(eventPayload)); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SmartContract) transferCash(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 4 {
		logger.Error("transferCash Incorrect arguments. Expecting 4 arguments")
		return shim.Error("Incorrect arguments. Expecting 4 arguments")
	}

	logger.Infof("transferCash Args: senderWalletAddress(%s), receiverWalletAddress(%s), coinAmount(%s), fee(%s)\n",
		args[0], args[1], args[2], args[3])

	r := new(model.Remittance)
	r.SenderWalletAddress = args[0]
	r.ReceiverWalletAddress = args[1]
	r.Amount, _ = strconv.ParseFloat(args[2], 64)
	r.Fee, _ = strconv.ParseFloat(args[3], 64)

	if err := r.Validate(); err != nil {
		return shim.Error(err.Error())
	}

	memberAccountBytes, err := t.getMemberAccount(stub, r.SenderWalletAddress)

	if err != nil {
		return shim.Error(err.Error())
	}

	if memberAccountBytes == nil {
		return shim.Error("Sender Account not found")
	}

	senderAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, senderAccount)

	if senderAccount.Frozen {
		return shim.Error("Cannot execute a transaction into frozen member")
	}

	if senderAccount.MemberWallet.CashBalance < r.Amount+r.Fee {
		return shim.Error("Insufficient Cash Balance")
	}

	memberAccountBytes, err = t.getMemberAccount(stub, r.ReceiverWalletAddress)

	if err != nil {
		return shim.Error(err.Error())
	}

	if memberAccountBytes == nil {
		return shim.Error("Receiver Account not found")
	}
	receiverAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, receiverAccount)

	if receiverAccount.Frozen {
		return shim.Error("Cannot execute a transaction into frozen member")
	}

	t.debitCash(stub, senderAccount, r.Amount)
	t.debitCash(stub, senderAccount, r.Fee)
	t.creditFee(stub, senderAccount.MemberWallet.WalletAddress, r.Fee)
	t.creditCash(stub, receiverAccount, r.Amount)

	// t.recordTransaction(stub, r, model.TxTransferCash, model.TxSuccess)

	return shim.Success(nil)
}

func (t *SmartContract) updateState(stub shim.ChaincodeStubInterface, a *model.MemberAccount) error {

	memberAccountBytes, _ := json.Marshal(a)
	err := stub.PutState(a.MemberWallet.WalletAddress, memberAccountBytes)

	if err != nil {
		logger.Infof("PutState Error: %s", err.Error())
		return err
	}

	return err
}

func (t *SmartContract) updateStateByCompositeKey(stub shim.ChaincodeStubInterface, uid string, a *model.MemberAccount) (string, error) {

	key, err := stub.CreateCompositeKey("", []string{a.MemberWallet.WalletAddress, uid})
	if err != nil {
		return uid, err
	}
	err = stub.DelState(key)
	if err != nil {
		logger.Infof("DelState Error: %s", err.Error())
		return uid, err
	}

	memberAccountBytes, _ := json.Marshal(a)

	uid = fmt.Sprintf("%x", model.GenTxID(memberAccountBytes))
	key, err = stub.CreateCompositeKey("", []string{a.MemberWallet.WalletAddress, uid})
	if err != nil {
		return uid, err
	}
	err = stub.PutState(key, memberAccountBytes)
	if err != nil {
		logger.Infof("PutState Error: %s", err.Error())
		return uid, err
	}

	return uid, err
}

func (t *SmartContract) debitCoin(stub shim.ChaincodeStubInterface, a *model.MemberAccount, amount float64) error {

	a.DebitCoin(amount)

	memberAccountBytes, _ := json.Marshal(a)
	err := stub.PutState(a.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		logger.Infof("PutState Error: %s", err.Error())
		return err
	}
	return err
}

func (t *SmartContract) debitCoinByCompositeKey(stub shim.ChaincodeStubInterface, uid string, a *model.MemberAccount, amount float64) (string, error) {

	key, err := stub.CreateCompositeKey("", []string{a.MemberWallet.WalletAddress, uid})
	if err != nil {
		return uid, err
	}
	err = stub.DelState(key)
	if err != nil {
		return uid, err
	}
	a.DebitCoin(amount)

	memberAccountBytes, _ := json.Marshal(a)
	uid = fmt.Sprintf("%x", model.GenTxID(memberAccountBytes))
	key, err = stub.CreateCompositeKey("", []string{a.MemberWallet.WalletAddress, uid})
	if err != nil {
		return uid, err
	}
	err = stub.PutState(key, memberAccountBytes)
	// err := stub.PutState(a.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		return uid, err
	}
	return uid, err
}

func (t *SmartContract) creditCoin(stub shim.ChaincodeStubInterface, a *model.MemberAccount, amount float64) error {

	err := a.CreditCoin(amount)
	if err != nil {
		logger.Error(`Wallet Validation Error:`, err.Error())
		return err
	}
	memberAccountBytes, _ := json.Marshal(a)

	err = stub.PutState(a.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		logger.Infof("PutState Error: %s", err.Error())
		return err
	}
	return err
}

func (t *SmartContract) creditCoinByCompositeKey(stub shim.ChaincodeStubInterface, uid string, a *model.MemberAccount, amount float64) (string, error) {

	key, err := stub.CreateCompositeKey("", []string{a.MemberWallet.WalletAddress, uid})
	if err != nil {
		return uid, err
	}
	err = stub.DelState(key)
	if err != nil {
		return uid, err
	}
	err = a.CreditCoin(amount)
	if err != nil {
		return uid, err
	}

	memberAccountBytes, _ := json.Marshal(a)
	uid = fmt.Sprintf("%x", model.GenTxID(memberAccountBytes))
	key, err = stub.CreateCompositeKey("", []string{a.MemberWallet.WalletAddress, uid})
	if err != nil {
		return uid, err
	}
	err = stub.PutState(key, memberAccountBytes)
	// err := stub.PutState(a.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		return uid, err
	}
	return uid, err
}

func (t *SmartContract) debitCash(stub shim.ChaincodeStubInterface, a *model.MemberAccount, amount float64) error {

	a.DebitCash(amount)

	memberAccountBytes, _ := json.Marshal(a)

	err := stub.PutState(a.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		return err
	}
	return err
}

func (t *SmartContract) debitCashByCompositeKey(stub shim.ChaincodeStubInterface, uid string, a *model.MemberAccount, amount float64) (string, error) {
	key, err := stub.CreateCompositeKey("", []string{a.MemberWallet.WalletAddress, uid})
	if err != nil {
		return uid, err
	}
	err = stub.DelState(key)
	if err != nil {
		return uid, err
	}
	a.DebitCash(amount)

	memberAccountBytes, _ := json.Marshal(a)
	uid = fmt.Sprintf("%x", model.GenTxID(memberAccountBytes))
	key, err = stub.CreateCompositeKey("", []string{a.MemberWallet.WalletAddress, uid})
	if err != nil {
		return uid, err
	}
	err = stub.PutState(key, memberAccountBytes)
	// err := stub.PutState(a.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		return uid, err
	}
	return uid, err
}

func (t *SmartContract) creditCash(stub shim.ChaincodeStubInterface, a *model.MemberAccount, amount float64) error {

	a.CreditCash(amount)

	memberAccountBytes, _ := json.Marshal(a)

	err := stub.PutState(a.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		return err
	}
	return err
}

func (t *SmartContract) creditCashByCompositeKey(stub shim.ChaincodeStubInterface, uid string, a *model.MemberAccount, amount float64) (string, error) {
	key, err := stub.CreateCompositeKey("", []string{a.MemberWallet.WalletAddress, uid})
	if err != nil {
		return uid, err
	}
	err = stub.DelState(key)
	if err != nil {
		return uid, err
	}
	a.CreditCash(amount)

	memberAccountBytes, _ := json.Marshal(a)
	uid = fmt.Sprintf("%x", model.GenTxID(memberAccountBytes))
	key, err = stub.CreateCompositeKey("", []string{a.MemberWallet.WalletAddress, uid})
	if err != nil {
		return uid, err
	}

	// err := stub.PutState(a.MemberWallet.WalletAddress, memberAccountBytes)
	err = stub.PutState(key, memberAccountBytes)
	if err != nil {
		return uid, err
	}
	return uid, err
}

func (t *SmartContract) creditFee(stub shim.ChaincodeStubInterface, walletAddress string, amount float64) error {

	txid := stub.GetTxID()
	logger.Infof("txid: %s", txid)
	amountStr := strconv.FormatFloat(amount, 'f', -1, 64)
	key, err := stub.CreateCompositeKey("ExchangeFee", []string{"exchange_platform", walletAddress, amountStr, txid})
	if err != nil {
		return err
	}

	err = stub.PutState(key, []byte{0x00})
	if err != nil {
		logger.Infof("PutState Error : %s", err.Error())
		return err
	}

	tx := model.CreateTxFeeLog("exchange_platform", amount, model.TxSuccess)
	txBytes, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	key, _ = t.createCompositeKey(tx.GetObjectType(), []string{"exchange_platform", walletAddress, amountStr, txid})
	logger.Infof("key : %s, tx : %v", key, tx)
	err = stub.PutState(key, txBytes)
	if err != nil {
		logger.Infof("PutState Error : %s", err.Error())
		return err
	}

	return err
}

func (t *SmartContract) getMemberAccountList(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		logger.Error("getMemberAccountList Incorrect arguments. Expecting 2 arguments")
		return shim.Error("Incorrect arguments. Expecting 2 arguments")
	}

	logger.Infof("getMemberAccountList: args[0](%s) args[1](%s)", args[0], args[1])
	resultsIter, err := stub.GetStateByRange(args[0], args[1])
	if err != nil {
		return shim.Error(err.Error())
	}

	defer resultsIter.Close()

	// var buffer bytes.Buffer
	// buffer.WriteString("[")

	// bArrayMemberAlreadyWritten := false

	memberAccountList := model.MemberAccountList{}

	for resultsIter.HasNext() {
		memberAccount, err := resultsIter.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// if bArrayMemberAlreadyWritten == true {
		// 	buffer.WriteString(",")
		// }
		// buffer.WriteString("{\"Key\":")
		// buffer.WriteString("\"")
		// buffer.WriteString(memberAccount.Key)
		// buffer.WriteString("\"")

		// buffer.WriteString(", \"Record\":")
		// // Record is a JSON object, so we write as-is
		// buffer.WriteString(string(memberAccount.Value))
		// buffer.WriteString("}")
		// bArrayMemberAlreadyWritten = true
		ma := new(model.MemberAccount)
		err = bytesToStruct(memberAccount.GetValue(), ma)
		if err != nil {
			logger.Errorf("Failed to get account details. Error: %s", err)
			continue
		}
		logger.Infof("ma : %v", ma)
		memberAccountList.MemberAccounts = append(memberAccountList.MemberAccounts, ma)
		logger.Infof("memberAccountList.MemberAccounts : %s", memberAccountList.MemberAccounts)
	}
	// buffer.WriteString("]")

	jsonList, _ := json.Marshal(memberAccountList)
	logger.Infof("member account List : %s", jsonList)

	// return shim.Success(buffer.Bytes())
	return shim.Success(jsonList)
}

func (t *SmartContract) getRemittanceLimit(stub shim.ChaincodeStubInterface, level string) ([]byte, error) {

	if level == "" {
		logger.Error("getRemittanceLimit Incorrect arguments. Expecting 1 arguments")
		return nil, errors.New("Missing required level")
	}

	transactionLimitBytes, err := stub.GetState(level)
	if err != nil {
		logger.Errorf("Failed to get Transaction Limit. Error: %s", err)
		return nil, err
	}

	return transactionLimitBytes, nil
}

func (t *SmartContract) getMemberAccount(stub shim.ChaincodeStubInterface, walletAddress string) ([]byte, error) {

	if walletAddress == "" {
		logger.Error("getMemberAccount Incorrect arguments. Expecting 1 arguments")
		return nil, errors.New("Missing required walletAddresss")
	}

	logger.Infof("getMemberAccount: walletAddress(%s)", walletAddress)

	memberAccountBytes, err := stub.GetState(walletAddress)
	if err != nil {
		logger.Errorf("Failed to get member account. Error: %s", err)
		return nil, err
	}

	return memberAccountBytes, nil
}

func (t *SmartContract) getMemberAccountByCompositeKey(stub shim.ChaincodeStubInterface, walletAddress string) (string, []byte, error) {

	if walletAddress == "" {
		logger.Error("getMemberAccount Incorrect arguments. Expecting 1 arguments")
		return "", nil, errors.New("Missing required walletAddresss")
	}

	logger.Infof("getMemberAccount: walletAddress(%s)", walletAddress)

	resultsIter, err := stub.GetStateByPartialCompositeKey("", []string{walletAddress})
	if err != nil {
		logger.Errorf("Failed to get member account. Error: %s", err)
		return "", nil, err
	}
	defer resultsIter.Close()
	var uid string

	for i := 0; resultsIter.HasNext(); i++ {
		key, err := resultsIter.Next()
		if err != nil {
			logger.Errorf("Failed to get member account. Error: %s", err)
			return "", nil, err
		}
		_, compositeKeyParts, err := stub.SplitCompositeKey(key.Key)
		if err != nil {
			logger.Errorf("Failed to get member account. Error: %s", err)
			return "", nil, err
		}
		walletAddr := compositeKeyParts[0]
		uid = compositeKeyParts[1]
		logger.Infof("walletAddress: (%s), uid : (%s)", walletAddr, uid)
	}

	var key string
	key, err = stub.CreateCompositeKey("", []string{walletAddress, uid})
	if err != nil {
		return "", nil, err
	}

	memberAccountBytes, err := stub.GetState(key)
	if err != nil {
		logger.Errorf("Failed to get member account. Error: %s", err)
		return "", nil, err
	}

	return uid, memberAccountBytes, nil
}

func (t *SmartContract) chargeCoin(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		logger.Error("chargeCoin Incorrect arguments. Expecting 2 arguments")
		return shim.Error("Incorrect arguments. Expecting 2 arguments")
	}

	logger.Infof("chargeCoin Args: %s, %s\n", args[0], args[1])

	memberAccountBytes, err := t.getMemberAccount(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	if memberAccountBytes == nil {
		return shim.Error("Member Accout not found")
	}

	memberAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, memberAccount)

	if memberAccount.Frozen {
		return shim.Error("Cannot execute a transaction into frozen member")
	}

	var coinAmount float64
	coinAmount, err = strconv.ParseFloat(args[1], 64)
	if err != nil {
		return shim.Error("Error parsing amount value")
	}

	if coinAmount <= 0 {
		return shim.Error("Invalid coin amount")
	}

	if err := t.creditCoin(stub, memberAccount, coinAmount); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SmartContract) depositCash(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		logger.Error("depositCash Incorrect arguments. Expecting 2 arguments")
		return shim.Error("Incorrect arguments. Expecting 2 arguments")
	}

	logger.Infof("depositCash Args: %s, %s\n", args[0], args[1])

	memberAccountBytes, err := t.getMemberAccount(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	if memberAccountBytes == nil {
		return shim.Error("Member Accout not found")
	}

	memberAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, memberAccount)

	if memberAccount.Frozen {
		return shim.Error("Cannot execute a transaction into frozen member")
	}

	var cashAmount float64
	cashAmount, err = strconv.ParseFloat(args[1], 64)
	if err != nil {
		return shim.Error("Error parsing amount value")
	}

	if cashAmount <= 0 {
		return shim.Error("Invalid cash amount")
	}

	t.creditCash(stub, memberAccount, cashAmount)

	return shim.Success(nil)
}

func (t *SmartContract) withdrawCash(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		logger.Error("withdrawCash Incorrect arguments. Expecting 2 arguments")
		return shim.Error("Incorrect arguments. Expecting 2 arguments")
	}

	logger.Infof("withdrawCash Args: %s, %s\n", args[0], args[1])

	memberAccountBytes, err := t.getMemberAccount(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	if memberAccountBytes == nil {
		return shim.Error("Member Accout not found")
	}

	memberAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, memberAccount)

	if memberAccount.Frozen {
		return shim.Error("Cannot execute a transaction into frozen member")
	}

	var cashAmount float64
	cashAmount, err = strconv.ParseFloat(args[1], 64)
	if err != nil {
		return shim.Error("Error parsing amount value")
	}

	if cashAmount <= 0 {
		return shim.Error("Invalid cash amount")
	}

	if memberAccount.MemberWallet.CashBalance < cashAmount {
		return shim.Error("Insufficient cash balance")
	}

	t.debitCash(stub, memberAccount, cashAmount)

	return shim.Success(nil)
}

func (t *SmartContract) setMemberLevel(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	ret := OK

	if len(args) != 1 {
		logger.Error("setMemberLevel Incorrect arguments. Expecting 1 arguments")
		ret = BadRequest
		return t.doResult(ret)
	}

	logger.Infof("setMemberLevel Args: %s\n", args[0])

	memberAccountBytes, err := t.getMemberAccount(stub, args[0])
	if err != nil {
		logger.Infof("GetState Error: %s", err.Error())
		ret = GetStateError
		return t.doResult(ret)
	}

	if memberAccountBytes == nil {
		ret = AccountNotFound
		return t.doResult(ret)
	}

	memberAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, memberAccount)

	if memberAccount.MemberLevel == args[0] {
		ret = AlreadySetMemberLevel
		return t.doResult(ret)
	}

	memberAccount.MemberLevel = args[0]
	memberAccountBytes, _ = json.Marshal(memberAccount)
	err = stub.PutState(memberAccount.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		logger.Infof("PutState Error: ", err.Error())
		ret = InternalServerError
		return t.doResult(ret)
	}

	return t.doResult(ret)
}

func (t *SmartContract) freezeMemberAccount(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	ret := OK

	if len(args) != 4 {
		logger.Error("freezeMemberAccount Incorrect arguments. Expecting 4 arguments")
		ret = BadRequest
		return t.doResult(ret)
	}

	logger.Infof("freezeMemberAccount Args: txID(%s), walletAddress(%s), balance(%s), txTime(%s)\n",
		args[0], args[1], args[2], args[3])

	key, _ := t.createCompositeKey(model.TxFreezeMemberAccountObjectType, []string{args[1], args[0]})
	freezeMemberAccountBytes, err := stub.GetState(key)
	if err != nil {
		logger.Infof("GetState Error: %s", err.Error())
		ret = GetStateError
		return t.doResult(ret)
	}

	if freezeMemberAccountBytes != nil {
		ret = DuplicateTxID
		return t.doResult(ret)
	}

	memberAccountBytes, err := t.getMemberAccount(stub, args[1])
	if err != nil {
		logger.Infof("GetState Error: %s", err.Error())
		ret = GetStateError
		return t.doResult(ret)
	}

	if memberAccountBytes == nil {
		ret = AccountNotFound
		return t.doResult(ret)
	}

	memberAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, memberAccount)

	balance, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		ret = InternalServerError
		return t.doResult(ret)
	}

	if !memberAccount.MemberWallet.ValidateBalance(balance) {
		ret = IncorrectBalance
		return t.doResult(ret)
	}

	if memberAccount.Frozen {
		ret = AlreadyFrozenMember
		return t.doResult(ret)
	}

	memberAccount.Frozen = true
	memberAccountBytes, _ = json.Marshal(memberAccount)
	err = stub.PutState(memberAccount.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		logger.Infof("PutState Error: ", err.Error())
		ret = InternalServerError
		return t.doResult(ret)
	}

	ret = t.recordTxFreezeMemberAccountLog(stub, model.TXID(args[0]), memberAccount.MemberWallet.WalletAddress, model.TxFreezeMember, OK)

	return t.doResult(ret)
}

func (t *SmartContract) recoverFrozenMemberAccount(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	ret := OK

	if len(args) != 4 {
		logger.Error("recoverFrozenMemberAccount Incorrect arguments. Expecting 4 arguments")
		ret = BadRequest
		return t.doResult(ret)
	}

	logger.Infof("recoverFrozenMemberAccount Args:  txID(%s), walletAddress(%s), balance(%s), txTime(%s)\n",
		args[0], args[1], args[2], args[3])

	key, _ := t.createCompositeKey(model.TxFreezeMemberAccountObjectType, []string{args[1], args[0]})
	freezeMemberAccountBytes, err := stub.GetState(key)
	if err != nil {
		logger.Infof("GetState Error: %s", err.Error())
		ret = GetStateError
		return t.doResult(ret)
	}

	if freezeMemberAccountBytes != nil {
		ret = DuplicateTxID
		return t.doResult(ret)
	}

	memberAccountBytes, err := t.getMemberAccount(stub, args[1])
	if err != nil {
		logger.Infof("GetState Error: %s", err.Error())
		ret = GetStateError
		return t.doResult(ret)
	}

	if memberAccountBytes == nil {
		ret = AccountNotFound
		return t.doResult(ret)
	}

	memberAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, memberAccount)

	balance, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		ret = InternalServerError
		return t.doResult(ret)
	}

	if !memberAccount.MemberWallet.ValidateBalance(balance) {
		ret = IncorrectBalance
		return t.doResult(ret)
	}

	if memberAccount.Frozen == false {
		ret = MethodNotAllowed
		return t.doResult(ret)
	}

	memberAccount.Frozen = false
	memberAccountBytes, _ = json.Marshal(memberAccount)
	err = stub.PutState(memberAccount.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		logger.Infof("PutState Error: ", err.Error())
		ret = InternalServerError
		return t.doResult(ret)
	}

	ret = t.recordTxFreezeMemberAccountLog(stub, model.TXID(args[0]), memberAccount.MemberWallet.WalletAddress, model.TxRecoverFrozenMember, OK)

	return t.doResult(ret)
}

func (t *SmartContract) createCompositeKey(objectType string, keyElements []string) (string, error) {

	const minKeyValue = ":"
	key := objectType + minKeyValue
	for _, ke := range keyElements {
		key += ke + minKeyValue
	}

	// logger.Infof("Created composite key: %s", key)

	return key, nil
}

func bytesToStruct(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		logger.Errorf("Error unmarshalling data for type %T", v)
		return err
	}

	return nil
}

func f2barr(f float64) []byte {
	str := strconv.FormatFloat(f, 'f', -1, 64)

	return []byte(str)
}
