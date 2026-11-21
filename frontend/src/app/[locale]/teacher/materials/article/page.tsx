'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from '@/i18n/routing';
import { useTranslations } from 'next-intl';
import {
    ChevronLeft, Sparkles, Plus,
    Calendar, XCircle, Loader2, BookOpen, FileText
} from 'lucide-react';
import api from '@/lib/api';
import { motion, AnimatePresence } from 'framer-motion';

interface Article {
    id: string;
    subject: string;
    grade: number;
    topic: string;
    language: string;
    description: string;
    style: string;
    complexity: string;
    title?: string;
    status: 'pending' | 'processing' | 'completed' | 'failed';
    error_message?: string;
    created_at: string;
}

export default function ArticlePage() {
    const router = useRouter();
    const t = useTranslations('articles');
    const tCommon = useTranslations('common');

    const [mode, setMode] = useState<'list' | 'create'>('list');
    const [loading, setLoading] = useState(false);
    const [submitting, setSubmitting] = useState(false);
    const [articles, setArticles] = useState<Article[]>([]);

    const [formData, setFormData] = useState({
        subject: '',
        grade: 8,
        topic: '',
        language: 'kaz',
        style: 'academic',
        complexity: 'medium',
        description: '',
    });

    const [helperLoading, setHelperLoading] = useState(false);

    // ─── Load list ───────────────────────────────────────────────────────────
    useEffect(() => {
        if (mode === 'list') fetchArticles();
    }, [mode]);

    // ─── Poll for pending/processing ─────────────────────────────────────────
    useEffect(() => {
        let interval: NodeJS.Timeout;
        const hasActive = articles.some(
            (a: Article) => a.status === 'pending' || a.status === 'processing'
        );
        if (mode === 'list' && hasActive) {
            interval = setInterval(fetchArticles, 5000);
        }
        return () => clearInterval(interval);
    }, [mode, articles]);

    const fetchArticles = async () => {
        setLoading(true);
        try {
            const res = await api.get('/articles');
            setArticles(res.data.data?.items || []);
        } catch (err) {
            console.error('Failed to fetch articles:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setSubmitting(true);
        try {
            await api.post('/articles', {
                ...formData,
                grade: parseInt(formData.grade.toString()),
            });
            setMode('list');
            fetchArticles();
        } catch (error: any) {
            console.error('Create failed:', error);
            alert(error.response?.data?.error?.message || tCommon('error'));
        } finally {
            setSubmitting(false);
        }
    };

    const handleDelete = async (id: string) => {
        if (!confirm(t('delete_confirm'))) return;
        try {
            await api.delete(`/articles/${id}`);
            setArticles((prev: Article[]) => prev.filter((a: Article) => a.id !== id));
        } catch (err) {
            console.error('Delete failed:', err);
        }
    };

    const handleMagicHelper = async () => {
        if (!formData.subject || !formData.topic) {
            alert(t('errors.required'));
            return;
        }

        setHelperLoading(true);
        try {
            const res = await api.post('/ai/suggest-article-instructions', {
                subject: formData.subject,
                grade: formData.grade,
                topic: formData.topic,
                language: formData.language,
            });

            const taskId = res.data.data.id;
            await pollHelperStatus(taskId);
        } catch (error) {
            console.error('Magic helper failed:', error);
            setHelperLoading(false);
        }
    };

    const pollHelperStatus = async (taskId: string) => {
        try {
            const res = await api.get(`/ai/status/${taskId}`);
            const task = res.data.data;

            if (task.status === 'completed') {
                setFormData(prev => ({
                    ...prev,
                    description: (prev.description ? prev.description + '\n\n' : '') + task.response
                }));
                setHelperLoading(false);
            } else if (task.status === 'failed') {
                console.error('AI Suggestion failed');
                setHelperLoading(false);
            } else {
                // Poll again in 2 seconds
                setTimeout(() => pollHelperStatus(taskId), 2000);
            }
        } catch (error) {
            console.error('Polling failed:', error);
            setHelperLoading(false);
        }
    };

    const statusColor = (status: Article['status']) => {
        switch (status) {
            case 'completed': return 'bg-green-500/20 text-green-400';
            case 'failed': return 'bg-red-500/20 text-red-400';
            case 'processing': return 'bg-blue-500/20 text-blue-400 animate-pulse';
            default: return 'bg-primary/20 text-primary animate-pulse';
        }
    };

    const statusLabel = (status: Article['status']) => {
        switch (status) {
            case 'completed': return t('status_completed');
            case 'failed': return t('status_failed');
            case 'processing': return t('status_processing');
            default: return t('status_pending');
        }
    };

    const statusEmoji = (status: Article['status']) => {
        switch (status) {
            case 'completed': return '📄';
            case 'failed': return '⚠️';
            default: return '🪄';
        }
    };

    const isActive = (status: Article['status']) =>
        status === 'pending' || status === 'processing';

    return (
        <div className="space-y-8 pb-20">
            {/* Header */}
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div className="flex items-center gap-4">
                    <button
                        onClick={() =>
                            mode === 'create'
                                ? setMode('list')
                                : router.push('/teacher/materials')
                        }
                        className="p-2 glass-card rounded-xl hover:bg-slate-200 dark:hover:bg-white/10 transition-all text-foreground/60"
                    >
                        <ChevronLeft size={20} />
                    </button>
                    <div>
                        <h1 className="text-3xl font-bold font-outfit">{t('title')}</h1>
                        <p className="text-foreground/60">{t('description')}</p>
                    </div>
                </div>

                {mode === 'list' && (
                    <button
                        onClick={() => setMode('create')}
                        className="px-6 py-3 bg-primary text-white rounded-xl font-bold hover:bg-primary-hover transition-all flex items-center justify-center gap-2 shadow-lg shadow-primary/25"
                    >
                        <Plus size={20} />
                        {t('create_new')}
                    </button>
                )}
            </div>

            {/* Content */}
            <AnimatePresence mode="wait">
                {mode === 'list' ? (
                    <motion.div
                        key="list"
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -20 }}
                        className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
                    >
                        {loading && articles.length === 0
                            ? Array(3).fill(0).map((_, i) => (
                                <div key={i} className="glass-card p-6 rounded-[32px] h-52 animate-pulse bg-slate-100 dark:bg-white/5" />
                            ))
                            : articles.length === 0
                                ? (
                                    <div className="col-span-full py-20 text-center glass-card rounded-[32px] bg-slate-50 dark:bg-white/5">
                                        <div className="w-20 h-20 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-4">
                                            <FileText className="text-primary" size={32} />
                                        </div>
                                        <h3 className="text-xl font-bold mb-2">{t('empty_title')}</h3>
                                        <p className="text-foreground/60 mb-6">{t('empty_description')}</p>
                                        <button
                                            onClick={() => setMode('create')}
                                            className="px-8 py-3 bg-primary text-white rounded-xl font-bold hover:bg-primary-hover transition-all"
                                        >
                                            {t('start_button')}
                                        </button>
                                    </div>
                                )
                                : articles.map((a) => (
                                    <div
                                        key={a.id}
                                        className={`
                                            glass-card p-6 rounded-[32px] group relative transition-all duration-500
                                            ${isActive(a.status)
                                                ? 'opacity-70 blur-[0.5px]'
                                                : 'hover:scale-[1.02] hover:shadow-2xl hover:shadow-primary/10'
                                            }
                                        `}
                                    >
                                        <div className="flex justify-between items-start mb-4">
                                            <div className={`
                                                w-12 h-12 rounded-2xl flex items-center justify-center text-2xl
                                                ${a.status === 'completed' ? 'bg-green-500/20' : a.status === 'failed' ? 'bg-red-500/20' : 'bg-primary/20'}
                                            `}>
                                                {statusEmoji(a.status)}
                                            </div>
                                            <div className="flex items-center gap-2">
                                                <span className={`px-3 py-1 text-xs font-bold rounded-full ${statusColor(a.status)}`}>
                                                    {statusLabel(a.status)}
                                                </span>
                                                <button
                                                    onClick={(e) => { e.stopPropagation(); handleDelete(a.id); }}
                                                    className="opacity-0 group-hover:opacity-100 p-2 hover:bg-red-500/20 hover:text-red-500 rounded-lg transition-all"
                                                >
                                                    <XCircle size={18} />
                                                </button>
                                            </div>
                                        </div>

                                        <h3 className="text-lg font-bold mb-1 line-clamp-2">{a.title || a.topic}</h3>
                                        <div className="space-y-1.5 text-sm text-foreground/60 mb-6">
                                            <p className="flex items-center gap-2">
                                                <BookOpen size={14} />
                                                {a.subject} · {a.grade}{t('form.grade_suffix')}
                                            </p>
                                            <p className="flex items-center gap-2">
                                                <FileText size={14} />
                                                {a.language.toUpperCase()}
                                            </p>
                                            <p className="flex items-center gap-2">
                                                <Calendar size={14} />
                                                {new Date(a.created_at).toLocaleDateString()}
                                            </p>
                                        </div>

                                        <button
                                            onClick={() =>
                                                a.status === 'completed' &&
                                                router.push(`/teacher/materials/article/${a.id}`)
                                            }
                                            disabled={a.status !== 'completed'}
                                            className={`
                                                w-full py-3 rounded-xl font-bold transition-all
                                                ${a.status === 'completed'
                                                    ? 'bg-primary text-white hover:bg-primary-hover shadow-lg shadow-primary/20'
                                                    : 'bg-slate-100 dark:bg-white/5 text-foreground/40 cursor-not-allowed'
                                                }
                                            `}
                                        >
                                            {isActive(a.status) ? t('wait') : t('open')}
                                        </button>
                                    </div>
                                ))
                        }
                    </motion.div>
                ) : (
                    /* Create Form */
                    <motion.div
                        key="create"
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{ opacity: 0, x: -20 }}
                    >
                        <form onSubmit={handleSubmit} className="space-y-6">
                            <div className="glass-card p-8 rounded-[32px] space-y-6">
                                <div className="flex items-center gap-3 px-5 py-4 bg-primary/10 border border-primary/20 rounded-2xl">
                                    <Sparkles className="text-primary shrink-0" size={20} />
                                    <p className="text-sm text-primary font-medium">
                                        Kimi K2 AI автоматически напишет структурированную статью на выбранную тему
                                    </p>
                                </div>

                                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                    <div>
                                        <label className="block text-sm font-semibold mb-2">{t('form.subject_label')}</label>
                                        <input
                                            type="text"
                                            required
                                            value={formData.subject}
                                            onChange={(e) => setFormData({ ...formData, subject: e.target.value })}
                                            className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                                            placeholder={t('form.subject_placeholder')}
                                        />
                                    </div>

                                    <div>
                                        <label className="block text-sm font-semibold mb-2">{t('form.grade_label')}</label>
                                        <select
                                            required
                                            value={formData.grade}
                                            onChange={(e) => setFormData({ ...formData, grade: parseInt(e.target.value) })}
                                            className="w-full px-4 py-3 glass-card bg-white dark:bg-slate-800 rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit text-foreground"
                                        >
                                            {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12].map((g) => (
                                                <option key={g} value={g} className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">
                                                    {g}{t('form.grade_suffix')}
                                                </option>
                                            ))}
                                        </select>
                                    </div>

                                    <div className="md:col-span-2">
                                        <label className="block text-sm font-semibold mb-2">{t('form.topic_label')}</label>
                                        <input
                                            type="text"
                                            required
                                            value={formData.topic}
                                            onChange={(e) => setFormData({ ...formData, topic: e.target.value })}
                                            className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                                            placeholder={t('form.topic_placeholder')}
                                        />
                                    </div>

                                    {/* Language */}
                                    <div>
                                        <label className="block text-sm font-semibold mb-2">{t('form.language_label')}</label>
                                        <select
                                            required
                                            value={formData.language}
                                            onChange={(e) => setFormData({ ...formData, language: e.target.value })}
                                            className="w-full px-4 py-3 glass-card bg-white dark:bg-slate-800 rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit text-foreground"
                                        >
                                            <option value="kaz" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{t('form.lang_kaz')}</option>
                                            <option value="rus" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{t('form.lang_rus')}</option>
                                            <option value="eng" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{t('form.lang_eng')}</option>
                                        </select>
                                    </div>

                                    {/* Style */}
                                    <div>
                                        <label className="block text-sm font-semibold mb-2">{t('form.style_label')}</label>
                                        <select
                                            required
                                            value={formData.style}
                                            onChange={(e) => setFormData({ ...formData, style: e.target.value })}
                                            className="w-full px-4 py-3 glass-card bg-white dark:bg-slate-800 rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit text-foreground"
                                        >
                                            <option value="academic" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{t('form.style_academic')}</option>
                                            <option value="popular" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{t('form.style_popular')}</option>
                                            <option value="creative" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{t('form.style_creative')}</option>
                                        </select>
                                    </div>

                                    {/* Complexity */}
                                    <div className="md:col-span-2">
                                        <label className="block text-sm font-semibold mb-2">{t('form.complexity_label')}</label>
                                        <select
                                            required
                                            value={formData.complexity}
                                            onChange={(e) => setFormData({ ...formData, complexity: e.target.value })}
                                            className="w-full px-4 py-3 glass-card bg-white dark:bg-slate-800 rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit text-foreground"
                                        >
                                            <option value="small" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{t('form.complexity_small')}</option>
                                            <option value="medium" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{t('form.complexity_medium')}</option>
                                            <option value="large" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{t('form.complexity_large')}</option>
                                        </select>
                                    </div>

                                    <div className="md:col-span-2">
                                        <div className="flex items-center justify-between mb-2">
                                            <label className="text-sm font-bold flex items-center gap-2">
                                                <Sparkles size={16} className="text-primary" />
                                                {t('form.description_label')}
                                            </label>
                                            <button
                                                type="button"
                                                onClick={handleMagicHelper}
                                                disabled={helperLoading || !formData.subject || !formData.topic}
                                                className="flex items-center gap-2 text-primary font-bold text-xs bg-primary/10 px-3 py-1.5 rounded-lg hover:bg-primary/20 transition-all disabled:opacity-50"
                                            >
                                                {helperLoading ? (
                                                    <Loader2 size={14} className="animate-spin" />
                                                ) : (
                                                    <Sparkles size={14} />
                                                )}
                                                {helperLoading ? t('form.magic_helper_loading') : t('form.magic_helper')}
                                            </button>
                                        </div>
                                        <textarea
                                            rows={4}
                                            value={formData.description}
                                            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                                            className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit resize-none"
                                            placeholder={t('form.description_placeholder')}
                                        />
                                    </div>
                                </div>

                                <div className="flex items-center justify-end gap-4 pt-6 border-t border-slate-300 dark:border-white/10">
                                    <button
                                        type="button"
                                        onClick={() => setMode('list')}
                                        className="px-6 py-3 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl font-semibold transition-all"
                                    >
                                        {t('form.cancel')}
                                    </button>

                                    <button
                                        type="submit"
                                        disabled={submitting}
                                        className="px-8 py-3 bg-primary text-white font-bold rounded-xl hover:bg-primary-hover disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg shadow-primary/25 flex items-center gap-2"
                                    >
                                        {submitting ? (
                                            <>
                                                <Loader2 className="animate-spin h-5 w-5" />
                                                <span>{t('wait')}</span>
                                            </>
                                        ) : (
                                            `${t('form.create_with_tokens')} (3 ${tCommon('tokens')})`
                                        )}
                                    </button>
                                </div>
                            </div>
                        </form>
                    </motion.div>
                )}
            </AnimatePresence>
        </div>
    );
}
