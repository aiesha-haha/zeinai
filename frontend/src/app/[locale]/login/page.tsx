'use client';

import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { LogIn, Mail, Lock, Phone, GraduationCap, Eye, EyeOff, ChevronLeft } from 'lucide-react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import api from '@/lib/api';
import { useTranslations } from 'next-intl';

import { formatPhoneInput, formatIINInput } from '@/lib/utils';

export default function LoginPage() {
    const tAuth = useTranslations('auth');
    const [isStudent, setIsStudent] = useState(false);
    const [showPassword, setShowPassword] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const router = useRouter();
    const searchParams = useSearchParams();

    React.useEffect(() => {
        const errorType = searchParams.get('error');
        const identifier = searchParams.get('identifier');

        if (errorType === 'email_exists') {
            setError(tAuth('errors.email_exists'));
        } else if (errorType === 'phone_exists') {
            setError(tAuth('errors.phone_exists'));
        } else if (searchParams.get('session_expired')) {
            setError(tAuth('errors.session_expired'));
        } else if (searchParams.get('registered')) {
            setError(tAuth('errors.reg_success'));
        }

        // Pre-fill email/phone if provided from registration
        if (identifier) {
            setFormData(prev => ({ ...prev, email: identifier }));
        }

        const token = searchParams.get('token');
        const refresh = searchParams.get('refresh');
        const roleParam = searchParams.get('role');

        if (token && refresh && roleParam && typeof window !== 'undefined') {
            localStorage.setItem('access_token', token);
            localStorage.setItem('refresh_token', refresh);
            localStorage.setItem('user_role', roleParam);

            if (roleParam === 'admin') router.push('/admin');
            else if (roleParam === 'teacher') router.push('/teacher');
            else router.push('/student');
            return;
        }

        if (typeof window !== 'undefined') {
            const existingToken = localStorage.getItem('access_token');
            const role = localStorage.getItem('user_role');
            if (existingToken && role) {
                if (role === 'admin') router.push('/admin');
                else if (role === 'teacher') router.push('/teacher');
                else router.push('/student');
            }
        }
    }, [router, searchParams, tAuth]);

    const [formData, setFormData] = useState({
        email: '',
        phone: '',
        password: '',
        iin: ''
    });

    const handleLogin = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError('');

        try {
            // Students can sign in with their IIN (12 digits) or, if they
            // self-registered, with the email / phone used at registration.
            const studentId = formData.iin.trim();
            const studentByIIN = isStudent && /^\d{12}$/.test(studentId);
            const identifier = isStudent ? studentId : formData.email;
            const endpoint = studentByIIN ? '/auth/login/student' : '/auth/login';
            const payload = studentByIIN
                ? { iin: studentId, password: formData.password }
                : {
                    email: identifier.includes('@') ? identifier : '',
                    phone: !identifier.includes('@') ? identifier : '',
                    password: formData.password
                };

            const res = await api.post(endpoint, payload);
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
            const errorMsg = err.response?.data?.error?.message;
            if (errorMsg === 'invalid credentials') {
                setError(tAuth('errors.invalid_credentials'));
            } else {
                setError(errorMsg || 'Login failed');
            }
        } finally {
            setLoading(false);
        }
    };

    return (
        <main className="min-h-screen flex items-center justify-center p-6">
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
                <Link
                    href="/"
                    className="absolute top-6 left-6 p-2 text-foreground/50 hover:text-foreground hover:bg-white/5 rounded-full transition-colors"
                >
                    <ChevronLeft className="w-6 h-6" />
                </Link>

                <div className="text-center mb-10">
                    <h1 className="text-3xl font-bold font-outfit mb-2">{tAuth('welcome_back')}</h1>
                    <p className="text-foreground/60">{tAuth(isStudent ? 'sign_in_subtitle_student' : 'sign_in_subtitle_teacher')}</p>
                </div>

                <div className="flex bg-slate-100 dark:bg-white/5 p-1 rounded-2xl mb-8">
                    <button
                        onClick={() => setIsStudent(false)}
                        className={`flex-1 py-3 rounded-xl text-sm font-semibold transition-all ${!isStudent ? 'bg-primary text-white shadow-lg' : 'hover:bg-slate-100 dark:hover:bg-white/5'}`}
                    >
                        {tAuth('teacher_role')}
                    </button>
                    <button
                        onClick={() => setIsStudent(true)}
                        className={`flex-1 py-3 rounded-xl text-sm font-semibold transition-all ${isStudent ? 'bg-primary text-white shadow-lg' : 'hover:bg-slate-100 dark:hover:bg-white/5'}`}
                    >
                        {tAuth('student_role')}
                    </button>
                </div>

                <form onSubmit={handleLogin} className="space-y-6" suppressHydrationWarning>
                    {error && (
                        <div className="p-4 bg-red-500/10 border border-red-500/20 text-red-500 text-sm rounded-xl">
                            {error}
                        </div>
                    )}

                    {!isStudent ? (
                        <div className="space-y-2">
                            <label className="text-sm font-medium ml-2">{tAuth('email_or_phone_label')}</label>
                            <div className="relative">
                                <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground/40" />
                                <input
                                    id="login"
                                    type="text"
                                    required
                                    autoComplete="username"
                                    name="email"
                                    value={formData.email}
                                    onChange={(e) => {
                                        const val = e.target.value;
                                        // Only format as phone if it starts with + or a digit and doesn't contain @
                                        if (!val.includes('@') && (val.startsWith('+') || /^\d/.test(val))) {
                                            setFormData({ ...formData, email: formatPhoneInput(val) });
                                        } else {
                                            setFormData({ ...formData, email: val });
                                        }
                                    }}
                                    className="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl py-4 pl-12 pr-4 outline-none focus:border-primary/50 transition-all text-foreground"
                                    placeholder="name@example.com"
                                    suppressHydrationWarning
                                />
                            </div>
                        </div>
                    ) : (
                        <div className="space-y-2">
                            <label htmlFor="student-login" className="text-sm font-medium ml-2">{tAuth('student_login_label')}</label>
                            <div className="relative">
                                <GraduationCap className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground/40" />
                                <input
                                    id="student-login"
                                    type="text"
                                    required
                                    autoComplete="username"
                                    name="iin"
                                    value={formData.iin}
                                    onChange={(e) => {
                                        const val = e.target.value;
                                        // Pure digits -> IIN; otherwise keep email / phone as typed
                                        if (/^\d*$/.test(val)) setFormData({ ...formData, iin: formatIINInput(val) });
                                        else if (!val.includes('@') && val.startsWith('+')) setFormData({ ...formData, iin: formatPhoneInput(val) });
                                        else setFormData({ ...formData, iin: val });
                                    }}
                                    className="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl py-4 pl-12 pr-4 outline-none focus:border-primary/50 transition-all text-foreground"
                                    placeholder={tAuth('student_login_placeholder')}
                                    suppressHydrationWarning
                                />
                            </div>
                        </div>
                    )}

                    <div className="space-y-2">
                        <div className="flex justify-between items-center ml-2">
                            <label className="text-sm font-medium">{tAuth('password_label')}</label>
                            <Link href={`/forgot-password?type=${isStudent ? 'student' : 'teacher'}`} className="text-primary text-xs hover:underline">{tAuth('forgot_password')}</Link>
                        </div>
                        <div className="relative">
                            <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground/40" />
                            <input
                                type={showPassword ? "text" : "password"}
                                required
                                autoComplete="current-password"
                                name="password"
                                value={formData.password}
                                onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                                className="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl py-4 pl-12 pr-12 outline-none focus:border-primary/50 transition-all text-foreground"
                                placeholder="••••••••"
                                suppressHydrationWarning
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

                    <button
                        disabled={loading}
                        className="w-full bg-primary text-white py-4 rounded-2xl font-bold text-lg hover:bg-primary-hover transition-all flex items-center justify-center gap-2 shadow-xl shadow-primary/20"
                    >
                        {loading ? tAuth('signing_in') : (
                            <>
                                {tAuth('login_button')}
                                <LogIn className="w-5 h-5" />
                            </>
                        )}
                    </button>
                </form>

                <div className="relative my-8">
                    <div className="absolute inset-0 flex items-center">
                        <div className="w-full border-t border-slate-200 dark:border-white/10"></div>
                    </div>
                    <div className="relative flex justify-center text-xs uppercase">
                        <span className="bg-background px-4 text-foreground/40 font-medium">{tAuth('or_continue_with')}</span>
                    </div>
                </div>

                <div className="grid grid-cols-1 gap-4">
                    <button
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
                </div>

                <p className="text-center mt-10 text-foreground/50 text-sm">
                    {tAuth('dont_have_account')} <Link href="/register" className="text-primary font-semibold hover:underline">{tAuth('create_one')}</Link>
                </p>
            </motion.div>
        </main>
    );
}
