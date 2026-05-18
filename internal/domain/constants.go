package domain

const (
	KleverImage = "kleverapp/klever-go:v1.7.16-0-gcf9f612c"

	KLVDelegation uint64 = 10_000_000_000_000
	KFISupply     uint64 = 21_000_000_000_000
	RootKLV       uint64 = 1_000_000_000_000
	RootKFI       uint64 = 1_000_000_000_000

	DefaultMaxSupply uint64 = 10_000_000_000_000_000

	SlotInterval                = 4000
	SlotsPerEpoch               = 20
	SlotRoundSeconds      int64 = 75
	KLVDenomination             = 6
	MinTransactionVersion       = 1

	// MinStartHeadroomSeconds is the minimum gap between "now" and genesis startTime
	// that `localnet start` enforces on a fresh-genesis run. Validators that boot after
	// genesis enter inSync state and the chain stalls indefinitely.
	MinStartHeadroomSeconds int64 = 30

	SeednodeStaticIP  = "172.25.0.5"
	SeednodeRESTPort  = 8799
	SeednodeP2PPort   = 37373
	ValidatorPortBase = 8800

	NetworkSubnet = "172.25.0.0/24"
	NetworkName   = "klever"

	ExpectedSeednodePeerID = "16Uiu2HAkvsiFSh3EDH66oT6cETfwf5JBHYsgKCKxJQUgq8Db8kFf"
)
