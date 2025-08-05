package main

// MultiChainWallet 多链钱包结构
type MultiChainWallet struct {
	Index          int    `json:"index"`
	Mnemonic       string `json:"mnemonic,omitempty"`
	PrivateKey     string `json:"private_key"`
	PublicKey      string `json:"public_key"`
	DerivePath     string `json:"derive_path,omitempty"`
	EthAddress     string `json:"eth_address"`
	BtcAddress     string `json:"btc_address"`
	BscAddress     string `json:"bsc_address"`
	PolygonAddress string `json:"polygon_address"`
	TronAddress    string `json:"tron_address"`
}

// GeneratorConfig 钱包生成配置
type GeneratorConfig struct {
	Count          int
	UseMnemonic    bool
	Mnemonic       string // 可选：使用指定助记词
	ConcurrentMode bool
	WorkerCount    int
}
