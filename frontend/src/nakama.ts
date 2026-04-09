import { Client, Session } from '@heroiclabs/nakama-js';
import type { Socket } from '@heroiclabs/nakama-js';

const USE_SSL = false;
const SERVER_KEY = 'defaultkey';
const SERVER_HOST = import.meta.env.VITE_NAKAMA_HOST || window.location.hostname;
const SERVER_PORT = import.meta.env.VITE_NAKAMA_PORT || '7350';

export class NakamaService {
    client: Client;
    session: Session | null = null;
    socket: Socket | null = null;
    private authInFlight: Promise<Session> | null = null;

    constructor() {
        this.client = new Client(SERVER_KEY, SERVER_HOST, SERVER_PORT, USE_SSL);
    }

    async authenticate(
        email: string,
        password: string,
        create: boolean = false,
        username?: string
    ): Promise<Session> {
        if (this.authInFlight) {
            return this.authInFlight;
        }

        const runAuth = async (): Promise<Session> => {
        try {
            const normalizedEmail = email.trim().toLowerCase();
            const normalizedUsername = username?.trim();
            const session = create
                ? await this.client.authenticateEmail(normalizedEmail, password, true, normalizedUsername)
                : await this.client.authenticateEmail(normalizedEmail, password, false);
            this.session = session;
            localStorage.setItem('nakama_session', session.token);

            this.socket = this.client.createSocket(USE_SSL, false);
            await this.socket.connect(session, true);

            return session;
        } catch (error) {
            console.error('Authentication Error:', error);
            throw error;
        } finally {
            this.authInFlight = null;
        }
        };

        this.authInFlight = runAuth();
        return this.authInFlight;
    }

    async restoreSession(): Promise<boolean> {
        const token = localStorage.getItem('nakama_session');
        if (!token) return false;

        try {
            const session = Session.restore(token, '');
            if (session.isexpired(Math.floor(Date.now() / 1000))) {
                localStorage.removeItem('nakama_session');
                return false;
            }

            this.session = session;
            this.socket = this.client.createSocket(USE_SSL, false);
            await this.socket.connect(session, true);
            return true;
        } catch (error) {
            console.error('Restore Session Error:', error);
            localStorage.removeItem('nakama_session');
            return false;
        }
    }

    async logout() {
        this.session = null;
        if (this.socket) {
            this.socket.disconnect(false);
            this.socket = null;
        }
        localStorage.removeItem('nakama_session');
    }
}

export const nakama = new NakamaService();
