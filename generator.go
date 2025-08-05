package main

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/ripemd160"
)

// GenerationResult 生成结果
type GenerationResult struct {
	Wallets       []MultiChainWallet
	Duration      time.Duration
	AvgTime       time.Duration
	WalletsPerSec float64
}

// WalletGenerator 钱包生成器
type WalletGenerator struct {
	config *Config
}

// NewWalletGenerator 创建钱包生成器
func NewWalletGenerator(config *Config) *WalletGenerator {
	return &WalletGenerator{
		config: config,
	}
}

// GenerateWallets 生成钱包的主函数
func (wg *WalletGenerator) GenerateWallets(config GeneratorConfig) []MultiChainWallet {
	if config.ConcurrentMode && config.Count > 1 {
		return wg.generateWalletsConcurrent(config)
	}
	return wg.generateWalletsSequential(config)
}

// GenerateBatch 批量生成钱包并返回结果统计
func (wg *WalletGenerator) GenerateBatch(config GeneratorConfig) *GenerationResult {
	start := time.Now()
	wallets := wg.GenerateWallets(config)
	duration := time.Since(start)

	var avgTime time.Duration
	if len(wallets) > 0 {
		avgTime = duration / time.Duration(len(wallets))
	}

	walletsPerSec := float64(len(wallets)) / duration.Seconds()

	return &GenerationResult{
		Wallets:       wallets,
		Duration:      duration,
		AvgTime:       avgTime,
		WalletsPerSec: walletsPerSec,
	}
}

// generateWalletsSequential 顺序生成钱包
func (wg *WalletGenerator) generateWalletsSequential(config GeneratorConfig) []MultiChainWallet {
	var wallets []MultiChainWallet
	var masterMnemonic string

	// 如果使用助记词且只生成一个钱包，生成新助记词
	if config.UseMnemonic && config.Count == 1 && config.Mnemonic == "" {
		entropy, _ := bip39.NewEntropy(128)
		masterMnemonic, _ = bip39.NewMnemonic(entropy)
	} else if config.Mnemonic != "" {
		masterMnemonic = config.Mnemonic
	}

	for i := 0; i < config.Count; i++ {
		var wallet MultiChainWallet
		var err error

		if config.UseMnemonic {
			wallet, err = wg.GenerateWalletFromMnemonic(masterMnemonic, i)
		} else {
			wallet, err = wg.GenerateRandomWallet()
		}

		if err != nil {
			log.Printf("生成第 %d 个钱包失败: %v", i+1, err)
			continue
		}

		wallet.Index = i
		wallets = append(wallets, wallet)
	}

	return wallets
}

// generateWalletsConcurrent 并发生成钱包
func (wg *WalletGenerator) generateWalletsConcurrent(config GeneratorConfig) []MultiChainWallet {
	walletChan := make(chan MultiChainWallet, config.Count)
	var wgSync sync.WaitGroup

	// 创建输出配置器
	var output *OutputConfig
	if wg.config != nil {
		output = &wg.config.Output
	}

	// 创建地址匹配器
	var matcher *AddressMatcher
	if wg.config != nil && wg.config.AddressMatching.Enabled {
		var err error
		matcher, err = NewAddressMatcher(wg.config)
		if err != nil {
			log.Printf("创建地址匹配器失败: %v", err)
			matcher = nil
		}
	}

	// 工作协程池
	jobs := make(chan int, config.Count)

	// 启动工作协程
	for w := 0; w < config.WorkerCount; w++ {
		wgSync.Add(1)
		go func() {
			defer wgSync.Done()
			for index := range jobs {
				var wallet MultiChainWallet
				var err error

				// 如果启用了地址匹配，循环生成直到匹配
				maxAttempts := 1
				if matcher != nil && matcher.IsEnabled() {
					maxAttempts = wg.config.AddressMatching.MaxAttempts
					if maxAttempts <= 0 {
						maxAttempts = 10000 // 默认最大尝试次数
					}
				}

				for attempt := 0; attempt < maxAttempts; attempt++ {
					if config.UseMnemonic {
						// 每个钱包生成独立的助记词
						entropy, _ := bip39.NewEntropy(128)
						mnemonic, _ := bip39.NewMnemonic(entropy)
						wallet, err = wg.GenerateWalletFromMnemonic(mnemonic, 0)
					} else {
						wallet, err = wg.GenerateRandomWallet()
					}

					if err != nil {
						log.Printf("协程生成第 %d 个钱包失败: %v", index+1, err)
						continue
					}

					// 检查地址匹配
					if matcher == nil || matcher.MatchWallet(wallet) {
						break
					}
				}

				if err == nil {
					wallet.Index = index
					walletChan <- wallet
				}
			}
		}()
	}

	// 发送任务
	go func() {
		for i := 0; i < config.Count; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	// 等待所有协程完成
	go func() {
		wgSync.Wait()
		close(walletChan)
	}()

	// 收集结果
	var wallets []MultiChainWallet
	for wallet := range walletChan {
		if output != nil && output.SaveToFile && output.OutputFile == "" {
			// 如果启用了输出配置，保存到文件
			if matcher != nil && matcher.IsEnabled() {
				saveWalletsToFile(config.UseMnemonic, matcher.config.AddressMatching.TargetChains[0], wallet, wg.config.Output.OutputFile)
			} else {
				saveWalletsToFile(config.UseMnemonic, "all", wallet, wg.config.Output.OutputFile)
			}
		}
		wallets = append(wallets, wallet)
	}

	// 显示匹配统计
	if matcher != nil && matcher.IsEnabled() {
		matcher.PrintStats()
	}

	return wallets
}

// GenerateRandomWallet 生成随机钱包
func (wg *WalletGenerator) GenerateRandomWallet() (MultiChainWallet, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return MultiChainWallet{}, fmt.Errorf("生成私钥失败: %v", err)
	}

	return wg.createWalletFromPrivateKey(privateKey, "", "")
}

// GenerateWalletFromMnemonic 从助记词生成钱包
func (wg *WalletGenerator) GenerateWalletFromMnemonic(mnemonic string, index int) (MultiChainWallet, error) {
	// 验证助记词
	if !bip39.IsMnemonicValid(mnemonic) {
		return MultiChainWallet{}, fmt.Errorf("无效的助记词")
	}

	// 生成种子
	seed := bip39.NewSeed(mnemonic, "")

	// 生成主密钥
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return MultiChainWallet{}, fmt.Errorf("生成主密钥失败: %v", err)
	}

	// 派生路径: m/44'/60'/0'/0/{index} (以太坊标准)
	derivePath := fmt.Sprintf("m/44'/60'/0'/0/%d", index)

	// 派生子密钥
	childKey, err := wg.deriveKeyFromPath(masterKey, derivePath)
	if err != nil {
		return MultiChainWallet{}, fmt.Errorf("派生密钥失败: %v", err)
	}

	// 转换为 ECDSA 私钥
	privateKey, err := crypto.ToECDSA(childKey.Key)
	if err != nil {
		return MultiChainWallet{}, fmt.Errorf("转换私钥失败: %v", err)
	}

	return wg.createWalletFromPrivateKey(privateKey, mnemonic, derivePath)
}

// deriveKeyFromPath 从路径派生密钥
func (wg *WalletGenerator) deriveKeyFromPath(masterKey *bip32.Key, path string) (*bip32.Key, error) {
	// 简化实现，实际应该解析完整路径
	// m/44'/60'/0'/0/{index}
	key := masterKey

	// 44' (hardened)
	key, _ = key.NewChildKey(bip32.FirstHardenedChild + 44)
	// 60' (hardened) - Ethereum
	key, _ = key.NewChildKey(bip32.FirstHardenedChild + 60)
	// 0' (hardened)
	key, _ = key.NewChildKey(bip32.FirstHardenedChild + 0)
	// 0 (non-hardened)
	key, _ = key.NewChildKey(0)

	// 从路径提取最后的索引
	var index uint32
	fmt.Sscanf(path, "m/44'/60'/0'/0/%d", &index)

	// index (non-hardened)
	return key.NewChildKey(index)
}

// createWalletFromPrivateKey 从私钥创建钱包
func (wg *WalletGenerator) createWalletFromPrivateKey(privateKey *ecdsa.PrivateKey, mnemonic, derivePath string) (MultiChainWallet, error) {
	privateKeyHex := hex.EncodeToString(crypto.FromECDSA(privateKey))
	publicKeyHex := hex.EncodeToString(crypto.FromECDSAPub(&privateKey.PublicKey))

	wallet := MultiChainWallet{
		Mnemonic:   mnemonic,
		PrivateKey: privateKeyHex,
		PublicKey:  publicKeyHex,
		DerivePath: derivePath,
	}

	// 生成各链地址
	wallet.EthAddress = wg.generateEthereumAddress(privateKey)
	wallet.BscAddress = wallet.EthAddress
	wallet.PolygonAddress = wallet.EthAddress

	btcAddr, err := wg.generateBitcoinAddress(privateKey)
	if err != nil {
		return wallet, fmt.Errorf("生成比特币地址失败: %v", err)
	}
	wallet.BtcAddress = btcAddr

	tronAddr, err := wg.generateTronAddress(&privateKey.PublicKey)
	if err != nil {
		return wallet, fmt.Errorf("生成波场地址失败: %v", err)
	}
	wallet.TronAddress = tronAddr

	return wallet, nil
}

// generateEthereumAddress 生成以太坊地址
func (wg *WalletGenerator) generateEthereumAddress(privateKey *ecdsa.PrivateKey) string {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return ""
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return address.Hex()
}

// generateBitcoinAddress 生成比特币地址
func (wg *WalletGenerator) generateBitcoinAddress(privateKey *ecdsa.PrivateKey) (string, error) {
	btcPrivKey, _ := btcec.PrivKeyFromBytes(crypto.FromECDSA(privateKey))
	pubKeyBytes := btcPrivKey.PubKey().SerializeCompressed()
	pubKeyHash := hash160(pubKeyBytes)
	address, err := btcutil.NewAddressPubKeyHash(pubKeyHash, &chaincfg.MainNetParams)
	if err != nil {
		return "", err
	}
	return address.EncodeAddress(), nil
}

// generateTronAddress 生成波场地址
func (wg *WalletGenerator) generateTronAddress(publicKey *ecdsa.PublicKey) (string, error) {
	pubKeyBytes := crypto.FromECDSAPub(publicKey)
	hash := crypto.Keccak256(pubKeyBytes[1:])
	address := hash[12:]
	tronAddress := append([]byte{0x41}, address...)
	return base58CheckEncode(tronAddress), nil
}

// 工具函数
func hash160(data []byte) []byte {
	sha256Hash := sha256.Sum256(data)
	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(sha256Hash[:])
	return ripemd160Hasher.Sum(nil)
}

func base58CheckEncode(data []byte) string {
	hash1 := sha256.Sum256(data)
	hash2 := sha256.Sum256(hash1[:])
	checksum := hash2[:4]
	fullData := append(data, checksum...)
	return base58Encode(fullData)
}

func base58Encode(data []byte) string {
	const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	zeroCount := 0
	for i := 0; i < len(data) && data[i] == 0; i++ {
		zeroCount++
	}

	var result []byte
	input := make([]byte, len(data))
	copy(input, data)

	for len(input) > 0 {
		remainder := 0
		for i := 0; i < len(input); i++ {
			temp := remainder*256 + int(input[i])
			input[i] = byte(temp / 58)
			remainder = temp % 58
		}
		result = append([]byte{alphabet[remainder]}, result...)
		input = trimLeadingZeros(input)
	}

	for i := 0; i < zeroCount; i++ {
		result = append([]byte{'1'}, result...)
	}

	return string(result)
}

func trimLeadingZeros(data []byte) []byte {
	for i, b := range data {
		if b != 0 {
			return data[i:]
		}
	}
	return []byte{}
}
