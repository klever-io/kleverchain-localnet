package domain

import (
	"bytes"
	"encoding/json"
)

type Delegation struct {
	Address string `json:"address"`
	Value   uint64 `json:"value"`
}

type GenesisEntry struct {
	Address    string     `json:"address"`
	Balance    uint64     `json:"balance"`
	KFIBalance uint64     `json:"kfiBalance"`
	Delegation Delegation `json:"delegation"`
}

type GenesisRootEntry struct {
	Address    string `json:"address"`
	Balance    uint64 `json:"balance"`
	KFIBalance uint64 `json:"kfiBalance"`
}

type Genesis struct {
	Entries []GenesisEntry
	Root    *GenesisRootEntry
}

func (g Genesis) MarshalJSON() ([]byte, error) {
	items := make([]any, 0, len(g.Entries)+1)
	for i := range g.Entries {
		items = append(items, g.Entries[i])
	}
	if g.Root != nil {
		items = append(items, *g.Root)
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(items); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}
