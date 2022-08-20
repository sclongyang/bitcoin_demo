package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"math/big"
	"strings"
	"time"
)

type Transaction struct {
	TxInputs  []TxInput //不使用引用是为了方便copy
	TxOutputs []*TxOutput
	TimeStamp int64
}

type TxInput struct {
	FromTxHash   []byte
	IdxOfUTXO    int
	UnLockScript UnLockScriptStruct //结构体更简单, 代替了二进制
}

type TxOutput struct {
	Value      uint64 //聪
	LockScript LockScriptStruct
}

type UnLockScriptStruct struct {
	Sig    *Signature
	PubKey []byte
}

type Signature struct {
	R *big.Int
	S *big.Int
}

type LockScriptStruct struct {
	PubKeyHash []byte
}

type UTXO struct {
	TxHash []byte
	Idx    int
	*TxOutput
}

func (u *UTXO) IsMatchTxInput(txInput *TxInput) bool {
	return bytes.Equal(u.TxHash, txInput.FromTxHash) && u.Idx == txInput.IdxOfUTXO
}

func NewCoinBaseTx(targetAddr string, value uint64, desc string) (tx *Transaction, err error) {
	txInput := TxInput{
		FromTxHash: nil,
		IdxOfUTXO:  -1,
		UnLockScript: UnLockScriptStruct{
			Sig:    nil,
			PubKey: []byte(desc), //coinBase的此字段用来存放desc
		},
	}

	//地址转 pubkeyHash
	pubkeyHash, err := Addr2PubKeyHash(targetAddr)
	if err != nil {
		return
	}

	txOutput := TxOutput{
		Value: value,
		LockScript: LockScriptStruct{
			PubKeyHash: pubkeyHash,
		},
	}

	tx = &Transaction{
		TxInputs:  []TxInput{txInput},
		TxOutputs: []*TxOutput{&txOutput},
		TimeStamp: time.Now().Unix(),
	}
	return
}

func NewTx(fromAddr string, targetAddr string, value uint64, chain *BlockChain, wallet *Wallet) (*Transaction, error) {
	allUtxos, utxosTotalAmount, err := chain.GetAllUTXOs(fromAddr)
	if err != nil {
		return nil, err
	}
	if utxosTotalAmount < value {
		return nil, fmt.Errorf("地址:%s 余额:%v 不足, 期望为:%v", fromAddr, utxosTotalAmount, value)
	}

	//根据amount筛选utxo
	var totalAmount uint64 = 0
	utxos := []*UTXO{}
	for _, utxo := range allUtxos {
		totalAmount += utxo.TxOutput.Value
		utxos = append(utxos, utxo)
		if totalAmount >= value {
			break
		}
	}

	if keypair, ok := wallet.KeyPairs[fromAddr]; ok {
		txInputs := make([]TxInput, 0)
		for _, utxo := range utxos {
			txInputs = append(txInputs, TxInput{
				FromTxHash: utxo.TxHash,
				IdxOfUTXO:  utxo.Idx,
				UnLockScript: UnLockScriptStruct{
					Sig:    nil,
					PubKey: nil,
				},
			})
		}

		targetPubkeyHash, err := Addr2PubKeyHash(targetAddr)
		if err != nil {
			return nil, err
		}
		fromAddrPubkeyHash, err := Addr2PubKeyHash(fromAddr)
		if err != nil {
			return nil, err
		}
		//target output
		targetOutput := &TxOutput{
			Value: value,
			LockScript: LockScriptStruct{
				PubKeyHash: targetPubkeyHash,
			},
		}
		//找零output
		changeOutput := &TxOutput{
			Value: totalAmount - value,
			LockScript: LockScriptStruct{
				PubKeyHash: fromAddrPubkeyHash,
			},
		}

		txOutputs := []*TxOutput{targetOutput, changeOutput}

		tx := &Transaction{
			TxInputs:  txInputs,
			TxOutputs: txOutputs,
			TimeStamp: time.Now().Unix(),
		}
		txHash, err := tx.GetTxHash()
		if err != nil {
			return nil, err
		}

		//签名
		r, s, err := ecdsa.Sign(rand.Reader, keypair.PrivKey, txHash)
		if err != nil {
			return nil, err
		}
		for i, _ := range tx.TxInputs {
			//input.UnLockScript.Sig = &Signature{
			tx.TxInputs[i].UnLockScript.Sig = &Signature{
				R: r,
				S: s,
			}
			//input.UnLockScript.PubKey = keypair.PubKey
			tx.TxInputs[i].UnLockScript.PubKey = keypair.PubKey
		}
		return tx, nil

	} else {
		return nil, fmt.Errorf("钱包不包含该 fromAddr:", fromAddr)
	}
}

func (t *Transaction) Check(chain *BlockChain) error {
	//是否coinbase
	if len(t.TxInputs) == 1 && t.TxInputs[0].FromTxHash == nil && t.TxInputs[0].IdxOfUTXO == -1 {
		return nil
	}
	var totalInputAmount uint64
	var totalOutputAmount uint64
	txCopy := *t
	//需要对TxInputs进行深度拷贝
	txCopy.TxInputs = make([]TxInput, len(t.TxInputs))
	copy(txCopy.TxInputs, t.TxInputs)
	//截断后再计算hash
	for i, _ := range txCopy.TxInputs {
		txCopy.TxInputs[i].UnLockScript.PubKey = nil
		txCopy.TxInputs[i].UnLockScript.Sig = nil
		//txInput.UnLockScript.PubKey = nil
		//txInput.UnLockScript.Sig = nil
	}
	txHash, err := txCopy.GetTxHash()
	if err != nil {
		return err
	}
	for _, input := range t.TxInputs {
		//inputs是否都是UTXO
		addr := PubKey2Addr(input.UnLockScript.PubKey)
		utxos, _, err := chain.GetAllUTXOs(addr)
		if err != nil {
			return err
		}
		idx := -1
		var utxoOfInput *UTXO
		for i, utxo := range utxos {
			if utxo.IsMatchTxInput(&input) {
				idx = i
				utxoOfInput = utxo
				break
			}
		}
		if idx == -1 {
			return fmt.Errorf("txInput:%v, %v 不是UTXO", input.FromTxHash, input.IdxOfUTXO)
		} else {
			utxos = append(utxos[:idx], utxos[idx+1:]...) //删除, 防止有两个相同的input指向同一个UTXO
		}

		//解锁
		pubkeyHash := Hash160(input.UnLockScript.PubKey)
		if !bytes.Equal(pubkeyHash, utxoOfInput.LockScript.PubKeyHash) {
			return fmt.Errorf("txinput pubkeyHash not eq to utxo pubkeyHash")
		}
		publicKey, err := PubKeyBytes2PubKey(input.UnLockScript.PubKey)
		if err != nil {
			return err
		}
		if !ecdsa.Verify(publicKey, txHash, input.UnLockScript.Sig.R, input.UnLockScript.Sig.S) {
			return fmt.Errorf("input 签名不正确")
		}

		//额度
		totalInputAmount += utxoOfInput.Value
	}

	for _, output := range t.TxOutputs {
		totalOutputAmount += output.Value
	}
	if totalInputAmount < totalOutputAmount {
		return fmt.Errorf("txInputs total amount < txOutput total amount")
	}
	return nil
}

func (t *Transaction) GetTxHash() ([]byte, error) {
	var buffer bytes.Buffer
	var encoder = gob.NewEncoder(&buffer)
	err := encoder.Encode(t)
	if err != nil {
		fmt.Println("encode err:", err)
		return nil, err
	}
	return Hash256(buffer.Bytes()), nil
}

func (tx *Transaction) String() string {
	var lines []string

	hash, _ := tx.GetTxHash()
	lines = append(lines, fmt.Sprintf("--- txHash %x:", hash))

	for i, input := range tx.TxInputs {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.FromTxHash))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.IdxOfUTXO))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.UnLockScript.Sig))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.UnLockScript.PubKey))
	}

	for i, output := range tx.TxOutputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %v", output.Value))
		lines = append(lines, fmt.Sprintf("       pubkeyhash: %x", output.LockScript.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}
