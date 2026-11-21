'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { Link } from '@/i18n/routing';
import api from '@/lib/api';
import { 
    ClipboardList, 
    Clock, 
    ArrowRight, 
    CheckCircle2, 
    AlertCircle, 
    Search,
    Filter,
    Calendar,
    GraduationCap
} from 'lucide-react';
import { motion } from 'framer-motion';
import { isGameType, testTypeKey } from '@/lib/testTypes';

export default function StudentTestsPage() {
    const t = useTranslations('student');
    const tCommon = useTranslations('common');
    const tD = useTranslations('dashboard');
    const tTypes = useTranslations('test_types');
    const [tests, setTests] = React.useState<any[]>([]);
    const [loading, setLoading] = React.useState(true);
    const [searchQuery, setSearchQuery] = React.useState('');

    React.useEffect(() => {
        const fetchTests = async () => {
            try {
                const res = await api.get('/student/tests');
                setTests(res.data.data || []);
            } catch (err) {
                console.error('Failed to fetch tests:', err);
            } finally {
                setLoading(false);
            }
        };

        fetchTests();
    }, []);

    const filteredTests = tests.filter(test => 
        test.topic.toLowerCase().includes(searchQuery.toLowerCase()) ||
        test.subject.toLowerCase().includes(searchQuery.toLowerCase())
    );

    const container = {
        hidden: { opacity: 0 },
        show: {
            opacity: 1,
            transition: {
                staggerChildren: 0.1
            }
        }
    };

    const item = {
        hidden: { y: 20, opacity: 0 },
        show: { y: 0, opacity: 1 }
    };

    return (
        <div className="space-y-8">
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-gradient">{t('assigned_tests')}</h1>
                    <p className="text-muted mt-1">{tD('tasks_subtitle')}</p>
                </div>

                <div className="flex items-center gap-3">
                    <div className="relative group">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-foreground/40 group-focus-within:text-primary transition-colors" size={18} />
                        <input 
                            type="text" 
                            placeholder={tCommon('search')}
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="pl-10 pr-4 py-2.5 rounded-xl border border-slate-200 dark:border-white/10 bg-white/50 dark:bg-slate-900/50 backdrop-blur-sm focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-all w-full md:w-64"
                        />
                    </div>
                </div>
            </div>

            {loading ? (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {[1, 2, 3].map((i) => (
                        <div key={i} className="h-64 rounded-3xl glass-card animate-pulse" />
                    ))}
                </div>
            ) : filteredTests.length > 0 ? (
                <motion.div 
                    variants={container}
                    initial="hidden"
                    animate="show"
                    className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
                >
                    {filteredTests.map((test) => (
                        <motion.div key={test.assignment_id} variants={item}>
                            <Link href={`/student/tests/${test.assignment_id}`}>
                                <div className="group relative glass-card p-6 rounded-3xl hover:shadow-2xl hover:shadow-primary/5 transition-all duration-500 overflow-hidden">
                                    {/* Decor */}
                                    <div className="absolute top-0 right-0 p-8 opacity-[0.03] group-hover:opacity-[0.06] transition-opacity">
                                        <ClipboardList size={120} />
                                    </div>

                                    <div className="relative z-10 space-y-4">
                                        <div className="flex justify-between items-start">
                                            <div className="p-3 rounded-2xl bg-primary/10 text-primary group-hover:scale-110 transition-transform">
                                                <GraduationCap size={24} />
                                            </div>
                                            <div className={`px-3 py-1 rounded-full text-[11px] font-bold ${test.status === 'completed'
                                                ? 'bg-slate-100 text-slate-700 dark:bg-white/10 dark:text-white/80'
                                                : 'bg-emerald-100 text-emerald-800 dark:bg-emerald-500/15 dark:text-emerald-300'}`}>
                                                {test.status === 'completed' ? tD('status_done') : tTypes(`${testTypeKey(test.test_type)}.title`)}
                                            </div>
                                        </div>

                                        <div>
                                            <p className="text-xs font-bold text-primary uppercase tracking-widest mb-1">{test.subject}</p>
                                            <h3 className="text-xl font-bold group-hover:text-primary transition-colors line-clamp-1">{test.topic}</h3>
                                        </div>

                                        <div className="grid grid-cols-2 gap-4 py-2">
                                            <div className="flex items-center gap-2 text-foreground/60 text-sm">
                                                <Calendar size={14} className="text-primary/60" />
                                                <span>{test.grade} {tCommon('grade_suffix')}</span>
                                            </div>
                                            <div className="flex items-center gap-2 text-foreground/60 text-sm">
                                                <Clock size={14} className="text-primary/60" />
                                                <span>{new Date(test.assigned_at).toLocaleDateString()}</span>
                                            </div>
                                        </div>

                                        <div className="pt-4 flex items-center justify-between border-t border-slate-200 dark:border-white/5">
                                            <span className="text-sm font-semibold text-primary">
                                                {test.status === 'completed' ? `${t('test_score')}: ${test.score}/${test.max_score}` : isGameType(test.test_type) ? tD('play') : t('start_test')}
                                            </span>
                                            <div className="w-10 h-10 rounded-full bg-slate-100 dark:bg-white/5 flex items-center justify-center group-hover:bg-primary group-hover:text-white transition-all">
                                                <ArrowRight size={18} />
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </Link>
                        </motion.div>
                    ))}
                </motion.div>
            ) : (
                <div className="flex flex-col items-center justify-center py-20 glass-card rounded-3xl">
                    <div className="w-20 h-20 rounded-full bg-primary/10 flex items-center justify-center mb-6">
                        <ClipboardList size={40} className="text-primary" />
                    </div>
                    <h3 className="text-xl font-bold mb-2">{t('no_tests')}</h3>
                    <p className="text-foreground/60 max-w-xs text-center">
                        {tD('student_no_tasks')}
                    </p>
                </div>
            )}
        </div>
    );
}
