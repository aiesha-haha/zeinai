'use client';

import React from 'react';
import { motion } from 'framer-motion';
import { 
    BarChart2, 
    Users, 
    Calendar, 
    ChevronRight, 
    Trash2, 
    Eye,
    Search,
    Filter,
    CheckCircle2,
    Clock,
    AlertCircle,
    ArrowLeft
} from 'lucide-react';
import api from '@/lib/api';
import { useTranslations } from 'next-intl';
import { Link, useRouter } from '@/i18n/routing';
import AssignmentResultsModal from '@/components/tests/AssignmentResultsModal';

const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric'
    });
};

export default function TeacherAssignmentsPage() {
    const t = useTranslations('tests');
    const tCommon = useTranslations('common');
    const [assignments, setAssignments] = React.useState<any[]>([]);
    const [loading, setLoading] = React.useState(true);
    const [selectedAssignment, setSelectedAssignment] = React.useState<any>(null);
    const [isResultsModalOpen, setIsResultsModalOpen] = React.useState(false);

    const fetchAssignments = async () => {
        try {
            setLoading(true);
            const res = await api.get('/tests/assignments/all');
            setAssignments(res.data.data || []);
        } catch (err) {
            console.error('Failed to fetch assignments:', err);
        } finally {
            setLoading(false);
        }
    };

    React.useEffect(() => {
        fetchAssignments();
    }, []);

    const handleDelete = async (id: string) => {
        if (!confirm(tCommon('confirm_delete') || 'Are you sure?')) return;
        try {
            await api.delete(`/tests/assignments/${id}`);
            setAssignments(assignments.filter(a => a.id !== id));
        } catch (err) {
            console.error('Failed to delete assignment:', err);
        }
    };

    const viewResults = (assignment: any) => {
        setSelectedAssignment(assignment);
        setIsResultsModalOpen(true);
    };

    return (
        <div className="space-y-8 pb-20">
            <Link 
                href="/teacher/materials/tests"
                className="inline-flex items-center gap-2 text-foreground/40 hover:text-primary transition-colors group"
            >
                <div className="p-2 rounded-xl bg-slate-100 dark:bg-white/5 group-hover:bg-primary group-hover:text-white transition-all">
                    <ArrowLeft size={20} />
                </div>
                <span className="font-bold text-sm">{tCommon('back') || 'Back'}</span>
            </Link>

            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-gradient mb-2">
                        {t('assigned_tests_title') || 'Assigned Tests'}
                    </h1>
                    <p className="text-foreground/60">
                        {t('assigned_tests_subtitle') || 'Track student progress and view results'}
                    </p>
                </div>
            </div>

            {loading ? (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {[1, 2, 3].map(i => (
                        <div key={i} className="glass-card p-6 h-48 animate-pulse bg-slate-200/50 dark:bg-white/5" />
                    ))}
                </div>
            ) : assignments.length === 0 ? (
                <div className="glass-card p-12 text-center">
                    <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center mx-auto mb-4">
                        <BarChart2 className="w-8 h-8 text-primary" />
                    </div>
                    <h3 className="text-xl font-bold mb-2">{t('no_assignments') || 'No Assigned Tests'}</h3>
                    <p className="text-foreground/60 mb-6">{t('no_assignments_desc') || 'You haven\'t assigned any tests to your classes yet.'}</p>
                    <Link href="/teacher/materials/generation">
                        <button className="premium-button">
                            {t('create_and_assign') || 'Create New Test'}
                        </button>
                    </Link>
                </div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {assignments.map((assignment, index) => (
                        <motion.div
                            key={assignment.id}
                            initial={{ opacity: 0, y: 20 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ delay: index * 0.1 }}
                            className="glass-card group hover:scale-[1.02] transition-all duration-300 overflow-hidden"
                        >
                            <div className="p-6">
                                <div className="flex justify-between items-start mb-4">
                                    <div className="p-2 rounded-xl bg-primary/10 text-primary">
                                        <Users size={20} />
                                    </div>
                                    <div className="flex gap-2">
                                        <button 
                                            onClick={() => viewResults(assignment)}
                                            className="p-2 rounded-lg hover:bg-primary/10 text-primary transition-colors"
                                            title="View Results"
                                        >
                                            <Eye size={18} />
                                        </button>
                                        <button 
                                            onClick={() => handleDelete(assignment.id)}
                                            className="p-2 rounded-lg hover:bg-red-500/10 text-red-500 transition-colors"
                                            title="Delete Assignment"
                                        >
                                            <Trash2 size={18} />
                                        </button>
                                    </div>
                                </div>

                                <h3 className="text-xl font-bold mb-1 line-clamp-1">{assignment.test_title}</h3>
                                <p className="text-primary font-bold text-sm mb-4">{assignment.class_name}</p>

                                <div className="space-y-4">
                                    <div>
                                        <div className="flex justify-between text-sm mb-1">
                                            <span className="text-foreground/60">{t('completion_rate') || 'Completion'}</span>
                                            <span className="font-bold">{assignment.finished_count} / {assignment.student_count}</span>
                                        </div>
                                        <div className="h-2 w-full bg-slate-100 dark:bg-white/5 rounded-full overflow-hidden">
                                            <motion.div 
                                                initial={{ width: 0 }}
                                                animate={{ width: `${(assignment.finished_count / Math.max(1, assignment.student_count)) * 100}%` }}
                                                className="h-full bg-primary"
                                            />
                                        </div>
                                    </div>

                                    <div className="flex items-center gap-4 text-xs text-foreground/40">
                                        <div className="flex items-center gap-1">
                                            <Calendar size={14} />
                                            <span>{formatDate(assignment.assigned_at)}</span>
                                        </div>
                                        <div className="flex items-center gap-1">
                                            <Clock size={14} />
                                            <span>{assignment.status}</span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                            
                            <button 
                                onClick={() => viewResults(assignment)}
                                className="w-full py-3 bg-slate-50 dark:bg-white/5 border-t border-slate-100 dark:border-white/5 font-bold text-sm text-primary hover:bg-primary hover:text-white transition-all flex items-center justify-center gap-2"
                            >
                                {t('view_details') || 'View Results'}
                                <ChevronRight size={16} />
                            </button>
                        </motion.div>
                    ))}
                </div>
            )}

            {selectedAssignment && (
                <AssignmentResultsModal
                    isOpen={isResultsModalOpen}
                    onClose={() => setIsResultsModalOpen(false)}
                    assignment={selectedAssignment}
                />
            )}
        </div>
    );
}
