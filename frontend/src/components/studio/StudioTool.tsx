'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { useSearchParams } from 'next/navigation';
import { AlertTriangle, Loader2, RotateCcw, Sparkles, Trash2, Wand2, CheckCircle2, Clock } from 'lucide-react';
import api from '@/lib/api';
import PageHeader from '@/components/ui/PageHeader';
import { EmptyState } from '@/components/dashboard/Charts';
import {
    StudioAvailability, StudioItem, StudioKind, isBusy, newRequestKey, studioApi, studioError,
} from '@/lib/studio';

export type FieldDef = {
    name: string;
    label: string;
    type: 'text' | 'textarea' | 'select' | 'number' | 'class';
    required?: boolean;
    options?: { value: string; label: string }[];
    placeholder?: string;
    defaultValue?: string;
    help?: string;
    min?: number;
    max?: number;
};

export type ResultHelpers = {
    /** Save edited content (text kinds) */
    save: (title: string, content: any) => Promise<void>;
    /** Generate another version with the same input */
    regenerate: () => void;
    saving: boolean;
};

type Props = {
    kind: StudioKind;
    title: string;
    description: string;
    backHref: string;
    fields: FieldDef[];
    submitLabel: string;
    busyHint?: string;
    renderResult: (item: StudioItem, h: ResultHelpers) => React.ReactNode;
};

export function StatusBadge({ status }: { status: StudioItem['status'] }) {
    const t = useTranslations('studio');
    const map = {
        pending: { cls: 'bg-amber-100 text-amber-800 dark:bg-amber-500/15 dark:text-amber-300', icon: <Clock size={12} /> },
        processing: { cls: 'bg-blue-100 text-blue-800 dark:bg-blue-500/15 dark:text-blue-300', icon: <Loader2 size={12} className="animate-spin" /> },
        completed: { cls: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-500/15 dark:text-emerald-300', icon: <CheckCircle2 size={12} /> },
        failed: { cls: 'bg-red-100 text-red-800 dark:bg-red-500/15 dark:text-red-300', icon: <AlertTriangle size={12} /> },
    }[status];
    return (
        <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-bold ${map.cls}`}>
            {map.icon}{t(`status_${status}`)}
        </span>
    );
}

export default function StudioTool({ kind, title, description, backHref, fields, submitLabel, busyHint, renderResult }: Props) {
    const t = useTranslations('studio');
    const tCommon = useTranslations('common');
    const searchParams = useSearchParams();

    const initialValues = React.useMemo(
        () => Object.fromEntries(fields.map((f) => [f.name, f.defaultValue ?? ''])) as Record<string, string>,
        [fields],
    );
    const [values, setValues] = React.useState<Record<string, string>>(initialValues);
    const [items, setItems] = React.useState<StudioItem[]>([]);
    const [selectedId, setSelectedId] = React.useState<string | null>(searchParams.get('id'));
    const [submitting, setSubmitting] = React.useState(false);
    const [saving, setSaving] = React.useState(false);
    const [error, setError] = React.useState<string | null>(null);
    const [notice, setNotice] = React.useState<string | null>(null);
    const [availability, setAvailability] = React.useState<StudioAvailability | null>(null);
    const [classes, setClasses] = React.useState<any[]>([]);
    const requestKey = React.useRef<string>(newRequestKey());

    const selected = items.find((i) => i.id === selectedId) || null;
    const hasClassField = fields.some((f) => f.type === 'class');

    const load = React.useCallback(async () => {
        try {
            const list = await studioApi.list(kind);
            setItems(list);
            setSelectedId((cur) => cur ?? list[0]?.id ?? null);
        } catch (e) {
            setError(studioError(e, tCommon('error_occurred')).message);
        }
    }, [kind, tCommon]);

    // Prefill handed over from another page (e.g. Achievements -> Certificate)
    React.useEffect(() => {
        try {
            const raw = sessionStorage.getItem(`studio_prefill_${kind}`);
            if (raw) {
                sessionStorage.removeItem(`studio_prefill_${kind}`);
                const data = JSON.parse(raw);
                setValues((v) => ({ ...v, ...Object.fromEntries(Object.entries(data).map(([k, val]) => [k, String(val)])) }));
            }
        } catch { /* storage unavailable */ }
    }, [kind]);

    React.useEffect(() => {
        load();
        studioApi.status().then(setAvailability).catch(() => setAvailability(null));
        if (hasClassField) api.get('/classes').then((r) => setClasses(r.data.data || [])).catch(() => { });
    }, [load, hasClassField]);

    // Poll unfinished items
    const busyIds = items.filter((i) => isBusy(i.status)).map((i) => i.id).join(',');
    React.useEffect(() => {
        if (!busyIds) return;
        const every = kind === 'video' ? 8000 : 4000;
        const timer = setInterval(async () => {
            for (const id of busyIds.split(',')) {
                try {
                    const fresh = await studioApi.get(id);
                    setItems((prev) => prev.map((p) => (p.id === id ? fresh : p)));
                } catch { /* keep polling */ }
            }
        }, every);
        return () => clearInterval(timer);
    }, [busyIds, kind]);

    const upsert = (item: StudioItem) =>
        setItems((prev) => [item, ...prev.filter((p) => p.id !== item.id)]);

    const submit = async (input: Record<string, any>) => {
        if (submitting) return; // double click guard (the server also de-duplicates by request key)
        setSubmitting(true);
        setError(null);
        setNotice(null);
        try {
            const item = await studioApi.create(kind, input, requestKey.current);
            upsert(item);
            setSelectedId(item.id);
            requestKey.current = newRequestKey(); // next click is a new request
        } catch (e) {
            const err = studioError(e, tCommon('error_occurred'));
            setError(err.code === 'AI_NOT_CONFIGURED' ? t('not_configured_error', { detail: err.message })
                : err.code === 'INSUFFICIENT_TOKENS' ? t('no_tokens') : err.message);
        } finally {
            setSubmitting(false);
        }
    };

    const onSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const input: Record<string, any> = {};
        for (const f of fields) {
            const v = (values[f.name] ?? '').trim();
            if (f.type === 'class') {
                const cls = classes.find((c) => String(c.id) === v);
                if (cls) input.class_name = cls.name;
                continue;
            }
            if (v !== '') input[f.name] = v;
        }
        submit(input);
    };

    const retry = async (item: StudioItem) => {
        setError(null);
        try {
            upsert(await studioApi.retry(item.id));
        } catch (e) {
            setError(studioError(e, tCommon('error_occurred')).message);
        }
    };

    const remove = async (item: StudioItem) => {
        if (!confirm(tCommon('confirm_delete'))) return;
        try {
            await studioApi.remove(item.id);
            setItems((prev) => prev.filter((p) => p.id !== item.id));
            if (selectedId === item.id) setSelectedId(null);
        } catch (e) {
            setError(studioError(e, tCommon('error_occurred')).message);
        }
    };

    const helpers: ResultHelpers = {
        saving,
        save: async (newTitle, content) => {
            if (!selected) return;
            setSaving(true);
            setError(null);
            try {
                upsert(await studioApi.update(selected.id, newTitle, content));
                setNotice(t('saved'));
            } catch (e) {
                setError(studioError(e, tCommon('error_occurred')).message);
            } finally {
                setSaving(false);
            }
        },
        regenerate: () => selected && submit(selected.input),
    };

    const fillFromItem = (item: StudioItem) => {
        const next: Record<string, string> = { ...initialValues };
        for (const f of fields) if (item.input?.[f.name] != null) next[f.name] = String(item.input[f.name]);
        setValues(next);
    };

    const unavailable = availability && availability[kind] && !availability[kind].available ? availability[kind] : null;

    return (
        <div>
            <PageHeader title={title} description={description} backHref={backHref} />

            {unavailable && (
                <div role="alert" className="mb-6 flex gap-3 p-4 rounded-2xl border border-amber-300 bg-amber-50 text-amber-900 dark:bg-amber-500/10 dark:text-amber-200 dark:border-amber-500/30">
                    <AlertTriangle className="shrink-0 mt-0.5" size={20} />
                    <p className="text-sm">{t('not_configured', { env: unavailable.missing_env })}</p>
                </div>
            )}

            <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
                {/* Form */}
                <form onSubmit={onSubmit} className="glass-card rounded-3xl p-6 lg:col-span-2 space-y-4 self-start">
                    {fields.map((f) => {
                        const id = `f-${f.name}`;
                        const common = {
                            id, name: f.name, required: f.required, value: values[f.name] ?? '',
                            onChange: (e: React.ChangeEvent<any>) => setValues((v) => ({ ...v, [f.name]: e.target.value })),
                            className: 'field', 'aria-describedby': f.help ? `${id}-help` : undefined,
                        };
                        return (
                            <div key={f.name}>
                                <label htmlFor={id} className="field-label">
                                    {f.label}{f.required && <span className="text-red-600" aria-hidden> *</span>}
                                </label>
                                {f.type === 'textarea' ? (
                                    <textarea {...common} rows={4} placeholder={f.placeholder} />
                                ) : f.type === 'select' ? (
                                    <select {...common}>
                                        {f.options?.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
                                    </select>
                                ) : f.type === 'class' ? (
                                    <select {...common}>
                                        <option value="">{classes.length ? t('choose_class') : t('no_classes_option')}</option>
                                        {classes.map((c) => <option key={c.id} value={String(c.id)}>{c.name}</option>)}
                                    </select>
                                ) : (
                                    <input {...common} type={f.type === 'number' ? 'number' : 'text'} min={f.min} max={f.max} placeholder={f.placeholder} />
                                )}
                                {f.help && <p id={`${id}-help`} className="text-xs text-muted mt-1">{f.help}</p>}
                            </div>
                        );
                    })}

                    {error && <p role="alert" className="text-sm text-red-700 dark:text-red-400 bg-red-50 dark:bg-red-500/10 p-3 rounded-xl">{error}</p>}

                    <button type="submit" className="btn-primary w-full" disabled={submitting || !!unavailable}>
                        {submitting ? <Loader2 size={18} className="animate-spin" /> : <Wand2 size={18} />}
                        {submitting ? t('sending') : submitLabel}
                    </button>
                </form>

                {/* Result */}
                <section className="glass-card rounded-3xl p-6 lg:col-span-3 min-h-[320px]" aria-live="polite">
                    {!selected ? (
                        <EmptyState icon={<Sparkles />} text={t('result_empty')} />
                    ) : isBusy(selected.status) ? (
                        <div className="flex flex-col items-center justify-center text-center gap-4 py-12" role="status">
                            <Loader2 size={44} className="text-primary animate-spin" />
                            <div>
                                <p className="font-bold text-lg">{t('processing_title')}</p>
                                <p className="text-sm text-muted mt-1">{busyHint || t('processing_hint')}</p>
                            </div>
                            <p className="text-sm font-medium">«{selected.title}»</p>
                        </div>
                    ) : selected.status === 'failed' ? (
                        <div className="flex flex-col items-center text-center gap-4 py-10" role="alert">
                            <span className="w-14 h-14 rounded-2xl bg-red-100 text-red-700 dark:bg-red-500/15 dark:text-red-300 flex items-center justify-center"><AlertTriangle /></span>
                            <div>
                                <p className="font-bold text-lg">{t('failed_title')}</p>
                                <p className="text-sm text-muted mt-1 max-w-md break-words">{selected.error_message || t('failed_unknown')}</p>
                            </div>
                            <div className="flex flex-wrap gap-2 justify-center">
                                <button className="btn-primary" onClick={() => retry(selected)}><RotateCcw size={18} /> {t('retry')}</button>
                                <button className="btn-ghost" onClick={() => fillFromItem(selected)}>{t('edit_input')}</button>
                            </div>
                        </div>
                    ) : (
                        <div>
                            {notice && <p role="status" className="mb-4 text-sm text-emerald-800 bg-emerald-50 dark:bg-emerald-500/10 dark:text-emerald-300 p-3 rounded-xl">{notice}</p>}
                            {renderResult(selected, helpers)}
                        </div>
                    )}
                </section>
            </div>

            {/* History */}
            <section className="mt-6 glass-card rounded-3xl p-6">
                <h2 className="text-lg font-bold font-outfit mb-4">{t('history')}</h2>
                {items.length === 0 ? (
                    <p className="text-sm text-muted">{t('history_empty')}</p>
                ) : (
                    <ul className="divide-y divide-line">
                        {items.map((it) => (
                            <li key={it.id} className="flex items-center gap-3 py-3">
                                <button onClick={() => { setSelectedId(it.id); setNotice(null); }}
                                    aria-current={it.id === selectedId}
                                    className={`flex-1 min-w-0 text-left rounded-xl px-2 py-1 ${it.id === selectedId ? 'bg-primary/10' : 'hover:bg-primary/[0.05]'}`}>
                                    <span className="block font-semibold truncate">{it.title}</span>
                                    <span className="block text-xs text-muted">{new Date(it.created_at).toLocaleString()}</span>
                                </button>
                                <StatusBadge status={it.status} />
                                <button onClick={() => remove(it)} className="p-2 rounded-lg text-red-600 hover:bg-red-500/10" aria-label={`${tCommon('delete')}: ${it.title}`}>
                                    <Trash2 size={16} />
                                </button>
                            </li>
                        ))}
                    </ul>
                )}
            </section>
        </div>
    );
}
