import { useState, useEffect } from 'react';
import { nakama } from '../nakama';

interface LobbyViewProps {
    onMatchJoined: (matchId: string) => void;
    onLogout: () => void;
}

export function LobbyView({ onMatchJoined, onLogout }: LobbyViewProps) {
    const [isSearching, setIsSearching] = useState(false);
    const [ticket, setTicket] = useState<string | null>(null);
    const [leaderboard, setLeaderboard] = useState<any[]>([]);

    useEffect(() => {
        if (!nakama.socket) return;

        // Listen for matchmaker matched event
        nakama.socket.onmatchmakermatched = (matched) => {
            console.log('Matchmaker matched:', matched);
            setIsSearching(false);
            setTicket(null);
            
            // Join the match using the token
            nakama.socket!.joinMatch(matched.match_id, matched.token).then(match => {
                onMatchJoined(match.match_id);
            });
        };

        // Fetch leaderboard records
        fetchLeaderboard();

        return () => {
            if (nakama.socket) {
                nakama.socket.onmatchmakermatched = () => {};
            }
        };
    }, []);

    const fetchLeaderboard = async () => {
        try {
            const result = await nakama.client.listLeaderboardRecords(
                nakama.session!,
                'tic_tac_toe_rank',
                undefined,
                10
            );       
            setLeaderboard(result.records || []);
        } catch (error) {
            console.error('Failed to fetch leaderboard:', error);
        }
    };

    const toggleMatchmaking = async () => {
        if (!nakama.socket) return;

        try {
            if (isSearching && ticket) {
                await nakama.socket.removeMatchmaker(ticket);
                setTicket(null);
                setIsSearching(false);
            } else {
                const matchmakerTicket = await nakama.socket.addMatchmaker('*', 2, 2);
                setTicket(matchmakerTicket.ticket);
                setIsSearching(true);
            }
        } catch (error) {
            console.error('Matchmaking error:', error);
            setIsSearching(false);
            setTicket(null);
        }
    };

    return (
        <div className="lobby-container">
            <header className="lobby-header">
                <div className="user-info">
                    <span className="avatar">{nakama.session?.username?.charAt(0).toUpperCase()}</span>
                    <h2>{nakama.session?.username}</h2>
                </div>
                <button onClick={onLogout} className="btn-secondary">Logout</button>
            </header>

            <main className="lobby-main">
                <div className="matchmaking-section">
                    <button 
                        className={`btn-play ${isSearching ? 'searching' : ''}`} 
                        onClick={toggleMatchmaking}
                    >
                        {isSearching ? 'Cancel Search' : 'Find Match'}
                    </button>
                    {isSearching && <p className="search-status">Looking for opponent...</p>}
                </div>

                <div className="leaderboard-section">
                    <h3>Global Rankings</h3>
                    {leaderboard.length === 0 ? (
                        <p className="empty-state">No ranking data yet</p>
                    ) : (
                        <ul className="leaderboard-list">
                            {leaderboard.map((record, index) => (
                                <li key={record.owner_id} className={`leaderboard-item ${record.owner_id === nakama.session?.user_id ? 'is-me' : ''}`}>
                                    <span className="rank">#{index + 1}</span>
                                    <span className="name">{record.username}</span>
                                    <span className="score">{record.score} Wins</span>
                                </li>
                            ))}
                        </ul>
                    )}
                </div>
            </main>
        </div>
    );
}
