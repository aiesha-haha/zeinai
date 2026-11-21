'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { motion, AnimatePresence } from 'framer-motion';
import { CheckCircle2, ChevronLeft, ChevronRight, Info, Loader2, Play, Timer, Trophy, XCircle } from 'lucide-react';
import { isGameType } from '@/lib/testTypes';

export type PlayQuestion = {
    id: string;
    question: string;
    options: string[];
    answer?: string;
    explanation?: string | null;
};

type Props = {
    title: string;
    subtitle?: string;
    testType: string;
    questions: PlayQuestion[];
    /** student: answers are sent to the server; preview: checked locally (teacher) */
    mode: 'student' | 'preview';
    onSubmit?: (answers: Record<string, string>) => Promise<void>;
    onExit: () => void;
    exitLabel: string;
};

/** Tug of war rope: the knot moves to the player's side as questions are answered */
function Rope({ progress }: { progress: number }) {
    const t = useTranslations('game');
    const pos = 50 - progress * 40; // 50% -> 10%
    return (
        <div className="glass-card rounded-3xl p-4" aria-hidden>
            <div className="flex justify-between text-xs font-bold mb-2">
                <span className="text-primary">{t('your_team')}</span>
                <span className="text-muted">{t('rival_team')}</span>
            </div>
            <div className="relative h-8">
                <div className="absolute top-1/2 left-0 right-0 h-2 -translate-y-1/2 rounded-full bg-amber-700/70" />
                <div className="absolute top-0 bottom-0 w-px bg-red-500 left-1/2" />
                <motion.div className="absolute top-1/2 w-6 h-6 -translate-y-1/2 -translate-x-1/2 rounded-full bg-red-500 border-4 border-white shadow"
                    animate={{ left: `${pos}%` }} transition={{ type: 'spring', stiffness: 120, damping: 14 }} />
            </div>
        </div>
    );
}

export function ResultsReview({ questions, answers, score, maxScore, onExit, exitLabel, testType }: {
    questions: PlayQuestion[];
    answers: Record<string, string>;
    score: number;
    maxScore: number;
    onExit: () => void;
    exitLabel: string;
    testType: string;
}) {
    const t = useTranslations('game');
    const tCommon = useTranslations('common');
    const percent = maxScore ? Math.round((score / maxScore) * 100) : 0;
    const praise = percent >= 85 ? t('praise_great') : percent >= 60 ? t('praise_good') : t('praise_try');
    const hasKey = questions.some((q) => q.answer);

    return (
        <div className="max-w-3xl mx-auto space-y-6">
            <section className="glass-card rounded-3xl p-6 md:p-8 text-center" aria-live="polite">
                <span className="w-16 h-16 mx-auto rounded-2xl bg-primary/10 text-primary flex items-center justify-center mb-4">
                    <Trophy size={32} />
                </span>
                <h2 className="text-2xl md:text-3xl font-bold font-outfit">{isGameType(testType) ? t('game_over') : t('test_done')}</h2>
                <p className="text-muted mt-1">{praise}</p>
                <div className="grid grid-cols-2 gap-4 mt-6 max-w-md mx-auto">
                    <div className="rounded-2xl bg-primary/[0.06] p-4">
                        <p className="text-sm text-muted">{t('points')}</p>
                        <p className="text-3xl font-bold text-primary tabular-nums">{score} / {maxScore}</p>
                    </div>
                    <div className="rounded-2xl bg-primary/[0.06] p-4">
                        <p className="text-sm text-muted">{t('percent')}</p>
                        <p className="text-3xl font-bold tabular-nums">{percent}%</p>
                    </div>
                </div>
                <button onClick={onExit} className="btn-primary mt-6">{exitLabel}</button>
            </section>

            {hasKey && (
                <section className="space-y-3">
                    <h3 className="text-lg font-bold font-outfit">{t('review_title')}</h3>
                    {questions.map((q, i) => {
                        const mine = answers[q.id];
                        const ok = mine === q.answer;
                        return (
                            <article key={q.id} className="glass-card rounded-2xl p-5">
                                <div className="flex gap-3">
                                    <span className={`shrink-0 mt-0.5 ${ok ? 'text-emerald-600' : 'text-red-600'}`} aria-hidden>
                                        {ok ? <CheckCircle2 size={22} /> : <XCircle size={22} />}
                                    </span>
                                    <div className="space-y-2 min-w-0">
                                        <p className="font-semibold">{i + 1}. {q.question}</p>
                                        <p className="text-sm">
                                            <span className="text-muted">{t('your_answer')}: </span>
                                            <span className={ok ? 'text-emerald-700 dark:text-emerald-400 font-semibold' : 'text-red-700 dark:text-red-400 font-semibold'}>
                                                {mine || t('no_answer')}
                                            </span>
                                            <span className="sr-only">{ok ? t('correct') : t('wrong')}</span>
                                        </p>
                                        {!ok && (
                                            <p className="text-sm">
                                                <span className="text-muted">{t('correct_answer')}: </span>
                                                <span className="text-emerald-700 dark:text-emerald-400 font-semibold">{q.answer}</span>
                                            </p>
                                        )}
                                        {q.explanation && (
                                            <p className="text-sm text-muted flex gap-2"><Info size={16} className="shrink-0 mt-0.5" />
                                                <span><span className="sr-only">{tCommon('explanation')}: </span>{q.explanation}</span></p>
                                        )}
                                    </div>
                                </div>
                            </article>
                        );
                    })}
                </section>
            )}
        </div>
    );
}

export default function GamePlayer({ title, subtitle, testType, questions, mode, onSubmit, onExit, exitLabel }: Props) {
    const t = useTranslations('game');
    const tTypes = useTranslations('test_types');
    const [started, setStarted] = React.useState(false);
    const [index, setIndex] = React.useState(0);
    const [answers, setAnswers] = React.useState<Record<string, string>>({});
    const [submitting, setSubmitting] = React.useState(false);
    const [error, setError] = React.useState<string | null>(null);
    const [previewDone, setPreviewDone] = React.useState(false);
    const [seconds, setSeconds] = React.useState(0);
    const isGame = isGameType(testType);
    const total = questions.length;
    const q = questions[index];
    const answeredCount = Object.keys(answers).length;

    React.useEffect(() => {
        if (!started || previewDone) return;
        const id = setInterval(() => setSeconds((s) => s + 1), 1000);
        return () => clearInterval(id);
    }, [started, previewDone]);

    const finish = async () => {
        if (submitting) return;
        if (mode === 'preview') {
            setPreviewDone(true);
            return;
        }
        setSubmitting(true);
        setError(null);
        try {
            await onSubmit?.(answers);
        } catch (e: any) {
            setError(e?.response?.data?.error?.message || t('submit_failed'));
            setSubmitting(false);
        }
    };

    const next = () => (index < total - 1 ? setIndex(index + 1) : finish());

    // Keyboard: 1-4 choose, Enter next
    React.useEffect(() => {
        if (!started || previewDone || !q) return;
        const onKey = (e: KeyboardEvent) => {
            const target = e.target as HTMLElement;
            if (target && ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName)) return;
            const n = parseInt(e.key, 10);
            if (n >= 1 && n <= q.options.length) setAnswers((a) => ({ ...a, [q.id]: q.options[n - 1] }));
        };
        window.addEventListener('keydown', onKey);
        return () => window.removeEventListener('keydown', onKey);
    }, [started, previewDone, q]);

    if (total === 0) {
        return <p className="glass-card rounded-3xl p-8 text-center text-muted">{t('no_questions')}</p>;
    }

    if (previewDone) {
        const score = questions.filter((x) => answers[x.id] && answers[x.id] === x.answer).length;
        return <ResultsReview questions={questions} answers={answers} score={score} maxScore={total}
            onExit={onExit} exitLabel={exitLabel} testType={testType} />;
    }

    if (!started) {
        return (
            <section className="max-w-2xl mx-auto glass-card rounded-3xl p-6 md:p-10 text-center space-y-5">
                {mode === 'preview' && <p className="inline-block text-xs font-bold px-3 py-1 rounded-full bg-amber-100 text-amber-900">{t('preview_badge')}</p>}
                <p className="text-sm font-semibold text-primary uppercase tracking-wider">{tTypes(`${testType}.title`)}</p>
                <h1 className="text-2xl md:text-3xl font-bold font-outfit">{title}</h1>
                {subtitle && <p className="text-muted">{subtitle}</p>}
                <ul className="text-left text-sm space-y-2 bg-primary/[0.05] rounded-2xl p-5">
                    <li>• {t('rule_count', { n: total })}</li>
                    <li>• {tTypes(`${testType}.rule`)}</li>
                    <li>• {t('rule_keys')}</li>
                    <li>• {t('rule_result')}</li>
                </ul>
                <div className="flex flex-col sm:flex-row gap-3 justify-center">
                    <button onClick={onExit} className="btn-ghost">{exitLabel}</button>
                    <button onClick={() => setStarted(true)} className="btn-primary text-lg px-8" autoFocus>
                        <Play size={20} /> {isGame ? t('start_game') : t('start_test')}
                    </button>
                </div>
            </section>
        );
    }

    const selected = answers[q.id];
    const isTF = testType === 'game_true_false';

    return (
        <div className="max-w-3xl mx-auto space-y-5">
            <div className="flex items-center justify-between gap-4">
                <h1 className="text-lg md:text-xl font-bold font-outfit truncate">{title}</h1>
                <span className="flex items-center gap-1.5 text-sm text-muted tabular-nums" aria-label={t('time')}>
                    <Timer size={16} /> {Math.floor(seconds / 60)}:{String(seconds % 60).padStart(2, '0')}
                </span>
            </div>

            <div>
                <div className="flex justify-between text-sm mb-1.5">
                    <span className="font-semibold text-primary">{t('question_of', { n: index + 1, total })}</span>
                    <span className="text-muted">{t('answered', { n: answeredCount, total })}</span>
                </div>
                <div className="h-2.5 rounded-full bg-primary/10 overflow-hidden" role="progressbar"
                    aria-valuenow={index + 1} aria-valuemin={1} aria-valuemax={total} aria-label={t('progress')}>
                    <div className="h-full bg-primary rounded-full transition-all" style={{ width: `${((index + 1) / total) * 100}%` }} />
                </div>
            </div>

            {testType === 'tug_of_war' && <Rope progress={answeredCount / total} />}

            <AnimatePresence mode="wait">
                <motion.section key={q.id} initial={{ opacity: 0, x: 16 }} animate={{ opacity: 1, x: 0 }} exit={{ opacity: 0, x: -16 }}
                    className="glass-card rounded-3xl p-6 md:p-8 space-y-6">
                    <h2 className="text-xl md:text-2xl font-bold leading-snug" id={`q-${q.id}`}>
                        {isTF && <span className="block text-sm font-semibold text-muted mb-2">{t('tf_prompt')}</span>}
                        {q.question}
                    </h2>
                    <div role="radiogroup" aria-labelledby={`q-${q.id}`} className={isTF ? 'grid grid-cols-2 gap-4' : 'grid gap-3'}>
                        {q.options.map((opt, i) => {
                            const on = selected === opt;
                            const tfColor = isTF ? (i === 0 ? 'emerald' : 'rose') : null;
                            return (
                                <button key={i} role="radio" aria-checked={on}
                                    onClick={() => setAnswers((a) => ({ ...a, [q.id]: opt }))}
                                    className={`flex items-center gap-4 text-left rounded-2xl border-2 transition-all ${isTF ? 'p-6 justify-center text-lg font-bold' : 'p-4 md:p-5'} ${on
                                        ? tfColor === 'emerald' ? 'border-emerald-600 bg-emerald-50 text-emerald-900 dark:bg-emerald-500/15 dark:text-emerald-200'
                                            : tfColor === 'rose' ? 'border-rose-600 bg-rose-50 text-rose-900 dark:bg-rose-500/15 dark:text-rose-200'
                                                : 'border-primary bg-primary/[0.07]'
                                        : 'border-line hover:border-primary/50'}`}>
                                    {!isTF && (
                                        <span className={`w-9 h-9 shrink-0 rounded-xl flex items-center justify-center font-bold ${on ? 'bg-primary text-white' : 'bg-primary/10 text-primary'}`}>
                                            {i + 1}
                                        </span>
                                    )}
                                    <span className="font-semibold">{opt}</span>
                                </button>
                            );
                        })}
                    </div>
                </motion.section>
            </AnimatePresence>

            {error && <p role="alert" className="text-sm text-red-700 bg-red-50 dark:bg-red-500/10 dark:text-red-300 p-3 rounded-xl">{error}</p>}

            <div className="flex items-center justify-between gap-3">
                <button onClick={() => setIndex(Math.max(0, index - 1))} disabled={index === 0} className="btn-ghost">
                    <ChevronLeft size={18} /> {t('prev')}
                </button>
                <button onClick={next} disabled={!selected || submitting} className="btn-primary px-8">
                    {submitting ? <Loader2 size={18} className="animate-spin" /> : null}
                    {index === total - 1 ? (isGame ? t('finish_game') : t('finish_test')) : t('next')}
                    {index < total - 1 && <ChevronRight size={18} />}
                </button>
            </div>
        </div>
    );
}
