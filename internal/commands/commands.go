package commands

import "context"

type CommandType string

const (
	CommandTypeResync  CommandType = "RESYNC"
	CommandTypeRestart CommandType = "RESTART"
	CommandTypeStop    CommandType = "STOP"
)

type Command struct {
	CommandType CommandType
}

type ReSyncChainFollowerCmd struct {
	Command
	BlockHash string // Block hash to re-sync from.
}

/** Restart the ChainFollower in case it becomes stuck. */
type RestartChainFollowerCmd struct {
	Command
}

/** Stop the ChainFollower. Ctx can have a timeout. */
type StopChainFollowerCmd struct {
	Command
	Ctx context.Context
}
