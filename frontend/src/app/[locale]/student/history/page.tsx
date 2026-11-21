'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { Link } from '@/i18n/routing';
import api from '@/lib/api';
import { 
    Clock, 
    ChevronRight, 
    ChevronLeft,
    Calendar,
    Search,
    BookOpen,
    Loader2
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

export default function StudentHistoryPage() {
    const t = useTranslations('student');
    const tCommon = useTranslations('common');
    const [history, setHistory] = React.useState<any[]>([]);
    const [loading, setLoading] = React.useState(true);
    const [page, setPage] = React.useState(1);
    const [totalPages, setTotalPages] = React.useState(1);
    const pageSize = 10;

    React.useEffect(() => {
        const fetchHistory = async () => {
            setLoading(true);
            try {
                // Assuming the backend supports pagination for history
                // If not, we'll slice the array on client side for now
                const res = await api.get('/whiteboard/student/history');
                const allHistory = res.data.data || [];
                
                // Client-side pagination if backend doesn't support it yet
                const start = (page - 1) * pageSize;
                const end = start + pageSize;
                setHistory(allHistory.slice(start, end));
                setTotalPages(Math.ceil(allHistory.length / pageSize));
            } catch (err) {
                console.error('Failed to fetch history:', err);
            } finally {
                setLoading(false);
            }
        };

        fetchHistory();
    }, [page]);

    const container = {
        hidden: { opacity: 0 },
        show: {
            opacity: 1,
            transition: {
                staggerChildren: 0.05
            }
        }
    };

    const item = {
        hidden: { y: 10, opacity: 0 },
        show: { y: 0, opacity: 1 }
    };

    return (
        <div className="space-y-8 pb-20">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-gradient">{t('past_lessons')}</h1>
                    <p className="text-foreground/60 mt-1">Review your previous sessions and summaries</p>
                </div>
            </div>

            {loading ? (
                <div className="space-y-4">
                    {[1, 2, 3, 4, 5].map((i) => (
                        <div key={i} className="h-20 rounded-2xl glass-card animate-pulse" />
                    ))}
                </div>
            ) : history.length > 0 ? (
                <>
                    <motion.div 
                        variants={container}
                        initial="hidden"
                        animate="show"
                        className="grid gap-4"
                    >
                        {history.map((session) => (
                            <motion.div
                                key={session.id}
                                variants={item}
                                className="p-5 glass-card rounded-3xl flex items-center justify-between hover:bg-white/5 transition-all group"
                            >
                                <div className="flex items-center gap-5">
                                    <div className="w-12 h-12 rounded-2xl bg-primary/10 flex items-center justify-center text-primary group-hover:scale-110 transition-transform">
                                        <BookOpen size={24} />
                                    </div>
                                    <div>
                                        <h4 className="font-bold text-lg">
                                            {session.title || `${t('lesson_label')} ${new Date(session.created_at).toLocaleDateString()}`}
                                        </h4>
                                        <div className="flex items-center gap-3 mt-1 text-sm text-foreground/40">
                                            <div className="flex items-center gap-1">
                                                <Calendar size={14} />
                                                {new Date(session.created_at).toLocaleDateString()}
                                            </div>
                                            <span>•</span>
                                            <div className="flex items-center gap-1">
                                                <Clock size={14} />
                                                {new Date(session.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                            </div>
                                        </div>
                                    </div>
                                </div>
                                <Link 
                                    href={`/student/whiteboard/summary/${session.id}`}
                                    className="px-6 py-2.5 bg-primary/10 text-primary rounded-xl text-sm font-bold hover:bg-primary hover:text-white transition-all flex items-center gap-2"
                                >
                                    {t('view_summary')}
                                    <ChevronRight size={16} />
                                </Link>
                            </motion.div>
                        ))}
                    </motion.div>

                    {/* Pagination */}
                    {totalPages > 1 && (
                        <div className="flex items-center justify-center gap-4 mt-12">
                            <button
                                onClick={() => setPage(p => Math.max(1, p - 1))}
                                disabled={page === 1}
                                className="p-2 rounded-xl glass-card disabled:opacity-20 hover:bg-white/10 transition-all"
                            >
                                <ChevronLeft />
                            </button>
                            <div className="flex items-center gap-2">
                                {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                                    <button
                                        key={p}
                                        onClick={() => setPage(p)}
                                        className={`w-10 h-10 rounded-xl font-bold transition-all ${
                                            page === p 
                                            ? 'bg-primary text-white shadow-lg shadow-primary/20' 
                                            : 'glass-card hover:bg-white/10'
                                        }`}
                                    >
                                        {p}
                                    </button>
                                ))}
                            </div>
                            <button
                                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                                disabled={page === totalPages}
                                className="p-2 rounded-xl glass-card disabled:opacity-20 hover:bg-white/10 transition-all"
                            >
                                <ChevronRight />
                            </button>
                        </div>
                    )}
                </>
            ) : (
                <div className="flex flex-col items-center justify-center py-20 glass-card rounded-[40px]">
                    <div className="w-20 h-20 rounded-full bg-primary/10 flex items-center justify-center mb-6">
                        <Clock size={40} className="text-primary" />
                    </div>
                    <h3 className="text-xl font-bold mb-2">No past lessons</h3>
                    <p className="text-foreground/60 max-w-xs text-center">
                        You haven't attended any lessons yet. Join your first session to see it here!
                    </p>
                </div>
            )}
        </div>
    );
}
