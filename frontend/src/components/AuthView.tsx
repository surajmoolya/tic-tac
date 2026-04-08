import { useState } from 'react';
import { nakama } from '../nakama';

interface AuthViewProps {
    onAuthenticated: () => void;
}

export function AuthView({ onAuthenticated }: AuthViewProps) {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [username, setUsername] = useState('');
    const [isRegistering, setIsRegistering] = useState(false);
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (loading) return;
        setError('');
        setLoading(true);

        try {
            if (isRegistering) {
                const usernamePattern = /^[a-zA-Z0-9_]{3,20}$/;
                if (!usernamePattern.test(username.trim())) {
                    setError('Username must be 3-20 characters and contain only letters, numbers, or underscores.');
                    setLoading(false);
                    return;
                }
            }

            await nakama.authenticate(email, password, isRegistering, isRegistering ? username : undefined);
            onAuthenticated();
        } catch (err: any) {
            console.error('Auth handler error:', err);
            setError(err.message || 'Authentication failed. Please try again.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="auth-container">
            <div className="auth-card">
                <h1>Tic Tac Toe</h1>
                <p className="subtitle">Real-time Multiplayer</p>
                
                {error && <div className="error-message">{error}</div>}
                
                <form onSubmit={handleSubmit} className="auth-form">
                    <div className="form-group">
                        <label>Email</label>
                        <input 
                            type="email" 
                            value={email} 
                            onChange={e => setEmail(e.target.value)} 
                            placeholder="player@example.com"
                            required 
                        />
                    </div>
                    
                    <div className="form-group">
                        <label>Password</label>
                        <input 
                            type="password" 
                            value={password} 
                            onChange={e => setPassword(e.target.value)} 
                            placeholder="Min 8 characters"
                            required 
                            minLength={8}
                        />
                    </div>

                    {isRegistering && (
                        <div className="form-group">
                            <label>Username</label>
                            <input
                                type="text"
                                value={username}
                                onChange={e => setUsername(e.target.value)}
                                placeholder="player_123"
                                required={isRegistering}
                                minLength={3}
                                maxLength={20}
                                pattern="[a-zA-Z0-9_]+"
                            />
                        </div>
                    )}
                    
                    <button type="submit" className="btn-primary" disabled={loading}>
                        {loading ? 'Processing...' : (isRegistering ? 'Create Account' : 'Login')}
                    </button>
                    
                    <p className="toggle-auth">
                        {isRegistering ? 'Already have an account? ' : 'Need an account? '}
                        <button
                            type="button"
                            className="btn-link"
                            onClick={() => {
                                setIsRegistering(!isRegistering);
                                setError('');
                            }}
                        >
                            {isRegistering ? 'Login' : 'Register'}
                        </button>
                    </p>
                </form>
            </div>
        </div>
    );
}
