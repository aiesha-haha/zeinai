'use client';

import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { UserPlus, Mail, Lock, Phone, Building, User, Shield, ArrowRight, CheckCircle2, Key, ChevronLeft, Eye, EyeOff, Check, X, Calendar, Users } from 'lucide-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import api from '@/lib/api';
import { formatEmailOrPhoneInput, formatIINInput, validatePassword } from '@/lib/utils';
import cookie from 'cookie-cutter';
import { useTranslations } from 'next-intl';

export default function RegisterPage() {
    const tAuth = useTranslations('auth');
    const tReg = useTranslations('register');
    const tVal = useTranslations('validation');

    const [step, setStep] = useState(1); // 1: Identifier, 2: OTP, 3: Full Profile
    const [showPassword, setShowPassword] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [timer, setTimer] = useState(0);
    const [canResend, setCanResend] = useState(true);
    const [status, setStatus] = useState<{ blocked: boolean; blockedUntil?: string; nextSendAt?: string; reason?: string }>({ blocked: false });
    const router = useRouter();

    React.useEffect(() => {
        const token = localStorage.getItem('access_token') || cookie.get('access_token');
        const role = localStorage.getItem('user_role') || cookie.get('user_role');
        if (token && role) {
            if (role === 'admin') router.push('/admin');
            else if (role === 'teacher') router.push('/teacher');
            else router.push('/student');
        }
    }, [router]);

    // Timer effect
    React.useEffect(() => {
        if (timer > 0) {
            const interval = setInterval(() => {
                setTimer((prev) => prev - 1);
            }, 1000);
            return () => clearInterval(interval);
        } else {
            setCanResend(true);
        }
    }, [timer]);

    const checkStatus = async (identifier: string) => {
        try {
            const res = await api.get(`/verification/status?identifier=${identifier}`);
            const data = res.data.data; // Fix: Use res.data.data
            setStatus(data);
            if (!data.can_send && data.next_send_at) {
                const nextAt = new Date(data.next_send_at).getTime();
                const now = new Date().getTime();
                const diff = Math.ceil((nextAt - now) / 1000);
                if (diff > 0) {
                    setTimer(diff);
                    setCanResend(false);
                }
            } else if (!data.can_send && data.blocked) {
                setCanResend(false);
            } else {
                setCanResend(true);
                setTimer(0);
            }
        } catch (err) {
            console.error("Failed to check status", err);
        }
    };

    const [formData, setFormData] = useState({
        identifier: '', // email or phone
        firstName: '',
        lastName: '',
        birthYear: 1990,
        gender: 'male',
        email: '',
        phone: '',
        password: '',
        school_name: '',
        code: '',
        role: 'teacher' as 'teacher' | 'student'
    });

    const isEmail = (str: string) => str.includes('@');

    // Auto verify when code reaches 6 digits
    React.useEffect(() => {
        if (formData.code.length === 6 && step === 2) {
            handleVerifyCode();
        }
    }, [formData.code]);

    const handleSendCode = async (e?: React.FormEvent) => {
        if (e) e.preventDefault();
        if (!canResend) return;

        setLoading(true);
        setError('');

        try {
            const payload = isEmail(formData.identifier)
                ? { email: formData.identifier, phone: '', type: 'registration' }
                : { email: '', phone: formData.identifier, type: 'registration' };

            await api.post('/verification/send-code', payload);

            if (isEmail(formData.identifier)) {
                setFormData(prev => ({ ...prev, email: formData.identifier, phone: '' }));
            } else {
                setFormData(prev => ({ ...prev, phone: formData.identifier, email: '' }));
            }

            await checkStatus(formData.identifier);
            setStep(2);
        } catch (err: any) {
            const errorMsg = err.response?.data?.error?.message;
            const errorRes = err.response?.data?.error;

            if (errorRes?.code === 'RATE_LIMIT_EXCEEDED' || err.response?.status === 429) {
                await checkStatus(formData.identifier);
            }

            if (errorRes?.code === 'SMS_UNAVAILABLE') {
                setError(tReg('sms_unavailable'));
            } else if (errorMsg === 'user already exists' || errorRes?.code === 'USER_EXISTS' || err.response?.status === 409) {
                const identifierType = isEmail(formData.identifier) ? 'email' : 'phone';
                router.push(`/login?error=${identifierType}_exists&identifier=${encodeURIComponent(formData.identifier)}`);
            } else {
                setError(errorMsg || tReg('error_sending'));
            }
        } finally {
            setLoading(false);
        }
    };

    const handleVerifyCode = async (e?: React.FormEvent) => {
        if (e) e.preventDefault();
        setLoading(true);
        setError('');

        try {
            await api.post('/verification/verify-code', {
                email: formData.email,
                phone: formData.phone,
                code: formData.code
            });
            setStep(3);
        } catch (err: any) {
            setError(err.response?.data?.error?.message === 'invalid verification code' ? tAuth('errors.invalid_code') : (err.response?.data?.error?.message || tReg('error_verifying')));
        } finally {
            setLoading(false);
        }
    };

    const handleFinalRegister = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError('');

        try {
            const res = await api.post('/auth/register', {
                first_name: formData.firstName,
                last_name: formData.lastName,
                birth_year: Number(formData.birthYear),
                gender: formData.gender,
                email: formData.email,
                phone: formData.phone,
                password: formData.password,
                school_name: formData.school_name,
                role: formData.role
            });
            
            const { access_token, refresh_token, user } = res.data.data;

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
            const errorRes = err.response?.data?.error;
            const errorMsg = err.response?.data?.error?.message;
            if (errorRes?.code === 'USER_EXISTS' || err.response?.status === 409 || errorMsg === 'user already exists') {
                const identifierType = formData.email ? 'email' : 'phone';
                const identifier = formData.email || formData.phone;
                router.push(`/login?error=${identifierType}_exists&identifier=${encodeURIComponent(identifier)}`);
            } else {
                setError(errorMsg || tReg('registration_failed'));
            }
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
                {step === 1 && (
                    <Link
                        href="/"
                        className="absolute top-6 left-6 p-2 text-foreground/50 hover:text-foreground hover:bg-white/5 rounded-full transition-colors"
                    >
                        <ChevronLeft className="w-6 h-6" />
                    </Link>
                )}
                <div className="text-center mb-10">
                    <div className="flex justify-center mb-6">
                        <div className="flex items-center gap-2">
                            {[1, 2, 3].map((s) => (
                                <div
                                    key={s}
                                    className={`h-1.5 rounded-full transition-all duration-500 ${step === s ? 'w-8 bg-primary' : (step > s ? 'w-4 bg-primary/40' : 'w-4 bg-white/10')
                                        }`}
                                />
                            ))}
                        </div>
                    </div>
                    <h1 className="text-3xl font-bold font-outfit mb-2">
                        {step === 1 && tReg('join_title')}
                        {step === 2 && tReg('verify_title')}
                        {step === 3 && tReg('complete_title')}
                    </h1>
                    <p className="text-foreground/60">
                        {step === 1 && tReg('step1_subtitle')}
                        {step === 2 && tReg('step2_subtitle', { identifier: formData.identifier })}
                        {step === 3 && tReg('step3_subtitle')}
                    </p>
                </div>

                <AnimatePresence mode="wait">
                    {step === 1 && (
                        <motion.form
                            key="step1"
                            initial={{ opacity: 0, x: -20 }}
                            animate={{ opacity: 1, x: 0 }}
                            exit={{ opacity: 0, x: 20 }}
                            onSubmit={handleSendCode}
                            className="space-y-6"
                        >
                            <fieldset className="space-y-2">
                                <legend className="text-sm font-medium ml-2 mb-2">{tReg('choose_role')}</legend>
                                <div className="grid grid-cols-2 gap-3" role="radiogroup">
                                    {(['teacher', 'student'] as const).map((r) => (
                                        <button
                                            key={r}
                                            type="button"
                                            role="radio"
                                            aria-checked={formData.role === r}
                                            onClick={() => setFormData({ ...formData, role: r })}
                                            className={`flex flex-col items-center gap-2 p-4 rounded-2xl border-2 transition-all focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-primary/30 ${formData.role === r ? 'border-primary bg-primary/10 text-primary' : 'border-slate-200 dark:border-white/10 text-foreground/70 hover:border-primary/40'}`}
                                        >
                                            {r === 'teacher' ? <Users className="w-6 h-6" /> : <User className="w-6 h-6" />}
                                            <span className="font-bold text-sm">{tAuth(r === 'teacher' ? 'teacher_role' : 'student_role')}</span>
                                            <span className="text-xs text-foreground/60 text-center leading-snug">{tReg(r === 'teacher' ? 'role_teacher_hint' : 'role_student_hint')}</span>
                                        </button>
                                    ))}
                                </div>
                            </fieldset>
                            <div className="space-y-2">
                                <label className="text-sm font-medium ml-2">{tAuth('email_or_phone_label')}</label>
                                <div className="relative">
                                    <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground/40" />
                                    <input
                                        type="text"
                                        required
                                        autoComplete="email"
                                        name="identifier"
                                        value={formData.identifier}
                                        onChange={(e) => setFormData({ ...formData, identifier: formatEmailOrPhoneInput(e.target.value) })}
                                        className="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl py-4 pl-12 pr-4 outline-none focus:border-primary/50 transition-all font-outfit text-foreground"
                                        placeholder="name@example.com or +7..."
                                    />
                                </div>
                            </div>
                            {error && <p className="text-red-500 text-xs ml-2 font-medium">{error}</p>}
                            {status.blocked && (
                                <div className="bg-red-500/10 border border-red-500/20 p-4 rounded-2xl text-red-500 text-sm font-medium">
                                    {tAuth('errors.blocked_label')}{status.reason || ''}
                                </div>
                            )}
                            <button
                                disabled={loading || (status.blocked && !status.blockedUntil)}
                                className="w-full bg-primary text-white py-4 rounded-2xl font-bold text-lg hover:bg-primary-hover transition-all flex items-center justify-center gap-2 shadow-xl shadow-primary/20 disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                {loading ? tReg('sending') : (status.blocked ? tReg('blocked_btn') : tReg('continue'))}
                                <ArrowRight className="w-5 h-5" />
                            </button>

                            <div className="relative my-6">
                                <div className="absolute inset-0 flex items-center">
                                    <div className="w-full border-t border-slate-200 dark:border-white/10"></div>
                                </div>
                                <div className="relative flex justify-center text-xs uppercase">
                                    <span className="bg-background px-4 text-foreground/40 font-medium">{tAuth('or_continue_with')}</span>
                                </div>
                            </div>

                            <button
                                type="button"
                                onClick={() => window.location.href = `${(process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000/api/v1').replace('/api/v1', '')}/api/v1/auth/google/login`}
                                className="w-full bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 text-foreground dark:text-white py-4 rounded-2xl font-semibold hover:bg-slate-50 dark:hover:bg-white/10 transition-all flex items-center justify-center gap-3 overflow-hidden group"
                            >
                                <svg className="w-5 h-5 group-hover:scale-110 transition-transform" viewBox="0 0 24 24">
                                    <path
                                        fill="currentColor"
                                        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                                    />
                                    <path
                                        fill="currentColor"
                                        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                                    />
                                    <path
                                        fill="currentColor"
                                        d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z"
                                    />
                                    <path
                                        fill="currentColor"
                                        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                                    />
                                </svg>
                                Google
                            </button>
                        </motion.form>
                    )}

                    {step === 2 && (
                        <motion.form
                            key="step2"
                            initial={{ opacity: 0, x: -20 }}
                            animate={{ opacity: 1, x: 0 }}
                            exit={{ opacity: 0, x: 20 }}
                            onSubmit={handleVerifyCode}
                            className="space-y-6"
                        >
                            <div className="space-y-2 text-center">
                                <label className="text-sm font-medium">{tReg('enter_code_label')}</label>
                                <div className="relative">
                                    <Key className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground/40" />
                                    <input
                                        type="text"
                                        required
                                        autoComplete="off"
                                        name="otp_code"
                                        maxLength={6}
                                        value={formData.code}
                                        onChange={(e) => setFormData({ ...formData, code: e.target.value.replace(/\D/g, '') })}
                                        className="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl py-6 pl-12 pr-4 outline-none focus:border-primary/50 transition-all text-center text-3xl font-bold tracking-[0.5em] font-outfit text-foreground"
                                        placeholder="000000"
                                    />
                                </div>
                            </div>
                            {error && <p className="text-red-500 text-xs text-center font-medium">{error}</p>}

                            <div className="text-center font-outfit">
                                {!canResend ? (
                                    <p className="text-sm text-foreground/50">
                                        {status.blocked ? (
                                            <span className="text-red-500 font-medium">{tAuth('errors.blocked_label')}{status.reason || ''}</span>
                                        ) : (
                                            <>{tAuth('errors.resend_in')}<span className="text-primary font-bold">{Math.floor(timer / 60)}:{(timer % 60).toString().padStart(2, '0')}</span></>
                                        )}
                                    </p>
                                ) : (
                                    <button
                                        type="button"
                                        onClick={() => handleSendCode()}
                                        className="text-primary text-sm font-semibold hover:underline"
                                    >
                                        {tAuth('errors.resend_now')}
                                    </button>
                                )}
                            </div>

                            <div className="flex gap-4">
                                <button
                                    type="button"
                                    onClick={() => setStep(1)}
                                    className="px-6 py-4 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 rounded-2xl transition-all"
                                >
                                    <ChevronLeft className="w-6 h-6" />
                                </button>
                                <button
                                    disabled={loading}
                                    className="flex-1 bg-primary text-white py-4 rounded-2xl font-bold text-lg hover:bg-primary-hover transition-all flex items-center justify-center gap-2 shadow-xl shadow-primary/20"
                                >
                                    {loading ? tReg('verifying') : tReg('verify_code')}
                                </button>
                            </div>
                        </motion.form>
                    )}

                    {step === 3 && (
                        <motion.form
                            key="step3"
                            initial={{ opacity: 0, x: -20 }}
                            animate={{ opacity: 1, x: 0 }}
                            onSubmit={handleFinalRegister}
                            className="grid grid-cols-1 md:grid-cols-2 gap-6"
                        >
                            <div className="space-y-2">
                                <label className="text-sm font-medium ml-2">{tReg('first_name')}</label>
                                <div className="relative">
                                    <User className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground/40" />
                                    <input
                                        type="text"
                                        required
                                        autoComplete="given-name"
                                        name="firstName"
                                        value={formData.firstName}
                                        onChange={(e) => setFormData({ ...formData, firstName: e.target.value })}
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
                                        autoComplete="family-name"
                                        name="lastName"
                                        value={formData.lastName}
                                        onChange={(e) => setFormData({ ...formData, lastName: e.target.value })}
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
                                        autoComplete="bday-year"
                                        name="birthYear"
                                        value={formData.birthYear}
                                        onChange={(e) => setFormData({ ...formData, birthYear: parseInt(e.target.value) })}
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
                                        autoComplete="sex"
                                        name="gender"
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
                                        autoComplete="organization"
                                        name="school_name"
                                        value={formData.school_name}
                                        onChange={(e) => setFormData({ ...formData, school_name: e.target.value })}
                                        className="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl py-4 pl-12 pr-4 outline-none focus:border-primary/50 transition-all text-sm text-foreground"
                                        placeholder="Science High School"
                                    />
                                </div>
                            </div>

                            <div className="md:col-span-2 space-y-2">
                                <label className="text-sm font-medium ml-2">{tReg('create_password')}</label>
                                <div className="relative">
                                    <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground/40" />
                                    <input
                                        type={showPassword ? "text" : "password"}
                                        required
                                        autoComplete="new-password"
                                        name="password"
                                        value={formData.password}
                                        onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                                        className="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl py-4 pl-12 pr-12 outline-none focus:border-primary/50 transition-all text-sm text-foreground"
                                        placeholder="••••••••"
                                    />
                                    <button
                                        type="button"
                                        onClick={() => setShowPassword(!showPassword)}
                                        className="absolute right-4 top-1/2 -translate-y-1/2 text-foreground/40 hover:text-foreground/60 transition-colors"
                                    >
                                        {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                                    </button>
                                </div>
                                <div className="grid grid-cols-2 gap-2 mt-3 px-2">
                                    <PasswordCheck label={tVal('min_8_chars')} met={validatePassword(formData.password).isLongEnough} />
                                    <PasswordCheck label={tVal('uppercase')} met={validatePassword(formData.password).hasUpper} />
                                    <PasswordCheck label={tVal('lowercase')} met={validatePassword(formData.password).hasLower} />
                                    <PasswordCheck label={tVal('digit')} met={validatePassword(formData.password).hasNumber} />
                                </div>
                            </div>

                            {error && <p className="md:col-span-2 text-red-500 text-xs ml-2">{error}</p>}

                            <button
                                disabled={loading || !validatePassword(formData.password).isValid}
                                className="md:col-span-2 bg-primary text-white py-4 rounded-2xl font-bold text-lg hover:bg-primary-hover transition-all flex items-center justify-center gap-2 shadow-xl shadow-primary/20 mt-4"
                            >
                                {loading ? tReg('creating_account') : tReg('complete_profile')}
                                <CheckCircle2 className="w-5 h-5" />
                            </button>
                        </motion.form>
                    )}
                </AnimatePresence>

                <p className="text-center mt-10 text-foreground/50 text-sm">
                    {tAuth('already_have_account')} <Link href="/login" className="text-primary font-semibold hover:underline">{tAuth('sign_in_link')}</Link>
                </p>
            </motion.div>
        </main>
    );
}

function PasswordCheck({ label, met }: { label: string, met: boolean }) {
    return (
        <div className={`flex items-center gap-1.5 text-[10px] font-medium transition-colors ${met ? 'text-green-500' : 'text-foreground/30'}`}>
            {met ? <Check className="w-3 h-3" /> : <X className="w-3 h-3" />}
            {label}
        </div>
    );
}
