'use client';

import { useState, useEffect, useRef } from 'react';
import { useRouter } from '@/i18n/routing';
import { useTranslations, useLocale } from 'next-intl';
import { ChevronLeft, Plus, BookOpen, Calendar, XCircle, Loader2, Sparkles, Pencil, CheckCircle2, ExternalLink, Trophy, Users } from 'lucide-react';
import api from '@/lib/api';
import { motion, AnimatePresence } from 'framer-motion';

interface CompletedToast {
    id: string;
    topic: string;
}

export default function TugOfWarPage() {
    const router = useRouter();
    const t = useTranslations('tug_of_war');
    const tCommon = useTranslations('common');
    const locale = useLocale();

    const [loading, setLoading] = useState(true);
    const [materials, setMaterials] = useState<any[]>([]);
    const [toasts, setToasts] = useState<CompletedToast[]>([]);
    const prevStatusesRef = useRef<Record<string, string>>({});

    const dismissToast = (id: string) => {
        setToasts(prev => prev.filter(t => t.id !== id));
    };

    const fetchMaterials = async () => {
        try {
            const response = await api.get('/tug-of-war');
            const newMaterials: any[] = response.data.data.materials || [];

            // Detect generating → completed transitions
            const newlyCompleted = newMaterials.filter(m =>
                prevStatusesRef.current[m.id] === 'generating' &&
                m.status === 'completed'
            );
            if (newlyCompleted.length > 0) {
                setToasts(prev => [
                    ...prev,
                    ...newlyCompleted.map(m => ({ id: m.id, topic: m.topic })),
                ]);
                // Auto-dismiss after 10s
                newlyCompleted.forEach(m => {
                    setTimeout(() => dismissToast(m.id), 10000);
                });
            }

            // Save current statuses for next comparison
            const statusMap: Record<string, string> = {};
            newMaterials.forEach(m => { statusMap[m.id] = m.status; });
            prevStatusesRef.current = statusMap;

            setMaterials(newMaterials);
        } catch (error) {
            console.error('Failed to fetch materials:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchMaterials();
    }, []);

    // Polling for generating materials
    useEffect(() => {
        let interval: NodeJS.Timeout;
        if (materials.some(m => m.status === 'generating')) {
            interval = setInterval(fetchMaterials, 5000);
        }
        return () => clearInterval(interval);
    }, [materials]);

    const handleDelete = async (id: string) => {
        if (!confirm(tCommon('confirm_delete'))) return;
        try {
            await api.delete(`/tug-of-war/${id}`);
            setMaterials(materials.filter(m => m.id !== id));
        } catch (error) {
            console.error('Failed to delete material:', error);
        }
    };

    return (
        <div className="space-y-8 pb-20">
            {/* Header */}
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div className="flex items-center gap-4">
                    <button
                        onClick={() => router.push('/teacher/materials/interactive')}
                        className="p-2 glass-card rounded-xl hover:bg-slate-200 dark:hover:bg-white/10 transition-all text-foreground/60"
                    >
                        <ChevronLeft size={20} />
                    </button>
                    <div>
                        <h1 className="text-3xl font-bold font-outfit">{t('title')}</h1>
                        <p className="text-foreground/60">{t('description')}</p>
                    </div>
                </div>

                <button
                    onClick={() => router.push('/teacher/materials/tug-of-war/create')}
                    className="px-6 py-3 bg-primary text-white rounded-xl font-bold hover:bg-primary-hover transition-all flex items-center justify-center gap-2 shadow-lg shadow-primary/25"
                >
                    <Plus size={20} />
                    {t('create_new')}
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
                    {loading && materials.length === 0 ? (
                        Array(3).fill(0).map((_, i) => (
                            <div key={i} className="glass-card p-6 rounded-[32px] h-48 animate-pulse bg-slate-100 dark:bg-white/5" />
                        ))
                    ) : materials.length === 0 ? (
                        <div className="col-span-full py-20 text-center glass-card rounded-[32px] bg-slate-50 dark:bg-white/5">
                            <div className="w-20 h-20 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-4">
                                <Sparkles className="text-primary" size={32} />
                            </div>
                            <h3 className="text-xl font-bold mb-2">{t('empty_title')}</h3>
                            <p className="text-foreground/60 mb-6">{t('empty_description')}</p>
                            <button
                                onClick={() => router.push('/teacher/materials/tug-of-war/create')}
                                className="px-8 py-3 bg-primary text-white rounded-xl font-bold transition-all hover:scale-105"
                            >
                                {t('create_new')}
                            </button>
                        </div>
                    ) : (
                        materials.map((m) => (
                            <div
                                key={m.id}
                                className={`
                                    glass-card p-6 rounded-[32px] group relative transition-all duration-500
                                    ${m.status === 'generating' ? 'opacity-70 blur-[0.5px]' : 'hover:scale-[1.02] hover:shadow-2xl hover:shadow-primary/10'}
                                `}
                            >
                                {/* Card Top Row */}
                                <div className="flex justify-between items-start mb-4">
                                    <div className={`
                                        w-12 h-12 rounded-2xl flex items-center justify-center text-2xl
                                        ${m.status === 'completed' ? 'bg-green-500/20' : m.status === 'failed' ? 'bg-red-500/20' : 'bg-primary/20 animate-pulse'}
                                    `}>
                                        {m.status === 'completed' ? '⚔️' : m.status === 'failed' ? '⚠️' : '🪄'}
                                    </div>

                                    {/* Action buttons */}
                                    <div className="flex gap-1">
                                        {m.status === 'generating' && (
                                            <span className="px-3 py-1 bg-primary/20 text-primary text-xs font-bold rounded-full animate-pulse">
                                                {t('status_generating')}
                                            </span>
                                        )}
                                        <button
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                handleDelete(m.id);
                                            }}
                                            className="p-2 hover:bg-red-500/20 hover:text-red-500 rounded-lg transition-all"
                                        >
                                            <XCircle size={18} />
                                        </button>
                                    </div>
                                </div>

                                {/* Card Info */}
                                <h3 className="text-xl font-bold mb-2 line-clamp-1">{m.topic}</h3>
                                <div className="space-y-2 text-sm text-foreground/60 mb-6">
                                    <p className="flex items-center gap-2">
                                        <BookOpen size={14} /> {m.subject} • {m.grade} {tCommon('grade_suffix')}
                                    </p>
                                    <p className="flex items-center gap-2 text-[10px] uppercase font-bold tracking-wider">
                                        <Users size={14} /> {m.group_1_name} vs {m.group_2_name}
                                    </p>
                                    <p className="flex items-center gap-2">
                                        <Calendar size={14} /> {new Date(m.created_at).toLocaleDateString()}
                                    </p>
                                </div>

                                {/* Card Actions */}
                                <button
                                    onClick={() => m.status === 'completed' && router.push(`/teacher/materials/tug-of-war/${m.id}`)}
                                    disabled={m.status === 'generating'}
                                    className={`
                                        w-full py-3 rounded-xl font-bold transition-all flex items-center justify-center gap-2
                                        ${m.status === 'completed' ? 'bg-primary text-white hover:bg-primary-hover shadow-lg shadow-primary/20' : 'bg-slate-100 dark:bg-white/5 text-foreground/40 cursor-not-allowed'}
                                    `}
                                >
                                    {m.status === 'completed' ? (
                                        <>
                                            <Trophy size={18} />
                                            {t('game.start')}
                                        </>
                                    ) : (
                                        <>
                                            <Loader2 size={18} className="animate-spin" />
                                            {t('status_generating')}
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
                                    Арқан тартыс дайын! 🎉
                                </p>
                                <p className="text-xs text-foreground/50 truncate mt-0.5">{toast.topic}</p>
                                <button
                                    onClick={() => {
                                        router.push(`/teacher/materials/tug-of-war/${toast.id}`);
                                        dismissToast(toast.id);
                                    }}
                                    className="mt-2 flex items-center gap-1 text-xs font-bold text-green-500 hover:text-green-400 transition-colors"
                                >
                                    <ExternalLink size={13} />
                                    Ашу
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
