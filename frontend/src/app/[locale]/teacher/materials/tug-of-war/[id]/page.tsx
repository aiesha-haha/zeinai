'use client';

import { useState, useEffect, useMemo } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { useTranslations, useLocale } from 'next-intl';
import { ChevronLeft, Trophy, Users, CheckCircle2, XCircle, RefreshCw, Play, SkipForward, Star, Loader2 } from 'lucide-react';
import api from '@/lib/api';
import { motion, AnimatePresence } from 'framer-motion';
import AssignTestModal from '@/components/materials/AssignTestModal';
import { Send } from 'lucide-react';

interface Question {
    id: string;
    group_index: number;
    question: string;
    options: string[];
    answer: string;
}

interface TugOfWarData {
    id: string;
    topic: string;
    subject: string;
    grade: number;
    language?: string;
    group_1_name: string;
    group_2_name: string;
    questions: Question[];
}

const GameScene = ({ isPulling, advantage, team1, team2 }: { isPulling: boolean, advantage: number, team1: string, team2: string }) => (
    <motion.div
        animate={{
            x: `${advantage * 10}%`,
            scale: isPulling ? [1, 1.01, 1] : 1,
            rotate: isPulling ? [0, -0.2, 0.2, 0] : 0
        }}
        transition={{
            x: { type: 'spring', stiffness: 50, damping: 20 },
            scale: { repeat: Infinity, duration: 0.6 },
            rotate: { repeat: Infinity, duration: 0.4 }
        }}
        className="relative flex flex-col items-center justify-center w-full h-full"
    >
        <div className="relative w-full h-full flex items-center justify-center max-w-4xl mx-auto px-4">
            <img
                src="/images/full-scene.png"
                alt="Tug of War"
                className="w-full h-full object-contain drop-shadow-2xl"
            />

            {/* VS Flag floating above the rope center */}
            <motion.div
                animate={isPulling ? { y: [-2, 2, -2], rotate: [-5, 5, -5] } : {}}
                transition={{ repeat: Infinity, duration: 2 }}
                className="absolute top-[45%] left-1/2 -translate-x-1/2 -translate-y-1/2 z-30"
            >
                <div className="w-10 h-8 bg-yellow-500 rounded-sm italic font-black text-[12px] flex items-center justify-center text-white shadow-xl border-2 border-white transform -rotate-6">
                    VS
                </div>
            </motion.div>

            {/* Team Labels positioned relatively */}
            <div className="absolute top-4 left-[15%]">
                <span className="px-3 py-0.5 rounded-full text-[9px] font-black uppercase tracking-widest bg-blue-500 text-white shadow-lg border border-white/20">
                    {team1}
                </span>
            </div>
            <div className="absolute top-4 right-[15%]">
                <span className="px-3 py-0.5 rounded-full text-[9px] font-black uppercase tracking-widest bg-red-500 text-white shadow-lg border border-white/20">
                    {team2}
                </span>
            </div>
        </div>
    </motion.div>
);

export default function TugOfWarGame() {
    const router = useRouter();
    const params = useParams();
    const t = useTranslations('tug_of_war');
    const tCommon = useTranslations('common');
    const tGames = useTranslations('games_hub');

    // "Give to students": the questions become a solo game test that can be assigned to a class
    const [assignTestId, setAssignTestId] = useState<string | null>(null);
    const [preparing, setPreparing] = useState(false);
    const giveToStudents = async () => {
        if (!gameData || preparing) return;
        setPreparing(true);
        try {
            const res = await api.post('/tests/manual', {
                subject: gameData.subject,
                grade: gameData.grade,
                topic: gameData.topic,
                language: ['kaz', 'rus', 'eng'].includes(gameData.language || '') ? gameData.language : 'kaz',
                test_type: 'tug_of_war',
                questions: gameData.questions.map((q) => ({ question: q.question, options: q.options, answer: q.answer })),
            });
            setAssignTestId(res.data.data.id);
        } catch (e: any) {
            alert(e.response?.data?.error?.message || tCommon('error_occurred'));
        } finally {
            setPreparing(false);
        }
    };

    const [loading, setLoading] = useState(true);
    const [gameData, setGameData] = useState<TugOfWarData | null>(null);

    // Game State
    const [started, setStarted] = useState(false);
    const [advantage, setAdvantage] = useState(0); // -3 to +3
    const [currentQuestionIndex, setCurrentQuestionIndex] = useState<{ 0: number, 1: number }>({ 0: 0, 1: 0 });
    const [scores, setScores] = useState({ 0: 0, 1: 0 });
    const [winner, setWinner] = useState<number | 'tie' | null>(null);

    const [showAnswer, setShowAnswer] = useState<{ 0: boolean, 1: boolean }>({ 0: false, 1: false });
    const [lastResult, setLastResult] = useState<{ 0: 'correct' | 'incorrect' | null, 1: 'correct' | 'incorrect' | null }>({ 0: null, 1: null });
    const [selectedOptionIdx, setSelectedOptionIdx] = useState<{ 0: number | null, 1: number | null }>({ 0: null, 1: null });

    const questionsByGroup = useMemo(() => {
        if (!gameData) return { 0: [], 1: [] };
        return {
            0: gameData.questions.filter((q: Question) => q.group_index === 0),
            1: gameData.questions.filter((q: Question) => q.group_index === 1)
        };
    }, [gameData]);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const res = await api.get(`/tug-of-war/${params.id}`);
                const data = res.data.data;
                // Flatten the material and questions structure
                const shuffledQuestions = data.questions.map((q: Question) => ({
                    ...q,
                    options: [...q.options].sort(() => Math.random() - 0.5)
                }));
                setGameData({
                    ...data.material,
                    questions: shuffledQuestions
                });
            } catch (err) {
                console.error('Failed to fetch game data:', err);
                router.push('/teacher/materials/tug-of-war');
            } finally {
                setLoading(false);
            }
        };
        fetchData();
    }, [params.id, router]);

    const handleAnswer = (teamIndex: 0 | 1, isCorrect: boolean, optionIdx: number) => {
        if (showAnswer[teamIndex]) return;

        setShowAnswer(prev => ({ ...prev, [teamIndex]: true }));
        setLastResult(prev => ({ ...prev, [teamIndex]: isCorrect ? 'correct' : 'incorrect' }));
        setSelectedOptionIdx(prev => ({ ...prev, [teamIndex]: optionIdx }));

        // Update scores and advantage if correct
        if (isCorrect) {
            setScores(prev => ({ ...prev, [teamIndex]: prev[teamIndex] + 1 }));
            const move = teamIndex === 0 ? -1 : 1; // Move towards Team 1 (left) or Team 2 (right)

            // Apply move with a small delay for feedback
            setTimeout(() => {
                setAdvantage(prev => {
                    const newAdv = Math.max(-3, Math.min(3, prev + move));
                    // Check instant win
                    if (newAdv === -3) setWinner(0);
                    else if (newAdv === 3) setWinner(1);
                    return newAdv;
                });
            }, 600);
        }

        // Automatic transition to next question
        const delay = isCorrect ? 1500 : 2500;
        setTimeout(() => {
            setShowAnswer(prev => ({ ...prev, [teamIndex]: false }));
            setLastResult(prev => ({ ...prev, [teamIndex]: null }));
            setSelectedOptionIdx(prev => ({ ...prev, [teamIndex]: null }));
            setCurrentQuestionIndex(prev => ({ ...prev, [teamIndex]: prev[teamIndex] + 1 }));
        }, delay);
    };

    useEffect(() => {
        if (!gameData || winner !== null) return;

        const total0 = questionsByGroup[0].length;
        const total1 = questionsByGroup[1].length;

        if (total0 > 0 && total1 > 0 && currentQuestionIndex[0] >= total0 && currentQuestionIndex[1] >= total1) {
            if (advantage < 0) setWinner(0);
            else if (advantage > 0) setWinner(1);
            else if (scores[0] > scores[1]) setWinner(0);
            else if (scores[0] < scores[1]) setWinner(1);
            else setWinner('tie');
        }
    }, [currentQuestionIndex, questionsByGroup, winner, advantage, scores]);

    if (loading) {
        return (
            <div className="h-[80vh] flex items-center justify-center">
                <Loader2 className="animate-spin text-primary" size={48} />
            </div>
        );
    }

    if (!gameData) return null;

    if (winner !== null) {
        return (
            <div className="h-[80vh] flex items-center justify-center p-4">
                <motion.div
                    initial={{ scale: 0.8, opacity: 0 }}
                    animate={{ scale: 1, opacity: 1 }}
                    className="glass-card p-12 rounded-[48px] text-center max-w-2xl w-full shadow-2xl relative overflow-hidden"
                >
                    <Trophy className="mx-auto text-yellow-500 mb-6" size={80} />
                    <h2 className="text-4xl font-bold mb-4 font-outfit">
                        {winner === 'tie' ? t('game.tie') : `${winner === 0 ? gameData.group_1_name : gameData.group_2_name} ${t('game.winner')}`}
                    </h2>

                    <div className="grid grid-cols-2 gap-8 mb-8 mt-8">
                        <div className="p-4 rounded-2xl bg-primary/5">
                            <p className="text-sm text-foreground/60 mb-1">{gameData.group_1_name}</p>
                            <p className="text-3xl font-bold">{scores[0]}</p>
                        </div>
                        <div className="p-4 rounded-2xl bg-primary/5">
                            <p className="text-sm text-foreground/60 mb-1">{gameData.group_2_name}</p>
                            <p className="text-3xl font-bold">{scores[1]}</p>
                        </div>
                    </div>

                    <div className="flex flex-col sm:flex-row gap-4 justify-center">
                        <button
                            onClick={() => window.location.reload()}
                            className="px-8 py-3 bg-primary text-white rounded-xl font-bold hover:bg-primary-hover transition-all flex items-center justify-center gap-2"
                        >
                            <RefreshCw size={20} />
                            {tCommon('retry')}
                        </button>
                        <button
                            onClick={() => router.push('/teacher/materials/tug-of-war')}
                            className="px-8 py-3 glass-card rounded-xl font-bold hover:bg-slate-100 dark:hover:bg-white/5 transition-all"
                        >
                            {tCommon('back')}
                        </button>
                    </div>
                </motion.div>
            </div>
        );
    }

    return (
        <div className="max-w-7xl mx-auto h-[calc(100vh-80px)] min-h-[600px] flex flex-col p-4 overflow-hidden">
            {/* Header */}
            <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-3">
                    <button
                        onClick={() => router.back()}
                        className="p-1.5 glass-card rounded-lg hover:bg-slate-200 dark:hover:bg-white/10 transition-all text-foreground/60"
                    >
                        <ChevronLeft size={18} />
                    </button>
                    <div>
                        <h1 className="text-xl font-bold font-outfit leading-tight">{gameData.topic}</h1>
                        <p className="text-xs text-foreground/60">{gameData.subject} • {gameData.grade} {tCommon('grade_suffix')}</p>
                    </div>
                </div>

                <button onClick={giveToStudents} disabled={preparing} className="btn-ghost text-sm !py-2 ml-auto mr-3">
                    {preparing ? <Loader2 size={16} className="animate-spin" /> : <Send size={16} />} {tGames('give_to_students')}
                </button>

                <div className="flex items-center gap-4 glass-card px-4 py-2 rounded-xl bg-primary/5">
                    <div className="text-center">
                        <p className="text-[10px] uppercase font-bold text-foreground/40 leading-none">{gameData.group_1_name}</p>
                        <p className="text-xl font-bold text-primary">{scores[0]}</p>
                    </div>
                    <div className="h-6 w-px bg-foreground/10" />
                    <div className="text-center">
                        <p className="text-[10px] uppercase font-bold text-foreground/40 leading-none">{gameData.group_2_name}</p>
                        <p className="text-xl font-bold text-primary">{scores[1]}</p>
                    </div>
                </div>
            </div>

            {/* Rope Area */}
            <div className="relative glass-card rounded-[32px] overflow-hidden bg-slate-100/50 dark:bg-slate-900/50 border-2 border-primary/10 mb-4 shrink-0 h-[30vh] min-h-[160px]">
                <div className="relative h-full flex items-center justify-center">
                    {/* The Main Rope */}
                    <div className="absolute w-full h-3 bg-[#8B4513] rounded-full shadow-inner overflow-hidden">
                        <div className="w-full h-full opacity-20 bg-[repeating-linear-gradient(45deg,transparent,transparent_10px,#000_10px,#000_20px)]" />
                    </div>

                    {/* Moving Game Scene (Single Image) */}
                    <GameScene
                        isPulling={started && !winner}
                        advantage={advantage}
                        team1={gameData.group_1_name}
                        team2={gameData.group_2_name}
                    />

                    {/* Ground/Steps Indicators */}
                    <div className="absolute bottom-6 w-full flex justify-around px-12">
                        {[-3, -2, -1, 0, 1, 2, 3].map(step => (
                            <div
                                key={step}
                                className={`flex flex-col items-center gap-2 transition-all duration-500 ${advantage === step ? 'scale-110' : 'opacity-40'}`}
                            >
                                <div className={`w-3 h-3 rounded-full ${advantage === step ? 'bg-primary shadow-[0_0_15px_rgba(var(--primary-rgb),0.5)]' : 'bg-foreground/20'}`} />
                                <span className="text-[10px] font-bold text-foreground/40">{step === 0 ? 'CENTER' : Math.abs(step)}</span>
                            </div>
                        ))}
                    </div>
                </div>
            </div>

            {/* Simultaneous Questions */}
            {!started ? (
                <div className="flex flex-col items-center justify-center flex-1 glass-card rounded-[32px] bg-slate-50/50 dark:bg-white/5 border border-primary/10">
                    <h2 className="text-2xl font-bold mb-6 font-outfit">{t('game.start')}?</h2>
                    <button
                        onClick={() => setStarted(true)}
                        className="px-12 py-4 bg-primary text-white rounded-2xl font-bold text-xl hover:bg-primary-hover transition-all flex items-center gap-3 shadow-xl active:scale-95"
                    >
                        <Play fill="currentColor" size={24} />
                        {t('game.start')}
                    </button>
                </div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 items-stretch flex-1 min-h-0 overflow-hidden">
                    {[0, 1].map((idx) => {
                        const teamIndex = idx as 0 | 1;
                        const isDone = currentQuestionIndex[teamIndex] >= questionsByGroup[teamIndex].length;
                        const currentQ = questionsByGroup[teamIndex][currentQuestionIndex[teamIndex]];

                        return (
                            <div key={teamIndex} className="flex flex-col space-y-3 min-h-0 overflow-hidden">
                                <div className={`p-3 px-5 rounded-xl glass-card flex justify-between items-center ${isDone ? 'opacity-50' : 'bg-primary/5'}`}>
                                    <h3 className="font-bold text-base flex items-center gap-3 italic">
                                        <div className={`w-6 h-6 rounded-md flex items-center justify-center text-[10px] font-black ${teamIndex === 0 ? 'bg-blue-500 text-white' : 'bg-red-500 text-white'}`}>
                                            {teamIndex === 0 ? 'A' : 'B'}
                                        </div>
                                        {teamIndex === 0 ? gameData.group_1_name : gameData.group_2_name}
                                    </h3>
                                    {!isDone && (
                                        <span className="text-[10px] font-black uppercase text-primary bg-primary/10 px-2 py-0.5 rounded-full">
                                            Q {currentQuestionIndex[teamIndex] + 1}
                                        </span>
                                    )}
                                </div>

                                <AnimatePresence mode="wait">
                                    {isDone ? (
                                        <motion.div
                                            initial={{ opacity: 0 }}
                                            animate={{ opacity: 1 }}
                                            className="glass-card p-6 rounded-[24px] text-center bg-slate-100/50 dark:bg-white/5 flex flex-col items-center justify-center h-full min-h-[200px]"
                                        >
                                            <CheckCircle2 size={32} className="text-green-500 mb-2" />
                                            <p className="font-bold text-foreground/40 text-xs uppercase tracking-widest">{tCommon('completed')}</p>
                                        </motion.div>
                                    ) : (
                                        <motion.div
                                            key={`${teamIndex}-${currentQuestionIndex[teamIndex]}`}
                                            initial={{ opacity: 0, y: 10 }}
                                            animate={{ opacity: 1, y: 0 }}
                                            exit={{ opacity: 0, scale: 0.98 }}
                                            className="glass-card p-5 rounded-[24px] shadow-lg bg-white dark:bg-white/5 h-full relative border-t-4 border-t-primary flex flex-col min-h-0"
                                        >
                                            <div className="mb-6 h-1 bg-foreground/5 rounded-full overflow-hidden">
                                                <motion.div
                                                    className="h-full bg-primary"
                                                    initial={{ width: 0 }}
                                                    animate={{ width: `${(currentQuestionIndex[teamIndex] / 10) * 100}%` }}
                                                />
                                            </div>

                                            <h2 className="text-lg font-bold mb-4 leading-tight font-outfit line-clamp-2">
                                                {currentQ?.question}
                                            </h2>

                                            <div className="grid grid-cols-1 gap-2 overflow-y-auto pr-1 custom-scrollbar flex-1 min-h-0">
                                                {currentQ?.options.map((option, oIdx) => {
                                                    const isCorrect = option === currentQ.answer;
                                                    const isSelected = showAnswer[teamIndex] && selectedOptionIdx[teamIndex] === oIdx;
                                                    const isWrong = isSelected && !isCorrect;

                                                    return (
                                                        <button
                                                            key={oIdx}
                                                            onClick={() => handleAnswer(teamIndex, isCorrect, oIdx)}
                                                            disabled={showAnswer[teamIndex]}
                                                            className={`
                                                                w-full text-left px-3 py-2 rounded-lg font-semibold border-2 transition-all duration-300 relative overflow-hidden
                                                                ${!showAnswer[teamIndex] ? 'border-foreground/5 hover:border-primary hover:bg-primary/5 active:scale-[0.98]' : ''}
                                                                ${showAnswer[teamIndex] && isCorrect ? 'bg-green-500 border-green-500 text-white shadow-lg' : ''}
                                                                ${isWrong ? 'bg-red-500 border-red-500 text-white shadow-lg' : ''}
                                                                ${showAnswer[teamIndex] && !isCorrect && !isSelected ? 'opacity-40 grayscale border-foreground/5 text-foreground/40' : ''}
                                                            `}
                                                        >
                                                            <div className="flex items-center justify-between relative z-10">
                                                                <span className="text-[11px] leading-tight">{option}</span>
                                                                {showAnswer[teamIndex] && isCorrect && <CheckCircle2 size={14} />}
                                                                {isWrong && <div className="text-white text-[9px] font-black uppercase tracking-tight">{t('game.incorrect')}</div>}
                                                            </div>
                                                        </button>
                                                    );
                                                })}
                                            </div>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                            </div>
                        );
                    })}
                </div>
            )}
            {assignTestId && (
                <AssignTestModal isOpen onClose={() => setAssignTestId(null)} testId={assignTestId} testTopic={gameData.topic} />
            )}
        </div>
    );
}
