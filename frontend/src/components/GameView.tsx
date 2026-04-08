import { useState, useEffect } from 'react';
import { nakama } from '../nakama';

interface GameViewProps {
    matchId: string;
    onLeaveMatch: () => void;
}

export function GameView({ matchId, onLeaveMatch }: GameViewProps) {
    const [board, setBoard] = useState<number[]>([0, 0, 0, 0, 0, 0, 0, 0, 0]);
    const [players, setPlayers] = useState<any>({});
    const [currentTurn, setCurrentTurn] = useState<string>('');
    const [turnTimer, setTurnTimer] = useState<number>(30);
    const [gameOver, setGameOver] = useState(false);
    const [winner, setWinner] = useState<string>('');
    const [winLine, setWinLine] = useState<number[]>([]);

    useEffect(() => {
        if (!nakama.socket) return;

        nakama.socket.onmatchdata = (matchState) => {
            if (matchState.match_id !== matchId) return;

            const opCode = matchState.op_code;
            const data = new TextDecoder().decode(matchState.data);

            if (opCode === 2) { // OpCodeStateUpdate
                const state = JSON.parse(data);
                setBoard(state.board);
                setCurrentTurn(state.currentTurn);
                setPlayers(state.players);
                setTurnTimer(state.turnTimer);
                setGameOver(state.gameOver);
                setWinner(state.winner);
                setWinLine(state.winLine);
            } else if (opCode === 3) { // OpCodeGameOver
                const state = JSON.parse(data);
                setGameOver(true);
                setWinner(state.winner);
                setWinLine(state.winLine);
            } else if (opCode === 4) { // OpCodePlayerJoined
                // We could show a toast here
            } else if (opCode === 5) { // OpCodeError
                const state = JSON.parse(data);
                console.error('Server error:', state.message);
                // Could show toast
            } else if (opCode === 6) { // OpCodeTimerUpdate
                const state = JSON.parse(data);
                setTurnTimer(state.turnTimer);
                setCurrentTurn(state.currentTurn);
            }
        };

        return () => {
             if (nakama.socket) nakama.socket.onmatchdata = () => {};
        };
    }, [matchId]);

    const handleCellClick = (index: number) => {
        if (gameOver || board[index] !== 0) return;
        if (currentTurn !== nakama.session?.user_id) return;

        const payload = JSON.stringify({ position: index });
        nakama.socket?.sendMatchState(matchId, 1, payload); // OpCodeMove = 1
    };

    const handleLeave = async () => {
        try {
            await nakama.socket?.leaveMatch(matchId);
        } catch (e) {
            console.error(e);
        }
        onLeaveMatch();
    };

    const myUserId = nakama.session?.user_id;
    const opponent = Object.values(players).find((p: any) => p.userId !== myUserId) as any;
    const me = players[myUserId!] as any;

    const isMyTurn = currentTurn === myUserId;
    
    // Determine status text
    let statusText = 'Waiting for opponent...';
    if (Object.keys(players).length === 2 && !gameOver) {
        statusText = isMyTurn ? 'Your Turn' : "Opponent's Turn";
    } else if (gameOver) {
        if (winner === 'draw') statusText = "It's a Draw!";
        else if (winner === myUserId) statusText = 'You Won!';
        else statusText = 'You Lost!';
    }

    return (
        <div className="game-container">
            <header className="game-header">
                <button onClick={handleLeave} className="btn-icon">← Leave</button>
                <div className="timer" style={{ color: turnTimer < 10 ? '#ef4444' : 'inherit'}}>
                    {Math.max(0, Math.ceil(turnTimer))}s
                </div>
            </header>

            <div className="players-hud">
                <div className={`player-card ${isMyTurn ? 'active' : ''}`}>
                    <div className="avatar me">{me?.mark === 1 ? 'X' : (me?.mark === 2 ? 'O' : '?')}</div>
                    <span className="name">You</span>
                </div>
                <div className="vs">VS</div>
                <div className={`player-card ${!isMyTurn && currentTurn ? 'active' : ''} ${!opponent ? 'waiting' : ''}`}>
                    <div className="avatar opponent">{opponent?.mark === 1 ? 'X' : (opponent?.mark === 2 ? 'O' : '?')}</div>
                    <span className="name">{opponent?.username || 'Waiting...'}</span>
                </div>
            </div>

            <div className="status-banner">
                <h2>{statusText}</h2>
            </div>

            <div className="board-container">
                <div className={`board ${!isMyTurn || gameOver ? 'disabled' : ''}`}>
                    {board.map((cell, index) => {
                        const isWinningCell = winLine?.includes(index);
                        return (
                            <div 
                                key={index} 
                                className={`cell ${cell !== 0 ? 'occupied' : ''} ${isWinningCell ? 'win-highlight' : ''}`}
                                onClick={() => handleCellClick(index)}
                            >
                                {cell === 1 ? 'X' : cell === 2 ? 'O' : ''}
                            </div>
                        );
                    })}
                </div>
            </div>

            {gameOver && (
                <div className="game-over-modal">
                    <div className="modal-content">
                        <h2>{winner === 'draw' ? 'Draw!' : (winner === myUserId ? 'Victory!' : 'Defeat')}</h2>
                        <button onClick={handleLeave} className="btn-primary">Back to Lobby</button>
                    </div>
                </div>
            )}
        </div>
    );
}
