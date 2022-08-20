package main

import (
	"fmt"
)

type API struct {
}

var BlockChainObj *BlockChain

func (a *API) createBlockChain(address string) {
	var err error
	if BlockChainObj == nil {
		BlockChainObj, err = NewBlockChain(address)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (a *API) GenKeyPair() {
	err := GenKeyPair()
	if err != nil {
		fmt.Println(err)
	}
}

func (a *API) printAllBlocks() {
	err := BlockChainObj.TraverseAllBlocks(func(block *Block) (bool, error) {
		fmt.Printf("\n------------------------------\n")
		fmt.Printf("Version : %d\n", block.Version)
		fmt.Printf("Hash : %x\n", block.GetBlockHash())
		fmt.Printf("PrevHash : %x\n", block.PrevBlockHash)
		fmt.Printf("MerkleRoot : %x\n", block.MerkleRoot)
		fmt.Printf("TimeStamp : %d\n", block.TimeStamp)
		fmt.Printf("Difficulty : %d\n", block.Difficulty)
		fmt.Printf("Nonce : %d\n", block.Nonce)
		fmt.Printf("Data : %s\n", string(block.Transactions[0].TxInputs[0].UnLockScript.PubKey)) //矿工写入的数据
		return false, nil
	})
	if err != nil {
		fmt.Println(err)
	}

}

func (a *API) getBalance(address string) {
	_, balance, err := BlockChainObj.GetAllUTXOs(address)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("balance:", balance, address)
}

func (a *API) send(from, to string, amount uint64, miner string) {
	err := WalletLocal.Send(from, to, amount, BlockChainObj, miner)
	if err != nil {
		fmt.Println("转账失败:", err)
		return
	}
	fmt.Println("转账成功:", amount, from, to)
}

func (a *API) listAddress() []string {
	return WalletLocal.ListAllAddr()
}

func (a *API) printTx() {
	err := BlockChainObj.TraverseAllBlocks(func(block *Block) (bool, error) {
		fmt.Println("\n+++++++++++++++++ 区块分割 +++++++++++++++")

		for _, tx := range block.Transactions {
			//直接打印交易
			fmt.Println(tx)
		}
		return false, nil
	})
	if err != nil {
		fmt.Println(err)
	}
}
