'use client';

import { useState, useEffect } from 'react';
import { useTranslations } from 'next-intl';
import { X, Users, Loader2, CheckCircle2, Send, Calendar } from 'lucide-react';
import api from '@/lib/api';
import { motion, AnimatePresence } from 'framer-motion';

interface AssignTestModalProps {
    isOpen: boolean;
    onClose: () => void;
    testId: string;
    testTopic: string;
}

export default function AssignTestModal({ isOpen, onClose, testId, testTopic }: AssignTestModalProps) {
    const t = useTranslations('classes');
    const tTests = useTranslations('tests');
    const tCommon = useTranslations('common');

    const [loading, setLoading] = useState(false);
    const [fetchingClasses, setFetchingClasses] = useState(true);
    const [classes, setClasses] = useState<any[]>([]);
    const [selectedClassId, setSelectedClassId] = useState<number | null>(null);
    const [success, setSuccess] = useState(false);

    useEffect(() => {
        if (isOpen) {
            fetchClasses();
            setSuccess(false);
            setSelectedClassId(null);
        }
    }, [isOpen]);

    const fetchClasses = async () => {
        setFetchingClasses(true);
        try {
            const [classesRes, assignmentsRes] = await Promise.all([
                api.get('/classes'),
                api.get('/tests/assignments/all')
            ]);
            
            const allClasses = classesRes.data.data || [];
            const allAssignments = assignmentsRes.data.data || [];
            
            // Get IDs of classes that already have this test assigned
            const assignedClassIds = allAssignments
                .filter((a: any) => a.test_id === testId)
                .map((a: any) => a.class_id);
            
            // Filter out those classes
            const availableClasses = allClasses.filter((c: any) => !assignedClassIds.includes(c.id));
            
            setClasses(availableClasses);
        } catch (error) {
            console.error('Failed to fetch classes or assignments:', error);
        } finally {
            setFetchingClasses(false);
        }
    };

    const handleAssign = async () => {
        if (!selectedClassId) return;

        setLoading(true);
        try {
            await api.post(`/tests/${testId}/assign`, {
                class_id: selectedClassId
            });
            setSuccess(true);
            setTimeout(() => {
                onClose();
            }, 2000);
        } catch (error: any) {
            console.error('Failed to assign test:', error);
            alert(error.response?.data?.error?.message || tCommon('error_occurred'));
        } finally {
            setLoading(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4">
            <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                onClick={onClose}
                className="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
            />

            <motion.div
                initial={{ opacity: 0, scale: 0.9, y: 20 }}
                animate={{ opacity: 1, scale: 1, y: 0 }}
                exit={{ opacity: 0, scale: 0.9, y: 20 }}
                className="relative w-full max-w-lg bg-white dark:bg-slate-900 rounded-[32px] shadow-2xl overflow-hidden"
            >
                <div className="p-8">
                    <div className="flex items-center justify-between mb-8">
                        <div>
                            <h2 className="text-2xl font-bold font-outfit">{tTests('assign')}</h2>
                            <p className="text-foreground/60 text-sm mt-1 truncate max-w-[300px]">{testTopic}</p>
                        </div>
                        <button
                            onClick={onClose}
                            className="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-all"
                        >
                            <X size={24} />
                        </button>
                    </div>

                    <AnimatePresence mode="wait">
                        {success ? (
                            <motion.div
                                key="success"
                                initial={{ opacity: 0, scale: 0.9 }}
                                animate={{ opacity: 1, scale: 1 }}
                                className="py-12 flex flex-col items-center justify-center text-center space-y-4"
                            >
                                <div className="w-20 h-20 bg-green-500/20 rounded-full flex items-center justify-center text-green-500">
                                    <CheckCircle2 size={40} />
                                </div>
                                <h3 className="text-xl font-bold">{tCommon('success') || 'Success!'}</h3>
                                <p className="text-foreground/60">{tTests('status_completed')}</p>
                            </motion.div>
                        ) : (
                            <motion.div key="form" className="space-y-6">
                                <div className="space-y-3">
                                    <label className="text-sm font-bold text-foreground/60 flex items-center gap-2">
                                        <Users size={16} />
                                        {t('my_classes')}
                                    </label>

                                    <div className="max-h-[300px] overflow-y-auto pr-2 space-y-2 custom-scrollbar">
                                        {fetchingClasses ? (
                                            Array(3).fill(0).map((_, i) => (
                                                <div key={i} className="h-16 bg-slate-50 dark:bg-white/5 rounded-2xl animate-pulse" />
                                            ))
                                        ) : classes.length === 0 ? (
                                            <div className="p-8 text-center bg-slate-50 dark:bg-white/5 rounded-2xl">
                                                <p className="text-foreground/40 text-sm">{tCommon('error_not_found')}</p>
                                            </div>
                                        ) : (
                                            classes.map((cls) => (
                                                <button
                                                    key={cls.id}
                                                    onClick={() => setSelectedClassId(cls.id)}
                                                    className={`
                                                        w-full p-4 rounded-2xl border transition-all flex items-center justify-between
                                                        ${selectedClassId === cls.id
                                                            ? 'bg-primary/10 border-primary text-primary'
                                                            : 'bg-slate-50 dark:bg-white/5 border-transparent text-foreground/60 hover:border-foreground/10'}
                                                    `}
                                                >
                                                    <div className="flex flex-col items-start">
                                                        <span className="font-bold">{cls.name}</span>
                                                        <span className="text-xs opacity-60">{cls.students_count || 0} {t('students_count_label').toLowerCase()}</span>
                                                    </div>
                                                    {selectedClassId === cls.id && <CheckCircle2 size={20} />}
                                                </button>
                                            ))
                                        )}
                                    </div>
                                </div>

                                <div className="pt-4 flex gap-4">
                                    <button
                                        onClick={onClose}
                                        className="flex-1 px-6 py-4 rounded-2xl font-bold bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 transition-all"
                                    >
                                        {tCommon('cancel')}
                                    </button>
                                    <button
                                        onClick={handleAssign}
                                        disabled={loading || !selectedClassId}
                                        className="flex-[2] px-6 py-4 bg-primary text-white rounded-2xl font-bold hover:bg-primary-hover transition-all flex items-center justify-center gap-2 shadow-xl shadow-primary/25 disabled:opacity-50 disabled:cursor-not-allowed"
                                    >
                                        {loading ? (
                                            <Loader2 size={20} className="animate-spin" />
                                        ) : (
                                            <>
                                                <Send size={20} />
                                                {tTests('assign')}
                                            </>
                                        )}
                                    </button>
                                </div>
                            </motion.div>
                        )}
                    </AnimatePresence>
                </div>
            </motion.div>
        </div>
    );
}
