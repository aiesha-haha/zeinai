'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import { useRouter } from '@/i18n/routing';
import { useTranslations } from 'next-intl';
import { ChevronLeft, Plus, X, Loader2, BookOpen, Check, Trash2 } from 'lucide-react';
import api from '@/lib/api';
import { motion, AnimatePresence } from 'framer-motion';
import toast from 'react-hot-toast';

interface Question {
    id: string;
    question: string;
    answer: string;
    order: number;
}

interface Task {
    id: string;
    topic: string;
    subject: string;
    grade: number;
}

export default function EditGroupTaskPage() {
    const params = useParams();
    const taskId = params.id as string;
    const router = useRouter();
    const t = useTranslations('group_tasks');
    const tCommon = useTranslations('common');

    const [task, setTask] = useState<Task | null>(null);
    const [questions, setQuestions] = useState<Question[]>([]);
    const [loading, setLoading] = useState(true);

    // Modal state
    const [modalQ, setModalQ] = useState<Question | null>(null);
    const [modalForm, setModalForm] = useState({ question: '', answer: '' });
    const [modalSaving, setModalSaving] = useState(false);
    const [modalDeleting, setModalDeleting] = useState(false);

    // Add new modal
    const [addOpen, setAddOpen] = useState(false);
    const [newForm, setNewForm] = useState({ question: '', answer: '' });
    const [addSaving, setAddSaving] = useState(false);

    useEffect(() => {
        api.get(`/group-tasks/${taskId}`)
            .then(res => {
                const data = res.data.data;
                setTask(data.task);
                setQuestions(data.questions || []);
            })
            .catch(() => toast.error(tCommon('error')))
            .finally(() => setLoading(false));
    }, [taskId]);

    const openEdit = (q: Question) => {
        setModalQ(q);
        setModalForm({ question: q.question, answer: q.answer });
    };

    const closeEdit = () => {
        setModalQ(null);
    };

    const saveEdit = async () => {
        if (!modalQ) return;
        setModalSaving(true);
        try {
            await api.patch(`/group-tasks/${taskId}/questions/${modalQ.id}`, modalForm);
            setQuestions(prev => prev.map(item =>
                item.id === modalQ.id ? { ...item, ...modalForm } : item
            ));
            toast.success(t('edit.saved'));
            closeEdit();
        } catch {
            toast.error(tCommon('error'));
        } finally {
            setModalSaving(false);
        }
    };

    const deleteQuestion = async () => {
        if (!modalQ) return;
        if (!confirm(tCommon('confirm_delete'))) return;
        setModalDeleting(true);
        try {
            await api.delete(`/group-tasks/${taskId}/questions/${modalQ.id}`);
            setQuestions(prev => prev.filter(q => q.id !== modalQ.id));
            toast.success(t('edit.deleted'));
            closeEdit();
        } catch {
            toast.error(tCommon('error'));
        } finally {
            setModalDeleting(false);
        }
    };

    const addQuestion = async () => {
        if (!newForm.question.trim() || !newForm.answer.trim()) return;
        setAddSaving(true);
        try {
            const res = await api.post(`/group-tasks/${taskId}/questions`, newForm);
            setQuestions(prev => [...prev, res.data.data]);
            setNewForm({ question: '', answer: '' });
            setAddOpen(false);
            toast.success(t('edit.added'));
        } catch {
            toast.error(tCommon('error'));
        } finally {
            setAddSaving(false);
        }
    };

    if (loading) return (
        <div className="flex items-center justify-center min-h-[60vh]">
            <Loader2 className="animate-spin text-primary" size={40} />
        </div>
    );

    return (
        <div className="space-y-6 pb-24">
            {/* Header */}
            <div className="flex items-center gap-4">
                <button
                    onClick={() => router.back()}
                    className="p-2 glass-card rounded-xl hover:bg-slate-200 dark:hover:bg-white/10 transition-all text-foreground/60"
                >
                    <ChevronLeft size={20} />
                </button>
                <div>
                    <h1 className="text-2xl font-bold font-outfit">{t('edit.title')}</h1>
                    {task && <p className="text-foreground/50 text-sm">{task.topic} · {task.subject} · {task.grade} {tCommon('grade_suffix')}</p>}
                </div>
                <div className="ml-auto text-sm font-semibold text-foreground/40">{questions.length} {t('edit.cards')}</div>
            </div>

            {/* Grid of question cards */}
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                <AnimatePresence>
                    {questions.map((q, idx) => (
                        <motion.div
                            key={q.id}
                            layout
                            initial={{ opacity: 0, scale: 0.9 }}
                            animate={{ opacity: 1, scale: 1 }}
                            exit={{ opacity: 0, scale: 0.85 }}
                            className="glass-card rounded-2xl overflow-hidden group hover:shadow-xl hover:shadow-primary/10 transition-all duration-200"
                        >
                            {/* Card header */}
                            <div className="flex items-center justify-between px-4 py-2 border-b border-foreground/5">
                                <span className="text-xs font-black text-foreground/30 uppercase tracking-widest">#{idx + 1}</span>
                                <button
                                    onClick={() => {
                                        if (!confirm(tCommon('confirm_delete'))) return;
                                        api.delete(`/group-tasks/${taskId}/questions/${q.id}`)
                                            .then(() => {
                                                setQuestions(prev => prev.filter(item => item.id !== q.id));
                                                toast.success(t('edit.deleted'));
                                            })
                                            .catch(() => toast.error(tCommon('error')));
                                    }}
                                    className="p-1.5 rounded-lg hover:bg-rose-50 dark:hover:bg-rose-900/30 text-rose-400 hover:text-rose-500 transition-all"
                                >
                                    <Trash2 size={14} />
                                </button>
                            </div>

                            {/* Card body — click to edit */}
                            <button
                                onClick={() => openEdit(q)}
                                className="w-full text-left px-4 py-4 space-y-3 hover:bg-primary/5 transition-all"
                            >
                                <p className="font-semibold text-sm text-foreground leading-snug line-clamp-3">{q.question}</p>
                                <div className="pt-1 border-t border-foreground/8">
                                    <span className="text-xs text-foreground/40 font-bold uppercase tracking-wider mr-1">{t('session.correct_answer')}:</span>
                                    <span className="text-sm font-semibold text-primary line-clamp-2">{q.answer}</span>
                                </div>
                            </button>
                        </motion.div>
                    ))}

                    {/* Add new card */}
                    <motion.button
                        layout
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        onClick={() => setAddOpen(true)}
                        className="
                            min-h-[140px] rounded-2xl border-2 border-dashed border-primary/30
                            flex flex-col items-center justify-center gap-2
                            text-primary hover:bg-primary/5 transition-all
                            hover:scale-[1.02] active:scale-95
                        "
                    >
                        <Plus size={28} />
                        <span className="text-xs font-bold">{t('edit.add_btn')}</span>
                    </motion.button>
                </AnimatePresence>
            </div>

            {/* Back to session */}
            <button
                onClick={() => router.back()}
                className="w-full py-4 rounded-2xl glass-card font-bold flex items-center justify-center gap-3 hover:bg-slate-100 dark:hover:bg-white/5 transition-all"
            >
                <BookOpen size={20} />
                {t('edit.back_to_task')}
            </button>

            {/* ─── Edit Modal ───────────────────────────────────── */}
            <AnimatePresence>
                {modalQ && (
                    <>
                        {/* Backdrop */}
                        <motion.div
                            key="backdrop"
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            exit={{ opacity: 0 }}
                            onClick={closeEdit}
                            className="fixed inset-0 bg-black/50 backdrop-blur-sm z-40"
                        />

                        {/* Modal */}
                        <motion.div
                            key="modal"
                            initial={{ opacity: 0, scale: 0.92, y: 20 }}
                            animate={{ opacity: 1, scale: 1, y: 0 }}
                            exit={{ opacity: 0, scale: 0.92, y: 20 }}
                            transition={{ type: 'spring', stiffness: 300, damping: 28 }}
                            className="
                                fixed inset-0 z-50 flex items-center justify-center p-4
                                pointer-events-none
                            "
                        >
                            <div
                                className="
                                    glass-card rounded-3xl p-6 w-full max-w-lg
                                    pointer-events-auto space-y-5 shadow-2xl
                                "
                                onClick={e => e.stopPropagation()}
                            >
                                {/* Modal header */}
                                <div className="flex items-center justify-between">
                                    <h2 className="text-lg font-black font-outfit">{t('edit.title')}</h2>
                                    <button onClick={closeEdit} className="p-2 rounded-xl hover:bg-slate-100 dark:hover:bg-white/10 transition-all text-foreground/50">
                                        <X size={20} />
                                    </button>
                                </div>

                                {/* Question */}
                                <div className="space-y-2">
                                    <label className="text-xs font-semibold text-foreground/50 uppercase tracking-wider">{t('edit.question_label')}</label>
                                    <textarea
                                        value={modalForm.question}
                                        onChange={e => setModalForm(f => ({ ...f, question: e.target.value }))}
                                        rows={3}
                                        className="w-full px-4 py-3 glass-card rounded-xl outline-none focus:ring-2 focus:ring-primary/30 resize-none font-outfit text-sm"
                                    />
                                </div>

                                {/* Answer */}
                                <div className="space-y-2">
                                    <label className="text-xs font-semibold text-foreground/50 uppercase tracking-wider">{t('edit.answer_label')}</label>
                                    <textarea
                                        value={modalForm.answer}
                                        onChange={e => setModalForm(f => ({ ...f, answer: e.target.value }))}
                                        rows={2}
                                        className="w-full px-4 py-3 glass-card rounded-xl outline-none focus:ring-2 focus:ring-primary/30 resize-none font-outfit text-sm"
                                    />
                                </div>

                                {/* Actions */}
                                <div className="flex gap-3 pt-1">
                                    <button
                                        onClick={deleteQuestion}
                                        disabled={modalDeleting}
                                        className="p-3 rounded-xl hover:bg-rose-50 dark:hover:bg-rose-900/30 text-rose-500 transition-all disabled:opacity-40"
                                    >
                                        {modalDeleting ? <Loader2 size={18} className="animate-spin" /> : <Trash2 size={18} />}
                                    </button>
                                    <div className="flex-1" />
                                    <button
                                        onClick={closeEdit}
                                        className="px-4 py-2.5 rounded-xl hover:bg-slate-100 dark:hover:bg-white/10 transition-all font-semibold text-foreground/60 text-sm"
                                    >
                                        {tCommon('cancel')}
                                    </button>
                                    <button
                                        onClick={saveEdit}
                                        disabled={modalSaving}
                                        className="px-5 py-2.5 bg-primary text-white rounded-xl font-bold hover:opacity-90 disabled:opacity-50 transition-all flex items-center gap-2 text-sm"
                                    >
                                        {modalSaving ? <Loader2 size={16} className="animate-spin" /> : <Check size={16} />}
                                        {tCommon('save')}
                                    </button>
                                </div>
                            </div>
                        </motion.div>
                    </>
                )}
            </AnimatePresence>

            {/* ─── Add Question Modal ───────────────────────────── */}
            <AnimatePresence>
                {addOpen && (
                    <>
                        <motion.div
                            key="add-backdrop"
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            exit={{ opacity: 0 }}
                            onClick={() => setAddOpen(false)}
                            className="fixed inset-0 bg-black/50 backdrop-blur-sm z-40"
                        />

                        <motion.div
                            key="add-modal"
                            initial={{ opacity: 0, scale: 0.92, y: 20 }}
                            animate={{ opacity: 1, scale: 1, y: 0 }}
                            exit={{ opacity: 0, scale: 0.92, y: 20 }}
                            transition={{ type: 'spring', stiffness: 300, damping: 28 }}
                            className="fixed inset-0 z-50 flex items-center justify-center p-4 pointer-events-none"
                        >
                            <div
                                className="glass-card rounded-3xl p-6 w-full max-w-lg pointer-events-auto space-y-5 shadow-2xl"
                                onClick={e => e.stopPropagation()}
                            >
                                <div className="flex items-center justify-between">
                                    <h2 className="text-lg font-black font-outfit">{t('edit.add_title')}</h2>
                                    <button onClick={() => setAddOpen(false)} className="p-2 rounded-xl hover:bg-slate-100 dark:hover:bg-white/10 transition-all text-foreground/50">
                                        <X size={20} />
                                    </button>
                                </div>

                                <div className="space-y-2">
                                    <label className="text-xs font-semibold text-foreground/50 uppercase tracking-wider">{t('edit.question_label')}</label>
                                    <textarea
                                        value={newForm.question}
                                        onChange={e => setNewForm(f => ({ ...f, question: e.target.value }))}
                                        rows={3}
                                        placeholder={t('edit.question_placeholder')}
                                        className="w-full px-4 py-3 glass-card rounded-xl outline-none focus:ring-2 focus:ring-primary/30 resize-none font-outfit text-sm"
                                        autoFocus
                                    />
                                </div>

                                <div className="space-y-2">
                                    <label className="text-xs font-semibold text-foreground/50 uppercase tracking-wider">{t('edit.answer_label')}</label>
                                    <textarea
                                        value={newForm.answer}
                                        onChange={e => setNewForm(f => ({ ...f, answer: e.target.value }))}
                                        rows={2}
                                        placeholder={t('edit.answer_placeholder')}
                                        className="w-full px-4 py-3 glass-card rounded-xl outline-none focus:ring-2 focus:ring-primary/30 resize-none font-outfit text-sm"
                                    />
                                </div>

                                <div className="flex justify-end gap-3 pt-1">
                                    <button
                                        onClick={() => setAddOpen(false)}
                                        className="px-4 py-2.5 rounded-xl hover:bg-slate-100 dark:hover:bg-white/10 transition-all font-semibold text-foreground/60 text-sm"
                                    >
                                        {tCommon('cancel')}
                                    </button>
                                    <button
                                        onClick={addQuestion}
                                        disabled={addSaving || !newForm.question.trim() || !newForm.answer.trim()}
                                        className="px-6 py-2.5 bg-primary text-white rounded-xl font-bold hover:opacity-90 disabled:opacity-40 transition-all flex items-center gap-2 text-sm"
                                    >
                                        {addSaving ? <Loader2 size={16} className="animate-spin" /> : <Plus size={16} />}
                                        {t('edit.add_btn')}
                                    </button>
                                </div>
                            </div>
                        </motion.div>
                    </>
                )}
            </AnimatePresence>
        </div>
    );
}
