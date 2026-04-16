package nodes

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/klever-io/kleverchain-localnet/internal/domain"
	"github.com/klever-io/kleverchain-localnet/internal/errs"
)

type Pair struct {
	NodeIndex int
	PubKey    string
	Address   string
}

func Build(state domain.LocalnetState, pairs []Pair) (domain.NodesSetup, error) {
	if err := state.Validate(); err != nil {
		return domain.NodesSetup{}, err
	}
	if len(pairs) != state.Validators {
		return domain.NodesSetup{}, fmt.Errorf("%w: pairs=%d does not match validators=%d", errs.ErrInvalidInput, len(pairs), state.Validators)
	}
	if state.ConsensusGroupSize > len(pairs) {
		return domain.NodesSetup{}, fmt.Errorf("%w: consensusGroupSize=%d exceeds pairs=%d", errs.ErrConsensusGroupTooLarge, state.ConsensusGroupSize, len(pairs))
	}

	ordered := make([]Pair, len(pairs))
	copy(ordered, pairs)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].NodeIndex < ordered[j].NodeIndex
	})

	raw := state.StartTime
	if raw == 0 {
		raw = time.Now().Unix()
	}
	startTime := RoundUpToSlotBoundary(raw)

	chainID := state.ChainID
	if chainID == "" {
		chainID = strconv.FormatInt(startTime/domain.SlotRoundSeconds, 10)
	}

	initial := make([]domain.InitialNode, 0, len(ordered))
	for _, p := range ordered {
		initial = append(initial, domain.InitialNode{
			PubKey:  p.PubKey,
			Address: p.Address,
		})
	}

	return domain.NodesSetup{
		StartTime:             startTime,
		SlotInterval:          domain.SlotInterval,
		SlotsPerEpoch:         domain.SlotsPerEpoch,
		ConsensusGroupSize:    state.ConsensusGroupSize,
		MinNodes:              state.EffectiveMinNodes(),
		ChainID:               chainID,
		MinTransactionVersion: domain.MinTransactionVersion,
		KLVDenomination:       domain.KLVDenomination,
		InitialNodes:          initial,
	}, nil
}

func RoundUpToSlotBoundary(unix int64) int64 {
	return unix + domain.SlotRoundSeconds - (unix % domain.SlotRoundSeconds)
}
