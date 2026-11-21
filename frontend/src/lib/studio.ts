import api from './api';

export type StudioKind = 'image' | 'video' | 'feedback' | 'lab' | 'news' | 'certificate';
export type StudioStatus = 'pending' | 'processing' | 'completed' | 'failed';

export type StudioItem = {
    id: string;
    kind: StudioKind;
    title: string;
    input: Record<string, any>;
    content?: any;
    file_url?: string;
    status: StudioStatus;
    error_message?: string;
    created_at: string;
    updated_at: string;
};

export type StudioAvailability = Record<string, { available: boolean; missing_env: string }>;

export function newRequestKey(): string {
    if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) return crypto.randomUUID();
    return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

/** Absolute URL for a file served from the API (/uploads/...) */
export function fileUrl(path?: string | null): string | null {
    if (!path) return null;
    if (path.startsWith('http')) return path;
    const base = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000/api/v1';
    return `${base}${path}`;
}

export function isBusy(s?: StudioStatus) {
    return s === 'pending' || s === 'processing';
}

/** Readable error text from an API error */
export function studioError(err: any, fallback: string): { code?: string; message: string } {
    const e = err?.response?.data?.error;
    return { code: e?.code, message: e?.message || err?.message || fallback };
}

export const studioApi = {
    status: () => api.get('/ai-studio/status').then((r) => r.data.data as StudioAvailability),
    list: (kind: StudioKind) => api.get('/ai-studio', { params: { kind } }).then((r) => (r.data.data || []) as StudioItem[]),
    get: (id: string) => api.get(`/ai-studio/${id}`).then((r) => r.data.data as StudioItem),
    create: (kind: StudioKind, input: Record<string, any>, requestKey: string, title?: string) =>
        api.post('/ai-studio', { kind, input, request_key: requestKey, title }).then((r) => r.data.data as StudioItem),
    update: (id: string, title: string, content: any) =>
        api.put(`/ai-studio/${id}`, { title, content }).then((r) => r.data.data as StudioItem),
    retry: (id: string) => api.post(`/ai-studio/${id}/retry`).then((r) => r.data.data as StudioItem),
    remove: (id: string) => api.delete(`/ai-studio/${id}`),
    /** Downloads the generated file through the authenticated API */
    download: async (item: StudioItem) => {
        const res = await api.get(`/ai-studio/${item.id}/download`, { responseType: 'blob' });
        const ext = item.kind === 'video' ? 'mp4' : 'png';
        const url = URL.createObjectURL(res.data);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${(item.title || item.kind).slice(0, 40).replace(/[\\/:*?"<>|]+/g, '_')}.${ext}`;
        document.body.appendChild(a);
        a.click();
        a.remove();
        setTimeout(() => URL.revokeObjectURL(url), 2000);
    },
};
