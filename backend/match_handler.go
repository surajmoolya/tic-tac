package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

// OpCodes for client-server communication
const (
	OpCodeMove          int64 = 1
	OpCodeStateUpdate   int64 = 2
	OpCodeGameOver      int64 = 3
	OpCodePlayerJoined  int64 = 4
	OpCodeError         int64 = 5
	OpCodeTimerUpdate   int64 = 6
	OpCodeForfeit       int64 = 7
	OpCodePlayerLeft    int64 = 8

	TurnTimeoutSeconds  float64 = 30.0
	MaxPlayers          int     = 2
)

// MatchState holds the full server-authoritative game state
type MatchState struct {
	Board           [9]int                       `json:"board"`           // 0=empty, 1=X, 2=O
	Players         map[string]*PlayerState      `json:"players"`
	PlayerOrder     []string                     `json:"playerOrder"`     // [0]=X, [1]=O
	CurrentTurn     int                          `json:"currentTurn"`     // index into PlayerOrder
	MoveCount       int                          `json:"moveCount"`
	GameOver        bool                         `json:"gameOver"`
	Winner          string                       `json:"winner"`          // userID or "draw"
	WinLine         []int                        `json:"winLine"`         // winning cell indices
	TurnTimer       float64                      `json:"turnTimer"`       // seconds remaining
	MatchLabel      string                       `json:"matchLabel"`
	StartedAt       int64                        `json:"startedAt"`
	Presences       map[string]runtime.Presence  `json:"-"`
}

// PlayerState tracks per-player information
type PlayerState struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Mark     int    `json:"mark"`  // 1=X, 2=O
}

// MoveMessage is the payload from the client
type MoveMessage struct {
	Position int `json:"position"` // 0-8
}

// StateUpdateMessage is broadcast to all clients
type StateUpdateMessage struct {
	Board       [9]int                  `json:"board"`
	CurrentTurn string                  `json:"currentTurn"` // userID
	Players     map[string]*PlayerState `json:"players"`
	MoveCount   int                     `json:"moveCount"`
	TurnTimer   float64                 `json:"turnTimer"`
	GameOver    bool                    `json:"gameOver"`
	Winner      string                  `json:"winner"`
	WinLine     []int                   `json:"winLine"`
}

// GameOverMessage is sent when the game ends
type GameOverMessage struct {
	Winner   string `json:"winner"` // userID or "draw"
	Reason   string `json:"reason"` // "win", "draw", "forfeit", "disconnect"
	WinLine  []int  `json:"winLine"`
}

// ErrorMessage is sent on invalid moves
type ErrorMessage struct {
	Message string `json:"message"`
}

// PlayerJoinedMessage announces a player
type PlayerJoinedMessage struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Mark     int    `json:"mark"`
}

// TimerUpdateMessage announces time remaining
type TimerUpdateMessage struct {
	TurnTimer   float64 `json:"turnTimer"`
	CurrentTurn string  `json:"currentTurn"`
}

// Win patterns (indices into the board)
var winPatterns = [][3]int{
	{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
	{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // columns
	{0, 4, 8}, {2, 4, 6},             // diagonals
}

// Match implements runtime.Match
type Match struct{}

func newMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
	return &Match{}, nil
}

func (m *Match) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	state := &MatchState{
		Board:       [9]int{},
		Players:     make(map[string]*PlayerState),
		PlayerOrder: make([]string, 0, 2),
		CurrentTurn: 0,
		MoveCount:   0,
		GameOver:    false,
		Winner:      "",
		WinLine:     []int{},
		TurnTimer:   TurnTimeoutSeconds,
		MatchLabel:  "waiting",
		StartedAt:   time.Now().Unix(),
		Presences:   make(map[string]runtime.Presence),
	}

	label := `{"open":true,"players":0}`
	return state, TickRate, label
}

func (m *Match) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	mState := state.(*MatchState)

	if mState.GameOver {
		return mState, false, "match is over"
	}
	if len(mState.Players) >= MaxPlayers {
		return mState, false, "match is full"
	}
	// Prevent duplicate joins
	if _, exists := mState.Players[presence.GetUserId()]; exists {
		return mState, false, "already joined"
	}

	return mState, true, ""
}

func (m *Match) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState := state.(*MatchState)

	for _, p := range presences {
		userID := p.GetUserId()
		username := p.GetUsername()

		// Assign mark (X=1 for first player, O=2 for second)
		mark := 1
		if len(mState.Players) == 1 {
			mark = 2
		}

		player := &PlayerState{
			UserID:   userID,
			Username: username,
			Mark:     mark,
		}
		mState.Players[userID] = player
		mState.PlayerOrder = append(mState.PlayerOrder, userID)
		mState.Presences[userID] = p

		logger.Info("Player %s (%s) joined as mark %d", username, userID, mark)

		// Notify all players about the new join
		joinMsg, _ := json.Marshal(PlayerJoinedMessage{
			UserID:   userID,
			Username: username,
			Mark:     mark,
		})
		dispatcher.BroadcastMessage(OpCodePlayerJoined, joinMsg, nil, nil, true)
	}

	// If two players have joined, randomize who goes first and start the game
	if len(mState.Players) == 2 {
		// Randomize turn order
		if rand.Intn(2) == 1 {
			mState.PlayerOrder[0], mState.PlayerOrder[1] = mState.PlayerOrder[1], mState.PlayerOrder[0]
		}
		mState.CurrentTurn = 0
		mState.TurnTimer = TurnTimeoutSeconds

		label := `{"open":false,"players":2}`
		dispatcher.MatchLabelUpdate(label)
		mState.MatchLabel = "playing"

		// Send initial state to all players
		broadcastState(mState, dispatcher)
	}

	return mState
}

func (m *Match) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState := state.(*MatchState)

	for _, p := range presences {
		userID := p.GetUserId()
		logger.Info("Player %s left the match", userID)

		delete(mState.Presences, userID)

		// If game is still in progress, the remaining player wins
		if !mState.GameOver && len(mState.Players) == 2 {
			// Find the other player
			var remainingUserID string
			for uid := range mState.Players {
				if uid != userID {
					remainingUserID = uid
					break
				}
			}

			mState.GameOver = true
			mState.Winner = remainingUserID

			gameOverMsg, _ := json.Marshal(GameOverMessage{
				Winner:  remainingUserID,
				Reason:  "disconnect",
				WinLine: []int{},
			})
			dispatcher.BroadcastMessage(OpCodeGameOver, gameOverMsg, nil, nil, true)

			// Record win on leaderboard
			recordLeaderboard(ctx, logger, nk, remainingUserID, mState.Players[remainingUserID].Username)
		}
	}

	// If no players left, terminate the match
	if len(mState.Presences) == 0 {
		return nil
	}

	return mState
}

func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	mState := state.(*MatchState)

	if mState.GameOver {
		return nil // End the match
	}

	// Don't process until we have 2 players
	if len(mState.Players) < 2 {
		return mState
	}

	// --- TIMER LOGIC ---
	// Decrement timer each tick (1/TickRate seconds per tick)
	mState.TurnTimer -= 1.0 / float64(TickRate)

	// Broadcast timer update every second (every TickRate ticks)
	if tick%int64(TickRate) == 0 {
		currentPlayerID := mState.PlayerOrder[mState.CurrentTurn]
		timerMsg, _ := json.Marshal(TimerUpdateMessage{
			TurnTimer:   mState.TurnTimer,
			CurrentTurn: currentPlayerID,
		})
		dispatcher.BroadcastMessage(OpCodeTimerUpdate, timerMsg, nil, nil, true)

		// Periodic full-state sync prevents clients from getting stuck if they miss
		// the initial state packet during match transition/join timing.
		broadcastState(mState, dispatcher)
	}

	// Check timeout
	if mState.TurnTimer <= 0 {
		// Current player forfeits due to timeout
		timedOutPlayerID := mState.PlayerOrder[mState.CurrentTurn]
		winnerIndex := 1 - mState.CurrentTurn
		winnerPlayerID := mState.PlayerOrder[winnerIndex]

		mState.GameOver = true
		mState.Winner = winnerPlayerID

		logger.Info("Player %s timed out. Player %s wins by forfeit.", timedOutPlayerID, winnerPlayerID)

		gameOverMsg, _ := json.Marshal(GameOverMessage{
			Winner:  winnerPlayerID,
			Reason:  "forfeit",
			WinLine: []int{},
		})
		dispatcher.BroadcastMessage(OpCodeGameOver, gameOverMsg, nil, nil, true)

		// Record win
		recordLeaderboard(ctx, logger, nk, winnerPlayerID, mState.Players[winnerPlayerID].Username)

		return mState
	}

	// --- PROCESS INCOMING MESSAGES ---
	for _, msg := range messages {
		if msg.GetOpCode() != OpCodeMove {
			continue
		}

		senderID := msg.GetUserId()
		currentPlayerID := mState.PlayerOrder[mState.CurrentTurn]

		// Validate it's the sender's turn
		if senderID != currentPlayerID {
			errMsg, _ := json.Marshal(ErrorMessage{Message: "Not your turn"})
			dispatcher.BroadcastMessage(OpCodeError, errMsg, []runtime.Presence{mState.Presences[senderID]}, nil, true)
			continue
		}

		// Parse move
		var move MoveMessage
		if err := json.Unmarshal(msg.GetData(), &move); err != nil {
			errMsg, _ := json.Marshal(ErrorMessage{Message: "Invalid move data"})
			dispatcher.BroadcastMessage(OpCodeError, errMsg, []runtime.Presence{mState.Presences[senderID]}, nil, true)
			continue
		}

		// Validate position bounds
		if move.Position < 0 || move.Position > 8 {
			errMsg, _ := json.Marshal(ErrorMessage{Message: "Position out of bounds (0-8)"})
			dispatcher.BroadcastMessage(OpCodeError, errMsg, []runtime.Presence{mState.Presences[senderID]}, nil, true)
			continue
		}

		// Validate cell is empty
		if mState.Board[move.Position] != 0 {
			errMsg, _ := json.Marshal(ErrorMessage{Message: "Cell already occupied"})
			dispatcher.BroadcastMessage(OpCodeError, errMsg, []runtime.Presence{mState.Presences[senderID]}, nil, true)
			continue
		}

		// Apply the move
		playerMark := mState.Players[senderID].Mark
		mState.Board[move.Position] = playerMark
		mState.MoveCount++

		logger.Info("Player %s placed mark %d at position %d", senderID, playerMark, move.Position)

		// Check win condition
		if winner, winLine := checkWin(mState.Board, playerMark); winner {
			mState.GameOver = true
			mState.Winner = senderID
			mState.WinLine = winLine

			gameOverMsg, _ := json.Marshal(GameOverMessage{
				Winner:  senderID,
				Reason:  "win",
				WinLine: winLine,
			})
			dispatcher.BroadcastMessage(OpCodeGameOver, gameOverMsg, nil, nil, true)

			// Record win
			recordLeaderboard(ctx, logger, nk, senderID, mState.Players[senderID].Username)

			broadcastState(mState, dispatcher)
			return mState
		}

		// Check draw
		if mState.MoveCount >= 9 {
			mState.GameOver = true
			mState.Winner = "draw"

			gameOverMsg, _ := json.Marshal(GameOverMessage{
				Winner:  "draw",
				Reason:  "draw",
				WinLine: []int{},
			})
			dispatcher.BroadcastMessage(OpCodeGameOver, gameOverMsg, nil, nil, true)

			broadcastState(mState, dispatcher)
			return mState
		}

		// Switch turns and reset timer
		mState.CurrentTurn = 1 - mState.CurrentTurn
		mState.TurnTimer = TurnTimeoutSeconds

		// Broadcast updated state
		broadcastState(mState, dispatcher)
	}

	return mState
}

func (m *Match) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	logger.Info("Match terminating")
	return nil
}

func (m *Match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, ""
}

// --- HELPER FUNCTIONS ---

func checkWin(board [9]int, mark int) (bool, []int) {
	for _, pattern := range winPatterns {
		if board[pattern[0]] == mark && board[pattern[1]] == mark && board[pattern[2]] == mark {
			return true, []int{pattern[0], pattern[1], pattern[2]}
		}
	}
	return false, nil
}

func broadcastState(mState *MatchState, dispatcher runtime.MatchDispatcher) {
	currentPlayerID := ""
	if len(mState.PlayerOrder) > mState.CurrentTurn {
		currentPlayerID = mState.PlayerOrder[mState.CurrentTurn]
	}

	stateMsg, _ := json.Marshal(StateUpdateMessage{
		Board:       mState.Board,
		CurrentTurn: currentPlayerID,
		Players:     mState.Players,
		MoveCount:   mState.MoveCount,
		TurnTimer:   mState.TurnTimer,
		GameOver:    mState.GameOver,
		Winner:      mState.Winner,
		WinLine:     mState.WinLine,
	})
	dispatcher.BroadcastMessage(OpCodeStateUpdate, stateMsg, nil, nil, true)
}

func recordLeaderboard(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, userID string, username string) {
	score := int64(1)
	subscore := int64(0)
	metadata := map[string]interface{}{
		"last_win": fmt.Sprintf("%d", time.Now().Unix()),
	}

	_, err := nk.LeaderboardRecordWrite(ctx, LeaderboardID, userID, username, score, subscore, metadata, nil)
	if err != nil {
		logger.Error("Error writing leaderboard record for %s: %v", userID, err)
	} else {
		logger.Info("Leaderboard updated for player %s (%s)", username, userID)
	}
}
