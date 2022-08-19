package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"time"
)

type Block struct {
	Version       uint
	PrevBlockHash []byte
	MerkleRoot    []byte
	TimeStamp     int64
	Difficulty    uint //简单起见, 代表前面有几个0
	Nonce         uint64
	Transactions  []*Transaction
}

func NewBlock(prevBlockHash []byte, txs []*Transaction, minerAddr string, minerDesc string) (*Block, error) {
	//coinbase
	coinBaseTx, err := NewCoinBaseTx(minerAddr, 50, minerDesc)
	if err != nil {
		return nil, err
	}
	txs = append([]*Transaction{coinBaseTx}, txs...)
	//mekle root
	mekleRoot, err := genMekleRoot(txs)
	if err != nil {
		return nil, err
	}

	var block = &Block{
		Version:       0,
		PrevBlockHash: prevBlockHash,
		MerkleRoot:    mekleRoot,
		TimeStamp:     time.Now().Unix(),
		Difficulty:    0,
		Nonce:         0,
		Transactions:  txs,
	}
	err = block.SetNonce()
	return block, err
}

func (b *Block) GetBlockHash() []byte {
	byteArray := [][]byte{
		Convert2Bytes(b.Version),
		b.PrevBlockHash,
		b.MerkleRoot,
		Convert2Bytes(b.TimeStamp),
		Convert2Bytes(b.Difficulty),
		Convert2Bytes(b.Nonce),
	}
	joinedData := bytes.Join(byteArray, []byte{})
	result := sha256.Sum256(joinedData)
	return result[:]
}

func (b *Block) SetNonce() error {
	b.Nonce = 0
	var targetHash = big.NewInt(1)
	temBigInt := big.NewInt(0)
	targetHash.Lsh(targetHash, 256-b.Difficulty)
	for b.Nonce < math.MaxUint64 {
		if temBigInt.SetBytes(b.GetBlockHash()[:]).Cmp(targetHash) == -1 {
			return nil
		}
		b.Nonce++
	}
	return fmt.Errorf("挖矿失败, 找不到合适的nonce")
}

func genMekleRoot(txs []*Transaction) (mekleRoot []byte, err error) {
	var tempHashArray [][]byte
	if len(txs) < 1 {
		err = fmt.Errorf("transaction len must >= 1")
		return
	}
	for i := 0; ; i++ {
		hash, err := txs[2*i].GetTxHash()
		if err != nil {
			return
		}
		tempHashArray = append(tempHashArray, hash)

		if 2*i+1 == len(txs) {
			//copy
			tempHashArray = append(tempHashArray, hash)
		} else {
			hash2, err := txs[2*i+1].GetTxHash()
			if err != nil {
				return
			}
			tempHashArray = append(tempHashArray, hash2)
		}

		if 2*i+2 == len(txs) {
			break
		}
	}

	var lastIdx = len(tempHashArray) - 1
	var curIdx = 0
	var second []byte
	for lastIdx > 1 {
		for i := 0; ; i++ {
			curIdx = 2 * i
			if curIdx == lastIdx+1 {
				break
			}
			if curIdx == lastIdx {
				//copy
				second = tempHashArray[curIdx]
			} else {
				second = tempHashArray[curIdx+1]
			}

			tempHashArray[i] = Hash256(bytes.Join([][]byte{tempHashArray[curIdx], second}, []byte{}))
			lastIdx = i
		}
	}
	return tempHashArray[0], nil
}
