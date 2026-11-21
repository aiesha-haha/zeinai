'use client';

import { useState, useEffect, use } from 'react';
import { useRouter } from '@/i18n/routing';
import { useTranslations, useLocale } from 'next-intl';
import { ChevronLeft, Rocket, Users, Award, CheckCircle2, Trophy, ArrowRight, RotateCcw, Play, Sparkles, Pencil } from 'lucide-react';
import api from '@/lib/api';
import { motion, AnimatePresence } from 'framer-motion';
import { toast } from 'react-hot-toast';

export default function GroupTaskSessionPage({ params }: { params: Promise<{ id: string }> }) {
    const resolvedParams = use(params);
    const id = resolvedParams.id;
    const router = useRouter();
    const t = useTranslations('group_tasks');
    const tCommon = useTranslations('common');
    const locale = useLocale();

    const [task, setTask] = useState<any>(null);
    const [loading, setLoading] = useState(true);
    const [sessionLoading, setSessionLoading] = useState(false);

    // Session Setup State
    const [gameState, setGameState] = useState<'setup' | 'playing' | 'finished'>('setup');
    const [groupCount, setGroupCount] = useState(2);
    const [session, setSession] = useState<any>(null);

    // Gameplay State
    const [currentCardIndex, setCurrentCardIndex] = useState(0);
    const [isRevealed, setIsRevealed] = useState(false);
    const [openedCards, setOpenedCards] = useState<number[]>([]);
    const [scoredCards, setScoredCards] = useState<{ [index: number]: string }>({}); // index -> groupId
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [selectedCardIndex, setSelectedCardIndex] = useState<number | null>(null);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const res = await api.get(`/group-tasks/${id}`);
                setTask(res.data.data);
            } catch (error) {
                console.error('Failed to load data:', error);
                toast.error(tCommon('error'));
            } finally {
                setLoading(false);
            }
        };
        fetchData();
    }, [id]);

    const handleStartSession = async () => {
        setSessionLoading(true);
        try {
            const res = await api.post('/group-tasks/sessions', {
                task_id: id,
                class_id: null,
                group_count: groupCount
            });
            setSession(res.data.data);
            setGameState('playing');
        } catch (error: any) {
            toast.error(error.response?.data?.error?.message || tCommon('error'));
        } finally {
            setSessionLoading(false);
        }
    };

    const handleAwardPoint = async (groupId: string) => {
        if (selectedCardIndex === null || scoredCards[selectedCardIndex]) return;

        try {
            await api.post(`/group-tasks/sessions/${session.session.id}/groups/${groupId}/award`);
            // Update session points locally for instant feedback
            setSession({
                ...session,
                groups: session.groups.map((g: any) =>
                    g.id === groupId ? { ...g, score: (g.score || 0) + 1 } : g
                )
            });

            // Mark card as scored
            setScoredCards(prev => ({ ...prev, [selectedCardIndex]: groupId }));

            toast.success('+1 ' + t('session.score'));
            setIsModalOpen(false); // Close modal once point is awarded
        } catch (error) {
            toast.error(tCommon('error'));
        }
    };

    const handleOpenCard = (index: number) => {
        if (!openedCards.includes(index)) {
            setOpenedCards(prev => [...prev, index]);
        }
        setSelectedCardIndex(index);
        setIsRevealed(false);
        setIsModalOpen(true);
    };

    const nextCard = () => {
        if (currentCardIndex < task.questions.length - 1) {
            setCurrentCardIndex(prev => prev + 1);
            setIsRevealed(false);
        } else {
            setGameState('finished');
        }
    };

    if (loading) return (
        <div className="flex h-[60vh] items-center justify-center">
            <motion.div animate={{ rotate: 360 }} transition={{ repeat: Infinity, duration: 1, ease: 'linear' }}>
                <Rocket className="text-primary w-12 h-12" />
            </motion.div>
        </div>
    );

    if (gameState === 'setup') {
        return (
            <div className="max-w-2xl mx-auto py-10">
                <button onClick={() => router.back()} className="flex items-center gap-2 mb-8 text-foreground/60 hover:text-foreground transition-colors">
                    <ChevronLeft size={20} />
                    {tCommon('back')}
                </button>

                <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} className="glass-card p-10 rounded-[40px] shadow-2xl overflow-hidden relative">
                    <div className="absolute top-0 right-0 p-8 opacity-10">
                        <Rocket size={120} />
                    </div>

                    <h1 className="text-4xl font-black mb-2 font-outfit uppercase tracking-tight">{task.task.topic}</h1>
                    <p className="text-foreground/60 mb-10 text-lg uppercase font-bold tracking-widest">{task.task.subject} • {task.task.grade} {tCommon('grade_suffix')}</p>

                    <div className="space-y-12">
                        {/* Group Count Selection */}
                        <div className="text-center">
                            <label className="block text-2xl font-black uppercase mb-8 tracking-widest text-primary/60">
                                {t('session.group_count')}
                            </label>

                            <div className="flex justify-center gap-4 flex-wrap">
                                {[2, 3, 4, 5, 6].map((num) => (
                                    <button
                                        key={num}
                                        onClick={() => setGroupCount(num)}
                                        className={`w-16 h-16 sm:w-20 sm:h-20 rounded-3xl text-2xl font-black transition-all duration-300 flex items-center justify-center border-4 ${groupCount === num
                                            ? 'bg-primary text-white border-primary shadow-xl shadow-primary/30 scale-110'
                                            : 'bg-slate-100 dark:bg-white/5 border-transparent hover:border-primary/20 text-foreground/40'
                                            }`}
                                    >
                                        {num}
                                    </button>
                                ))}
                            </div>

                            <p className="mt-6 text-sm font-bold text-foreground/40 uppercase tracking-widest">
                                {t('session.setup_hint')}
                            </p>
                        </div>

                        <button
                            onClick={handleStartSession}
                            disabled={sessionLoading}
                            className="w-full py-6 bg-primary text-white rounded-[24px] font-black text-xl hover:bg-primary-hover shadow-2xl shadow-primary/30 transition-all active:scale-95 disabled:opacity-50 flex items-center justify-center gap-3"
                        >
                            {sessionLoading ? <RotateCcw className="animate-spin" /> : <Play fill="currentColor" />}
                            {t('session.start')}
                        </button>

                    </div >
                </motion.div >
            </div >
        );
    }

    if (gameState === 'playing') {
        const currentQuestion = selectedCardIndex !== null ? task.questions[selectedCardIndex] : null;

        return (
            <div className="h-full flex flex-col gap-6 pb-6">
                {/* Top Bar - Compact */}
                <div className="flex items-center justify-between glass-card p-4 px-8 rounded-[24px]">
                    <div className="flex items-center gap-4">
                        <button onClick={() => setGameState('setup')} className="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-full transition-all">
                            <ChevronLeft size={20} />
                        </button>
                        <div>
                            <span className="text-[10px] font-black uppercase tracking-widest text-primary leading-none block mb-1">{task.task.topic}</span>
                            <div className="flex items-center gap-2">
                                <span className="text-xl font-bold font-outfit uppercase">{t('session.board_title')}</span>
                                <span className="text-xs font-bold text-foreground/40">{openedCards.length} / {task.questions.length}</span>
                            </div>
                        </div>
                    </div>

                    <div className="flex gap-3">
                        {session.groups.map((g: any, i: number) => (
                            <div key={g.id} className="flex items-center gap-3 p-2 px-5 glass-card rounded-xl border-b-2 border-primary/40">
                                <span className="text-[10px] font-black opacity-40 uppercase">{t('session.group')} {i + 1}</span>
                                <span className="text-lg font-black text-primary leading-none">{g.score || 0}</span>
                            </div>
                        ))}
                    </div>
                </div>

                {/* Grid Area - Scrollable if needed but optimized for 1 screen */}
                <div className="flex-1 grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4 content-start overflow-auto pr-2 pb-4 pt-2">
                    {task.questions.map((q: any, idx: number) => (
                        <motion.button
                            key={q.id}
                            initial={{ opacity: 0, scale: 0.9 }}
                            animate={{ opacity: 1, scale: 1 }}
                            transition={{ delay: idx * 0.05 }}
                            onClick={() => handleOpenCard(idx)}
                            className={`
                                relative aspect-[4/3] rounded-[24px] border-b-[4px] transition-all flex flex-col items-center justify-center gap-2 group
                                ${openedCards.includes(idx)
                                    ? 'bg-slate-100 dark:bg-white/5 border-slate-200 dark:border-white/10 opacity-60'
                                    : 'glass-card border-primary/20 hover:scale-[1.05] hover:border-primary/50 shadow-lg hover:shadow-primary/10 active:scale-95'}
                            `}
                        >
                            <span className={`text-4xl font-black font-outfit ${openedCards.includes(idx) ? 'text-foreground/20' : 'text-primary'}`}>
                                {idx + 1}
                            </span>
                            {scoredCards[idx] && (
                                <div className="mt-1 flex flex-col items-center gap-1">
                                    <div className="px-2 py-0.5 bg-primary/10 rounded-lg">
                                        <span className="text-[8px] font-black text-primary uppercase leading-tight">
                                            {t('session.group')} {session.groups.findIndex((g: any) => g.id === scoredCards[idx]) + 1}
                                        </span>
                                    </div>
                                    <CheckCircle2 className="text-primary" size={14} />
                                </div>
                            )}
                            {openedCards.includes(idx) && !scoredCards[idx] && (
                                <CheckCircle2 className="text-foreground/10" size={16} />
                            )}
                        </motion.button>
                    ))}
                </div>

                {/* Footer Controls */}
                <div className="flex justify-between items-center pt-6 border-t-2 border-primary/10 mt-8">
                    <div className="text-sm font-black text-foreground/60 uppercase tracking-widest leading-none">
                        {t('session.total_cards')}: {task.questions.length}
                    </div>
                    <button
                        onClick={() => setGameState('finished')}
                        className="px-8 py-4 bg-primary text-white rounded-[24px] font-black text-xl hover:scale-105 active:scale-95 transition-all flex items-center gap-3 shadow-2xl shadow-primary/30"
                    >
                        {t('session.finish')}
                        <Trophy size={24} />
                    </button>
                </div>

                {/* Card Modal */}
                <AnimatePresence>
                    {isModalOpen && currentQuestion && (
                        <div className="fixed inset-0 z-50 flex items-center justify-center p-6 sm:p-12">
                            <motion.div
                                initial={{ opacity: 0 }}
                                animate={{ opacity: 1 }}
                                exit={{ opacity: 0 }}
                                onClick={() => setIsModalOpen(false)}
                                className="absolute inset-0 bg-background/80 backdrop-blur-md"
                            />

                            <motion.div
                                initial={{ scale: 0.9, y: 20, opacity: 0 }}
                                animate={{ scale: 1, y: 0, opacity: 1 }}
                                exit={{ scale: 0.9, y: 20, opacity: 0 }}
                                className="relative w-full max-w-4xl glass-card rounded-[40px] shadow-2xl border-2 border-primary/20 overflow-hidden"
                            >
                                {/* Modal Header */}
                                <div className="p-6 border-b border-foreground/5 flex justify-between items-center">
                                    <span className="text-sm font-black uppercase tracking-[0.2em] text-primary">{t('session.card_label')} {selectedCardIndex! + 1}</span>
                                    <button
                                        onClick={() => setIsModalOpen(false)}
                                        className="p-2 hover:bg-foreground/5 rounded-full transition-all"
                                    >
                                        <ChevronLeft size={24} className="rotate-90" />
                                    </button>
                                </div>

                                {/* Modal Content - The Card Activity */}
                                <div className="p-12 text-center flex flex-col items-center">
                                    <div className={`relative w-full min-h-[300px] flex items-center justify-center transition-all duration-700 preserve-3d ${isRevealed ? 'rotate-y-180' : ''}`}>
                                        {/* Front: Question */}
                                        <div className="w-full backface-hidden flex flex-col items-center">
                                            <Sparkles className="text-primary/10 mb-8 w-16 h-16" />
                                            <h3 className="text-4xl sm:text-5xl font-black font-outfit leading-tight mb-10 max-w-3xl">
                                                {currentQuestion.question}
                                            </h3>
                                            {!isRevealed && (
                                                <button
                                                    onClick={() => setIsRevealed(true)}
                                                    className="px-8 py-3 bg-primary text-white rounded-full text-xs font-black uppercase tracking-widest hover:scale-110 active:scale-95 transition-all shadow-lg shadow-primary/20"
                                                >
                                                    {t('session.show_answer')}
                                                </button>
                                            )}
                                        </div>

                                        {/* Back: Answer */}
                                        <div className="absolute inset-0 backface-hidden flex flex-col items-center justify-center rotate-y-180">
                                            <Award className="text-primary mb-6 w-16 h-16" />
                                            <div className="space-y-4 mb-10">
                                                <span className="text-[10px] font-black uppercase tracking-[0.4em] text-primary opacity-60 leading-none">{t('session.correct_answer')}</span>
                                                <p className="text-5xl font-black font-outfit text-primary-dark max-w-2xl leading-none italic">
                                                    {currentQuestion.answer}
                                                </p>
                                            </div>

                                            {/* Scoring Buttons in Modal */}
                                            <div className="flex gap-3 w-full max-w-xl">
                                                {session.groups.map((g: any, i: number) => (
                                                    <button
                                                        key={g.id}
                                                        onClick={(e) => {
                                                            e.stopPropagation();
                                                            handleAwardPoint(g.id);
                                                        }}
                                                        disabled={scoredCards[selectedCardIndex!] !== undefined}
                                                        className="flex-1 py-4 bg-primary text-white rounded-2xl font-black text-[10px] uppercase hover:scale-105 active:scale-95 transition-all shadow-xl shadow-primary/20 flex flex-col items-center gap-1 disabled:opacity-50 disabled:grayscale disabled:hover:scale-100"
                                                    >
                                                        <span>{t('session.group')} {i + 1}</span>
                                                        <span className="text-base font-black">+1</span>
                                                    </button>
                                                ))}
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </motion.div>
                        </div>
                    )}
                </AnimatePresence>
            </div>
        );
    }

    if (gameState === 'finished') {
        const sortedGroups = [...session.groups].sort((a, b) => b.score - a.score);
        const winner = sortedGroups[0];

        return (
            <div className="max-w-2xl mx-auto py-20 text-center">
                <motion.div initial={{ scale: 0 }} animate={{ scale: 1 }} transition={{ type: 'spring' }} className="mb-10 inline-block relative">
                    <div className="absolute inset-0 bg-yellow-400 blur-3xl opacity-20 animate-pulse" />
                    <Trophy size={160} className="text-yellow-400 relative" />
                </motion.div>

                <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.2 }}>
                    <h1 className="text-6xl font-black font-outfit mb-4">{t('session.winner')}</h1>
                    <div className="text-3xl font-black text-primary mb-12 uppercase">
                        {t('session.group')} {session.groups.findIndex((g: any) => g.id === winner.id) + 1}
                        <div className="text-lg text-foreground/40 mt-2 font-normal">{winner.score} {t('session.score')}</div>
                    </div>

                    <div className="space-y-4 mb-16">
                        {sortedGroups.map((group: any, index: number) => (
                            <div key={group.id} className="glass-card p-6 px-10 rounded-[24px] flex items-center justify-between">
                                <div className="flex items-center gap-6">
                                    <span className="text-4xl font-black opacity-20">{index + 1}</span>
                                    <span className="text-xl font-bold uppercase tracking-tight">{t('session.group')} {session.groups.findIndex((g: any) => g.id === group.id) + 1}</span>
                                </div>
                                <span className="text-2xl font-black">{group.score}</span>
                            </div>
                        ))}
                    </div>

                    <button
                        onClick={() => router.push('/teacher/materials/group-tasks')}
                        className="px-12 py-5 bg-primary text-white rounded-3xl font-black text-xl hover:bg-primary-hover shadow-2xl shadow-primary/20 transition-all"
                    >
                        {t('session.finish')}
                    </button>
                </motion.div>
            </div>
        );
    }

    return null;
}
