'use client';

import React from 'react';
import { useParams, useRouter } from 'next/navigation';
import api from '@/lib/api';
import { motion } from 'framer-motion';
import { Users, Check, AlertCircle, Loader2, ArrowRight } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { saveSession, clearSession, getSession } from '@/lib/auth';
import { formatPhoneInput, formatIINInput } from '@/lib/utils';

export default function JoinClassPage() {
    const params = useParams();
    const router = useRouter();
    const code = params.code as string;
    const t = useTranslations('classes');
    const tAuth = useTranslations('auth');
    const tReg = useTranslations('register');
    const tProfile = useTranslations('profile');
    const tCommon = useTranslations('common');

    // States
    const [step, setStep] = React.useState<'loading_invite' | 'input_iin' | 'login' | 'register' | 'joining' | 'success' | 'role_mismatch' | 'error'>('loading_invite');
    const [inviteInfo, setInviteInfo] = React.useState<{ class_name: string, teacher_name: string } | null>(null);
    const [iin, setIin] = React.useState('');
    const [checkLoading, setCheckLoading] = React.useState(false);
    const [errorMsg, setErrorMsg] = React.useState('');

    // User Data for Login/Register
    const [password, setPassword] = React.useState('');
    const [userInfo, setUserInfo] = React.useState<{ first_name: string, last_name: string } | null>(null);

    // Registration Data
    const [regData, setRegData] = React.useState({
        first_name: '',
        last_name: '',
        phone: '',
        password: '',
        gender: 'male',
        birth_year: new Date().getFullYear() - 10
    });

    const joinClass = async () => {
        setStep('joining');
        try {
            await api.post(`/invites/${code}/accept`);
            setStep('success');
        } catch (e: any) {
            const status = e.response?.status;
            const errCode = e.response?.data?.error?.code;
            const errMsg = e.response?.data?.error?.message;

            // Only log unexpected errors
            if (status !== 401 && status !== 403 && errCode !== "LIMIT_REACHED") {
                console.error('Join error:', e);
            }

            if (status === 401) {
                clearSession();
                setStep('input_iin');
                setErrorMsg(tAuth('errors.session_expired'));
            } else if (errMsg === "ALREADY_IN_CLASS") {
                // Special case: if already in class, show success
                setStep('success');
            } else if (errCode === "LIMIT_REACHED" || status === 403) {
                console.log("Limit reached detected", e.response?.data);
                setStep('error');
                setErrorMsg(t('join.limit_reached'));
            } else {
                setStep('error');
                setErrorMsg(errMsg || tCommon('generic'));
            }
        }
    };

    // 1. Fetch Invite Info and Logout previous session
    React.useEffect(() => {
        setErrorMsg('');
        const fetchInfoAndState = async () => {
            try {
                // Always fetch invite info first
                const res = await api.get(`/invites/${code}`);
                setInviteInfo(res.data.data);

                // Check logged in state
                const { token, role } = getSession();

                if (token && role) {
                    if (role === 'student') {
                        // Auto join
                        joinClass();
                    } else {
                        // Warn mismatch
                        setStep('role_mismatch');
                    }
                } else {
                    // Not logged in
                    setStep('input_iin');
                }
            } catch (e: any) {
                if (e.response?.status === 410 || e.response?.status === 404) {
                    setStep('error');
                    setErrorMsg(t('no_invite_found'));
                } else {
                    console.error('Invite fetch error:', e);
                    setStep('error');
                    setErrorMsg(tCommon('generic'));
                }
            }
        };
        if (code) fetchInfoAndState();
    }, [code]);

    const handleCheckIIN = async () => {
        const cleanIIN = iin.replace(/\s/g, '');
        if (cleanIIN.length !== 12) {
            setErrorMsg(tAuth('iin_label'));
            return;
        }

        setCheckLoading(true);
        setErrorMsg('');
        try {
            const res = await api.get(`/auth/check-iin?iin=${cleanIIN}`);
            if (res.data.data.exists) {
                setUserInfo({
                    first_name: res.data.data.first_name,
                    last_name: res.data.data.last_name
                });
                setStep('login');
            } else {
                setStep('register');
            }
        } catch (e) {
            console.error(e);
            setErrorMsg(tCommon('generic'));
        } finally {
            setCheckLoading(false);
        }
    };

    const handleLogin = async () => {
        setCheckLoading(true);
        setErrorMsg('');
        const cleanIIN = iin.replace(/\s/g, '');
        try {
            const res = await api.post('/auth/login/student', {
                iin: cleanIIN,
                password: password
            });

            const { access_token, refresh_token, user } = res.data.data;
            saveSession(access_token, refresh_token, user.role);

            if (user.role === 'student') {
                joinClass();
            } else {
                setStep('role_mismatch');
            }
        } catch (e: any) {
            setErrorMsg(e.response?.data?.error?.message || tCommon('generic'));
        } finally {
            setCheckLoading(false);
        }
    };

    const handleRegister = async () => {
        setCheckLoading(true);
        setErrorMsg('');
        const cleanIIN = iin.replace(/\s/g, '');
        try {
            const res = await api.post('/auth/register', {
                iin: cleanIIN,
                role: 'student',
                invite_code: code,
                ...regData
            });

            const { access_token, refresh_token, user } = res.data.data;
            saveSession(access_token, refresh_token, user.role);

            setStep('success');
        } catch (e: any) {
            const msg = e.response?.data?.error?.message;
            if (msg === 'user with this phone already exists') {
                setErrorMsg(tAuth('errors.phone_exists'));
            } else if (msg === 'user with this email already exists') {
                setErrorMsg(tAuth('errors.email_exists'));
            } else {
                setErrorMsg(msg || tReg('registration_failed'));
            }
            setCheckLoading(false);
        }
    };

    const handleSwitchAccount = () => {
        clearSession();
        setStep('input_iin');
    };

    // --- Renders ---

    if (step === 'loading_invite') {
        return (
            <div className="min-h-screen flex items-center justify-center bg-[#050505]">
                <Loader2 className="w-10 h-10 text-primary animate-spin" />
            </div>
        );
    }

    if (step === 'error') {
        return (
            <div className="min-h-screen flex items-center justify-center p-4 bg-[#050505] text-white">
                <div className="text-center max-w-md">
                    <div className="w-16 h-16 bg-red-500/10 rounded-full flex items-center justify-center mx-auto mb-6 text-red-500">
                        <AlertCircle className="w-8 h-8" />
                    </div>
                    <h2 className="text-2xl font-bold mb-2">Oops!</h2>
                    <p className="text-white/60 mb-6">{errorMsg}</p>
                    <button onClick={() => router.push('/')} className="px-6 py-2 bg-white/10 rounded-xl hover:bg-white/20 transition-all">{tCommon('back')}</button>
                    {errorMsg.includes('already') && (
                        <button onClick={() => router.push('/student')} className="block mt-4 text-primary hover:underline w-full">{tAuth('welcome_back')}</button>
                    )}
                </div>
            </div>
        );
    }

    if (step === 'role_mismatch') {
        return (
            <div className="min-h-screen flex items-center justify-center p-4 bg-[#050505] text-white">
                <div className="text-center max-w-md">
                    <div className="w-16 h-16 bg-yellow-500/10 rounded-full flex items-center justify-center mx-auto mb-6 text-yellow-500">
                        <Users className="w-8 h-8" />
                    </div>
                    <h2 className="text-2xl font-bold mb-2">{tAuth('student_role')} Access Only</h2>
                    <p className="text-white/60 mb-6">
                        This invite link is for joining the class <strong>{inviteInfo?.class_name}</strong> as a {tAuth('student_role')}.
                        You are currently logged in as a different role.
                    </p>
                    <button
                        onClick={handleSwitchAccount}
                        className="w-full py-3 bg-primary text-white rounded-xl font-bold hover:bg-primary-hover transition-all mb-4"
                    >
                        {tAuth('logout')} & {tCommon('confirm')}
                    </button>
                    <button onClick={() => router.push('/')} className="w-full text-sm text-white/40 hover:text-white">{tCommon('cancel')}</button>
                </div>
            </div>
        );
    }

    if (step === 'success') {
        return (
            <div className="min-h-screen flex items-center justify-center p-4 bg-[#050505] text-white">
                <motion.div
                    initial={{ opacity: 0, scale: 0.9 }}
                    animate={{ opacity: 1, scale: 1 }}
                    className="glass-card p-8 rounded-3xl w-full max-w-md text-center border-green-500/20"
                >
                    <div className="w-16 h-16 bg-green-500/10 rounded-full flex items-center justify-center mx-auto mb-6 text-green-500">
                        <Check className="w-8 h-8" />
                    </div>
                    <h1 className="text-2xl font-bold font-outfit mb-2">{t('add_form.success_title')}</h1>
                    <p className="text-white/60 mb-8">
                        {t('join.success_joined', { className: inviteInfo?.class_name || '' })}
                    </p>
                    <button
                        onClick={() => router.push('/student')}
                        className="w-full py-3 bg-primary text-white rounded-xl font-bold hover:bg-primary-hover transition-all flex items-center justify-center gap-2"
                    >
                        {t('join.dashboard_btn')} <ArrowRight size={20} />
                    </button>
                </motion.div>
            </div>
        );
    }

    if (step === 'joining') {
        return (
            <div className="min-h-screen flex items-center justify-center p-4 bg-[#050505] text-white">
                <div className="text-center">
                    <Loader2 className="w-10 h-10 text-primary animate-spin mx-auto mb-4" />
                    <p className="text-lg">{tCommon('loading')} {inviteInfo?.class_name}...</p>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen flex flex-col items-center justify-center p-4 bg-[#050505] text-white">
            {/* Invite Header */}
            <div className="text-center mb-8">
                <p className="text-white/40 mb-2 uppercase tracking-widest text-xs font-bold">{t('join.invited_to')}</p>
                <h1 className="text-3xl font-bold font-outfit mb-2">{inviteInfo?.class_name}</h1>
                <div className="flex items-center justify-center gap-2 text-white/60 bg-white/5 py-1 px-3 rounded-full mx-auto w-fit">
                    <Users size={14} />
                    <span className="text-sm">{tAuth('teacher_role')}: {inviteInfo?.teacher_name}</span>
                </div>
            </div>

            <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                className="glass-card p-8 rounded-3xl w-full max-w-md border-white/10"
            >
                {step === 'input_iin' && (
                    <div className="space-y-6">
                        <div className="text-center">
                            <h2 className="text-2xl font-bold mb-2 font-outfit">{tAuth('iin_label')}</h2>
                            <p className="text-white/40 text-sm">{t('join.enter_iin_desc')}</p>
                        </div>
                        <div className="relative">
                            <input
                                placeholder="000000 000000"
                                className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-4 text-center text-2xl font-mono tracking-widest outline-none focus:border-primary/50 transition-all"
                                value={iin}
                                onChange={(e) => {
                                    let val = e.target.value.replace(/\D/g, '');
                                    if (val.length > 12) val = val.slice(0, 12);
                                    if (val.length > 6) {
                                        val = val.slice(0, 6) + ' ' + val.slice(6);
                                    }
                                    setIin(val);
                                }}
                            />
                        </div>
                        {errorMsg && <p className="text-red-500 text-sm text-center font-bold bg-red-500/10 p-2 rounded-lg">{errorMsg}</p>}
                        <button
                            onClick={handleCheckIIN}
                            disabled={iin.replace(/\s/g, '').length !== 12 || checkLoading}
                            className="w-full py-4 bg-primary text-white rounded-xl font-bold hover:bg-primary-hover transition-all flex items-center justify-center gap-2 disabled:opacity-50"
                        >
                            {checkLoading ? <Loader2 className="animate-spin" /> : <><ArrowRight size={20} /> {tCommon('next')}</>}
                        </button>
                    </div>
                )}

                {step === 'login' && (
                    <div className="space-y-6">
                        <div className="text-center">
                            <h2 className="text-2xl font-bold mb-1 font-outfit">{tAuth('welcome_back')}, {userInfo?.first_name}!</h2>
                            <p className="text-white/40 text-sm">{t('join.enter_pass_desc')}</p>
                        </div>
                        <input
                            type="password"
                            placeholder={tAuth('password_label')}
                            className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 outline-none focus:border-primary/50"
                            value={password}
                            onChange={e => setPassword(e.target.value)}
                        />
                        {errorMsg && <p className="text-red-500 text-sm font-bold bg-red-500/10 p-2 rounded-lg">{errorMsg}</p>}
                        <button
                            onClick={handleLogin}
                            disabled={!password || checkLoading}
                            className="w-full py-3 bg-primary text-white rounded-xl font-bold hover:bg-primary-hover transition-all"
                        >
                            {checkLoading ? <Loader2 className="animate-spin mx-auto" /> : tAuth('login_button')}
                        </button>
                        <button onClick={() => setStep('input_iin')} className="w-full text-sm text-white/40 hover:text-white mt-2">{tCommon('back')}</button>
                    </div>
                )}

                {step === 'register' && (
                    <div className="space-y-4">
                        <div className="text-center mb-6">
                            <h2 className="text-2xl font-bold mb-1 font-outfit">{tReg('complete_title')}</h2>
                            <p className="text-white/40 text-sm">{t('join.review_details')}</p>
                        </div>

                        <div className="p-3 bg-white/5 rounded-xl border border-white/10 text-center mb-4">
                            <p className="text-xs text-white/40 uppercase">{tAuth('iin_label')}</p>
                            <p className="font-mono text-lg font-bold tracking-widest">{iin}</p>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <input
                                placeholder={tReg('first_name')}
                                className="bg-white/5 border border-white/10 rounded-xl px-4 py-3 outline-none focus:border-primary/50"
                                value={regData.first_name}
                                onChange={e => setRegData({ ...regData, first_name: e.target.value })}
                            />
                            <input
                                placeholder={tReg('last_name')}
                                className="bg-white/5 border border-white/10 rounded-xl px-4 py-3 outline-none focus:border-primary/50"
                                value={regData.last_name}
                                onChange={e => setRegData({ ...regData, last_name: e.target.value })}
                            />
                        </div>
                        <input
                            placeholder={tProfile('phone')}
                            className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 outline-none focus:border-primary/50"
                            value={regData.phone}
                            onChange={e => setRegData({ ...regData, phone: formatPhoneInput(e.target.value) })}
                        />
                        <input
                            type="password"
                            placeholder={tReg('create_password')}
                            className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 outline-none focus:border-primary/50"
                            value={regData.password}
                            onChange={e => setRegData({ ...regData, password: e.target.value })}
                        />
                        <div className="grid grid-cols-2 gap-4">
                            <select
                                className="bg-white/5 border border-white/10 rounded-xl px-4 py-3 outline-none focus:border-primary/50 bg-[#0a0a0a] text-white"
                                value={regData.gender}
                                onChange={e => setRegData({ ...regData, gender: e.target.value })}
                            >
                                <option value="male" className="bg-[#0a0a0a] text-white">{tReg('male')}</option>
                                <option value="female" className="bg-[#0a0a0a] text-white">{tReg('female')}</option>
                            </select>
                            <input
                                type="number"
                                placeholder={tReg('birth_year')}
                                className="bg-white/5 border border-white/10 rounded-xl px-4 py-3 outline-none focus:border-primary/50"
                                value={regData.birth_year}
                                onChange={e => setRegData({ ...regData, birth_year: parseInt(e.target.value) })}
                            />
                        </div>

                        {errorMsg && <p className="text-red-500 text-sm">{errorMsg}</p>}
                        <button
                            onClick={handleRegister}
                            disabled={!regData.first_name || !regData.password || !regData.phone || checkLoading}
                            className="w-full py-3 bg-primary text-white rounded-xl font-bold hover:bg-primary-hover transition-all disabled:opacity-50 mt-2"
                        >
                            {checkLoading ? <Loader2 className="animate-spin mx-auto" /> : t('join.register_and_join')}
                        </button>
                        <button onClick={() => setStep('input_iin')} className="w-full text-sm text-white/40 hover:text-white mt-2 text-center">{tCommon('back')}</button>
                    </div>
                )}
            </motion.div>
        </div>
    );
}
