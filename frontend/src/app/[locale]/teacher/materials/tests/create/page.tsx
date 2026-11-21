'use client';

import React, { Suspense, useState } from 'react';
import { useRouter } from '@/i18n/routing';
import { useSearchParams } from 'next/navigation';
import { useTranslations, useLocale } from 'next-intl';
import { Sparkles, Loader2 } from 'lucide-react';
import api from '@/lib/api';
import PageHeader from '@/components/ui/PageHeader';
import { TEST_TYPES, isGameType } from '@/lib/testTypes';

const CREATABLE = TEST_TYPES.filter((t) => t !== 'tug_of_war');

function localeToLang(locale: string) {
    return locale === 'ru' ? 'rus' : locale === 'en' ? 'eng' : 'kaz';
}

function CreateTestForm() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const tTests = useTranslations('tests');
    const tTypes = useTranslations('test_types');
    const tCommon = useTranslations('common');
    const tKMJ = useTranslations('kmj.form');
    const locale = useLocale();

    const preset = searchParams.get('type');
    const initialType = (CREATABLE as readonly string[]).includes(preset || '') ? (preset as string) : 'standard';
    const backHref = searchParams.get('back') || (isGameType(initialType) ? '/teacher/materials/games' : '/teacher/materials/tests');

    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [formData, setFormData] = useState({
        subject: '',
        grade: 5,
        topic: '',
        language: localeToLang(locale),
        test_type: initialType,
        question_count: isGameType(initialType) ? 8 : 10,
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (loading) return; // one request per click
        setLoading(true);
        setError(null);
        try {
            await api.post('/tests', formData);
            router.push(isGameType(formData.test_type) ? '/teacher/materials/games' : backHref);
        } catch (err: any) {
            setError(err.response?.data?.error?.message || tCommon('error_occurred'));
            setLoading(false);
        }
    };

    const set = (name: string, value: string | number) => setFormData((p) => ({ ...p, [name]: value }));
    const typeTitle = tTypes(`${formData.test_type}.title`);

    return (
        <div className="max-w-3xl mx-auto">
            <PageHeader title={initialType === 'standard' ? tTests('create_new') : typeTitle}
                description={tTypes(`${formData.test_type}.desc`)} backHref={backHref} />

            <form onSubmit={handleSubmit} className="glass-card rounded-3xl p-6 md:p-8 space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                    <div>
                        <label htmlFor="t-subject" className="field-label">{tTests('form.subject')} *</label>
                        <input id="t-subject" required value={formData.subject} onChange={(e) => set('subject', e.target.value)}
                            placeholder={tKMJ('subject_placeholder')} className="field" />
                    </div>
                    <div>
                        <label htmlFor="t-grade" className="field-label">{tTests('form.grade')}</label>
                        <select id="t-grade" value={formData.grade} onChange={(e) => set('grade', parseInt(e.target.value))} className="field">
                            {Array.from({ length: 11 }, (_, i) => (
                                <option key={i + 1} value={i + 1}>{i + 1} {tCommon('grade_suffix')}</option>
                            ))}
                        </select>
                    </div>
                    <div className="md:col-span-2">
                        <label htmlFor="t-topic" className="field-label">{tTests('form.topic')} *</label>
                        <textarea id="t-topic" required rows={3} value={formData.topic} onChange={(e) => set('topic', e.target.value)}
                            placeholder={tKMJ('topic_placeholder')} className="field resize-none" />
                    </div>
                    <div>
                        <label htmlFor="t-type" className="field-label">{tTests('test_type')}</label>
                        <select id="t-type" value={formData.test_type} onChange={(e) => set('test_type', e.target.value)} className="field">
                            {CREATABLE.map((tp) => <option key={tp} value={tp}>{tTypes(`${tp}.title`)}</option>)}
                        </select>
                    </div>
                    <div>
                        <label htmlFor="t-count" className="field-label">{tTests('form.question_count')}</label>
                        <input id="t-count" type="number" min={1} max={40} value={formData.question_count}
                            onChange={(e) => set('question_count', parseInt(e.target.value) || 1)} className="field" />
                    </div>
                    <fieldset className="md:col-span-2">
                        <legend className="field-label">{tTests('form.language')}</legend>
                        <div className="flex gap-2" role="radiogroup">
                            {['kaz', 'rus', 'eng'].map((lang) => (
                                <button key={lang} type="button" role="radio" aria-checked={formData.language === lang}
                                    onClick={() => set('language', lang)}
                                    className={`flex-1 py-3 rounded-2xl font-semibold border transition-colors ${formData.language === lang
                                        ? 'bg-primary/10 border-primary text-primary'
                                        : 'border-line text-muted hover:border-primary/40'}`}>
                                    {lang === 'kaz' ? 'Қазақша' : lang === 'rus' ? 'Русский' : 'English'}
                                </button>
                            ))}
                        </div>
                    </fieldset>
                </div>

                {error && <p role="alert" className="text-sm text-red-700 bg-red-50 dark:bg-red-500/10 dark:text-red-300 p-3 rounded-xl">{error}</p>}

                <div className="flex flex-col sm:flex-row gap-3">
                    <button type="button" onClick={() => router.push(backHref)} className="btn-ghost flex-1">{tCommon('cancel')}</button>
                    <button type="submit" disabled={loading} className="btn-primary flex-[2]">
                        {loading ? <><Loader2 size={20} className="animate-spin" /> {tTests('status_generating')}</>
                            : <><Sparkles size={20} /> {isGameType(formData.test_type) ? tTypes('create_game') : tTests('form.generate_btn')}</>}
                    </button>
                </div>
            </form>

            <p className="text-sm text-muted mt-4">{tTypes('tip')}</p>
        </div>
    );
}

export default function CreateTestPage() {
    return <Suspense><CreateTestForm /></Suspense>;
}
