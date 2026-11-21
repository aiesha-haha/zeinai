'use client';

import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { 
    X, 
    User, 
    CheckCircle2, 
    Clock, 
    AlertCircle,
    Trophy,
    Search,
    Loader2
} from 'lucide-react';
import api from '@/lib/api';
import { useTranslations } from 'next-intl';

const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
};

interface AssignmentResultsModalProps {
    isOpen: boolean;
    onClose: () => void;
    assignment: any;
}

export default function AssignmentResultsModal({ isOpen, onClose, assignment }: AssignmentResultsModalProps) {
    const t = useTranslations('tests');
    const [results, setResults] = React.useState<any[]>([]);
    const [loading, setLoading] = React.useState(true);
    const [searchTerm, setSearchTerm] = React.useState('');

    React.useEffect(() => {
        if (isOpen && assignment) {
            fetchResults();
        }
    }, [isOpen, assignment]);

    const fetchResults = async () => {
        try {
            setLoading(true);
            const res = await api.get(`/tests/assignments/${assignment.id}/results`);
            setResults(res.data.data || []);
        } catch (err) {
            console.error('Failed to fetch results:', err);
        } finally {
            setLoading(false);
        }
    };

    const filteredResults = results.filter(res => 
        res.student_name.toLowerCase().includes(searchTerm.toLowerCase())
    );

    const getStatusStyle = (status: string) => {
        switch (status) {
            case 'completed': return 'bg-green-500/10 text-green-500';
            case 'in_progress': return 'bg-blue-500/10 text-blue-500';
            case 'not_started': return 'bg-slate-500/10 text-slate-500';
            default: return 'bg-slate-500/10 text-slate-500';
        }
    };

    return (
        <AnimatePresence>
            {isOpen && (
                <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 md:p-6">
                    <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        onClick={onClose}
                        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
                    />

                    <motion.div
                        initial={{ opacity: 0, scale: 0.95, y: 20 }}
                        animate={{ opacity: 1, scale: 1, y: 0 }}
                        exit={{ opacity: 0, scale: 0.95, y: 20 }}
                        className="relative w-full max-w-4xl glass-card bg-white dark:bg-[#020617] p-0 overflow-hidden max-h-[90vh] flex flex-col shadow-2xl"
                    >
                        {/* Header */}
                        <div className="p-6 md:p-8 border-b border-slate-100 dark:border-white/5 bg-gradient-to-r from-primary/5 to-secondary/5">
                            <div className="flex justify-between items-start mb-4">
                                <div>
                                    <h2 className="text-2xl font-bold tracking-tight mb-1">{assignment.test_title}</h2>
                                    <p className="text-primary font-bold">{assignment.class_name}</p>
                                </div>
                                <button
                                    onClick={onClose}
                                    className="p-2 rounded-xl bg-white/50 dark:bg-white/5 hover:bg-red-500/10 hover:text-red-500 transition-all"
                                >
                                    <X size={20} />
                                </button>
                            </div>

                            <div className="flex flex-wrap gap-4 mt-6">
                                <div className="glass-card py-3 px-5 bg-white dark:bg-white/5">
                                    <p className="text-[10px] uppercase tracking-wider text-foreground/40 font-bold mb-1">Total Students</p>
                                    <p className="text-xl font-bold">{assignment.student_count}</p>
                                </div>
                                <div className="glass-card py-3 px-5 bg-white dark:bg-white/5">
                                    <p className="text-[10px] uppercase tracking-wider text-foreground/40 font-bold mb-1">Completed</p>
                                    <p className="text-xl font-bold text-green-500">{assignment.finished_count}</p>
                                </div>
                                <div className="glass-card py-3 px-5 bg-white dark:bg-white/5">
                                    <p className="text-[10px] uppercase tracking-wider text-foreground/40 font-bold mb-1">Avg Score</p>
                                    <p className="text-xl font-bold text-primary">
                                        {results.length > 0 && results.some(r => r.status === 'completed')
                                            ? Math.round(results.reduce((acc, curr) => acc + (curr.score || 0), 0) / results.filter(r => r.status === 'completed').length)
                                            : 0}%
                                    </p>
                                </div>
                            </div>
                        </div>

                        {/* Search and Table */}
                        <div className="p-6 flex-1 overflow-hidden flex flex-col">
                            <div className="relative mb-6">
                                <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-foreground/30" size={18} />
                                <input
                                    type="text"
                                    placeholder={t('search_students') || 'Search students...'}
                                    value={searchTerm}
                                    onChange={(e) => setSearchTerm(e.target.value)}
                                    className="w-full pl-12 pr-4 py-3 rounded-2xl bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:ring-2 focus:ring-primary/20 outline-none transition-all"
                                />
                            </div>

                            <div className="flex-1 overflow-y-auto custom-scrollbar">
                                {loading ? (
                                    <div className="flex flex-col items-center justify-center py-20">
                                        <Loader2 className="w-10 h-10 text-primary animate-spin mb-4" />
                                        <p className="text-foreground/60">Loading results...</p>
                                    </div>
                                ) : filteredResults.length === 0 ? (
                                    <div className="text-center py-20">
                                        <AlertCircle className="w-12 h-12 text-foreground/20 mx-auto mb-4" />
                                        <p className="text-xl font-bold text-foreground/40">No students found</p>
                                    </div>
                                ) : (
                                    <div className="grid gap-3">
                                        {filteredResults.map((res, index) => (
                                            <motion.div
                                                key={res.student_id}
                                                initial={{ opacity: 0, x: -20 }}
                                                animate={{ opacity: 1, x: 0 }}
                                                transition={{ delay: index * 0.05 }}
                                                className="flex items-center justify-between p-4 rounded-2xl bg-slate-50 dark:bg-white/5 border border-slate-100 dark:border-white/5"
                                            >
                                                <div className="flex items-center gap-4">
                                                    <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary/20 to-secondary/20 flex items-center justify-center text-primary font-bold">
                                                        {res.student_name[0]}
                                                    </div>
                                                    <div>
                                                        <h4 className="font-bold">{res.student_name}</h4>
                                                        <div className={`inline-block text-[10px] px-2 py-0.5 rounded-full font-bold uppercase tracking-wider ${getStatusStyle(res.status)}`}>
                                                            {res.status.replace('_', ' ')}
                                                        </div>
                                                    </div>
                                                </div>

                                                <div className="text-right">
                                                    {res.status === 'completed' ? (
                                                        <div>
                                                            <p className="text-xl font-black text-primary">
                                                                {res.score}<span className="text-xs text-foreground/40 font-normal"> / {res.max_score}</span>
                                                            </p>
                                                            <p className="text-[10px] text-foreground/40 font-bold uppercase tracking-tighter">
                                                                {res.completed_at ? formatDate(res.completed_at) : ''}
                                                            </p>
                                                        </div>
                                                    ) : (
                                                        <p className="text-sm font-bold text-foreground/20">-- / --</p>
                                                    )}
                                                </div>
                                            </motion.div>
                                        ))}
                                    </div>
                                )}
                            </div>
                        </div>
                    </motion.div>
                </div>
            )}
        </AnimatePresence>
    );
}
