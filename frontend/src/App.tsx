import { useState, useEffect } from 'react';
import { AuthView } from './components/AuthView';
import { LobbyView } from './components/LobbyView';
import { GameView } from './components/GameView';
import { nakama } from './nakama';

function App() {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [currentMatchId, setCurrentMatchId] = useState<string | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const init = async () => {
            const restored = await nakama.restoreSession();
            setIsAuthenticated(restored);
            setLoading(false);
        };
        init();
    }, []);

    const handleLogout = async () => {
        await nakama.logout();
        setIsAuthenticated(false);
        setCurrentMatchId(null);
    };

    if (loading) {
        return <div className="loading-screen">Loading...</div>;
    }

    if (!isAuthenticated) {
        return <AuthView onAuthenticated={() => setIsAuthenticated(true)} />;
    }

    if (currentMatchId) {
        return <GameView matchId={currentMatchId} onLeaveMatch={() => setCurrentMatchId(null)} />;
    }

    return <LobbyView onMatchJoined={setCurrentMatchId} onLogout={handleLogout} />;
}

export default App;
