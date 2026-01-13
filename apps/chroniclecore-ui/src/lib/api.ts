export const API_BASE = '/api/v1';

export async function fetchApi(endpoint: string, options: RequestInit = {}) {
    const url = `${API_BASE}${endpoint.startsWith('/') ? endpoint : '/' + endpoint}`;

    try {
        const res = await fetch(url, {
            ...options,
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            }
        });

        if (!res.ok) {
            const errorData = await res.json().catch(() => ({}));
            console.error('API Error Response:', res.status, errorData);
            throw new Error(errorData.error?.message || `API Error: ${res.status}`);
        }

        // Handle empty responses (e.g. 204)
        if (res.status === 204) return null;

        return res.json();
    } catch (err) {
        console.error(`API Fetch Error (${endpoint}):`, err);
        throw err;
    }
}
