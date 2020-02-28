package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	peer "github.com/hyperledger/fabric/protos/peer"
)

type HandlerFunc func(stub shim.ChaincodeStubInterface, args []string) peer.Response

type FuncMap struct {
	handlers map[string]HandlerFunc
}

func NewHandlerMap() *FuncMap {
	return &FuncMap{make(map[string]HandlerFunc)}
}

func (p *FuncMap) Add(name string, handler HandlerFunc) {
	p.handlers[name] = handler
}

func (p *FuncMap) Handle(stub shim.ChaincodeStubInterface, function string, args []string) peer.Response {
	handlerFunc := p.handlers[function]

	return handlerFunc(stub, args)

}
