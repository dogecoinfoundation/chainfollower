package store

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dogecoinfoundation/chainfollower/pkg/state"
)

func LoadChainPos(path string) (*state.ChainPos, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &state.ChainPos{
			BlockHash:          "",
			BlockHeight:        0,
			WaitingForNextHash: false,
		}, nil
	}

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, err
	}
	defer file.Close()

	// Decode JSON into struct
	var chainState state.ChainPos
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&chainState)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return nil, err
	}

	return &chainState, nil
}

func SaveChainPos(path string, chainState *state.ChainPos) error {
	file, err := os.Create("data.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(chainState)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return err
	}

	return nil
}
