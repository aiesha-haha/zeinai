'use client';

import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { User, Shield, Calendar, Users, Building, CheckCircle2, Eye, EyeOff, Check, X } from 'lucide-react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import api from '@/lib/api';
import { formatIINInput } from '@/lib/utils';
import cookie from 'cookie-cutter';
import { useTranslations } from 'next-intl';

export default function GoogleCompleteProfilePage() {
    const tReg = useTranslations('register');
    const tAuth = useTranslations('auth');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const router = useRouter();
    const searchParams = useSearchParams();

    const [formData, setFormData] = useState({
        google_id: '',
        email: '',
        first_name: '',
        last_name: '',
        picture: '',
        gender: 'male',
        birth_year: 1990,
        school_name: '',
        role: 'teacher'
    });

    useEffect(() => {
        // Get data from URL params
        const googleId = searchParams.get('google_id');
        const email = searchParams.get('email');
        const firstName = searchParams.get('first_name');
        const lastName = searchParams.get('last_name');
        const picture = searchParams.get('picture');

        if (!googleId || !email) {
            router.push('/login');
            return;
        }

        setFormData(prev => ({
            ...prev,
            google_id: googleId,
            email: email,
            first_name: firstName || '',
            last_name: lastName || '',
            picture: picture || ''
        }));
    }, [searchParams, router]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError('');

        try {
            const res = await api.post('/auth/google/register/complete', formData);
            const { access_token, refresh_token, user } = res.data.data;

            // Set localStorage instead of cookies to match api.ts
            if (typeof window !== 'undefined') {
                localStorage.setItem('access_token', access_token);
                localStorage.setItem('refresh_token', refresh_token);
                localStorage.setItem('user_role', user.role);
            }

            // Redirect based on role
            if (user.role === 'admin') router.push('/admin');
            else if (user.role === 'teacher') router.push('/teacher');
            else router.push('/student');

        } catch (err: any) {
            const errorMsg = err.response?.data?.error?.message;
            setError(errorMsg || tReg('registration_failed'));
        } finally {
            setLoading(false);
        }
    };

    return (
        <main className="min-h-screen flex items-center justify-center p-6 py-20">
            <div className="absolute top-8 left-8">
                <Link href="/" className="text-2xl font-bold tracking-tighter text-gradient">
                    ZeinAI
                </Link>
            </div>

            <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                className="glass-card w-full max-w-xl p-10 rounded-3xl relative"
            >
                <div className="text-center mb-10">
                    {formData.picture && (
                        <div className="flex justify-center mb-4">
                            <img
                                src={formData.picture}
                                alt="Profile"
                                className="w-20 h-20 rounded-full border-4 border-primary/20"
                            />
                        </div>
                    )}
                    <h1 className="text-3xl font-bold font-outfit mb-2">{tReg('complete_title')}</h1>
                    <p className="text-foreground/60">{tReg('google_complete_subtitle')}</p>
                </div>

                <form onSubmit={handleSubmit} className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <fieldset className="md:col-span-2 space-y-2">
                        <legend className="text-sm font-medium ml-2 mb-2">{tReg('choose_role')}</legend>
                        <div className="grid grid-cols-2 gap-3" role="radiogroup">
                            {(['teacher', 'student'] as const).map((r) => (
                                <button
                                    key={r}
                                    type="button"
                                    role="radio"
                                    aria-checked={formData.role === r}
                                    onClick={() => setFormData({ ...formData, role: r })}
                                    className={`p-4 rounded-2xl border-2 font-bold text-sm transition-all focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-primary/30 ${formData.role === r ? 'border-primary bg-primary/10 text-primary' : 'border-slate-200 dark:border-white/10 text-foreground/70 hover:border-primary/40'}`}
                                >
                                    {tAuth(r === 'teacher' ? 'teacher_role' : 'student_role')}
                                </button>
                            ))}
                        </div>
                    </fieldset>
                    <div className="md:col-span-2 space-y-2">
                        <label className="text-sm font-medium ml-2">{tAuth('email_label')}</label>
                        <div className="relative">
                            <input
                                type="email"
                                disabled
                                value={formData.email}
                                className="w-full bg-slate-100 dark:bg-white/10 border border-slate-200 dark:border-white/10 rounded-2xl py-4 px-4 outline-none text-sm text-foreground/50 cursor-not-allowed"
                            />
                        </div>
                    </div>

                    <div className="space-y-2">
                        <label className="text-sm font-medium ml-2">{tReg('first_name')}</label>
                        <div className="relative">
                            <User className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground/40" />
                            <input
                                type="text"
                                required
                                value={formData.first_name}
                                onChange={(e) => setFormData({ ...formData, first_name: e.target.value })}
                                className="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl py-4 pl-12 pr-4 outline-none focus:border-primary/50 transition-all text-sm text-foreground"
                                placeholder="First Name"
                            />
                        </div>
                    </div>

                    <div className="space-y-2">
                        <label className="text-sm font-medium ml-2">{tReg('last_name')}</label>
                        <div className="relative">
                            <User className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground/40" />
                            <input
                                type="text"
                                required
                                value={formData.last_name}
                                onChange={(e) => setFormData({ ...formData, last_name: e.target.value })}
                                className="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl py-4 pl-12 pr-4 outline-none focus:border-primary/50 transition-all text-sm text-foreground"
                                placeholder="Last Name"
                            />
                        </div>
                    </div>

                    <div className="space-y-2">
                        <label className="text-sm font-medium ml-2">{tReg('birth_year')}</label>
                        <div className="relative">
                            <Calendar className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground/40" />
                            <input
                                type="number"
                                required
                                min="1950"
                                max="2015"
                                value={formData.birth_year}
                                onChange={(e) => setFormData({ ...formData, birth_year: parseInt(e.target.value) })}
                                className="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl py-4 pl-12 pr-4 outline-none focus:border-primary/50 transition-all text-sm text-foreground"
                            />
                        </div>
                    </div>

                    <div className="space-y-2">
                        <label className="text-sm font-medium ml-2">{tReg('gender')}</label>
                        <div className="relative">
                            <Users className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground/40" />
                            <select
                                value={formData.gender}
                                onChange={(e) => setFormData({ ...formData, gender: e.target.value })}
                                className="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl py-4 pl-12 pr-4 outline-none focus:border-primary/50 transition-all text-sm appearance-none text-foreground dark:text-white"
                            >
                                <option value="male" className="bg-background text-foreground">{tReg('male')}</option>
                                <option value="female" className="bg-background text-foreground">{tReg('female')}</option>
                            </select>
                        </div>
                    </div>


                    <div className="md:col-span-2 space-y-2">
                        <label className="text-sm font-medium ml-2">{tReg('school_name')}</label>
                        <div className="relative">
                            <Building className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground/40" />
                            <input
                                type="text"
                                value={formData.school_name}
                                onChange={(e) => setFormData({ ...formData, school_name: e.target.value })}
                                className="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl py-4 pl-12 pr-4 outline-none focus:border-primary/50 transition-all text-sm text-foreground"
                                placeholder="Science High School"
                            />
                        </div>
                    </div>

                    {error && <p className="md:col-span-2 text-red-500 text-xs ml-2">{error}</p>}

                    <button
                        disabled={loading}
                        className="md:col-span-2 bg-primary text-white py-4 rounded-2xl font-bold text-lg hover:bg-primary-hover transition-all flex items-center justify-center gap-2 shadow-xl shadow-primary/20 mt-4 disabled:opacity-50"
                    >
                        {loading ? tReg('creating_account') : tReg('complete_profile')}
                        <CheckCircle2 className="w-5 h-5" />
                    </button>
                </form>

                <p className="text-center mt-10 text-foreground/50 text-sm">
                    {tAuth('already_have_account')} <Link href="/login" className="text-primary font-semibold hover:underline">{tAuth('sign_in_link')}</Link>
                </p>
            </motion.div>
        </main>
    );
}
