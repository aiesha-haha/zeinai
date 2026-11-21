import api from './api';

export const logout = async () => {
    try {
        // Try calling the backend logout API to log the action
        await api.post('/auth/logout');
    } catch (err) {
        console.error('Backend logout failed:', err);
    } finally {
        // Always clear local session
        if (typeof window !== 'undefined') {
            localStorage.removeItem('access_token');
            localStorage.removeItem('refresh_token');
            localStorage.removeItem('user_role');
        }

        // Redirect to login
        window.location.replace('/login');
    }
};

/**
 * Stores a login session. The API client reads localStorage; a few older pages
 * read cookies, so both are written to keep every page in sync.
 */
export const saveSession = (accessToken: string, refreshToken: string, role: string) => {
    if (typeof window === 'undefined') return;
    localStorage.setItem('access_token', accessToken);
    localStorage.setItem('refresh_token', refreshToken);
    localStorage.setItem('user_role', role);
    const secure = window.location.protocol === 'https:' ? '; Secure' : '';
    document.cookie = `access_token=${accessToken}; path=/; SameSite=Lax${secure}`;
    document.cookie = `refresh_token=${refreshToken}; path=/; SameSite=Lax${secure}`;
    document.cookie = `user_role=${role}; path=/; SameSite=Lax${secure}`;
};

export const clearSession = () => {
    if (typeof window === 'undefined') return;
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user_role');
    for (const k of ['access_token', 'refresh_token', 'user_role']) {
        document.cookie = `${k}=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT`;
    }
};

/** Current session as seen by the API client */
export const getSession = () => {
    if (typeof window === 'undefined') return { token: null as string | null, role: null as string | null };
    return { token: localStorage.getItem('access_token'), role: localStorage.getItem('user_role') };
};
