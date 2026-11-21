'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import { useRouter } from '@/i18n/routing';
import { useTranslations, useLocale } from 'next-intl';
import { ChevronLeft, Loader2, BookOpen, Calendar, FileText, ClipboardList, Send, CheckCircle2, AlertCircle, Trash2, Edit2, Info } from 'lucide-react';
import api from '@/lib/api';
import { motion, AnimatePresence } from 'framer-motion';
import AssignTestModal from '@/components/materials/AssignTestModal';
import GamePlayer from '@/components/game/GamePlayer';
import { isGameType } from '@/lib/testTypes';
import { Play } from 'lucide-react';

export default function TestDetailPage() {
    const params = useParams();
    const id = params.id;
    const router = useRouter();
    const t = useTranslations('education_materials');
    const tTests = useTranslations('tests');
    const tCommon = useTranslations('common');
    const locale = useLocale();

    const [loading, setLoading] = useState(true);
    const [testData, setTestData] = useState<any>(null);
    const [isAssignModalOpen, setIsAssignModalOpen] = useState(false);
    const [previewing, setPreviewing] = useState(false);
    const tTypes = useTranslations('test_types');
    const tGame = useTranslations('game');

    const fetchTestData = async () => {
        try {
            const response = await api.get(`/tests/${id}`);
            setTestData(response.data.data);
        } catch (error) {
            console.error('Failed to fetch test data:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (id) fetchTestData();
    }, [id]); // eslint-disable-line react-hooks/exhaustive-deps

    // While the AI is still generating, refresh every 5 s
    const status = testData?.test?.status;
    useEffect(() => {
        if (status !== 'generating' && status !== 'draft') return;
        const timer = setInterval(fetchTestData, 5000);
        return () => clearInterval(timer);
    }, [status]); // eslint-disable-line react-hooks/exhaustive-deps

    if (loading) {
        return (
            <div className="h-[60vh] flex items-center justify-center">
                <Loader2 size={48} className="animate-spin text-primary opacity-20" />
            </div>
        );
    }

    if (!testData) {
        return (
            <div className="h-[60vh] flex flex-col items-center justify-center gap-4">
                <AlertCircle size={48} className="text-red-500 opacity-20" />
                <p className="text-foreground/60">{tCommon('error_not_found')}</p>
                <button onClick={() => router.back()} className="text-primary font-bold">{tCommon('back')}</button>
            </div>
        );
    }

    const { test, questions } = testData;
    const backHref = isGameType(test.test_type) ? '/teacher/materials/games'
        : ['pisa', 'unt', 'olympiad'].includes(test.test_type) ? `/teacher/materials/${test.test_type}` : '/teacher/materials/tests';

    if (previewing) {
        return (
            <GamePlayer
                mode="preview"
                title={test.topic}
                subtitle={`${test.subject} · ${test.grade} ${tCommon('grade_suffix')}`}
                testType={test.test_type}
                questions={questions}
                exitLabel={tGame('back_to_editor')}
                onExit={() => setPreviewing(false)}
            />
        );
    }

    return (
        <div className="max-w-5xl mx-auto space-y-8 pb-20">
            {/* Header */}
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
                <div className="flex items-center gap-4">
                    <button
                        onClick={() => router.push(backHref)}
                        aria-label={tCommon('back')}
                        className="p-2 glass-card rounded-xl hover:bg-slate-200 dark:hover:bg-white/10 transition-all text-foreground/60"
                    >
                        <ChevronLeft size={20} />
                    </button>
                    <div>
                        <div className="flex items-center gap-3 mb-1">
                            <h1 className="text-3xl font-bold font-outfit">{test.topic}</h1>
                            <span className="px-3 py-1 bg-indigo-500/10 text-indigo-500 text-[10px] font-bold rounded-full uppercase tracking-widest">
                                {tTypes(`${test.test_type}.title`)}
                            </span>
                        </div>
                        <div className="flex items-center gap-4 text-sm text-foreground/60">
                            <p className="flex items-center gap-1.5"><BookOpen size={14} /> {test.subject} • {test.grade} {tCommon('grade_suffix')}</p>
                            <p className="flex items-center gap-1.5"><Calendar size={14} /> {new Date(test.created_at).toLocaleDateString()}</p>
                        </div>
                    </div>
                </div>

                <div className="flex flex-wrap gap-3">
                    {questions.length > 0 && (
                        <button onClick={() => setPreviewing(true)} className="btn-ghost">
                            <Play size={18} /> {isGameType(test.test_type) ? tGame('preview_game') : tGame('preview_test')}
                        </button>
                    )}
                    <button
                        onClick={() => router.push('/teacher/tests/assignments')}
                        className="px-6 py-3 glass-card bg-white dark:bg-white/5 rounded-xl font-bold hover:bg-slate-50 dark:hover:bg-white/10 transition-all flex items-center justify-center gap-2"
                    >
                        <ClipboardList size={18} className="text-primary" />
                        {tTests('assigned_tests_title')}
                    </button>
                    <button
                        onClick={() => setIsAssignModalOpen(true)}
                        disabled={questions.length === 0}
                        className="px-6 py-3 bg-primary text-white rounded-xl font-bold hover:bg-primary-hover transition-all flex items-center justify-center gap-2 shadow-lg shadow-primary/25"
                    >
                        <Send size={18} />
                        {tTests('assign')}
                    </button>
                </div>
            </div>

            {questions.length === 0 && (
                <div className="glass-card rounded-3xl p-8 text-center" role="status">
                    {test.status === 'failed' ? (
                        <p className="text-red-700">{tTests('status_failed')}{test.error_message ? `: ${test.error_message}` : ''}</p>
                    ) : (
                        <p className="flex items-center justify-center gap-2 text-primary"><Loader2 className="animate-spin" size={18} /> {tTests('status_generating')}</p>
                    )}
                </div>
            )}

            {/* Questions List */}
            <div className="space-y-6">
                <h2 className="text-2xl font-bold font-outfit px-2 flex items-center gap-3">
                    <ClipboardList className="text-primary" />
                    {tTests('questions_count')} ({questions.length})
                </h2>

                <div className="grid grid-cols-1 gap-6">
                    {questions.map((q: any, idx: number) => (
                        <motion.div
                            key={q.id}
                            initial={{ opacity: 0, y: 20 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ delay: idx * 0.05 }}
                            className="glass-card rounded-[32px] overflow-hidden group border-transparent hover:border-primary/20 transition-all"
                        >
                            <div className="p-8 space-y-6">
                                <div className="flex items-start justify-between gap-4">
                                    <div className="flex items-start gap-4">
                                        <div className="w-10 h-10 rounded-xl bg-slate-100 dark:bg-white/5 flex items-center justify-center font-bold text-primary shrink-0">
                                            {idx + 1}
                                        </div>
                                        <h3 className="text-xl font-bold leading-tight pt-1">{q.question}</h3>
                                    </div>
                                </div>

                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    {q.options.map((opt: string, optIdx: number) => (
                                        <div
                                            key={optIdx}
                                            className={`
                                                p-4 rounded-2xl border flex items-center gap-3 transition-all
                                                ${opt === q.answer
                                                    ? 'bg-green-500/10 border-green-500/30 text-green-600 dark:text-green-400 font-bold'
                                                    : 'bg-slate-50 dark:bg-white/5 border-transparent text-foreground/60'}
                                            `}
                                        >
                                            <div className={`
                                                w-6 h-6 rounded-lg flex items-center justify-center text-xs font-bold
                                                ${opt === q.answer ? 'bg-green-500 text-white' : 'bg-slate-200 dark:bg-white/10'}
                                            `}>
                                                {String.fromCharCode(65 + optIdx)}
                                            </div>
                                            {opt}
                                            {opt === q.answer && <CheckCircle2 size={16} className="ml-auto" />}
                                        </div>
                                    ))}
                                </div>

                                {q.explanation && (
                                    <div className="p-4 bg-indigo-500/5 rounded-2xl border border-indigo-500/10 flex items-start gap-3">
                                        <Info size={18} className="text-indigo-500 shrink-0 mt-0.5" />
                                        <div className="space-y-1">
                                            <p className="text-xs font-bold text-indigo-500 uppercase tracking-widest">{tCommon('explanation') || 'Explanation'}</p>
                                            <p className="text-sm text-foreground/60 italic leading-relaxed">{q.explanation}</p>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </motion.div>
                    ))}
                </div>
            </div>

            <AssignTestModal
                isOpen={isAssignModalOpen}
                onClose={() => setIsAssignModalOpen(false)}
                testId={test.id}
                testTopic={test.topic}
            />
        </div>
    );
}
