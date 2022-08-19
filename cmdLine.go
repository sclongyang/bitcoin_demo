package main

import (
	"fmt"
	"os"
	"strconv"
)

//处理用户输入命令，完成具体函数的调用
//cli : command line 命令行
type CLI struct {
	//不需要字段
}

//使用说明，帮助用户正确使用
const Usage = `
正确使用方法：
	./blockchain create <地址> "创建区块链"	
	./blockchain print "打印区块链"
	./blockchain getBalance <地址> "获取余额"
	./blockchain send <FROM> <TO> <AMOUNT> <MINER>
	./blockchain listAddress "列举所有的钱包地址"
	./blockchain printTx "打印区块的所有交易"
`

//负责解析命令的方法
func (cli *CLI) Run() {
	cmds := os.Args
	//用户至少输入两个参数
	if len(cmds) < 2 {
		fmt.Println("输入参数无效，请检查!")
		fmt.Println(Usage)
		return
	}

	switch cmds[1] {
	case "createBlockChain":
		fmt.Println("创建区块被调用!")
		if len(cmds) != 3 {
			fmt.Println("输入参数无效，请检查!")
			return
		}
		address := cmds[2]
		cli.createBlockChain(address)
	case "print":
		fmt.Println("打印区块链被调用!")
		cli.print()
	case "getBalance":
		fmt.Println("获取余额命令被调用!")
		if len(cmds) != 3 {
			fmt.Println("输入参数无效，请检查!")
			return
		}
		address := cmds[2] //需要检验个数
		cli.getBalance(address)
	case "send":
		fmt.Println("send命令被调用")
		if len(cmds) != 6 {
			fmt.Println("输入参数无效，请检查!")
			return
		}

		from := cmds[2]
		to := cmds[3]
		//这个是金额，float64，命令接收都是字符串，需要转换
		amount, _ := strconv.ParseUint(cmds[4], 10, 64)
		miner := cmds[5]
		cli.send(from, to, amount, miner)
	case "listAddress":
		fmt.Println("listAddress 被调用")
		cli.listAddress()
	case "printTx":
		cli.printTx()
	default:
		fmt.Println("输入参数无效，请检查!")
		fmt.Println(Usage)
	}
}

var BlockChainObj *BlockChain

func (cli *CLI) createBlockChain(address string) {
	var err error
	if BlockChainObj == nil {
		BlockChainObj, err = NewBlockChain(address)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (cli *CLI) print() {
	err := BlockChainObj.TraverseAllBlocks(func(block *Block) (bool, error) {
		fmt.Printf("\n++++++++++++++++++++++\n")
		fmt.Printf("Version : %d\n", block.Version)
		fmt.Printf("PrevHash : %x\n", block.PrevBlockHash)
		fmt.Printf("MerkleRoot : %x\n", block.MerkleRoot)
		fmt.Printf("TimeStamp : %d\n", block.TimeStamp)
		fmt.Printf("Bits : %d\n", block.Difficulty)
		fmt.Printf("Nonce : %d\n", block.Nonce)
		fmt.Printf("Hash : %x\n", block.GetBlockHash())
		fmt.Printf("Data : %s\n", string(block.Transactions[0].TxInputs[0].UnLockScript.PubKey)) //矿工写入的数据
		return false, nil
	})
	if err != nil {
		fmt.Println(err)
	}

}

func (cli *CLI) getBalance(address string) {
	_, balance, err := BlockChainObj.GetAllUTXOs(address)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("balance:", balance)
}

func (cli *CLI) send(from, to string, amount uint64, miner string) {
	err := WalletLocal.Send(from, to, amount, BlockChainObj, miner)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("转账成功")
}

func (cli *CLI) listAddress() {
	WalletLocal.ListAllAddr()
}

func (cli *CLI) printTx() {
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
