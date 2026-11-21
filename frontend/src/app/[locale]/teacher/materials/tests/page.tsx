'use client';

import { useState, useEffect, useRef } from 'react';
import { useRouter } from '@/i18n/routing';
import { useTranslations, useLocale } from 'next-intl';
import { ChevronLeft, Plus, BookOpen, Calendar, XCircle, Loader2, Sparkles, CheckCircle2, ExternalLink, FileText, ClipboardList } from 'lucide-react';
import api from '@/lib/api';
import { motion, AnimatePresence } from 'framer-motion';

interface CompletedToast {
    id: string;
    topic: string;
}

export default function TestsPage() {
    const router = useRouter();
    const t = useTranslations('education_materials');
    const tTests = useTranslations('tests');
    const tCommon = useTranslations('common');
    const locale = useLocale();

    const [loading, setLoading] = useState(true);
    const [tests, setTests] = useState<any[]>([]);
    const [toasts, setToasts] = useState<CompletedToast[]>([]);
    const prevStatusesRef = useRef<Record<string, string>>({});

    const dismissToast = (id: string) => {
        setToasts((prev: CompletedToast[]) => prev.filter((t: CompletedToast) => t.id !== id));
    };

    const fetchTests = async () => {
        try {
            const response = await api.get('/tests');
            const newTests: any[] = response.data.data.tests || [];

            // Detect generating → completed transitions
            const newlyCompleted = newTests.filter(t =>
                prevStatusesRef.current[t.id] === 'generating' &&
                t.status === 'completed'
            );
            if (newlyCompleted.length > 0) {
                setToasts((prev: CompletedToast[]) => [
                    ...prev,
                    ...newlyCompleted.map((t: any) => ({ id: t.id, topic: t.topic })),
                ]);
                // Auto-dismiss after 10s
                newlyCompleted.forEach(t => {
                    setTimeout(() => dismissToast(t.id), 10000);
                });
            }

            // Save current statuses for next comparison
            const statusMap: Record<string, string> = {};
            newTests.forEach((t: any) => { statusMap[t.id] = t.status; });
            prevStatusesRef.current = statusMap;

            setTests(newTests);
        } catch (error) {
            console.error('Failed to fetch tests:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchTests();
    }, []);

    // Polling for generating tests
    useEffect(() => {
        let interval: NodeJS.Timeout;
        if (tests.some((t: any) => t.status === 'generating')) {
            interval = setInterval(fetchTests, 5000);
        }
        return () => clearInterval(interval);
    }, [tests]);

    const handleDelete = async (id: string) => {
        if (!confirm(tCommon('confirm_delete'))) return;
        try {
            await api.delete(`/tests/${id}`);
            setTests(tests.filter(t => t.id !== id));
        } catch (error) {
            console.error('Failed to delete test:', error);
        }
    };

    return (
        <div className="space-y-8 pb-20">
            {/* Header */}
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div className="flex items-center gap-4">
                    <button
                        onClick={() => router.push('/teacher/materials/assignments')}
                        className="p-2 glass-card rounded-xl hover:bg-slate-200 dark:hover:bg-white/10 transition-all text-foreground/60"
                    >
                        <ChevronLeft size={20} />
                    </button>
                    <div>
                        <h1 className="text-3xl font-bold font-outfit">{tTests('title')}</h1>
                        <p className="text-foreground/60">{tTests('description')}</p>
                    </div>
                </div>

                <button
                    onClick={() => router.push('/teacher/materials/tests/create')}
                    className="px-6 py-3 bg-primary text-white rounded-xl font-bold hover:bg-primary-hover transition-all flex items-center justify-center gap-2 shadow-lg shadow-primary/25"
                >
                    <Plus size={20} />
                    {tTests('create_new')}
                </button>
            </div>

            <AnimatePresence mode="wait">
                <motion.div
                    key="list"
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -20 }}
                    className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
                >
                    {loading && tests.length === 0 ? (
                        Array(3).fill(0).map((_, i) => (
                            <div key={i} className="glass-card p-6 rounded-[32px] h-48 animate-pulse bg-slate-100 dark:bg-white/5" />
                        ))
                    ) : tests.length === 0 ? (
                        <div className="col-span-full py-20 text-center glass-card rounded-[32px] bg-slate-50 dark:bg-white/5">
                            <div className="w-20 h-20 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-4">
                                <Sparkles className="text-primary" size={32} />
                            </div>
                            <h3 className="text-xl font-bold mb-2">{tTests('empty_title')}</h3>
                            <p className="text-foreground/60 mb-6">{tTests('empty_description')}</p>
                            <button
                                onClick={() => router.push('/teacher/materials/tests/create')}
                                className="px-8 py-3 bg-primary text-white rounded-xl font-bold transition-all hover:scale-105"
                            >
                                {tTests('create_new')}
                            </button>
                        </div>
                    ) : (
                        tests.map((test) => (
                            <div
                                key={test.id}
                                className={`
                                    glass-card p-6 rounded-[32px] group relative transition-all duration-500
                                    ${test.status === 'generating' ? 'opacity-70 blur-[0.5px]' : 'hover:scale-[1.02] hover:shadow-2xl hover:shadow-primary/10'}
                                `}
                            >
                                {/* Card Top Row */}
                                <div className="flex justify-between items-start mb-4">
                                    <div className={`
                                        w-12 h-12 rounded-2xl flex items-center justify-center text-2xl
                                        ${test.status === 'completed' ? 'bg-indigo-500/20' : test.status === 'failed' ? 'bg-red-500/20' : 'bg-primary/20 animate-pulse'}
                                    `}>
                                        {test.status === 'completed' ? <FileText className="text-indigo-500" /> : test.status === 'failed' ? '⚠️' : '🪄'}
                                    </div>

                                    {/* Action buttons */}
                                    <div className="flex gap-1">
                                        {test.status === 'generating' && (
                                            <span className="px-3 py-1 bg-primary/20 text-primary text-xs font-bold rounded-full animate-pulse">
                                                {tTests('status_generating')}
                                            </span>
                                        )}
                                        {test.status === 'failed' && (
                                            <span className="px-3 py-1 bg-red-500/20 text-red-500 text-xs font-bold rounded-full">
                                                {tTests('status_failed')}
                                            </span>
                                        )}
                                        <button
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                handleDelete(test.id);
                                            }}
                                            className="p-2 hover:bg-red-500/20 hover:text-red-500 rounded-lg transition-all"
                                        >
                                            <XCircle size={18} />
                                        </button>
                                    </div>
                                </div>

                                {/* Card Info */}
                                <h3 className="text-xl font-bold mb-2 line-clamp-1">{test.topic}</h3>
                                <div className="space-y-2 text-sm text-foreground/60 mb-6">
                                    <p className="flex items-center gap-2">
                                        <BookOpen size={14} /> {test.subject} • {test.grade} {tCommon('grade_suffix')}
                                    </p>
                                    <p className="flex items-center gap-2 text-[10px] uppercase font-bold tracking-wider">
                                        <ClipboardList size={14} /> {test.test_type.toUpperCase()}
                                    </p>
                                    <p className="flex items-center gap-2">
                                        <Calendar size={14} /> {new Date(test.created_at).toLocaleDateString()}
                                    </p>
                                </div>

                                {/* Card Actions */}
                                {test.status === 'completed' && (
                                    <button
                                        onClick={() => router.push('/teacher/tests/assignments')}
                                        className="w-full py-3 mb-3 rounded-xl font-bold bg-slate-100 dark:bg-white/5 text-primary hover:bg-slate-200 dark:hover:bg-white/10 transition-all flex items-center justify-center gap-2 border border-primary/10"
                                    >
                                        <ClipboardList size={18} />
                                        {tTests('assigned_tests_title')}
                                    </button>
                                )}

                                <button
                                    onClick={() => test.status === 'completed' && router.push(`/teacher/materials/tests/${test.id}`)}
                                    disabled={test.status === 'generating'}
                                    className={`
                                        w-full py-3 rounded-xl font-bold transition-all flex items-center justify-center gap-2
                                        ${test.status === 'completed' ? 'bg-primary text-white hover:bg-primary-hover shadow-lg shadow-primary/20' : 'bg-slate-100 dark:bg-white/5 text-foreground/40 cursor-not-allowed'}
                                    `}
                                >
                                    {test.status === 'completed' ? (
                                        <>
                                            <ExternalLink size={18} />
                                            {tTests('open')}
                                        </>
                                    ) : (
                                        <>
                                            <Loader2 size={18} className="animate-spin" />
                                            {tTests('status_generating')}
                                        </>
                                    )}
                                </button>
                            </div>
                        ))
                    )}
                </motion.div>
            </AnimatePresence>

            {/* Toast Stack */}
            <div className="fixed bottom-6 right-6 z-[200] flex flex-col gap-3 pointer-events-none">
                <AnimatePresence>
                    {toasts.map((toast) => (
                        <motion.div
                            key={toast.id}
                            initial={{ opacity: 0, x: 80, scale: 0.95 }}
                            animate={{ opacity: 1, x: 0, scale: 1 }}
                            exit={{ opacity: 0, x: 80, scale: 0.95 }}
                            transition={{ type: 'spring', stiffness: 300, damping: 25 }}
                            className="pointer-events-auto flex items-start gap-3 w-80 bg-background border border-green-500/30 shadow-2xl shadow-green-500/10 rounded-2xl p-4"
                        >
                            <div className="shrink-0 mt-0.5 w-9 h-9 rounded-xl bg-green-500/20 flex items-center justify-center">
                                <CheckCircle2 size={18} className="text-green-500" />
                            </div>
                            <div className="flex-1 min-w-0">
                                <p className="text-sm font-bold text-foreground leading-snug">
                                    {tTests('notification_title')}
                                </p>
                                <p className="text-xs text-foreground/50 truncate mt-0.5">{toast.topic}</p>
                                <button
                                    onClick={() => {
                                        router.push(`/teacher/materials/tests/${toast.id}`);
                                        dismissToast(toast.id);
                                    }}
                                    className="mt-2 flex items-center gap-1 text-xs font-bold text-green-500 hover:text-green-400 transition-colors"
                                >
                                    <ExternalLink size={13} />
                                    {tTests('notification_open')}
                                </button>
                            </div>
                            <button
                                onClick={() => dismissToast(toast.id)}
                                className="shrink-0 text-foreground/30 hover:text-foreground/60 transition-colors mt-0.5"
                            >
                                <XCircle size={16} />
                            </button>
                        </motion.div>
                    ))}
                </AnimatePresence>
            </div>
        </div>
    );
}
