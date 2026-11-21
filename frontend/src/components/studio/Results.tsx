'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { Copy, Download, Loader2, Plus, Printer, RefreshCw, Save, Trash2, Check } from 'lucide-react';
import { StudioItem, fileUrl, studioApi } from '@/lib/studio';
import type { ResultHelpers } from './StudioTool';

function useCopy() {
    const [copied, setCopied] = React.useState(false);
    const copy = async (text: string) => {
        try {
            await navigator.clipboard.writeText(text);
        } catch {
            const ta = document.createElement('textarea');
            ta.value = text; document.body.appendChild(ta); ta.select();
            document.execCommand('copy'); ta.remove();
        }
        setCopied(true);
        setTimeout(() => setCopied(false), 1800);
    };
    return { copied, copy };
}

function DownloadButton({ item }: { item: StudioItem }) {
    const t = useTranslations('studio');
    const [busy, setBusy] = React.useState(false);
    const [err, setErr] = React.useState(false);
    return (
        <>
            <button className="btn-primary" disabled={busy} onClick={async () => {
                setBusy(true); setErr(false);
                try { await studioApi.download(item); } catch { setErr(true); } finally { setBusy(false); }
            }}>
                {busy ? <Loader2 size={18} className="animate-spin" /> : <Download size={18} />} {t('download')}
            </button>
            {err && <span role="alert" className="text-sm text-red-700">{t('download_failed')}</span>}
        </>
    );
}

/* ------------------------------------------------------------ image / video */

export function MediaResult({ item, h }: { item: StudioItem; h: ResultHelpers }) {
    const t = useTranslations('studio');
    const src = fileUrl(item.file_url);
    return (
        <div className="space-y-4">
            <h2 className="font-bold text-lg">{item.title}</h2>
            <div className="rounded-2xl overflow-hidden bg-black/5 dark:bg-white/5 flex items-center justify-center">
                {item.kind === 'video' ? (
                    <video src={src || undefined} controls playsInline className="w-full max-h-[480px] bg-black" aria-label={item.title} />
                ) : (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img src={src || ''} alt={item.input?.prompt || item.title} className="w-full max-h-[520px] object-contain" />
                )}
            </div>
            <div className="flex flex-wrap items-center gap-2">
                <DownloadButton item={item} />
                <button className="btn-ghost" onClick={h.regenerate}><RefreshCw size={18} /> {t('new_version')}</button>
            </div>
        </div>
    );
}

/* ------------------------------------------------------------ feedback template */

type FeedbackContent = { title?: string; instructions?: string; questions: { text: string }[] };

export function FeedbackEditor({ item, h }: { item: StudioItem; h: ResultHelpers }) {
    const t = useTranslations('studio');
    const tCommon = useTranslations('common');
    const initial: FeedbackContent = { questions: [], ...(item.content || {}) };
    const [title, setTitle] = React.useState(item.title);
    const [instructions, setInstructions] = React.useState(initial.instructions || '');
    const [questions, setQuestions] = React.useState<string[]>((initial.questions || []).map((q: any) => (typeof q === 'string' ? q : q.text || '')));
    const { copied, copy } = useCopy();

    React.useEffect(() => {
        const c: FeedbackContent = { questions: [], ...(item.content || {}) };
        setTitle(item.title);
        setInstructions(c.instructions || '');
        setQuestions((c.questions || []).map((q: any) => (typeof q === 'string' ? q : q.text || '')));
    }, [item.id, item.updated_at]); // eslint-disable-line react-hooks/exhaustive-deps

    const asText = () => [title, instructions, '', ...questions.filter(Boolean).map((q, i) => `${i + 1}. ${q}`)].filter((x) => x !== undefined).join('\n');

    return (
        <div className="space-y-4">
            <div className="print-area space-y-4">
                <div>
                    <label htmlFor="fb-title" className="field-label">{t('template_title')}</label>
                    <input id="fb-title" className="field font-bold" value={title} onChange={(e) => setTitle(e.target.value)} />
                </div>
                <div>
                    <label htmlFor="fb-instr" className="field-label">{t('instructions')}</label>
                    <textarea id="fb-instr" className="field" rows={2} value={instructions} onChange={(e) => setInstructions(e.target.value)} />
                </div>
                <fieldset>
                    <legend className="field-label">{t('questions')}</legend>
                    <ol className="space-y-2">
                        {questions.map((q, i) => (
                            <li key={i} className="flex gap-2 items-start">
                                <span className="w-7 h-11 flex items-center justify-center font-bold text-primary">{i + 1}.</span>
                                <textarea aria-label={`${t('question')} ${i + 1}`} className="field flex-1" rows={2} value={q}
                                    onChange={(e) => setQuestions((qs) => qs.map((x, j) => (j === i ? e.target.value : x)))} />
                                <button onClick={() => setQuestions((qs) => qs.filter((_, j) => j !== i))}
                                    className="p-2.5 rounded-lg text-red-600 hover:bg-red-500/10" aria-label={`${tCommon('delete')} ${i + 1}`}>
                                    <Trash2 size={16} />
                                </button>
                            </li>
                        ))}
                    </ol>
                    <button onClick={() => setQuestions((qs) => [...qs, ''])} className="btn-ghost text-sm mt-3"><Plus size={16} /> {t('add_question')}</button>
                </fieldset>
            </div>
            <div className="flex flex-wrap gap-2 pt-2 border-t border-line">
                <button className="btn-primary" disabled={h.saving || !title.trim()}
                    onClick={() => h.save(title, { title, instructions, questions: questions.filter((q) => q.trim()).map((text) => ({ text })) })}>
                    {h.saving ? <Loader2 size={18} className="animate-spin" /> : <Save size={18} />} {tCommon('save')}
                </button>
                <button className="btn-ghost" onClick={() => copy(asText())}>{copied ? <Check size={18} /> : <Copy size={18} />} {copied ? t('copied') : t('copy')}</button>
                <button className="btn-ghost" onClick={() => window.print()}><Printer size={18} /> {t('print')}</button>
                <button className="btn-ghost" onClick={h.regenerate}><RefreshCw size={18} /> {t('new_version')}</button>
            </div>
        </div>
    );
}

/* ------------------------------------------------------------ documents: lab / news */

type Section = { heading: string; body: string };

export function SectionsEditor({ item, h }: { item: StudioItem; h: ResultHelpers }) {
    const t = useTranslations('studio');
    const tCommon = useTranslations('common');
    const [title, setTitle] = React.useState(item.title);
    const [sections, setSections] = React.useState<Section[]>(item.content?.sections || []);
    const [editing, setEditing] = React.useState(false);
    const { copied, copy } = useCopy();

    React.useEffect(() => {
        setTitle(item.content?.title || item.title);
        setSections(item.content?.sections || []);
        setEditing(false);
    }, [item.id, item.updated_at]); // eslint-disable-line react-hooks/exhaustive-deps

    const asText = () => [title, '', ...sections.map((s) => (s.heading ? `${s.heading}\n${s.body}` : s.body))].join('\n\n');

    return (
        <div className="space-y-4">
            {editing ? (
                <div className="space-y-4">
                    <div>
                        <label htmlFor="doc-title" className="field-label">{t('doc_title')}</label>
                        <input id="doc-title" className="field font-bold" value={title} onChange={(e) => setTitle(e.target.value)} />
                    </div>
                    {sections.map((s, i) => (
                        <div key={i} className="p-4 rounded-2xl border border-line space-y-2">
                            <div className="flex gap-2">
                                <input aria-label={`${t('section_heading')} ${i + 1}`} className="field font-semibold" value={s.heading}
                                    onChange={(e) => setSections((ss) => ss.map((x, j) => (j === i ? { ...x, heading: e.target.value } : x)))} />
                                <button onClick={() => setSections((ss) => ss.filter((_, j) => j !== i))}
                                    className="p-2.5 rounded-lg text-red-600 hover:bg-red-500/10" aria-label={`${tCommon('delete')} ${i + 1}`}>
                                    <Trash2 size={16} />
                                </button>
                            </div>
                            <textarea aria-label={`${t('section_body')} ${i + 1}`} className="field" rows={5} value={s.body}
                                onChange={(e) => setSections((ss) => ss.map((x, j) => (j === i ? { ...x, body: e.target.value } : x)))} />
                        </div>
                    ))}
                    <button onClick={() => setSections((ss) => [...ss, { heading: '', body: '' }])} className="btn-ghost text-sm"><Plus size={16} /> {t('add_section')}</button>
                </div>
            ) : (
                <article className="print-area space-y-4 leading-relaxed">
                    <h2 className="text-xl font-bold font-outfit">{title}</h2>
                    {sections.map((s, i) => (
                        <section key={i}>
                            {s.heading && <h3 className="font-bold text-primary mb-1">{s.heading}</h3>}
                            <p className="whitespace-pre-line">{s.body}</p>
                        </section>
                    ))}
                </article>
            )}
            <div className="flex flex-wrap gap-2 pt-2 border-t border-line">
                {editing ? (
                    <button className="btn-primary" disabled={h.saving || !title.trim()}
                        onClick={async () => { await h.save(title, { title, sections }); setEditing(false); }}>
                        {h.saving ? <Loader2 size={18} className="animate-spin" /> : <Save size={18} />} {tCommon('save')}
                    </button>
                ) : (
                    <button className="btn-primary" onClick={() => setEditing(true)}>{tCommon('edit')}</button>
                )}
                <button className="btn-ghost" onClick={() => copy(asText())}>{copied ? <Check size={18} /> : <Copy size={18} />} {copied ? t('copied') : t('copy')}</button>
                <button className="btn-ghost" onClick={() => window.print()}><Printer size={18} /> {t('print')}</button>
                <button className="btn-ghost" onClick={h.regenerate}><RefreshCw size={18} /> {t('new_version')}</button>
            </div>
        </div>
    );
}

/* ------------------------------------------------------------ certificate */

export function CertificateView({ item, h }: { item: StudioItem; h: ResultHelpers }) {
    const t = useTranslations('studio');
    const tCommon = useTranslations('common');
    const pick = (name: string) => (item.content?.sections || []).find((s: Section) => s.heading?.toLowerCase() === name)?.body || '';
    const [heading, setHeading] = React.useState('');
    const [recipient, setRecipient] = React.useState('');
    const [text, setText] = React.useState('');
    const [signature, setSignature] = React.useState('');

    React.useEffect(() => {
        setHeading(item.content?.title || item.title);
        setRecipient(pick('recipient') || item.input?.student_name || '');
        setText(pick('text') || (item.content?.sections?.[1]?.body ?? ''));
        setSignature(pick('signature') || [item.input?.teacher_name, item.input?.school].filter(Boolean).join(', '));
    }, [item.id, item.updated_at]); // eslint-disable-line react-hooks/exhaustive-deps

    return (
        <div className="space-y-5">
            <div className="print-area rounded-3xl p-2 bg-gradient-to-br from-primary to-secondary">
                <div className="rounded-[20px] bg-white text-slate-900 px-6 py-10 md:px-12 text-center space-y-5 border-4 border-white">
                    <p className="text-3xl md:text-4xl font-extrabold tracking-wide text-primary font-outfit uppercase">{heading}</p>
                    <p className="text-2xl md:text-3xl font-bold">{recipient}</p>
                    <p className="text-base md:text-lg leading-relaxed max-w-xl mx-auto whitespace-pre-line">{text}</p>
                    <div className="pt-6 flex flex-col sm:flex-row items-center justify-between gap-3 text-sm text-slate-600">
                        <span>{new Date().toLocaleDateString()}</span>
                        <span className="font-semibold">{signature}</span>
                    </div>
                </div>
            </div>

            <details className="rounded-2xl border border-line p-4">
                <summary className="cursor-pointer font-semibold">{t('edit_text')}</summary>
                <div className="grid gap-3 mt-4">
                    <label className="field-label" htmlFor="c-h">{t('cert_heading')}</label>
                    <input id="c-h" className="field" value={heading} onChange={(e) => setHeading(e.target.value)} />
                    <label className="field-label" htmlFor="c-r">{t('cert_recipient')}</label>
                    <input id="c-r" className="field" value={recipient} onChange={(e) => setRecipient(e.target.value)} />
                    <label className="field-label" htmlFor="c-t">{t('cert_text')}</label>
                    <textarea id="c-t" className="field" rows={4} value={text} onChange={(e) => setText(e.target.value)} />
                    <label className="field-label" htmlFor="c-s">{t('cert_signature')}</label>
                    <input id="c-s" className="field" value={signature} onChange={(e) => setSignature(e.target.value)} />
                </div>
            </details>

            <div className="flex flex-wrap gap-2">
                <button className="btn-primary" disabled={h.saving}
                    onClick={() => h.save(item.title, {
                        title: heading,
                        sections: [{ heading: 'recipient', body: recipient }, { heading: 'text', body: text }, { heading: 'signature', body: signature }],
                    })}>
                    {h.saving ? <Loader2 size={18} className="animate-spin" /> : <Save size={18} />} {tCommon('save')}
                </button>
                <button className="btn-ghost" onClick={() => window.print()}><Printer size={18} /> {t('print_pdf')}</button>
                <button className="btn-ghost" onClick={h.regenerate}><RefreshCw size={18} /> {t('new_version')}</button>
            </div>
        </div>
    );
}
