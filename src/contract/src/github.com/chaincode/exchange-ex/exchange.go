package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/chaincode/model"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	peer "github.com/hyperledger/fabric/protos/peer"
)

var (
	logger     = shim.NewLogger("exchange")
	handlerMap = NewHandlerMap()
	txLimit    = model.NewTransactionLimit()
)

const (
	OK    = 200
	ERROR = 500
)

const (
	START_IDX            = 1
	TX_ID_OFFSET         = 1
	TX_ARGS_OFFSET       = 2
	REGISTER_MEMBER_ARGS = 7
	TRANSFER_COIN_ARGS   = 5
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
	idx := START_IDX
	i := 0

	type Result struct {
		UUID string `json:"uuid`
		Code string `json:"code"`
	}

	type ResultList struct {
		Results []*Result `json:"results"`
	}

	resultList := ResultList{}

	for idxArrayList < txArrayListCount {
		startIdx := idx + TX_ARGS_OFFSET
		id := idx + TX_ID_OFFSET
		uuididx := idx
		switch args[id] {
		case "1":
			txArgsCount = REGISTER_MEMBER_ARGS
		case "2":
			txArgsCount = TRANSFER_COIN_ARGS
		}

		idx += TX_ARGS_OFFSET + txArgsCount
		i = 0
		txArgs := make([]string, txArgsCount)
		for startIdx < idx {
			txArgs[i] = args[startIdx]
			startIdx++
			i++
		}

		logger.Infof("args[%d]: %s", id, args[id])

		result := new(Result)

		switch args[id] {
		case "1":
			result.Code = t.batchRegisterMember(stub, txArgs)
		case "2":
			result.Code = t.batchTransferCoin(stub, txArgs)
		}

		result.UUID = args[uuididx]
		resultList.Results = append(resultList.Results, result)

		idxArrayList++

	}

	eventPayload := "Test Event"
	if err := stub.SetEvent(args[1], []byte(eventPayload)); err != nil {
		return shim.Error(err.Error())
	}

	jsonList, _ := json.Marshal(resultList)

	return shim.Success(jsonList)
}

func (t *SmartContract) _batchProcess(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	logger.Infof("batchProcess Args: txCount(%s)", args[0])

	txArrayListCount, _ := strconv.Atoi(args[0])
	txArgsCount := 0
	idxArrayList := 0
	idx := 1
	i := 0

	type Result struct {
		Code        int    `json:"code"`
		Description string `json:"description"`
	}

	type ResultList struct {
		Results []*Result `json:"results"`
	}

	resultList := ResultList{}

	for idxArrayList < txArrayListCount {
		startIdx := idx + 2
		id := idx + 1
		switch args[id] {
		case "1":
			txArgsCount = 7
		case "2":
			txArgsCount = 5
		}

		idx += 2 + txArgsCount
		i = 0
		txArgs := make([]string, txArgsCount)
		for startIdx < idx {
			txArgs[i] = args[startIdx]
			startIdx++
			i++
		}

		logger.Infof("args[%d]: %s", id, args[id])

		result := new(Result)

		switch args[id] {
		case "1":
			result.Code, result.Description = t._batchRegisterMember(stub, txArgs)
		case "2":
			result.Code, result.Description = t._batchTransferCoin(stub, txArgs)
		}

		resultList.Results = append(resultList.Results, result)

		idxArrayList++

	}

	jsonList, _ := json.Marshal(resultList)

	return shim.Success(jsonList)
}

func (t *SmartContract) batchRegisterMember(stub shim.ChaincodeStubInterface, args []string) string {

	if len(args) != 7 {
		logger.Error("batchRegisterMember Incorrect arguments. Expecting 7 arguments")
		return "500"
	}

	logger.Infof(`batchRegisterMember Args: MemberId(%s), VSCode(%s), CountryCode(%s), 
	CurrencyCode(%s), MemberRole(%s), Wallet Address(%s), CreatedDate(%s)\n`,
		args[0], args[1], args[2], args[3], args[4], args[5], args[6])

	memberBytes, err := stub.GetState(args[5])
	if err != nil {
		return "500"
	}

	if memberBytes != nil {
		return "400"
	}

	account := new(model.MemberAccount)

	account.MemberId = args[0]
	account.VSCode = args[1]
	account.CountryCode = args[2]
	account.CurrencyCode = args[3]
	account.MemberRole = args[4]
	account.CreatedDate = args[6]
	account.MemberWallet.WalletAddress = args[5]
	account.MemberWallet.CoinBalance = 10000000000
	account.MemberWallet.CashBalance = 0
	account.MemberLevel = "TxLimit:" + "1"
	account.CustomOneTimeTransferLimit = 0
	account.CustomOneTimeWithdrawLimit = 0
	account.CustomOneDayTransferLimit = 0
	account.CustomOneDayWithdrawLimit = 0
	account.OneDayTransferSum = 0
	//account.OneDayTransferDate
	account.OneDayWithdrawSum = 0
	//account.OneDayWithdrawDate

	if err := account.Validate(); err != nil {
		return "500"
	}

	memberAsBytes, _ := json.Marshal(account)

	//@@ need to define exchange memberRole identifier
	// if account.MemberRole == "exchange" {
	// 	uid := fmt.Sprintf("%x", model.GenTxId(memberAsBytes))
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
		return "500"
	}

	return "201"
}

func (t *SmartContract) _batchRegisterMember(stub shim.ChaincodeStubInterface, args []string) (int, string) {

	if len(args) != 7 {
		logger.Error("batchRegisterMember Incorrect arguments. Expecting 7 arguments")
		return -1, "Incorrect arguments. Expecting 7 arguments"
	}

	logger.Infof(`batchRegisterMember Args: MemberId(%s), VSCode(%s), CountryCode(%s), 
	CurrencyCode(%s), MemberRole(%s), Wallet Address(%s), CreatedDate(%s)\n`,
		args[0], args[1], args[2], args[3], args[4], args[5], args[6])

	account := new(model.MemberAccount)

	account.MemberId = args[0]
	account.VSCode = args[1]
	account.CountryCode = args[2]
	account.CurrencyCode = args[3]
	account.MemberRole = args[4]
	account.CreatedDate = args[6]
	account.MemberWallet.WalletAddress = args[5]
	account.MemberWallet.CoinBalance = 10000000000
	account.MemberWallet.CashBalance = 0
	account.MemberLevel = "TxLimit:" + "1"
	account.CustomOneTimeTransferLimit = 0
	account.CustomOneTimeWithdrawLimit = 0
	account.CustomOneDayTransferLimit = 0
	account.CustomOneDayWithdrawLimit = 0
	account.OneDayTransferSum = 0
	//account.OneDayTransferDate
	account.OneDayWithdrawSum = 0
	//account.OneDayWithdrawDate

	if err := account.Validate(); err != nil {
		return -1, err.Error()
	}

	memberAsBytes, _ := json.Marshal(account)

	//@@ need to define exchange memberRole identifier
	// if account.MemberRole == "exchange" {
	// 	uid := fmt.Sprintf("%x", model.GenTxId(memberAsBytes))
	// 	key, err := stub.CreateCompositeKey("", []string{account.MemberWallet.WalletAddress, uid})
	// 	if err != nil {
	// 		return shim.Error(err.Error())
	// 	}

	// 	err = stub.PutState(key, memberAsBytes)
	// 	if err != nil {
	// 		return shim.Error("Failed to set member account")
	// 	}
	// } else {
	err := stub.PutState(account.MemberWallet.WalletAddress, memberAsBytes)
	if err != nil {
		return 0, "Failed to set member account"
	}

	return 0, "success"
}

func (t *SmartContract) registerMember(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 7 {
		logger.Error("registerMember Incorrect arguments. Expecting 7 arguments")
		return shim.Error("Incorrect arguments. Expecting 7 arguments")
	}

	logger.Infof(`registerMember Args: MemberId(%s), VSCode(%s), CountryCode(%s), 
	CurrencyCode(%s), MemberRole(%s), Wallet Address(%s), CreatedDate(%s)\n`,
		args[0], args[1], args[2], args[3], args[4], args[5], args[6])

	memberBytes, err := stub.GetState(args[5])
	if err != nil {
		return shim.Error(err.Error())
	}

	if memberBytes != nil {
		return shim.Error("Already registered")
	}

	account := new(model.MemberAccount)

	account.MemberId = args[0]
	account.VSCode = args[1]
	account.CountryCode = args[2]
	account.CurrencyCode = args[3]
	account.MemberRole = args[4]
	account.CreatedDate = args[6]
	account.MemberWallet.WalletAddress = args[5]
	account.MemberWallet.CoinBalance = 10000000000
	account.MemberWallet.CashBalance = 0
	account.MemberLevel = "TxLimit:" + "1"
	account.CustomOneTimeTransferLimit = 0
	account.CustomOneTimeWithdrawLimit = 0
	account.CustomOneDayTransferLimit = 0
	account.CustomOneDayWithdrawLimit = 0
	account.OneDayTransferSum = 0
	//account.OneDayTransferDate
	account.OneDayWithdrawSum = 0
	//account.OneDayWithdrawDate

	if err := account.Validate(); err != nil {
		return shim.Error(err.Error())
	}

	memberAsBytes, _ := json.Marshal(account)

	//@@ need to define exchange memberRole identifier
	// if account.MemberRole == "exchange" {
	// 	uid := fmt.Sprintf("%x", model.GenTxId(memberAsBytes))
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
		return shim.Error("Failed to set member account")
	}

	eventPayload := "Test Event"
	if err := stub.SetEvent(args[0], []byte(eventPayload)); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SmartContract) setTransactionLimit(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 5 {
		logger.Error("setTransactionLimit Incorrect arguments. Expecting 5 arguments")
		return shim.Error("Incorrect arguments. Expecting 5 arguments")
	}

	logger.Infof("setTransactionLimit Args: Level(%s), TransactionLimit(%s), WithdrawLimit(%s)\n", args[0], args[1], args[2])

	keyStr := "TxLimit:" + args[0]
	transactionLimitBytes, err := stub.GetState(keyStr)
	if err != nil {
		return shim.Error(err.Error())
	}

	transactionLimit := new(model.TransactionLimit)

	if transactionLimitBytes == nil {

		transactionLimit.Level = "TxLimit:" + args[0]
		transactionLimit.OneTimeTransferLimit, _ = strconv.ParseFloat(args[1], 64)
		transactionLimit.OneTimeWithdrawLimit, _ = strconv.ParseFloat(args[2], 64)
		transactionLimit.OneDayTransferLimit, _ = strconv.ParseFloat(args[3], 64)
		transactionLimit.OneDayWithdrawLimit, _ = strconv.ParseFloat(args[4], 64)
	} else {

		bytesToStruct(transactionLimitBytes, transactionLimit)

		oneTimeTransferLimit, _ := strconv.ParseFloat(args[1], 64)
		oneTimeWithdrawLimit, _ := strconv.ParseFloat(args[2], 64)
		oneDayTransferLimit, _ := strconv.ParseFloat(args[3], 64)
		oneDayWithdrawLimit, _ := strconv.ParseFloat(args[4], 64)

		if oneTimeTransferLimit != 0 {
			transactionLimit.OneTimeTransferLimit = oneTimeTransferLimit
		}
		if oneTimeWithdrawLimit != 0 {
			transactionLimit.OneTimeWithdrawLimit = oneTimeWithdrawLimit
		}
		if oneDayTransferLimit != 0 {
			transactionLimit.OneDayTransferLimit = oneDayTransferLimit
		}
		if oneDayWithdrawLimit != 0 {
			transactionLimit.OneDayWithdrawLimit = oneDayWithdrawLimit
		}
	}

	transactionLimitAsBytes, _ := json.Marshal(transactionLimit)
	err = stub.PutState(transactionLimit.Level, transactionLimitAsBytes)
	if err != nil {
		return shim.Error("Failed to set Transaction Limit")
	}

	return shim.Success(nil)
}

func (t *SmartContract) issueCoin(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		logger.Error("issueCoin Incorrect arguments. Expecting 2 arguments")
		return shim.Error("Incorrect arguments. Expecting 2 arguments")
	}

	walletAddress := args[0]
	amount, _ := strconv.ParseFloat(args[1], 64)

	memberAccountBytes, err := t.getMemberAccount(stub, walletAddress)
	if err != nil {
		return shim.Error(err.Error())
	}

	if memberAccountBytes == nil {
		return shim.Error("Account not found")
	}

	account := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, account)

	t.creditCoin(stub, account, amount)

	return shim.Success(nil)
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
		t.recordTransaction(stub, sr, model.TxSellCoin, model.TxSuccess)

		sellerAccount.DebitCash(sr.Fee)
		t.creditFee(stub, sellerAccount.MemberWallet.WalletAddress, sr.Fee)
		t.recordTransaction(stub, sr, model.TxSellCoin, model.TxSuccess)

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
		buyerAccount.CreditCoin(sr.Amount)

		t.recordTransaction(stub, sr, model.TxSellCoin, model.TxSuccess)

		buyerAccount.DebitCash(br.Amount)
		buyerAccount.DebitCash(br.Fee)
		t.creditFee(stub, buyerAccount.MemberWallet.WalletAddress, br.Fee)

		sellerAccount.CreditCash(br.Amount)

		t.recordTransaction(stub, br, model.TxBuyCoin, model.TxSuccess)

		t.updateState(stub, sellerAccount)
		t.updateState(stub, buyerAccount)
	}

	return shim.Success(nil)
}

func (t *SmartContract) batchTransferCoin(stub shim.ChaincodeStubInterface, args []string) string {

	if len(args) != 5 {
		logger.Error("transferCoin Incorrect arguments. Expecting 5 arguments")
		return "500"
	}

	logger.Infof("transferCoin Args: senderWalletAddress(%s), receiverWalletAddress(%s), coinAmount(%s), fee(%s), txFlag(%s)\n",
		args[0], args[1], args[2], args[3], args[4])

	r := new(model.Remittance)
	r.SenderWalletAddress = args[0]
	r.ReceiverWalletAddress = args[1]
	r.Amount, _ = strconv.ParseFloat(args[2], 64)
	r.Fee, _ = strconv.ParseFloat(args[3], 64)
	r.TxFlag = args[4]

	if err := r.Validate(); err != nil {
		return "500"
	}

	memberAccountBytes, err := t.getMemberAccount(stub, r.SenderWalletAddress)

	if err != nil {
		return "500"
	}

	if memberAccountBytes == nil {
		return "500"
	}

	senderAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, senderAccount)

	if senderAccount.Frozen {
		return "500"
	}

	if senderAccount.MemberWallet.CoinBalance < (r.Amount + r.Fee) {
		return "500"
	}

	// deprecated
	// if senderAccount.MemberWallet.CashBalance < r.Fee {
	// 	return shim.Error("Insufficient Cash Balance")
	// }

	if senderAccount.MemberRole != "exchange" {
		if senderAccount.CustomOneTimeTransferLimit != 0 {
			if senderAccount.CustomOneTimeTransferLimit < r.Amount {
				return "500"
			}
		} else {
			txLimitBytes, _ := t.getTransactionLimit(stub, senderAccount.MemberLevel)
			txLimit := new(model.TransactionLimit)
			bytesToStruct(txLimitBytes, txLimit)

			if txLimit.OneTimeTransferLimit < r.Amount {
				return "500"
			}
		}

		tm := time.Now()
		txDate := model.TxDate{tm.Year(), tm.Month(), tm.Day()}
		if txDate != senderAccount.OneDayTransferDate {
			senderAccount.OneDayTransferDate = txDate
			senderAccount.OneDayTransferSum = 0
		}

		if senderAccount.CustomOneDayTransferLimit != 0 {
			if senderAccount.CustomOneDayTransferLimit < senderAccount.OneDayTransferSum+r.Amount {
				return "500"
			}
		} else {
			txLimitBytes, _ := t.getTransactionLimit(stub, senderAccount.MemberLevel)
			txLimit := new(model.TransactionLimit)
			bytesToStruct(txLimitBytes, txLimit)

			if txLimit.OneDayTransferLimit < senderAccount.OneDayTransferSum+r.Amount {
				return "500"
			}
		}
	}

	senderAccount.OneDayTransferSum += r.Amount

	memberAccountBytes, _ = json.Marshal(senderAccount)

	err = stub.PutState(senderAccount.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		return "500"
	}

	memberAccountBytes, err = t.getMemberAccount(stub, r.ReceiverWalletAddress)

	if err != nil {
		return "500"
	}

	if memberAccountBytes == nil {
		return "500"
	}
	receiverAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, receiverAccount)

	if receiverAccount.Frozen {
		return "500"
	}

	t.debitCoin(stub, senderAccount, r.Amount+r.Fee)
	//t.debitCash(stub, senderAccount, r.Fee) //deprecated
	t.creditFee(stub, senderAccount.MemberWallet.WalletAddress, r.Fee)
	t.creditCoin(stub, receiverAccount, r.Amount)

	t.recordTransaction(stub, r, model.TxTransferCoin, model.TxSuccess)

	return "200"
}

func (t *SmartContract) _batchTransferCoin(stub shim.ChaincodeStubInterface, args []string) (int, string) {

	if len(args) != 5 {
		logger.Error("transferCoin Incorrect arguments. Expecting 5 arguments")
		return -1, "Incorrect arguments. Expecting 5 arguments"
	}

	logger.Infof("transferCoin Args: senderWalletAddress(%s), receiverWalletAddress(%s), coinAmount(%s), fee(%s), txFlag(%s)\n",
		args[0], args[1], args[2], args[3], args[4])

	r := new(model.Remittance)
	r.SenderWalletAddress = args[0]
	r.ReceiverWalletAddress = args[1]
	r.Amount, _ = strconv.ParseFloat(args[2], 64)
	r.Fee, _ = strconv.ParseFloat(args[3], 64)
	r.TxFlag = args[4]

	if err := r.Validate(); err != nil {
		return -1, err.Error()
	}

	memberAccountBytes, err := t.getMemberAccount(stub, r.SenderWalletAddress)

	if err != nil {
		return -1, err.Error()
	}

	if memberAccountBytes == nil {
		return -1, "Sender Account not found"
	}

	senderAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, senderAccount)

	if senderAccount.Frozen {
		return -1, "Cannot execute a transaction into frozen member"
	}

	if senderAccount.MemberWallet.CoinBalance < (r.Amount + r.Fee) {
		return -1, "Insufficient Coin Balance"
	}

	// deprecated
	// if senderAccount.MemberWallet.CashBalance < r.Fee {
	// 	return shim.Error("Insufficient Cash Balance")
	// }

	if senderAccount.MemberRole != "exchange" {
		if senderAccount.CustomOneTimeTransferLimit != 0 {
			if senderAccount.CustomOneTimeTransferLimit < r.Amount {
				return -1, "One Time Transfer Limit Over"
			}
		} else {
			txLimitBytes, _ := t.getTransactionLimit(stub, senderAccount.MemberLevel)
			txLimit := new(model.TransactionLimit)
			bytesToStruct(txLimitBytes, txLimit)

			if txLimit.OneTimeTransferLimit < r.Amount {
				return -1, "One Time Transfer Limit Over"
			}
		}

		tm := time.Now()
		txDate := model.TxDate{tm.Year(), tm.Month(), tm.Day()}
		if txDate != senderAccount.OneDayTransferDate {
			senderAccount.OneDayTransferDate = txDate
			senderAccount.OneDayTransferSum = 0
		}

		if senderAccount.CustomOneDayTransferLimit != 0 {
			if senderAccount.CustomOneDayTransferLimit < senderAccount.OneDayTransferSum+r.Amount {
				return -1, "One Day Transfer Limit Over"
			}
		} else {
			txLimitBytes, _ := t.getTransactionLimit(stub, senderAccount.MemberLevel)
			txLimit := new(model.TransactionLimit)
			bytesToStruct(txLimitBytes, txLimit)

			if txLimit.OneDayTransferLimit < senderAccount.OneDayTransferSum+r.Amount {
				return -1, "One Day Transfer Limit Over"
			}
		}
	}

	senderAccount.OneDayTransferSum += r.Amount

	memberAccountBytes, _ = json.Marshal(senderAccount)

	err = stub.PutState(senderAccount.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		return -1, err.Error()
	}

	memberAccountBytes, err = t.getMemberAccount(stub, r.ReceiverWalletAddress)

	if err != nil {
		return -1, err.Error()
	}

	if memberAccountBytes == nil {
		return -1, "Receiver Account not found"
	}
	receiverAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, receiverAccount)

	if receiverAccount.Frozen {
		return -1, "Cannot execute a transaction into frozen member"
	}

	t.debitCoin(stub, senderAccount, r.Amount+r.Fee)
	//t.debitCash(stub, senderAccount, r.Fee) //deprecated
	t.creditFee(stub, senderAccount.MemberWallet.WalletAddress, r.Fee)
	t.creditCoin(stub, receiverAccount, r.Amount)

	t.recordTransaction(stub, r, model.TxTransferCoin, model.TxSuccess)

	return 0, "success"
}

func (t *SmartContract) transferCoin(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 5 {
		logger.Error("transferCoin Incorrect arguments. Expecting 5 arguments")
		return shim.Error("Incorrect arguments. Expecting 5 arguments")
	}

	logger.Infof("transferCoin Args: senderWalletAddress(%s), receiverWalletAddress(%s), coinAmount(%s), fee(%s), txFlag(%s)\n",
		args[0], args[1], args[2], args[3], args[4])

	r := new(model.Remittance)
	r.SenderWalletAddress = args[0]
	r.ReceiverWalletAddress = args[1]
	r.Amount, _ = strconv.ParseFloat(args[2], 64)
	r.Fee, _ = strconv.ParseFloat(args[3], 64)
	r.TxFlag = args[4]

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

	if senderAccount.MemberWallet.CoinBalance < (r.Amount + r.Fee) {
		return shim.Error("Insufficient Coin Balance")
	}

	// deprecated
	// if senderAccount.MemberWallet.CashBalance < r.Fee {
	// 	return shim.Error("Insufficient Cash Balance")
	// }

	if senderAccount.MemberRole != "exchange" {
		if senderAccount.CustomOneTimeTransferLimit != 0 {
			if senderAccount.CustomOneTimeTransferLimit < r.Amount {
				return shim.Error("One Time Transfer Limit Over")
			}
		} else {
			txLimitBytes, _ := t.getTransactionLimit(stub, senderAccount.MemberLevel)
			txLimit := new(model.TransactionLimit)
			bytesToStruct(txLimitBytes, txLimit)

			if txLimit.OneTimeTransferLimit < r.Amount {
				return shim.Error("One Time Transfer Limit Over")
			}
		}

		tm := time.Now()
		txDate := model.TxDate{tm.Year(), tm.Month(), tm.Day()}
		if txDate != senderAccount.OneDayTransferDate {
			senderAccount.OneDayTransferDate = txDate
			senderAccount.OneDayTransferSum = 0
		}

		if senderAccount.CustomOneDayTransferLimit != 0 {
			if senderAccount.CustomOneDayTransferLimit < senderAccount.OneDayTransferSum+r.Amount {
				return shim.Error("One Day Transfer Limit Over")
			}
		} else {
			txLimitBytes, _ := t.getTransactionLimit(stub, senderAccount.MemberLevel)
			txLimit := new(model.TransactionLimit)
			bytesToStruct(txLimitBytes, txLimit)

			if txLimit.OneDayTransferLimit < senderAccount.OneDayTransferSum+r.Amount {
				return shim.Error("One Day Transfer Limit Over")
			}
		}
	}

	senderAccount.OneDayTransferSum += r.Amount

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
	//t.debitCash(stub, senderAccount, r.Fee) //deprecated
	t.creditFee(stub, senderAccount.MemberWallet.WalletAddress, r.Fee)
	t.creditCoin(stub, receiverAccount, r.Amount)

	t.recordTransaction(stub, r, model.TxTransferCoin, model.TxSuccess)

	eventPayload := "Test Event"
	if err := stub.SetEvent(args[2], []byte(eventPayload)); err != nil {
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

	t.recordTransaction(stub, r, model.TxTransferCash, model.TxSuccess)

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

	uid = fmt.Sprintf("%x", model.GenTxId(memberAccountBytes))
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
	uid = fmt.Sprintf("%x", model.GenTxId(memberAccountBytes))
	key, err = stub.CreateCompositeKey("", []string{a.MemberWallet.WalletAddress, uid})
	if err != nil {
		return uid, err
	}
	err = stub.PutState(key, memberAccountBytes)
	//err := stub.PutState(a.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		return uid, err
	}
	return uid, err
}

func (t *SmartContract) creditCoin(stub shim.ChaincodeStubInterface, a *model.MemberAccount, amount float64) error {

	a.CreditCoin(amount)

	memberAccountBytes, _ := json.Marshal(a)

	err := stub.PutState(a.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
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
	a.CreditCoin(amount)

	memberAccountBytes, _ := json.Marshal(a)
	uid = fmt.Sprintf("%x", model.GenTxId(memberAccountBytes))
	key, err = stub.CreateCompositeKey("", []string{a.MemberWallet.WalletAddress, uid})
	if err != nil {
		return uid, err
	}
	err = stub.PutState(key, memberAccountBytes)
	//err := stub.PutState(a.MemberWallet.WalletAddress, memberAccountBytes)
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
	uid = fmt.Sprintf("%x", model.GenTxId(memberAccountBytes))
	key, err = stub.CreateCompositeKey("", []string{a.MemberWallet.WalletAddress, uid})
	if err != nil {
		return uid, err
	}
	err = stub.PutState(key, memberAccountBytes)
	//err := stub.PutState(a.MemberWallet.WalletAddress, memberAccountBytes)
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
	uid = fmt.Sprintf("%x", model.GenTxId(memberAccountBytes))
	key, err = stub.CreateCompositeKey("", []string{a.MemberWallet.WalletAddress, uid})
	if err != nil {
		return uid, err
	}

	//err := stub.PutState(a.MemberWallet.WalletAddress, memberAccountBytes)
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
		return err
	}

	tx, _ := model.CreateTransactionFee("exchange_platform", amount, model.TxFee, model.TxSuccess)
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
	//buffer.WriteString("]")

	jsonList, _ := json.Marshal(memberAccountList)
	logger.Infof("member account List : %s", jsonList)

	//return shim.Success(buffer.Bytes())
	return shim.Success(jsonList)
}

func (t *SmartContract) getTransactionLimit(stub shim.ChaincodeStubInterface, level string) ([]byte, error) {

	if level == "" {
		logger.Error("getTransactionLimit Incorrect arguments. Expecting 1 arguments")
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

	t.creditCoin(stub, memberAccount, coinAmount)

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

	if len(args) != 1 {
		logger.Error("setMemberLevel Incorrect arguments. Expecting 1 arguments")
		return shim.Error("Incorrect arguments. Expecting 1 arguments")
	}

	logger.Infof("setMemberLevel Args: %s\n", args[0])

	memberAccountBytes, err := t.getMemberAccount(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	if memberAccountBytes == nil {
		return shim.Error("Member Accout not found")
	}

	memberAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, memberAccount)

	if memberAccount.MemberLevel == args[0] {
		return shim.Error("Already set")
	}

	memberAccount.MemberLevel = args[0]
	memberAccountBytes, _ = json.Marshal(memberAccount)
	err = stub.PutState(memberAccount.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		return shim.Error("PutState error")
	}

	return shim.Success(nil)
}

func (t *SmartContract) freezeMemberAccount(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		logger.Error("freezeMemberAccount Incorrect arguments. Expecting 1 arguments")
		return shim.Error("Incorrect arguments. Expecting 1 arguments")
	}

	logger.Infof("freezeMemberAccount Args: %s\n", args[0])

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
		return shim.Error("Already frozen member")
	}

	memberAccount.Frozen = true
	memberAccountBytes, _ = json.Marshal(memberAccount)
	err = stub.PutState(memberAccount.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		return shim.Error("PutState error")
	}

	return shim.Success(nil)
}

func (t *SmartContract) recoverFrozenMemberAccount(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		logger.Error("recoverFrozenMemberAccount Incorrect arguments. Expecting 1 arguments")
		return shim.Error("Incorrect arguments. Expecting 1 arguments")
	}

	logger.Infof("recoverFrozenMemberAccount Args: %s\n", args[0])

	memberAccountBytes, err := t.getMemberAccount(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	if memberAccountBytes == nil {
		return shim.Error("Member Accout not found")
	}

	memberAccount := new(model.MemberAccount)
	bytesToStruct(memberAccountBytes, memberAccount)

	if memberAccount.Frozen == false {
		return shim.Error("Normal member cannot recover")
	}

	memberAccount.Frozen = false
	memberAccountBytes, _ = json.Marshal(memberAccount)
	err = stub.PutState(memberAccount.MemberWallet.WalletAddress, memberAccountBytes)
	if err != nil {
		return shim.Error("PutState error")
	}

	return shim.Success(nil)
}

func (t *SmartContract) recordTransaction(stub shim.ChaincodeStubInterface, r *model.Remittance,
	status model.TxStatus, failureCode model.TxFailureCode) peer.Response {

	tx, _ := model.CreateTransaction(r, status, failureCode)
	txBytes, err := json.Marshal(tx)
	if err != nil {
		return shim.Error("Error marshalling")
	}

	key, _ := t.createCompositeKey(tx.GetObjectType(), []string{r.SenderWalletAddress, r.ReceiverWalletAddress, string(tx.TxId)})
	logger.Infof("key : %s, tx : %v", key, tx)
	err = stub.PutState(key, txBytes)
	if err != nil {
		logger.Infof("PutState Error : %s", err.Error())
		return shim.Error("PutState Error")
	}

	return shim.Success(nil)
}

func (t *SmartContract) createCompositeKey(objectType string, keyElements []string) (string, error) {

	const minKeyValue = ":"
	key := objectType + minKeyValue
	for _, ke := range keyElements {
		key += ke + minKeyValue
	}

	//logger.Infof("Created composite key: %s", key)

	return key, nil
}

func bytesToStruct(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		logger.Errorf("Error unmarshalling data for type %T", v)
		return err
	}

	return nil
}

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

func f2barr(f float64) []byte {
	str := strconv.FormatFloat(f, 'f', -1, 64)

	return []byte(str)
}
