'use client';

import { useState } from 'react';
import { useRouter } from '@/i18n/routing';
import { useTranslations, useLocale } from 'next-intl';
import { ChevronLeft, Sparkles, Loader2, Users } from 'lucide-react';
import api from '@/lib/api';
import { motion } from 'framer-motion';

export default function CreateTugOfWarPage() {
    const router = useRouter();
    const t = useTranslations('tug_of_war');
    const tCommon = useTranslations('common');
    const locale = useLocale();

    const [loading, setLoading] = useState(false);
    const [formData, setFormData] = useState({
        subject: '',
        grade: 8,
        topic: '',
        group_1_name: locale === 'kk' ? '1-топ' : locale === 'ru' ? 'Команда 1' : 'Team 1',
        group_2_name: locale === 'kk' ? '2-топ' : locale === 'ru' ? 'Команда 2' : 'Team 2',
        language: locale === 'en' ? 'eng' : locale === 'kk' ? 'kaz' : 'rus'
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);

        try {
            // 1. Create Draft
            const creationRes = await api.post('/tug-of-war', {
                ...formData,
                grade: parseInt(formData.grade.toString())
            });

            const materialId = creationRes.data.data.id;

            // 2. Trigger Generation
            await api.post(`/tug-of-war/${materialId}/generate`);

            // 3. Redirect back to list
            router.push('/teacher/materials/tug-of-war');
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
                <form onSubmit={handleSubmit} className="glass-card p-8 rounded-[32px] space-y-8">
                    {/* Topic & Subject */}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div className="md:col-span-2">
                            <label className="block text-sm font-semibold mb-2">{t('form.topic')}</label>
                            <input
                                type="text"
                                required
                                value={formData.topic}
                                onChange={(e) => setFormData({ ...formData, topic: e.target.value })}
                                className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                                placeholder={t('form.topic')}
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
                                placeholder={t('form.subject')}
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
                    </div>

                    {/* Team Names */}
                    <div className="pt-6 border-t border-slate-300 dark:border-white/10">
                        <h3 className="text-lg font-bold mb-4 flex items-center gap-2">
                            <Users size={20} className="text-primary" />
                            {t('game.team_1')} & {t('game.team_2')}
                        </h3>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <div>
                                <label className="block text-sm font-semibold mb-2">{t('form.group_1_name')}</label>
                                <input
                                    type="text"
                                    required
                                    value={formData.group_1_name}
                                    onChange={(e) => setFormData({ ...formData, group_1_name: e.target.value })}
                                    className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-semibold mb-2">{t('form.group_2_name')}</label>
                                <input
                                    type="text"
                                    required
                                    value={formData.group_2_name}
                                    onChange={(e) => setFormData({ ...formData, group_2_name: e.target.value })}
                                    className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                                />
                            </div>
                        </div>
                    </div>

                    {/* Language */}
                    <div className="pt-6 border-t border-slate-300 dark:border-white/10">
                        <div>
                            <label className="block text-sm font-semibold mb-2">{t('form.language')}</label>
                            <select
                                required
                                value={formData.language}
                                onChange={(e) => setFormData({ ...formData, language: e.target.value })}
                                className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                            >
                                <option value="kaz" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">Қазақша</option>
                                <option value="rus" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">Русский</option>
                                <option value="eng" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">English</option>
                            </select>
                        </div>
                    </div>

                    <div className="flex items-center justify-end gap-4 pt-8 border-t border-slate-300 dark:border-white/10">
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
                                    <span>{t('form.create_btn')} (5 {tCommon('tokens')})</span>
                                </>
                            )}
                        </button>
                    </div>
                </form>
            </motion.div>
        </div>
    );
}
