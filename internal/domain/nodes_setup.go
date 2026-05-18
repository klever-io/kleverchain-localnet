package domain

type InitialNode struct {
	PubKey  string `json:"pubkey"`
	Address string `json:"address"`
}

type NodesSetup struct {
	StartTime             int64         `json:"startTime"`
	SlotInterval          int           `json:"slotInterval"`
	SlotsPerEpoch         int           `json:"slotsPerEpoch"`
	ConsensusGroupSize    int           `json:"consensusGroupSize"`
	MinNodes              int           `json:"minNodes"`
	ChainID               string        `json:"chainID"`
	MinTransactionVersion int           `json:"minTransactionVersion"`
	KLVDenomination       int           `json:"klvDenomination"`
	InitialNodes          []InitialNode `json:"initialNodes"`
}
