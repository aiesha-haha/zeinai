'use client';

import { useState, useEffect } from 'react';
import { useRouter } from '@/i18n/routing';
import { useTranslations, useLocale } from 'next-intl';
import { ChevronLeft, Sparkles, Plus, BookOpen, Target, Calendar, Check, AlertCircle, Search, XCircle, Loader2, X } from 'lucide-react';
import api from '@/lib/api';
import { motion, AnimatePresence } from 'framer-motion';

interface TeachingMethod {
    id: string;
    name_kaz: string;
    name_rus: string;
    name_eng: string;
    purpose_kaz?: string;
    purpose_rus?: string;
    purpose_eng?: string;
    type_kaz?: string;
    type_rus?: string;
    type_eng?: string;
}

interface NationalValue {
    id: string;
    name_kaz: string;
    name_rus: string;
    name_eng: string;
    month_number: number;
}

export default function ShortTermPlanPage() {
    const router = useRouter();
    const t = useTranslations('education_materials');
    const tCommon = useTranslations('common');
    const tKmj = useTranslations('kmj');
    const locale = useLocale();

    // Helper function to get localized field
    const getLocalizedField = (obj: any, field: string) => {
        const langSuffix = locale === 'kk' ? 'kaz' : locale === 'ru' ? 'rus' : 'eng';
        return obj[`${field}_${langSuffix}`] || obj[`${field}_kaz`];
    };

    const getMethodTypeLabel = (type: string | undefined) => {
        if (!type) return null;
        const lowerType = type.toLowerCase();
        // Use translation from JSON if available, otherwise fallback to the field itself
        const translationKey = `method_types.${lowerType}`;
        const translated = tKmj(translationKey);
        // next-intl returns the key if translation is missing
        return translated === translationKey ? type : translated;
    };

    const [mode, setMode] = useState<'list' | 'create'>('list');
    const [loading, setLoading] = useState(false);
    const [plans, setPlans] = useState<any[]>([]);
    const [methods, setMethods] = useState<TeachingMethod[]>([]);
    const [nationalValue, setNationalValue] = useState<NationalValue | null>(null);
    const [selectedMethods, setSelectedMethods] = useState<string[]>([]);
    const [searchTerm, setSearchTerm] = useState('');
    const [isMethodsModalOpen, setIsMethodsModalOpen] = useState(false);

    const [formData, setFormData] = useState({
        subject: '',
        grade: 8,
        lesson_number: 1,
        lesson_topic: '',
        teacher_name: '',
        lesson_date: new Date().toISOString().split('T')[0],
        content_lang: 'kaz',
        learning_objective: '',
        duration_minutes: 45
    });

    // Fetch plans
    useEffect(() => {
        if (mode === 'list') {
            fetchPlans();
        }
    }, [mode]);

    // Poll for generating plans
    useEffect(() => {
        let interval: NodeJS.Timeout;
        if (mode === 'list' && plans.some(p => p.status === 'generating')) {
            interval = setInterval(fetchPlans, 5000);
        }
        return () => clearInterval(interval);
    }, [mode, plans]);

    const fetchPlans = async () => {
        setLoading(true);
        try {
            const response = await api.get('/kmj');
            setPlans(response.data.data.plans || []);
        } catch (error) {
            console.error('Failed to fetch plans:', error);
        } finally {
            setLoading(false);
        }
    };

    // Fetch teaching methods
    useEffect(() => {
        const fetchMethods = async () => {
            try {
                const response = await api.get('/teaching-methods');
                setMethods(response.data.data || []);
            } catch (error) {
                console.error('Failed to fetch methods:', error);
            }
        };
        fetchMethods();
    }, []);

    // Fetch profile for auto-filling teacher name
    useEffect(() => {
        const fetchProfile = async () => {
            try {
                const res = await api.get('/users/me');
                if (res.data.data) {
                    const { first_name, last_name } = res.data.data;
                    setFormData(prev => ({
                        ...prev,
                        teacher_name: `${first_name} ${last_name}`.trim()
                    }));
                }
            } catch (error) {
                console.error('Failed to fetch profile for auto-fill:', error);
            }
        };
        fetchProfile();
    }, []);

    // Fetch national value for selected month
    useEffect(() => {
        const fetchNationalValue = async () => {
            if (!formData.lesson_date) return;

            try {
                const month = new Date(formData.lesson_date).getMonth() + 1;
                const response = await api.get(`/national-values/month/${month}`);
                if (response.data.data && response.data.data.length > 0) {
                    setNationalValue(response.data.data[0]);
                }
            } catch (error) {
                console.error('Failed to fetch national value:', error);
            }
        };
        fetchNationalValue();
    }, [formData.lesson_date]);

    const filteredMethods = methods.filter(m => {
        const query = searchTerm.toLowerCase();
        return (
            m.name_kaz.toLowerCase().includes(query) ||
            m.name_rus.toLowerCase().includes(query) ||
            m.name_eng.toLowerCase().includes(query) ||
            (m.purpose_kaz && m.purpose_kaz.toLowerCase().includes(query)) ||
            (m.purpose_rus && m.purpose_rus.toLowerCase().includes(query)) ||
            (m.purpose_eng && m.purpose_eng.toLowerCase().includes(query))
        );
    }).sort((a, b) => {
        const aSelected = selectedMethods.includes(a.id);
        const bSelected = selectedMethods.includes(b.id);
        if (aSelected && !bSelected) return -1;
        if (!aSelected && bSelected) return 1;
        return 0;
    });

    const displayedMethods = filteredMethods.slice(0, 6);

    const handleMethodToggle = (methodId: string) => {
        setSelectedMethods(prev => {
            if (prev.includes(methodId)) {
                return prev.filter(id => id !== methodId);
            } else if (prev.length < 4) {
                return [...prev, methodId];
            }
            return prev;
        });
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (loading) return; // one plan per click
        setLoading(true);

        try {
            const response = await api.post('/kmj', {
                ...formData,
                grade: parseInt(formData.grade.toString()),
                lesson_number: parseInt(formData.lesson_number.toString()),
                duration_minutes: parseInt(formData.duration_minutes.toString()),
                lesson_date: `${formData.lesson_date}T00:00:00Z`,
                method_ids: selectedMethods
            });

            const planId = response.data.data.plan.id;

            // Start generation
            await api.post(`/kmj/${planId}/generate`);

            setMode('list');
            fetchPlans();
        } catch (error: any) {
            console.error('Creation failed:', error);
            alert(error.response?.data?.error?.message || tCommon('error'));
        } finally {
            setLoading(false);
        }
    };

    // A failed generation can be started again (no new plan is created)
    const handleRetryPlan = async (id: string) => {
        try {
            await api.post(`/kmj/${id}/generate`);
            fetchPlans();
        } catch (error: any) {
            alert(error.response?.data?.error?.message || tCommon('error'));
        }
    };

    const handleDeletePlan = async (id: string) => {
        if (!confirm(tKmj('delete_confirm'))) return;
        try {
            await api.delete(`/kmj/${id}`);
            setPlans(plans.filter(p => p.id !== id));
        } catch (error) {
            console.error('Failed to delete plan:', error);
        }
    };

    const monthNames = tKmj.raw('months') as string[];

    const selectedMonth = new Date(formData.lesson_date).getMonth();

    return (
        <div className="space-y-8 pb-20">
            {/* Header */}
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div className="flex items-center gap-4">
                    <button
                        onClick={() => mode === 'create' ? setMode('list') : router.push('/teacher/materials/generation')}
                        aria-label={tCommon('back')}
                        className="p-2 glass-card rounded-xl hover:bg-slate-200 dark:hover:bg-white/10 transition-all text-foreground/60"
                    >
                        <ChevronLeft size={20} />
                    </button>
                    <div>
                        <h1 className="text-3xl font-bold font-outfit">{tKmj('title')}</h1>
                        <p className="text-foreground/60">{tKmj('description')}</p>
                    </div>
                </div>

                {mode === 'list' && (
                    <button
                        onClick={() => setMode('create')}
                        className="px-6 py-3 bg-primary text-white rounded-xl font-bold hover:bg-primary-hover transition-all flex items-center justify-center gap-2 shadow-lg shadow-primary/25"
                    >
                        <Plus size={20} />
                        {tKmj('create_new')}
                    </button>
                )}
            </div>

            <AnimatePresence mode="wait">
                {mode === 'list' ? (
                    <motion.div
                        key="list"
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -20 }}
                        className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
                    >
                        {loading && plans.length === 0 ? (
                            Array(3).fill(0).map((_, i) => (
                                <div key={i} className="glass-card p-6 rounded-[32px] h-48 animate-pulse bg-slate-100 dark:bg-white/5" />
                            ))
                        ) : plans.length === 0 ? (
                            <div className="col-span-full py-20 text-center glass-card rounded-[32px] bg-slate-50 dark:bg-white/5">
                                <div className="w-20 h-20 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-4">
                                    <BookOpen className="text-primary" size={32} />
                                </div>
                                <h3 className="text-xl font-bold mb-2">{tKmj('empty_title')}</h3>
                                <p className="text-foreground/60 mb-6">{tKmj('empty_description')}</p>
                                <button
                                    onClick={() => setMode('create')}
                                    className="px-8 py-3 bg-primary text-white rounded-xl font-bold"
                                >
                                    {tKmj('start_button')}
                                </button>
                            </div>
                        ) : (
                            plans.map((plan) => (
                                <div
                                    key={plan.id}
                                    className={`
                                        glass-card p-6 rounded-[32px] group relative transition-all duration-500
                                        ${plan.status === 'generating' ? 'opacity-70 blur-[1px]' : 'hover:scale-[1.02] hover:shadow-2xl hover:shadow-primary/10'}
                                    `}
                                >
                                    <div className="flex justify-between items-start mb-4">
                                        <div className={`
                                            w-12 h-12 rounded-2xl flex items-center justify-center text-2xl
                                            ${plan.status === 'completed' ? 'bg-green-500/20' : plan.status === 'failed' ? 'bg-red-500/20' : 'bg-primary/20 animate-pulse'}
                                        `}>
                                            {plan.status === 'completed' ? '📝' : plan.status === 'failed' ? '⚠️' : '🪄'}
                                        </div>
                                        <div className="flex gap-2">
                                            {plan.status === 'generating' && (
                                                <span className="px-3 py-1 bg-primary/20 text-primary text-xs font-bold rounded-full animate-pulse">
                                                    {tKmj('status_generating')}
                                                </span>
                                            )}
                                            <button
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    handleDeletePlan(plan.id);
                                                }}
                                                className="opacity-60 group-hover:opacity-100 focus-visible:opacity-100 p-2 hover:bg-red-500/20 hover:text-red-600 rounded-lg transition-all"
                                                aria-label={`${tCommon('delete')}: ${plan.lesson_topic}`}
                                            >
                                                <XCircle size={18} />
                                            </button>
                                        </div>
                                    </div>

                                    <h3 className="text-xl font-bold mb-2 line-clamp-1">{plan.lesson_topic}</h3>
                                    <div className="space-y-2 text-sm text-foreground/60 mb-6">
                                        <p className="flex items-center gap-2">
                                            <BookOpen size={14} /> {plan.subject}
                                        </p>
                                        <p className="flex items-center gap-2">
                                            <Calendar size={14} /> {new Date(plan.created_at).toLocaleDateString()}
                                        </p>
                                    </div>

                                    {plan.status === 'failed' ? (
                                        <button onClick={() => handleRetryPlan(plan.id)} className="btn-primary w-full">
                                            {tKmj('retry_generation')}
                                        </button>
                                    ) : (
                                    <button
                                        onClick={() => plan.status !== 'generating' && router.push(`/teacher/materials/short-term-plan/${plan.id}`)}
                                        disabled={plan.status === 'generating'}
                                        className={`
                                            w-full py-3 rounded-xl font-bold transition-all
                                            ${plan.status === 'completed' ? 'bg-primary text-white hover:bg-primary-hover shadow-lg shadow-primary/20' : 'bg-slate-100 dark:bg-white/5 text-foreground/40 cursor-not-allowed'}
                                        `}
                                    >
                                        {plan.status === 'completed' ? tKmj('open_kmj') : tKmj('wait')}
                                    </button>
                                    )}
                                </div>
                            ))
                        )}
                    </motion.div>
                ) : (
                    <motion.div
                        key="create"
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{ opacity: 0, x: -20 }}
                    >
                        <form onSubmit={handleSubmit} className="space-y-6">
                            <div className="glass-card p-8 rounded-[32px] space-y-6">
                                {/* Basic Info */}
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                    <div>
                                        <label className="block text-sm font-semibold mb-2">{tKmj('form.subject_label')}</label>
                                        <input
                                            type="text"
                                            required
                                            value={formData.subject}
                                            onChange={(e) => setFormData({ ...formData, subject: e.target.value })}
                                            className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                                            placeholder={tKmj('form.subject_placeholder')}
                                        />
                                    </div>

                                    <div>
                                        <label className="block text-sm font-semibold mb-2">{tKmj('form.teacher_name_label')}</label>
                                        <input
                                            type="text"
                                            required
                                            value={formData.teacher_name}
                                            onChange={(e) => setFormData({ ...formData, teacher_name: e.target.value })}
                                            className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                                            placeholder={tKmj('form.teacher_placeholder')}
                                        />
                                    </div>

                                    <div>
                                        <label className="block text-sm font-semibold mb-2">{tKmj('form.grade_label')}</label>
                                        <select
                                            required
                                            value={formData.grade}
                                            onChange={(e) => setFormData({ ...formData, grade: parseInt(e.target.value) })}
                                            className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                                        >
                                            {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12].map(g => (
                                                <option key={g} value={g} className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{g}{tKmj('form.grade_suffix')}</option>
                                            ))}
                                        </select>
                                    </div>

                                    <div>
                                        <label className="block text-sm font-semibold mb-2">{tKmj('form.lesson_number_label')}</label>
                                        <input
                                            type="number"
                                            required
                                            min="1"
                                            value={formData.lesson_number}
                                            onChange={(e) => setFormData({ ...formData, lesson_number: parseInt(e.target.value) })}
                                            className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                                        />
                                    </div>

                                    <div className="md:col-span-2">
                                        <label className="block text-sm font-semibold mb-2">{tKmj('form.lesson_topic_label')}</label>
                                        <input
                                            type="text"
                                            required
                                            value={formData.lesson_topic}
                                            onChange={(e) => setFormData({ ...formData, lesson_topic: e.target.value })}
                                            className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                                            placeholder={tKmj('form.topic_placeholder')}
                                        />
                                    </div>

                                    <div className="md:col-span-2">
                                        <label htmlFor="kmj-objective" className="block text-sm font-semibold mb-2">{tKmj('form.learning_objective_label')}</label>
                                        <textarea
                                            id="kmj-objective"
                                            rows={2}
                                            maxLength={2000}
                                            value={formData.learning_objective}
                                            onChange={(e) => setFormData({ ...formData, learning_objective: e.target.value })}
                                            className="field"
                                            placeholder={tKmj('form.learning_objective_placeholder')}
                                            aria-describedby="kmj-objective-help"
                                        />
                                        <p id="kmj-objective-help" className="text-xs text-muted mt-1">{tKmj('form.learning_objective_help')}</p>
                                    </div>

                                    <div>
                                        <label htmlFor="kmj-duration" className="block text-sm font-semibold mb-2">{tKmj('form.duration_label')}</label>
                                        <select
                                            id="kmj-duration"
                                            value={formData.duration_minutes}
                                            onChange={(e) => setFormData({ ...formData, duration_minutes: parseInt(e.target.value) })}
                                            className="field"
                                        >
                                            {[30, 35, 40, 45, 60, 80, 90].map((m) => (
                                                <option key={m} value={m}>{tKmj('form.duration_option', { m })}</option>
                                            ))}
                                        </select>
                                    </div>

                                    <div>
                                        <label className="block text-sm font-semibold mb-2">{tKmj('form.lesson_date_label')}</label>
                                        <input
                                            type="date"
                                            required
                                            value={formData.lesson_date}
                                            onChange={(e) => setFormData({ ...formData, lesson_date: e.target.value })}
                                            className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                                        />
                                    </div>

                                    <div>
                                        <label className="block text-sm font-semibold mb-2">{tKmj('form.content_lang_label')}</label>
                                        <select
                                            required
                                            value={formData.content_lang}
                                            onChange={(e) => setFormData({ ...formData, content_lang: e.target.value })}
                                            className="w-full px-4 py-3 glass-card rounded-xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                                        >
                                            <option value="kaz" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{tKmj('form.lang_kaz')}</option>
                                            <option value="rus" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{tKmj('form.lang_rus')}</option>
                                            <option value="eng" className="bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100">{tKmj('form.lang_eng')}</option>
                                        </select>
                                    </div>
                                </div>

                                {/* National Value */}
                                {nationalValue && (
                                    <div className="bg-amber-500/10 border border-amber-500/20 rounded-2xl p-6">
                                        <div className="flex items-start space-x-4">
                                            <div className="flex-shrink-0 w-12 h-12 bg-amber-500/20 rounded-full flex items-center justify-center">
                                                <span className="text-2xl">✨</span>
                                            </div>
                                            <div className="flex-1">
                                                <h3 className="text-lg font-bold text-amber-600 dark:text-amber-400 mb-2">
                                                    {monthNames[selectedMonth]} {tKmj('form.month_value')}
                                                </h3>
                                                <p className="text-xl font-bold mb-2">
                                                    {getLocalizedField(nationalValue, 'name')}
                                                </p>
                                                <p className="text-sm opacity-70">
                                                    {tKmj('form.value_auto_added')}
                                                </p>
                                            </div>
                                        </div>
                                    </div>
                                )}

                                {/* Teaching Methods */}
                                <div>
                                    <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-4">
                                        <label className="text-lg font-bold">{tKmj('form.methods_label')}</label>
                                        <div className="flex items-center gap-4">
                                            <div className="relative">
                                                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-foreground/40" />
                                                <input
                                                    type="text"
                                                    placeholder={tKmj('form.search_methods')}
                                                    value={searchTerm}
                                                    onChange={(e) => setSearchTerm(e.target.value)}
                                                    className="pl-9 pr-4 py-2 glass-card rounded-xl text-sm outline-none focus:border-primary transition-all w-full md:w-64"
                                                />
                                            </div>
                                            <span className="text-sm font-semibold px-3 py-1 bg-primary/10 text-primary rounded-full shrink-0">
                                                {selectedMethods.length}/4 {tKmj('form.selected')}
                                            </span>
                                        </div>
                                    </div>

                                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                        {displayedMethods.map((method) => (
                                            <button
                                                key={method.id}
                                                type="button"
                                                onClick={() => handleMethodToggle(method.id)}
                                                disabled={!selectedMethods.includes(method.id) && selectedMethods.length >= 4}
                                                className={`
                                                    p-4 rounded-xl border text-left transition-all
                                                    ${selectedMethods.includes(method.id)
                                                        ? 'border-primary bg-primary/10'
                                                        : 'glass-card hover:border-primary/50'
                                                    }
                                                    ${!selectedMethods.includes(method.id) && selectedMethods.length >= 4
                                                        ? 'opacity-50 cursor-not-allowed'
                                                        : 'cursor-pointer'
                                                    }
                                                `}
                                            >
                                                <div className="flex items-start justify-between">
                                                    <div className="flex-1">
                                                        <h4 className="font-bold mb-1">{getLocalizedField(method, 'name')}</h4>
                                                        {getLocalizedField(method, 'purpose') && (
                                                            <p className="text-sm opacity-70 line-clamp-2">{getLocalizedField(method, 'purpose')}</p>
                                                        )}
                                                        {getLocalizedField(method, 'type') && (
                                                            <span className="inline-block mt-2 text-xs px-2 py-1 bg-slate-100 dark:bg-white/5 rounded">
                                                                {getMethodTypeLabel(method.type_kaz)}
                                                            </span>
                                                        )}
                                                    </div>
                                                    {selectedMethods.includes(method.id) && (
                                                        <div className="ml-3 w-6 h-6 bg-primary rounded-full flex items-center justify-center flex-shrink-0 text-white">
                                                            <Check size={14} />
                                                        </div>
                                                    )}
                                                </div>
                                            </button>
                                        ))}
                                    </div>

                                    {filteredMethods.length > 6 && (
                                        <button
                                            type="button"
                                            onClick={() => setIsMethodsModalOpen(true)}
                                            className="w-full mt-6 py-4 border-2 border-dashed border-slate-400 dark:border-white/10 rounded-2xl text-foreground/60 hover:border-primary/50 hover:text-primary transition-all font-semibold flex items-center justify-center gap-2"
                                        >
                                            <Plus size={20} />
                                            {tKmj('form.show_all_methods')} ({filteredMethods.length - 6} {tKmj('form.methods_remaining')})
                                        </button>
                                    )}

                                    {/* Methods Modal */}
                                    <AnimatePresence>
                                        {isMethodsModalOpen && (
                                            <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 md:p-8">
                                                <motion.div
                                                    initial={{ opacity: 0 }}
                                                    animate={{ opacity: 1 }}
                                                    exit={{ opacity: 0 }}
                                                    onClick={() => setIsMethodsModalOpen(false)}
                                                    className="absolute inset-0 bg-slate-900/80 backdrop-blur-sm"
                                                />
                                                <motion.div
                                                    initial={{ opacity: 0, scale: 0.95, y: 20 }}
                                                    animate={{ opacity: 1, scale: 1, y: 0 }}
                                                    exit={{ opacity: 0, scale: 0.95, y: 20 }}
                                                    className="relative w-full max-w-5xl h-[90vh] glass-card rounded-[32px] shadow-2xl overflow-hidden flex flex-col"
                                                >
                                                    {/* Modal Header */}
                                                    <div className="p-6 md:p-8 border-b border-white/10 flex items-center justify-between bg-slate-900/50">
                                                        <div>
                                                            <h3 className="text-2xl font-bold font-outfit mb-1">{tKmj('form.teaching_methods')}</h3>
                                                            <p className="text-sm opacity-60">{selectedMethods.length}/4 {tKmj('form.selected')}</p>
                                                        </div>
                                                        <button
                                                            type="button"
                                                            onClick={() => setIsMethodsModalOpen(false)}
                                                            className="p-2 hover:bg-white/10 rounded-full transition-all"
                                                        >
                                                            <X size={24} />
                                                        </button>
                                                    </div>

                                                    {/* Search in Modal */}
                                                    <div className="p-6 md:px-8 py-4 bg-slate-900/30">
                                                        <div className="relative">
                                                            <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-foreground/40" size={20} />
                                                            <input
                                                                type="text"
                                                                placeholder={tKmj('form.search_methods')}
                                                                value={searchTerm}
                                                                onChange={(e) => setSearchTerm(e.target.value)}
                                                                className="w-full pl-12 pr-4 py-4 glass-card rounded-2xl focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all font-outfit"
                                                                autoFocus
                                                            />
                                                        </div>
                                                    </div>

                                                    {/* Methods Grid in Modal */}
                                                    <div className="flex-1 overflow-y-auto p-6 md:p-8 custom-scrollbar">
                                                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                                                            {filteredMethods.map((method) => (
                                                                <button
                                                                    key={method.id}
                                                                    type="button"
                                                                    onClick={() => handleMethodToggle(method.id)}
                                                                    disabled={!selectedMethods.includes(method.id) && selectedMethods.length >= 4}
                                                                    className={`
                                                                        p-6 rounded-2xl border text-left transition-all relative group
                                                                        ${selectedMethods.includes(method.id)
                                                                            ? 'border-primary bg-primary/10 shadow-lg shadow-primary/5'
                                                                            : 'glass-card hover:border-primary/50'
                                                                        }
                                                                        ${!selectedMethods.includes(method.id) && selectedMethods.length >= 4
                                                                            ? 'opacity-50 cursor-not-allowed'
                                                                            : 'cursor-pointer hover:scale-[1.02]'
                                                                        }
                                                                    `}
                                                                >
                                                                    <div className="flex items-start justify-between gap-4">
                                                                        <div className="flex-1">
                                                                            <h4 className="font-bold mb-2 text-lg">{getLocalizedField(method, 'name')}</h4>
                                                                            {getLocalizedField(method, 'purpose') && (
                                                                                <p className="text-sm opacity-70 mb-3">{getLocalizedField(method, 'purpose')}</p>
                                                                            )}
                                                                            {getLocalizedField(method, 'type') && (
                                                                                <span className="inline-block text-xs px-3 py-1.5 bg-slate-100 dark:bg-white/10 rounded-lg font-medium">
                                                                                    {getMethodTypeLabel(method.type_kaz)}
                                                                                </span>
                                                                            )}
                                                                        </div>
                                                                        {selectedMethods.includes(method.id) && (
                                                                            <div className="w-8 h-8 bg-primary rounded-full flex items-center justify-center flex-shrink-0 text-white shadow-lg shadow-primary/20">
                                                                                <Check size={18} />
                                                                            </div>
                                                                        )}
                                                                    </div>
                                                                </button>
                                                            ))}
                                                        </div>
                                                    </div>

                                                    {/* Modal Footer */}
                                                    <div className="p-6 md:p-8 border-t border-white/10 bg-slate-900/50 flex justify-end items-center gap-4">
                                                        <button
                                                            type="button"
                                                            onClick={() => setIsMethodsModalOpen(false)}
                                                            className="px-8 py-3 bg-primary text-white rounded-xl font-bold hover:bg-primary-hover transition-all"
                                                        >
                                                            {tKmj('save')}
                                                        </button>
                                                    </div>
                                                </motion.div>
                                            </div>
                                        )}
                                    </AnimatePresence>
                                </div>

                                {/* Submit */}
                                <div className="flex items-center justify-end gap-4 pt-6 border-t border-slate-300 dark:border-white/10">
                                    <button
                                        type="button"
                                        onClick={() => setMode('list')}
                                        className="px-6 py-3 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl font-semibold transition-all"
                                    >
                                        {tKmj('form.cancel')}
                                    </button>

                                    <button
                                        type="submit"
                                        disabled={loading || selectedMethods.length === 0}
                                        className="px-8 py-3 bg-primary text-white font-bold rounded-xl hover:bg-primary-hover disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg shadow-primary/25"
                                    >
                                        {loading ? (
                                            <span className="flex items-center space-x-2">
                                                <Loader2 className="animate-spin h-5 w-5" />
                                                <span>{tKmj('status_generating')}</span>
                                            </span>
                                        ) : (
                                            `${tKmj('form.create_with_tokens')} (5 ${tCommon('tokens')})`
                                        )}
                                    </button>
                                </div>
                            </div>
                        </form>
                    </motion.div>
                )}
            </AnimatePresence>
        </div>
    );
}
