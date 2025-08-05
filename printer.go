package main

import (
	"fmt"
)

// PrintWallet 打印完整钱包信息
func PrintWallet(wallet MultiChainWallet) {
	fmt.Println("\n🔐 钱包信息")
	fmt.Println("=============================================================")

	if wallet.Mnemonic != "" {
		fmt.Printf("助记词 (Mnemonic): %s\n", wallet.Mnemonic)
		fmt.Printf("派生路径 (Path):   %s\n", wallet.DerivePath)
		fmt.Println("-------------------------------------------------------------")
	}

	fmt.Printf("私钥 (Private Key): %s\n", wallet.PrivateKey)
	fmt.Printf("公钥 (Public Key):  %s\n", wallet.PublicKey)
	fmt.Println("-------------------------------------------------------------")
	fmt.Printf("🔹 Ethereum:  %s\n", wallet.EthAddress)
	fmt.Printf("🔹 Bitcoin:   %s\n", wallet.BtcAddress)
	fmt.Printf("🔹 BSC:       %s\n", wallet.BscAddress)
	fmt.Printf("🔹 Polygon:   %s\n", wallet.PolygonAddress)
	fmt.Printf("🔹 Tron:      %s\n", wallet.TronAddress)
	fmt.Println("=============================================================")
	fmt.Println("⚠️  请安全保存私钥和助记词!")
}

// PrintWalletSimple 打印简化钱包信息
func PrintWalletSimple(wallet MultiChainWallet) {
	fmt.Printf("\n💼 钱包 #%d\n", wallet.Index+1)
	if wallet.DerivePath != "" {
		fmt.Printf("路径: %s\n", wallet.DerivePath)
	}
	if wallet.Mnemonic != "" {
		fmt.Printf("助记词: %s\n", wallet.Mnemonic)
	} else {
		fmt.Printf("私钥: %s\n", wallet.PrivateKey)
	}
	fmt.Printf("ETH: %s\n", wallet.EthAddress)
	fmt.Printf("BTC: %s\n", wallet.BtcAddress)
	fmt.Printf("TRX: %s\n", wallet.TronAddress)
}
