'use client';

import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { User, Mail, Phone, Calendar, Shield, CreditCard, Edit2, Lock, Save, Camera, CheckCircle2, ChevronRight, X, Sun, Moon, Monitor, Globe, Check, RefreshCcw, LogOut, Trash2 } from 'lucide-react';
import { useTranslations, useLocale } from 'next-intl';
import { useRouter } from '@/i18n/routing';
import { useTheme } from 'next-themes';
import api from '@/lib/api';
import { logout } from '@/lib/auth';
import cookie from 'cookie-cutter';
import ImageCropper from './ImageCropper';
import { formatIINInput } from '@/lib/utils';
import { toast } from 'react-hot-toast';

interface UserProfile {
    id: string;
    phone: string;
    email: string;
    iin?: string;
    first_name: string;
    last_name: string;
    birth_year: number;
    gender: string;
    role: string;
    avatar_url?: string;
    subscription: {
        plan: string;
        tokens_limit: number;
        tokens_used: number;
        students_limit: number;
        students_count: number;
        expires_at?: string;
        limits_reset_at?: string;
    };
}

const languages = [
    { code: 'en', label: 'English' },
    { code: 'ru', label: 'Русский' },
    { code: 'kk', label: 'Қазақша' }
];

export default function ProfileContent() {
    const t = useTranslations('profile');
    const tCommon = useTranslations('common');
    const locale = useLocale();
    const router = useRouter();
    const { theme, setTheme } = useTheme();
    const [profile, setProfile] = useState<UserProfile | null>(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [isEditing, setIsEditing] = useState(false);
    const [successMessage, setSuccessMessage] = useState('');
    const [errorMessage, setErrorMessage] = useState('');

    const [isPassModalOpen, setIsPassModalOpen] = useState(false);
    const [passData, setPassData] = useState({
        old_password: '',
        new_password: '',
        confirm_password: ''
    });
    const [passLoading, setPassLoading] = useState(false);
    const [passError, setPassError] = useState('');

    const [editData, setEditData] = useState({
        first_name: '',
        last_name: '',
        birth_year: 0,
        gender: '',
        iin: ''
    });

    // Verification state
    const [isVerifyModalOpen, setIsVerifyModalOpen] = useState(false);
    const [verifyType, setVerifyType] = useState<'email' | 'phone' | null>(null);
    const [verifyStep, setVerifyStep] = useState(1);
    const [newVal, setNewVal] = useState('');
    const [code, setCode] = useState('');
    const [verifyLoading, setVerifyLoading] = useState(false);
    const [verifyError, setVerifyError] = useState('');

    // Image Upload State
    const [selectedImage, setSelectedImage] = useState<string | null>(null);
    const [isCropping, setIsCropping] = useState(false);
    const [uploadingAvatar, setUploadingAvatar] = useState(false);
    const fileInputRef = React.useRef<HTMLInputElement>(null);

    useEffect(() => {
        fetchProfile();
    }, []);

    const fetchProfile = async () => {
        try {
            const res = await api.get('/users/me');
            const data = res.data.data;
            setProfile(data);
            setEditData({
                first_name: data.first_name,
                last_name: data.last_name,
                birth_year: data.birth_year,
                gender: data.gender,
                iin: data.iin || ''
            });
        } catch (err) {
            console.error('Failed to fetch profile:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async () => {
        setSaving(true);
        setSuccessMessage('');
        setErrorMessage('');
        try {
            const res = await api.put('/users/me', editData);
            setProfile(res.data.data);
            setIsEditing(false);
            setSuccessMessage(t('update_success'));
            setTimeout(() => setSuccessMessage(''), 3000);
        } catch (err) {
            setErrorMessage(t('update_error'));
        } finally {
            setSaving(false);
        }
    };

    const handlePassChange = async (e: React.FormEvent) => {
        e.preventDefault();
        if (passData.new_password !== passData.confirm_password) {
            setPassError('Passwords do not match');
            return;
        }

        setPassLoading(true);
        setPassError('');
        try {
            await api.put('/users/me/password', {
                old_password: passData.old_password,
                new_password: passData.new_password
            });
            setSuccessMessage(t('password_updated'));
            setIsPassModalOpen(false);
            setPassData({ old_password: '', new_password: '', confirm_password: '' });
            setTimeout(() => setSuccessMessage(''), 3000);
        } catch (err: any) {
            setPassError(err.response?.data?.error?.message || 'Failed to change password');
        } finally {
            setPassLoading(false);
        }
    };

    const switchLanguage = (nextLocale: string) => {
        cookie.set('NEXT_LOCALE', nextLocale, { path: '/' });
        router.refresh();
    };

    const handleOpenVerify = (type: 'email' | 'phone') => {
        setVerifyType(type);
        setVerifyStep(1);
        setNewVal('');
        setCode('');
        setVerifyError('');
        setIsVerifyModalOpen(true);
    };

    const handleSendCode = async () => {
        setVerifyLoading(true);
        setVerifyError('');
        try {
            const endpoint = verifyType === 'email' ? '/users/me/email/request' : '/users/me/phone/request';
            await api.post(endpoint, { [verifyType!]: newVal });
            setVerifyStep(2);
        } catch (err: any) {
            setVerifyError(err.response?.data?.error?.message || 'Failed to send code');
        } finally {
            setVerifyLoading(false);
        }
    };

    const handleVerifyCode = async () => {
        setVerifyLoading(true);
        setVerifyError('');
        try {
            const endpoint = verifyType === 'email' ? '/users/me/email/verify' : '/users/me/phone/verify';
            await api.post(endpoint, { [verifyType!]: newVal, code });
            setSuccessMessage(t('update_success'));
            setIsVerifyModalOpen(false);
            fetchProfile();
            setTimeout(() => setSuccessMessage(''), 3000);
        } catch (err: any) {
            setVerifyError(err.response?.data?.error?.message || 'Verification failed');
        } finally {
            setVerifyLoading(false);
        }
    };

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = () => {
                setSelectedImage(reader.result as string);
                setIsCropping(true);
            };
            reader.readAsDataURL(file);
        }
    };

    const handleDeleteAvatar = async () => {
        if (!window.confirm(t('confirm_delete_avatar') || 'Are you sure you want to delete your profile photo?')) {
            return;
        }

        setUploadingAvatar(true);
        try {
            await api.delete('/users/me/avatar');
            if (profile) {
                setProfile({ ...profile, avatar_url: undefined });
            }
            toast.success(t('avatar_deleted') || 'Avatar deleted successfully');
        } catch (err) {
            console.error('Failed to delete avatar:', err);
            toast.error(t('avatar_delete_error') || 'Failed to delete avatar');
        } finally {
            setUploadingAvatar(false);
        }
    };

    const handleCropComplete = async (croppedBlob: Blob) => {
        setIsCropping(false);
        setUploadingAvatar(true);
        
        const formData = new FormData();
        formData.append('avatar', croppedBlob, 'avatar.jpg');

        try {
            const res = await api.post('/users/me/avatar', formData, {
                headers: {
                    'Content-Type': 'multipart/form-data',
                },
            });
            
            if (profile) {
                setProfile({ ...profile, avatar_url: res.data.data.avatar_url });
            }
            toast.success(t('avatar_updated') || 'Avatar updated successfully');
        } catch (err) {
            console.error('Failed to upload avatar:', err);
            toast.error(t('avatar_upload_error') || 'Failed to upload avatar');
        } finally {
            setUploadingAvatar(false);
            setSelectedImage(null);
        }
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center py-20">
                <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
            </div>
        );
    }

    if (!profile) return null;

    return (
        <div className="space-y-10">
            <header className="text-left">
                <motion.h1
                    initial={{ opacity: 0, y: -20 }}
                    animate={{ opacity: 1, y: 0 }}
                    className="text-4xl md:text-5xl font-bold tracking-tight text-gradient mb-2"
                >
                    {t('title')}
                </motion.h1>
                <p className="text-foreground/60">{t('personal_info')}</p>
            </header>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                <motion.div initial={{ opacity: 0, x: -20 }} animate={{ opacity: 1, x: 0 }}>
                    <div className="glass-card rounded-3xl p-8">
                        <div className="relative w-32 h-32 mx-auto mb-6 group">
                            <div className="absolute inset-0 bg-gradient-to-tr from-primary to-purple-500 rounded-full blur-xl opacity-20 group-hover:opacity-40 transition-opacity" />
                            <div className="relative w-full h-full rounded-full bg-slate-100 dark:bg-white/5 border-2 border-slate-200 dark:border-white/10 flex items-center justify-center overflow-hidden">
                                {profile.avatar_url ? (
                                    <img 
                                        src={`${process.env.NEXT_PUBLIC_API_URL || ''}${profile.avatar_url}`} 
                                        alt="Avatar" 
                                        className="w-full h-full object-cover"
                                    />
                                ) : (
                                    <User className="w-16 h-16 text-primary/60" />
                                )}
                                <button 
                                    onClick={() => fileInputRef.current?.click()}
                                    disabled={uploadingAvatar}
                                    className="absolute inset-0 bg-black/40 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
                                >
                                    {uploadingAvatar ? (
                                        <RefreshCcw className="w-6 h-6 text-white animate-spin" />
                                    ) : (
                                        <Camera className="w-6 h-6 text-white" />
                                    )}
                                </button>
                            </div>
                            {profile.avatar_url && !uploadingAvatar && (
                                <button
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        handleDeleteAvatar();
                                    }}
                                    className="absolute -top-1 -right-1 w-8 h-8 bg-white dark:bg-slate-800 rounded-full shadow-lg border border-slate-200 dark:border-white/10 flex items-center justify-center text-red-500 hover:bg-red-50 dark:hover:bg-red-500/10 transition-all z-10"
                                    title={t('delete_avatar') || 'Delete photo'}
                                >
                                    <Trash2 className="w-4 h-4" />
                                </button>
                            )}
                            <input 
                                type="file" 
                                ref={fileInputRef} 
                                onChange={handleFileChange} 
                                accept="image/*" 
                                className="hidden" 
                            />
                        </div>

                        <div className="text-center mb-8">
                            <h2 className="text-2xl font-bold">{profile.first_name} {profile.last_name}</h2>
                            <span className="inline-block px-3 py-1 bg-primary/10 text-primary text-xs font-bold rounded-full mt-2 uppercase tracking-wider">
                                {profile.role === 'teacher' ? t('role_teacher') :
                                    profile.role === 'student' ? t('role_student') :
                                        profile.role === 'admin' ? t('role_admin') : profile.role}
                            </span>
                        </div>

                        {profile.role !== 'student' && (
                            <>
                                <div className="space-y-6 border-t border-slate-200 dark:border-white/10 pt-6">
                                    <div className="flex items-center justify-between text-sm">
                                        <span className="text-foreground/40">{t('subscription')}</span>
                                        <span className="font-bold text-primary uppercase">{profile.subscription.plan}</span>
                                    </div>

                                    {/* Tokens limit */}
                                    <div className="space-y-2">
                                        <div className="flex justify-between text-[10px] text-foreground/40 font-bold uppercase tracking-widest">
                                            <span>{t('tokens_used')}</span>
                                            <span>{profile.subscription.tokens_limit === -1 ? '∞' : t('tokens_limit')}</span>
                                        </div>
                                        <div className="w-full bg-primary/5 rounded-full h-2.5 overflow-hidden border border-primary/10 p-0.5">
                                            <div
                                                className="bg-gradient-to-r from-primary to-purple-500 h-full rounded-full transition-all duration-1000"
                                                style={{ width: profile.subscription.tokens_limit === -1 ? '100%' : `${Math.min((profile.subscription.tokens_used / profile.subscription.tokens_limit) * 100, 100)}%` }}
                                            />
                                        </div>
                                        <div className="flex justify-between text-[10px] text-foreground/30 font-mono">
                                            <span>{profile.subscription.tokens_used}</span>
                                            <span>{profile.subscription.tokens_limit === -1 ? 'UNLIMITED' : profile.subscription.tokens_limit}</span>
                                        </div>
                                    </div>

                                    {/* Students limit */}
                                    <div className="space-y-2">
                                        <div className="flex justify-between text-[10px] text-foreground/40 font-bold uppercase tracking-widest">
                                            <span>{t('students_used')}</span>
                                            <span>{profile.subscription.students_limit === -1 ? '∞' : t('students_limit')}</span>
                                        </div>
                                        <div className="w-full bg-emerald-500/5 rounded-full h-2.5 overflow-hidden border border-emerald-500/10 p-0.5">
                                            <div
                                                className="bg-gradient-to-r from-emerald-500 to-teal-500 h-full rounded-full transition-all duration-1000"
                                                style={{ width: profile.subscription.students_limit === -1 ? '100%' : `${Math.min((profile.subscription.students_count / profile.subscription.students_limit) * 100, 100)}%` }}
                                            />
                                        </div>
                                        <div className="flex justify-between text-[10px] text-foreground/30 font-mono">
                                            <span>{profile.subscription.students_count}</span>
                                            <span>{profile.subscription.students_limit === -1 ? 'UNLIMITED' : profile.subscription.students_limit}</span>
                                        </div>
                                    </div>

                                    {/* Expiration and Reset Dates */}
                                    <div className="pt-4 border-t border-slate-200 dark:border-white/10 space-y-3">
                                        {profile.subscription.expires_at && (
                                            <div className="flex items-center gap-2 text-xs text-foreground/60">
                                                <Calendar size={14} className="text-primary/60" />
                                                <span>{t('expires_label')}: <span className="font-bold">{new Date(profile.subscription.expires_at).toLocaleDateString()}</span></span>
                                            </div>
                                        )}
                                        {profile.subscription.limits_reset_at && (
                                            <div className="flex items-center gap-2 text-xs text-foreground/60">
                                                <RefreshCcw size={14} className="text-emerald-500/60" />
                                                <span>{t('reset_label')}: <span className="font-bold">{new Date(profile.subscription.limits_reset_at).toLocaleDateString()}</span></span>
                                            </div>
                                        )}
                                    </div>
                                </div>

                                <button
                                    onClick={() => {
                                        const msg = encodeURIComponent(t('upgrade_plan_msg'));
                                        window.open(`https://wa.me/77088710495?text=${msg}`, '_blank');
                                    }}
                                    className="w-full mt-8 flex items-center justify-center gap-2 py-3 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 border border-slate-200 dark:border-white/10 rounded-2xl transition-all font-medium"
                                >
                                    <CreditCard className="w-4 h-4" />
                                    {t('upgrade_plan_btn')}
                                </button>
                            </>
                        )}
                        {profile.role === 'student' && (
                            <div className="mt-8 pt-6 border-t border-slate-200 dark:border-white/10 space-y-4">
                                <p className="text-xs font-bold text-foreground/40 uppercase tracking-widest text-center">{t('achievements')}</p>
                                <div className="grid grid-cols-3 gap-2">
                                    {['🌅', '🧩', '⚡'].map((emoji, i) => (
                                        <div key={i} className="flex flex-col items-center p-3 bg-slate-100 dark:bg-white/5 rounded-2xl border border-slate-200 dark:border-white/5 opacity-40 grayscale">
                                            <span className="text-2xl">{emoji}</span>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}
                    </div>
                </motion.div>

                <motion.div initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} className="lg:col-span-2 space-y-6">
                    {successMessage && (
                        <motion.div initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }} className="bg-emerald-500/10 border border-emerald-500/20 text-emerald-500 p-4 rounded-2xl flex items-center gap-3 font-medium">
                            <CheckCircle2 className="w-5 h-5" />
                            {successMessage}
                        </motion.div>
                    )}

                    <div className="glass-card rounded-3xl p-8 md:p-10">
                        <div className="flex items-center justify-between mb-8">
                            <h3 className="text-xl font-bold flex items-center gap-2">
                                <Shield className="w-5 h-5 text-primary" />
                                {t('account_details')}
                            </h3>
                            {!isEditing && (
                                <button onClick={() => setIsEditing(true)} className="text-sm font-semibold text-primary flex items-center gap-1 hover:underline">
                                    <Edit2 className="w-3 h-3" />
                                    {t('edit_profile')}
                                </button>
                            )}
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                            <div className="space-y-6">
                                <div className="space-y-1">
                                    <label className="text-xs font-bold text-foreground/40 uppercase tracking-widest">{t('first_name')}</label>
                                    {isEditing ? <input value={editData.first_name} onChange={e => setEditData({ ...editData, first_name: e.target.value })} className="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-3 outline-none focus:border-primary/50" /> : <p className="text-lg font-medium">{profile.first_name}</p>}
                                </div>
                                <div className="space-y-1">
                                    <label className="text-xs font-bold text-foreground/40 uppercase tracking-widest">{t('last_name')}</label>
                                    {isEditing ? <input value={editData.last_name} onChange={e => setEditData({ ...editData, last_name: e.target.value })} className="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-3 outline-none focus:border-primary/50" /> : <p className="text-lg font-medium">{profile.last_name}</p>}
                                </div>
                                <div className="space-y-1">
                                    <label className="text-xs font-bold text-foreground/40 uppercase tracking-widest">{t('birth_year')}</label>
                                    {isEditing ? <input type="number" value={editData.birth_year} onChange={e => setEditData({ ...editData, birth_year: parseInt(e.target.value) })} className="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-3 outline-none focus:border-primary/50" /> : <p className="text-lg font-medium">{profile.birth_year}</p>}
                                </div>
                                <div className="space-y-1">
                                    <label className="text-xs font-bold text-foreground/40 uppercase tracking-widest">{t('gender')}</label>
                                    {isEditing ? (
                                        <select
                                            value={editData.gender}
                                            onChange={e => setEditData({ ...editData, gender: e.target.value })}
                                            className="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-3 outline-none focus:border-primary/50 bg-background text-foreground"
                                        >
                                            <option value="male" className="bg-background text-foreground">{t('male')}</option>
                                            <option value="female" className="bg-background text-foreground">{t('female')}</option>
                                        </select>
                                    ) : (
                                        <p className="text-lg font-medium">{profile.gender === 'male' ? t('male') : t('female')}</p>
                                    )}
                                </div>
                            </div>
                            <div className="space-y-6">
                                <div className="space-y-1">
                                    <div className="flex justify-between">
                                        <label className="text-xs font-bold text-foreground/40 uppercase tracking-widest">{t('email')}</label>
                                        <button onClick={() => handleOpenVerify('email')} className="text-[10px] font-bold text-primary hover:underline">{t('change')}</button>
                                    </div>
                                    <div className="flex items-center gap-3 text-lg font-medium"><Mail className="w-4 h-4 text-primary/60" />{profile.email || '—'}</div>
                                </div>
                                <div className="space-y-1">
                                    <div className="flex justify-between">
                                        <label className="text-xs font-bold text-foreground/40 uppercase tracking-widest">{t('phone')}</label>
                                        <button onClick={() => handleOpenVerify('phone')} className="text-[10px] font-bold text-primary hover:underline">{t('change')}</button>
                                    </div>
                                    <div className="flex items-center gap-3 text-lg font-medium"><Phone className="w-4 h-4 text-emerald-500/60" />{profile.phone}</div>
                                </div>
                                <div className="space-y-1">
                                    <label className="text-xs font-bold text-foreground/40 uppercase tracking-widest">{t('iin')}</label>
                                    {isEditing ? (
                                        <div className="relative">
                                            <Shield className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-purple-500/60" />
                                            <input
                                                value={editData.iin}
                                                onChange={e => setEditData({ ...editData, iin: formatIINInput(e.target.value) })}
                                                className="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl pl-10 pr-4 py-3 outline-none focus:border-primary/50"
                                                placeholder="000000000000"
                                                maxLength={12}
                                            />
                                        </div>
                                    ) : (
                                        <div className="flex items-center gap-3 text-lg font-medium">
                                            <Shield className="w-4 h-4 text-purple-500/60" />
                                            {profile.iin || '—'}
                                        </div>
                                    )}
                                </div>
                            </div>
                        </div>

                        {isEditing && (
                            <div className="mt-12 flex items-center gap-4">
                                <button onClick={handleSave} disabled={saving} className="flex items-center gap-2 px-8 py-3 bg-primary text-white font-bold rounded-2xl hover:bg-primary-hover transition-all shadow-lg shadow-primary/20">
                                    {saving ? '...' : <><Save className="w-4 h-4" /> {t('save_changes')}</>}
                                </button>
                                <button onClick={() => setIsEditing(false)} className="px-8 py-3 bg-slate-100 dark:bg-white/5 font-bold rounded-2xl hover:bg-slate-200 dark:hover:bg-white/10 transition-all border border-slate-200 dark:border-white/10">{tCommon('cancel')}</button>
                            </div>
                        )}
                    </div>

                    <div className="glass-card rounded-3xl p-8 md:p-10">
                        <h3 className="text-xl font-bold flex items-center gap-2 mb-8 uppercase tracking-widest text-foreground/40 scale-90 origin-left">{t('settings_title')}</h3>
                        <div className="space-y-10">
                            <div className="space-y-4">
                                <div className="flex items-center gap-2 text-sm font-bold text-foreground/60"><Sun className="w-4 h-4" />{t('appearance')}</div>
                                <div className="flex gap-4 p-1.5 bg-slate-100 dark:bg-black/10 rounded-2xl max-w-sm">
                                    {[{ id: 'light', icon: Sun, label: t('theme_light') }, { id: 'dark', icon: Moon, label: t('theme_dark') }, { id: 'system', icon: Monitor, label: t('theme_system') }].map((item) => (
                                        <button key={item.id} onClick={() => setTheme(item.id)} className={`flex-1 flex flex-col items-center gap-2 py-3 px-2 rounded-xl transition-all ${theme === item.id ? 'bg-white dark:bg-white/10 shadow-sm' : 'hover:bg-slate-200 dark:hover:bg-white/5 opacity-40 hover:opacity-100'}`}>
                                            <item.icon className={`w-5 h-5 ${theme === item.id ? 'text-primary' : ''}`} />
                                            <span className="text-[10px] font-bold uppercase tracking-wider">{item.label}</span>
                                        </button>
                                    ))}
                                </div>
                            </div>

                            <div className="space-y-4">
                                <div className="flex items-center gap-2 text-sm font-bold text-foreground/60"><Globe className="w-4 h-4" />{t('language')}</div>
                                <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-3">
                                    {languages.map((lang) => (
                                        <button key={lang.code} onClick={() => switchLanguage(lang.code)} className={`flex items-center justify-center p-3 rounded-2xl border transition-all text-sm font-medium ${locale === lang.code ? 'bg-primary/10 border-primary/50 text-primary' : 'bg-slate-100 dark:bg-white/5 border-slate-200 dark:border-white/10 hover:border-slate-300 dark:hover:border-white/20'}`}>{lang.label}{locale === lang.code && <Check className="ml-2 w-3 h-3" />}</button>
                                    ))}
                                </div>
                            </div>

                            <div className="pt-6 border-t border-slate-200 dark:border-white/10">
                                <button onClick={() => setIsPassModalOpen(true)} className="flex items-center justify-between w-full group hover:text-primary transition-colors">
                                    <div className="flex items-center gap-4">
                                        <div className="w-12 h-12 rounded-2xl bg-primary/10 flex items-center justify-center"><Lock className="w-6 h-6 text-primary" /></div>
                                        <div className="text-left"><p className="font-bold">{t('change_password')}</p><p className="text-xs text-foreground/40">{t('keep_account_secure')}</p></div>
                                    </div>
                                    <ChevronRight className="w-5 h-5 text-foreground/20 group-hover:text-primary transition-colors" />
                                </button>
                            </div>

                            <div className="pt-6 border-t border-slate-200 dark:border-white/10">
                                <button onClick={() => logout()} className="flex items-center justify-between w-full group hover:text-red-500 transition-colors">
                                    <div className="flex items-center gap-4">
                                        <div className="w-12 h-12 rounded-2xl bg-red-500/10 flex items-center justify-center">
                                            <LogOut className="w-6 h-6 text-red-500" />
                                        </div>
                                        <div className="text-left">
                                            <p className="font-bold">{t('logout_btn')}</p>
                                            <p className="text-xs text-foreground/40">{t('logout_desc')}</p>
                                        </div>
                                    </div>
                                    <ChevronRight className="w-5 h-5 text-foreground/20 group-hover:text-red-500 transition-colors" />
                                </button>
                            </div>
                        </div>
                    </div>
                </motion.div>
            </div>

            <AnimatePresence>
                {/* Password Change Modal */}
                {isPassModalOpen && (
                    <div className="fixed inset-0 z-[60] flex items-center justify-center p-6">
                        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} onClick={() => setIsPassModalOpen(false)} className="absolute inset-0 bg-black/80 backdrop-blur-sm" />
                        <motion.div initial={{ opacity: 0, scale: 0.95, y: 20 }} animate={{ opacity: 1, scale: 1, y: 0 }} exit={{ opacity: 0, scale: 0.95, y: 20 }} className="relative w-full max-w-md bg-background border border-slate-200 dark:border-white/10 rounded-[32px] p-8 md:p-10 shadow-2xl">
                            <button onClick={() => setIsPassModalOpen(false)} className="absolute top-6 right-6 p-2 rounded-xl hover:bg-slate-100 dark:hover:bg-white/5 text-foreground/40 transition-colors"><X className="w-5 h-5" /></button>
                            <div className="mb-8"><h3 className="text-2xl font-bold mb-2">{t('change_password')}</h3><p className="text-xs text-foreground/40 uppercase tracking-widest font-bold">{t('security_update')}</p></div>
                            <form onSubmit={handlePassChange} className="space-y-6">
                                <div className="space-y-2"><label className="text-xs font-bold text-foreground/40 uppercase ml-2 tracking-widest">{t('current_password')}</label><input type="password" required value={passData.old_password} onChange={e => setPassData({ ...passData, old_password: e.target.value })} className="w-full bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl px-5 py-4 outline-none focus:border-primary/50 transition-all" /></div>
                                <div className="space-y-2"><label className="text-xs font-bold text-foreground/40 uppercase ml-2 tracking-widest">{t('new_password')}</label><input type="password" required value={passData.new_password} onChange={e => setPassData({ ...passData, new_password: e.target.value })} className="w-full bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl px-5 py-4 outline-none focus:border-primary/50 transition-all" /></div>
                                <div className="space-y-2"><label className="text-xs font-bold text-foreground/40 uppercase ml-2 tracking-widest">{t('confirm_new_password')}</label><input type="password" required value={passData.confirm_password} onChange={e => setPassData({ ...passData, confirm_password: e.target.value })} className="w-full bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl px-5 py-4 outline-none focus:border-primary/50 transition-all" /></div>
                                {passError && <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-500 text-xs font-bold">{passError}</div>}
                                <button type="submit" disabled={passLoading} className="w-full bg-primary text-white py-4 rounded-2xl font-bold text-lg hover:bg-primary-hover transition-all shadow-xl shadow-primary/20 disabled:opacity-50">{passLoading ? '...' : t('save_changes')}</button>
                            </form>
                        </motion.div>
                    </div>
                )}

                {/* Email/Phone Verification Modal */}
                {isVerifyModalOpen && (
                    <div className="fixed inset-0 z-[60] flex items-center justify-center p-6">
                        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} onClick={() => setIsVerifyModalOpen(false)} className="absolute inset-0 bg-black/80 backdrop-blur-sm" />
                        <motion.div initial={{ opacity: 0, scale: 0.95, y: 20 }} animate={{ opacity: 1, scale: 1, y: 0 }} exit={{ opacity: 0, scale: 0.95, y: 20 }} className="relative w-full max-w-md bg-background border border-slate-200 dark:border-white/10 rounded-[32px] p-8 md:p-10 shadow-2xl">
                            <button onClick={() => setIsVerifyModalOpen(false)} className="absolute top-6 right-6 p-2 rounded-xl hover:bg-slate-100 dark:hover:bg-white/5 text-foreground/40 transition-colors"><X className="w-5 h-5" /></button>

                            <div className="mb-8">
                                <h3 className="text-2xl font-bold mb-2">
                                    {verifyType === 'email' ? t('new_email') : t('new_phone')}
                                </h3>
                                <p className="text-xs text-foreground/40 uppercase tracking-widest font-bold">Verification</p>
                            </div>

                            <div className="space-y-6">
                                {verifyStep === 1 ? (
                                    <>
                                        <div className="space-y-2">
                                            <label className="text-xs font-bold text-foreground/40 uppercase ml-2 tracking-widest">
                                                {verifyType === 'email' ? t('new_email') : t('new_phone')}
                                            </label>
                                            <input
                                                value={newVal}
                                                onChange={e => setNewVal(e.target.value)}
                                                placeholder={verifyType === 'email' ? 'example@mail.com' : '+77000000000'}
                                                className="w-full bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl px-5 py-4 outline-none focus:border-primary/50 transition-all"
                                            />
                                        </div>
                                        {verifyError && <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-500 text-xs font-bold">{verifyError}</div>}
                                        <button
                                            onClick={handleSendCode}
                                            disabled={verifyLoading || !newVal}
                                            className="w-full bg-primary text-white py-4 rounded-2xl font-bold text-lg hover:bg-primary-hover transition-all shadow-xl shadow-primary/20 disabled:opacity-50 flex items-center justify-center gap-2"
                                        >
                                            {verifyLoading ? <RefreshCcw size={20} className="animate-spin" /> : t('send_code')}
                                        </button>
                                    </>
                                ) : (
                                    <>
                                        <div className="space-y-2">
                                            <label className="text-xs font-bold text-foreground/40 uppercase ml-2 tracking-widest">{t('enter_code')}</label>
                                            <input
                                                value={code}
                                                onChange={e => setCode(e.target.value)}
                                                placeholder="000000"
                                                maxLength={6}
                                                className="w-full bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl px-5 py-4 outline-none focus:border-primary/50 transition-all text-center tracking-[0.5em] text-2xl font-bold"
                                            />
                                        </div>
                                        {verifyError && <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-500 text-xs font-bold">{verifyError}</div>}
                                        <button
                                            onClick={handleVerifyCode}
                                            disabled={verifyLoading || code.length < 6}
                                            className="w-full bg-emerald-500 text-white py-4 rounded-2xl font-bold text-lg hover:bg-emerald-600 transition-all shadow-xl shadow-emerald-500/20 disabled:opacity-50 flex items-center justify-center gap-2"
                                        >
                                            {verifyLoading ? <RefreshCcw size={20} className="animate-spin" /> : t('verify_and_save')}
                                        </button>
                                        <button
                                            onClick={() => setVerifyStep(1)}
                                            className="w-full text-xs font-bold text-foreground/40 hover:text-foreground transition-colors"
                                        >
                                            {t('edit_resend')}
                                        </button>
                                    </>
                                )}
                            </div>
                        </motion.div>
                    </div>
                )}
            </AnimatePresence>

            <AnimatePresence>
                {isCropping && selectedImage && (
                    <ImageCropper 
                        image={selectedImage}
                        onCropComplete={handleCropComplete}
                        onCancel={() => {
                            setIsCropping(false);
                            setSelectedImage(null);
                        }}
                    />
                )}
            </AnimatePresence>
        </div>
    );
}
