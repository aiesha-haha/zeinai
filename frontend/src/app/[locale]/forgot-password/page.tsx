'use client';

import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Mail, Lock, Key, ArrowRight, CheckCircle2, ChevronLeft, Eye, EyeOff, Shield, GraduationCap } from 'lucide-react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import api from '@/lib/api';
import { formatPhoneInput, formatIINInput, validatePassword } from '@/lib/utils';
import { useTranslations } from 'next-intl';

export default function ForgotPasswordPage() {
    const tAuth = useTranslations('auth');
    const tPass = useTranslations('password');
    const tVal = useTranslations('validation');
    const tReg = useTranslations('register');

    const [step, setStep] = useState(1);
    const searchParams = useSearchParams();
    const router = useRouter();

    const [isStudent, setIsStudent] = useState(false);
    const [showPassword, setShowPassword] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState(false);
    const [timer, setTimer] = useState(0);
    const [canResend, setCanResend] = useState(true);
    const [status, setStatus] = useState<{ blocked: boolean; blockedUntil?: string; nextSendAt?: string; reason?: string }>({ blocked: false });

    React.useEffect(() => {
        const type = searchParams.get('type');
        if (type === 'student') {
            setIsStudent(true);
        }
    }, [searchParams]);

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
        iin: '',
        code: '',
        newPassword: '',
        confirmPassword: '',
        maskedPhone: ''
    });

    const isEmail = (str: string) => str.includes('@');

    // Auto verify when code length is met
    React.useEffect(() => {
        const requiredLength = isStudent ? 4 : 6;
        if (formData.code.length === requiredLength && step === 2) {
            if (isStudent) {
                setStep(3); // For students, step 2 is just code entry, step 3 is new password
            } else {
                setStep(3); // General flow: step 2 verifies and moves to step 3
            }
        }
    }, [formData.code, step, isStudent]);

    const handleSendCode = async (e?: React.FormEvent) => {
        if (e) e.preventDefault();
        if (!canResend) return;

        setLoading(true);
        setError('');

        try {
            if (isStudent) {
                // Student IIN-based flow
                const res = await api.post('/auth/reset-password/student/init', { iin: formData.iin });
                setFormData(prev => ({ ...prev, maskedPhone: res.data.data.masked_phone }));
                setStep(2);
            } else {
                // General Email/Phone flow
                const payload = isEmail(formData.identifier)
                    ? { email: formData.identifier, phone: '', type: 'password_reset' }
                    : { email: '', phone: formData.identifier, type: 'password_reset' };

                await api.post('/verification/send-code', payload);
                await checkStatus(formData.identifier);
                setStep(2);
            }
        } catch (err: any) {
            const errorMsg = err.response?.data?.error?.message;
            if (err.response?.status === 429) {
                await checkStatus(formData.identifier);
            }
            setError(errorMsg || tPass('error_initiating'));
        } finally {
            setLoading(false);
        }
    };

    const handleVerifyAndReset = async (e: React.FormEvent) => {
        e.preventDefault();
        if (formData.newPassword !== formData.confirmPassword) {
            setError(tVal('passwords_match'));
            return;
        }

        setLoading(true);
        setError('');

        try {
            if (isStudent) {
                // Student confirmation
                await api.post('/auth/reset-password/student/confirm', {
                    iin: formData.iin,
                    code: formData.code,
                    new_password: formData.newPassword
                });
            } else {
                // General confirmation
                const payload = isEmail(formData.identifier)
                    ? { email: formData.identifier, phone: '', code: formData.code, new_password: formData.newPassword }
                    : { email: '', phone: formData.identifier, code: formData.code, new_password: formData.newPassword };

                await api.post('/verification/reset-password', payload);
            }
            setSuccess(true);
            setTimeout(() => router.push('/login'), 3000);
        } catch (err: any) {
            setError(err.response?.data?.error?.message || 'Verification failed');
        } finally {
            setLoading(false);
        }
    };

    if (success) {
        return (
            <main className="min-h-screen flex items-center justify-center p-6">
                <motion.div
                    initial={{ opacity: 0, scale: 0.9 }}
                    animate={{ opacity: 1, scale: 1 }}
                    className="glass-card w-full max-w-md p-10 rounded-3xl text-center"
                >
                    <div className="w-20 h-20 bg-green-500/20 rounded-full flex items-center justify-center mx-auto mb-6">
                        <CheckCircle2 className="w-10 h-10 text-green-500" />
                    </div>
                    <h1 className="text-3xl font-bold mb-4">{tPass('success_title')}</h1>
                    <p className="text-foreground/60 mb-8">{tPass('success_msg')}</p>
                    <Link href="/login" className="text-primary font-bold hover:underline">{tPass('click_here')}</Link>
                </motion.div>
            </main>
        );
    }

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
                className="glass-card w-full max-w-md p-10 rounded-3xl relative"
            >
                {step === 1 && (
                    <Link
                        href="/login"
                        className="absolute top-6 left-6 p-2 text-foreground/50 hover:text-foreground hover:bg-white/5 rounded-full transition-colors"
                    >
                        <ChevronLeft className="w-6 h-6" />
                    </Link>
                )}
                <div className="text-center mb-10">
                    <h1 className="text-3xl font-bold font-outfit mb-2">{tPass('reset_title')}</h1>
                    <p className="text-foreground/60">
                        {step === 1 && tPass('recover_access')}
                        {step === 2 && (isStudent ? tPass('enter_hidden_digits', { phone: formData.maskedPhone }) : tPass('enter_code'))}
                        {step === 3 && tPass('create_new')}
                    </p>
                </div>

                {step === 1 && (
                    <div className="flex bg-white/5 p-1 rounded-2xl mb-8">
                        <button
                            onClick={() => setIsStudent(false)}
                            className={`flex-1 py-3 rounded-xl text-sm font-semibold transition-all ${!isStudent ? 'bg-primary text-white shadow-lg' : 'hover:bg-white/5'}`}
                        >
                            {tPass('teacher_admin')}
                        </button>
                        <button
                            onClick={() => setIsStudent(true)}
                            className={`flex-1 py-3 rounded-xl text-sm font-semibold transition-all ${isStudent ? 'bg-primary text-white shadow-lg' : 'hover:bg-white/5'}`}
                        >
                            {tAuth('student_role')}
                        </button>
                    </div>
                )}

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
                            {!isStudent ? (
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
                                            onChange={(e) => {
                                                const val = e.target.value;
                                                if (!val.includes('@') && (val.startsWith('+') || /^\d/.test(val))) {
                                                    setFormData({ ...formData, identifier: formatPhoneInput(val) });
                                                } else {
                                                    setFormData({ ...formData, identifier: val });
                                                }
                                            }}
                                            className="w-full bg-white/5 border border-white/10 rounded-2xl py-4 pl-12 pr-4 outline-none focus:border-primary/50 transition-all font-outfit"
                                            placeholder="name@example.com or +7..."
                                        />
                                    </div>
                                </div>
                            ) : (
                                <div className="space-y-2">
                                    <label className="text-sm font-medium ml-2">{tAuth('iin_label')}</label>
                                    <div className="relative">
                                        <GraduationCap className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground/40" />
                                        <input
                                            type="text"
                                            required
                                            autoComplete="off"
                                            name="iin"
                                            maxLength={12}
                                            value={formData.iin}
                                            onChange={(e) => setFormData({ ...formData, iin: formatIINInput(e.target.value) })}
                                            className="w-full bg-white/5 border border-white/10 rounded-2xl py-4 pl-12 pr-4 outline-none focus:border-primary/50 transition-all font-outfit"
                                            placeholder="000000000000"
                                        />
                                    </div>
                                </div>
                            )}
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
                                {loading ? tPass('checking') : (status.blocked ? tReg('blocked_btn') : tReg('continue'))}
                                <ArrowRight className="w-5 h-5" />
                            </button>
                        </motion.form>
                    )}

                    {step === 2 && (
                        <motion.form
                            key="step2"
                            initial={{ opacity: 0, x: -20 }}
                            animate={{ opacity: 1, x: 0 }}
                            onSubmit={(e) => { e.preventDefault(); setStep(3); }}
                            className="space-y-6"
                        >
                            <div className="space-y-2 text-center">
                                <label className="text-sm font-medium">
                                    {isStudent ? tPass('enter_hidden_digits', { phone: '****' }) : tReg('enter_code_label')}
                                </label>
                                <div className="relative">
                                    <Key className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground/40" />
                                    <input
                                        type="text"
                                        required
                                        maxLength={isStudent ? 4 : 6}
                                        value={formData.code}
                                        onChange={(e) => setFormData({ ...formData, code: e.target.value })}
                                        className="w-full bg-white/5 border border-white/10 rounded-2xl py-6 pl-12 pr-4 outline-none focus:border-primary/50 transition-all text-center text-3xl font-bold tracking-[0.5em] font-outfit"
                                        placeholder={isStudent ? "0000" : "000000"}
                                    />
                                </div>
                                {isStudent && (
                                    <p className="text-xs text-foreground/40 mt-2 italic">
                                        {tPass('check_phone', { phone: formData.maskedPhone })}
                                    </p>
                                )}
                            </div>

                            {!isStudent && (
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
                            )}

                            <div className="flex gap-4">
                                <button
                                    type="button"
                                    onClick={() => setStep(1)}
                                    className="px-6 py-4 bg-white/5 hover:bg-white/10 rounded-2xl transition-all"
                                >
                                    <ChevronLeft className="w-6 h-6" />
                                </button>
                                <button
                                    className="flex-1 bg-primary text-white py-4 rounded-2xl font-bold text-lg hover:bg-primary-hover transition-all flex items-center justify-center gap-2 shadow-xl shadow-primary/20"
                                >
                                    {tReg('verify_code')}
                                </button>
                            </div>
                        </motion.form>
                    )}

                    {step === 3 && (
                        <motion.form
                            key="step3"
                            initial={{ opacity: 0, x: -20 }}
                            animate={{ opacity: 1, x: 0 }}
                            onSubmit={handleVerifyAndReset}
                            className="space-y-6"
                        >
                            <div className="space-y-2">
                                <label className="text-sm font-medium ml-2">{tPass('new_password')}</label>
                                <div className="relative">
                                    <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground/40" />
                                    <input
                                        type={showPassword ? "text" : "password"}
                                        required
                                        value={formData.newPassword}
                                        onChange={(e) => setFormData({ ...formData, newPassword: e.target.value })}
                                        className="w-full bg-white/5 border border-white/10 rounded-2xl py-4 pl-12 pr-12 outline-none focus:border-primary/50 transition-all text-sm font-outfit"
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
                            </div>

                            <div className="space-y-2">
                                <label className="text-sm font-medium ml-2">{tPass('confirm_password')}</label>
                                <div className="relative">
                                    <Shield className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground/40" />
                                    <input
                                        type="password"
                                        required
                                        value={formData.confirmPassword}
                                        onChange={(e) => setFormData({ ...formData, confirmPassword: e.target.value })}
                                        className="w-full bg-white/5 border border-white/10 rounded-2xl py-4 pl-12 pr-4 outline-none focus:border-primary/50 transition-all text-sm font-outfit"
                                        placeholder="••••••••"
                                    />
                                </div>
                            </div>

                            {error && <p className="text-red-500 text-xs ml-2">{error}</p>}

                            <div className="flex gap-4">
                                <button
                                    type="button"
                                    onClick={() => setStep(2)}
                                    className="px-6 py-4 bg-white/5 hover:bg-white/10 rounded-2xl transition-all"
                                >
                                    <ChevronLeft className="w-6 h-6" />
                                </button>
                                <button
                                    disabled={loading || !validatePassword(formData.newPassword).isValid}
                                    className="flex-1 bg-primary text-white py-4 rounded-2xl font-bold text-lg hover:bg-primary-hover transition-all flex items-center justify-center gap-2 shadow-xl shadow-primary/20"
                                >
                                    {loading ? tPass('updating') : tPass('set_password')}
                                </button>
                            </div>
                        </motion.form>
                    )}
                </AnimatePresence>

                <p className="text-center mt-10 text-foreground/50 text-sm">
                    {tPass('remember_password')} <Link href="/login" className="text-primary font-semibold hover:underline">{tAuth('sign_in_link')}</Link>
                </p>
            </motion.div>
        </main >
    );
}

