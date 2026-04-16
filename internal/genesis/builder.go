package genesis

import (
	"fmt"

	"github.com/klever-io/kleverchain-localnet/internal/domain"
	"github.com/klever-io/kleverchain-localnet/internal/errs"
)

func Build(state domain.LocalnetState, wallets []domain.Wallet, root *domain.Wallet) (domain.Genesis, error) {
	if err := state.Validate(); err != nil {
		return domain.Genesis{}, err
	}
	if len(wallets) == 0 {
		return domain.Genesis{}, fmt.Errorf("%w: at least one wallet is required", errs.ErrInvalidInput)
	}
	if len(wallets) != state.Validators {
		return domain.Genesis{}, fmt.Errorf("%w: wallets=%d does not match validators=%d", errs.ErrInvalidInput, len(wallets), state.Validators)
	}

	numWallets := uint64(len(wallets))

	var rootKLV, rootKFI uint64
	if root != nil {
		rootKLV = domain.RootKLV
		rootKFI = domain.RootKFI
	}

	totalStaking := domain.KLVDelegation * numWallets
	if state.MaxSupply <= totalStaking+rootKLV {
		return domain.Genesis{}, fmt.Errorf("%w: maxSupply=%d must exceed totalStaking+rootKLV=%d", errs.ErrInvalidInput, state.MaxSupply, totalStaking+rootKLV)
	}

	eachKLV := (state.MaxSupply - totalStaking - rootKLV) / numWallets
	eachKFI := (domain.KFISupply - rootKFI) / numWallets

	entries := make([]domain.GenesisEntry, 0, len(wallets))
	for _, w := range wallets {
		entries = append(entries, domain.GenesisEntry{
			Address:    w.Address,
			Balance:    eachKLV,
			KFIBalance: eachKFI,
			Delegation: domain.Delegation{
				Address: w.Address,
				Value:   domain.KLVDelegation,
			},
		})
	}

	g := domain.Genesis{Entries: entries}
	if root != nil {
		g.Root = &domain.GenesisRootEntry{
			Address:    root.Address,
			Balance:    domain.RootKLV,
			KFIBalance: domain.RootKFI,
		}
	}
	return g, nil
}
