package main

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/runtime"
)

const (
	LeaderboardID = "tic_tac_toe_rank"
	TickRate      = 5
	ModuleName    = "tic_tac_toe"
)

// main is a no-op so generic `go build ./...` succeeds in CI/local checks.
// Nakama loads this module via InitModule.
func main() {}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	logger.Info("Tic-Tac-Toe module loaded")

	// Create or reset leaderboard
	if err := createLeaderboard(ctx, logger, nk); err != nil {
		return err
	}

	// Register the match handler
	if err := initializer.RegisterMatch(ModuleName, newMatch); err != nil {
		logger.Error("Unable to register match: %v", err)
		return err
	}

	// Register RPC to create a match (used by matchmaker hook)
	if err := initializer.RegisterRpc("create_match", rpcCreateMatch); err != nil {
		logger.Error("Unable to register RPC: %v", err)
		return err
	}

	// Register matchmaker matched hook
	if err := initializer.RegisterMatchmakerMatched(matchmakerMatched); err != nil {
		logger.Error("Unable to register matchmaker matched hook: %v", err)
		return err
	}

	logger.Info("Tic-Tac-Toe module initialized successfully")
	return nil
}

func createLeaderboard(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) error {
	authoritative := false
	sortOrder := "desc"
	operator := "incr"
	resetSchedule := ""
	metadata := map[string]interface{}{
		"game": "tic_tac_toe",
	}

	if err := nk.LeaderboardCreate(ctx, LeaderboardID, authoritative, sortOrder, operator, resetSchedule, metadata); err != nil {
		logger.Error("Error creating leaderboard: %v", err)
		return err
	}
	logger.Info("Leaderboard '%s' created/verified", LeaderboardID)
	return nil
}

func rpcCreateMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	params := map[string]interface{}{}
	matchID, err := nk.MatchCreate(ctx, ModuleName, params)
	if err != nil {
		logger.Error("Error creating match: %v", err)
		return "", err
	}
	logger.Info("Match created: %s", matchID)
	return matchID, nil
}

func matchmakerMatched(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, entries []runtime.MatchmakerEntry) (string, error) {
	logger.Info("Matchmaker matched %d players", len(entries))

	// Keep params JSON-safe and simple. The invited players are already delivered
	// by Nakama's matchmaker token to each client, so extra complex params are not needed.
	params := map[string]interface{}{}

	matchID, err := nk.MatchCreate(ctx, ModuleName, params)
	if err != nil {
		logger.Error("Error creating match from matchmaker: %v", err)
		return "", err
	}

	logger.Info("Match created from matchmaker: %s", matchID)
	return matchID, nil
}
