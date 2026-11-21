'use client';

import { useState } from 'react';
import { useRouter } from '@/i18n/routing';
import { useTranslations } from 'next-intl';
import { ChevronLeft, Sparkles, Loader2 } from 'lucide-react';
import api from '@/lib/api';
import { motion } from 'framer-motion';

export default function CreateGroupTaskPage() {
    const router = useRouter();
    const t = useTranslations('group_tasks');
    const tCommon = useTranslations('common');

    const [loading, setLoading] = useState(false);
    const [formData, setFormData] = useState({
        subject: '',
        grade: 8,
        topic: '',
        card_count: 8,
        language: 'kaz'
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);

        try {
            // 1. Create Draft
            const creationRes = await api.post('/group-tasks', {
                ...formData,
                card_count: parseInt(formData.card_count.toString()),
                grade: parseInt(formData.grade.toString())
            });

            const taskId = creationRes.data.data.id;

            // 2. Trigger Generation
            await api.post(`/group-tasks/${taskId}/generate`);

            // 3. Redirect back to list
            router.push('/teacher/materials/group-tasks');
        } catch (error: any) {
            console.error('Creation failed:', error);
            alert(error.response?.data?.error?.message || tCommon('error'));
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="max-w-4xl mx-auto space-y-8 pb-20">
            {/* Header */}
            <div className="flex items-center gap-4">
                <button
                    onClick={() => router.back()}
                    className="p-2 glass-card rounded-xl hover:bg-slate-200 dark:hover:bg-white/10 transition-all text-foreground/60"
                >
                    <ChevronLeft size={20} />
                </button>
                <div>
                    <h1 className="text-3xl font-bold font-outfit">{t('create_new')}</h1>
                    <p className="text-foreground/60">{t('description')}</p>
                </div>
            </div>

            <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
            >
                <form onSubmit={handleSubmit} className="glass-card p-8 rounded-[32px] space-y-6">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div className="md:col-span-2">
                            <label className="block text-sm font-semibold mb-2">{t('form.topic')}</label>
                            <input
                                type="text"
                                required
                                value={formData.topic}
                                onChange={(e) => setFormData({ ...formData, topic: e.target.value })}
                                className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                                placeholder={t('form.topic_placeholder')}
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-semibold mb-2">{t('form.subject')}</label>
                            <input
                                type="text"
                                required
                                value={formData.subject}
                                onChange={(e) => setFormData({ ...formData, subject: e.target.value })}
                                className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                                placeholder={t('form.subject_placeholder')}
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-semibold mb-2">{t('form.grade')}</label>
                            <select
                                required
                                value={formData.grade}
                                onChange={(e) => setFormData({ ...formData, grade: parseInt(e.target.value) })}
                                className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                            >
                                {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12].map(g => (
                                    <option key={g} value={g} className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{g} {tCommon('grade_suffix')}</option>
                                ))}
                            </select>
                        </div>

                        <div>
                            <label className="block text-sm font-semibold mb-2">{t('form.card_count')}</label>
                            <select
                                required
                                value={formData.card_count}
                                onChange={(e) => setFormData({ ...formData, card_count: parseInt(e.target.value) })}
                                className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                            >
                                {[4, 6, 8, 10, 12, 16].map(c => (
                                    <option key={c} value={c} className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{c} {t('form.card_suffix')}</option>
                                ))}
                            </select>
                        </div>

                        <div>
                            <label className="block text-sm font-semibold mb-2">{t('form.language')}</label>
                            <select
                                required
                                value={formData.language}
                                onChange={(e) => setFormData({ ...formData, language: e.target.value })}
                                className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                            >
                                <option value="kaz" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{t('form.lang_kaz')}</option>
                                <option value="rus" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{t('form.lang_rus')}</option>
                                <option value="eng" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{t('form.lang_eng')}</option>
                            </select>
                        </div>
                    </div>

                    <div className="flex items-center justify-end gap-4 pt-6 border-t border-slate-300 dark:border-white/10">
                        <button
                            type="button"
                            onClick={() => router.back()}
                            className="px-6 py-3 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl font-semibold transition-all"
                        >
                            {tCommon('cancel')}
                        </button>

                        <button
                            type="submit"
                            disabled={loading}
                            className="px-8 py-3 bg-primary text-white font-bold rounded-xl hover:bg-primary-hover disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg shadow-primary/25 flex items-center gap-2"
                        >
                            {loading ? (
                                <>
                                    <Loader2 className="animate-spin h-5 w-5" />
                                    <span>{tCommon('loading')}</span>
                                </>
                            ) : (
                                <>
                                    <Sparkles size={20} />
                                    <span>{t('form.create_btn')} (4 {tCommon('tokens')})</span>
                                </>
                            )}
                        </button>
                    </div>
                </form>
            </motion.div>
        </div>
    );
}
