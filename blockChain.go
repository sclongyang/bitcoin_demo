package main

import (
	"bytes"
	"fmt"
)

const BlockHashBytesCount = 32

type BlockChain struct {
	Blocks        map[[BlockHashBytesCount]byte]*Block
	LastBlockHash [BlockHashBytesCount]byte
}

func NewBlockChain(minerAddr string) (*BlockChain, error) {
	chain := BlockChain{
		Blocks: make(map[[BlockHashBytesCount]byte]*Block, 5),
	}
	//add 创世块
	err := chain.AddBlock([]*Transaction{}, minerAddr, "Genesis block!!")
	return &chain, err
}

func (b *BlockChain) TraverseAllBlocks(handler func(block *Block) (bool, error)) error {
	block := b.Blocks[b.LastBlockHash]
	for {
		needReturn, err := handler(block)
		if err != nil || needReturn {
			return err
		}
		//退出条件
		if block.PrevBlockHash == nil {
			fmt.Println("区块链遍历结束!")
			break
		}
		var hash [BlockHashBytesCount]byte
		copy(hash[:], block.PrevBlockHash)
		block = BlockChainObj.Blocks[hash]
	}
	return nil
}

func (b *BlockChain) CheckTxs(txs []*Transaction) error {
	for _, tx := range txs {
		err := tx.Check(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BlockChain) AddBlock(txs []*Transaction, minnerAddr string, minerDesc string) error {
	err := b.CheckTxs(txs)
	if err != nil {
		return err
	}

	var prevBlockHash []byte
	lastBlock := b.GetLastBlock()
	if lastBlock != nil {
		prevBlockHash = lastBlock.GetBlockHash()
	}
	block, err := NewBlock(prevBlockHash, txs, minnerAddr, minerDesc)
	if err != nil {
		return err
	}
	var curBlockHash [BlockHashBytesCount]byte
	copy(curBlockHash[:], block.GetBlockHash())
	b.Blocks[curBlockHash] = block
	b.LastBlockHash = curBlockHash
	return nil
}

func (b *BlockChain) GetLastBlock() *Block {
	return b.Blocks[b.LastBlockHash]
}

//获取刚好 >= amount 对应的UTXOs
func (b *BlockChain) GetAllUTXOs(addr string) (UTXOs []*UTXO, utxosTotalAmount uint64, err error) {
	pubKeyHash, err := Addr2PubKeyHash(addr)
	if err != nil {
		return
	}
	curBlock := b.GetLastBlock()
	UTXOs = make([]*UTXO, 0)
	spentTxInput := make([]*TxInput, 0)
	for {
		for _, transaction := range curBlock.Transactions {
			txHash, err2 := transaction.GetTxHash()
			if err2 != nil {
				err = err2
				return
			}
			//处理outputs
			for i, output := range transaction.TxOutputs {
				if bytes.Equal(output.LockScript.PubKeyHash, pubKeyHash) {
					//先把output当做UTXO
					candidateUTXO := UTXO{
						TxHash:   txHash,
						Idx:      i,
						TxOutput: output,
					}
					idxOfSpentInput := -1
					for i, input := range spentTxInput {
						if candidateUTXO.IsMatchTxInput(input) {
							idxOfSpentInput = i
							break
						}
					}
					if idxOfSpentInput == -1 {
						UTXOs = append(UTXOs, &candidateUTXO)
						utxosTotalAmount += candidateUTXO.Value
					} else {
						spentTxInput = append(spentTxInput[:idxOfSpentInput], spentTxInput[idxOfSpentInput+1:]...)
					}
				}
			}

			//处理inputs
			spentTxInput = append(spentTxInput, transaction.TxInputs...)
		}

		if curBlock.PrevBlockHash == nil {
			break
		}
		var prevBlockId [BlockHashBytesCount]byte
		copy(prevBlockId[:], curBlock.PrevBlockHash)
		curBlock = b.Blocks[prevBlockId]
	}
	return
}
