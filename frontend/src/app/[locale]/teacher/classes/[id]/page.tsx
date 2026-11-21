'use client';

import React from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import api from '@/lib/api';
import { motion } from 'framer-motion';
import {
    Users, Settings, UserPlus, Trash2, Copy, Check,
    ChevronLeft, RefreshCw, X, Plus, Key, Lock, AlertTriangle, UserCheck, Bell, Clock, LayoutPanelLeft
} from 'lucide-react';
import { formatPhoneInput } from '@/lib/utils';
import SendNotificationModal from '@/components/notifications/SendNotificationModal';

export default function ClassDetailsPage() {
    const t = useTranslations('classes');
    const tCommon = useTranslations('common');
    const tTeacher = useTranslations('education_materials');
    const params = useParams();
    const router = useRouter();
    const classId = params.id as string;

    const [classData, setClassData] = React.useState<any>(null);
    const [students, setStudents] = React.useState<any[]>([]);
    const [loading, setLoading] = React.useState(true);
    const [isNotFound, setIsNotFound] = React.useState(false);
    const [activeTab, setActiveTab] = React.useState('students');

    // Invite Modal
    const [isInviteModalOpen, setIsInviteModalOpen] = React.useState(false);
    const [inviteData, setInviteData] = React.useState<any>(null);
    const [inviteCopied, setInviteCopied] = React.useState(false);
    const [modalTab, setModalTab] = React.useState<'invite_link' | 'manual_add'>('invite_link');
    const [usageLimit, setUsageLimit] = React.useState(30);
    const [generatingInvite, setGeneratingInvite] = React.useState(false);
    const [fetchingInvite, setFetchingInvite] = React.useState(false);

    // Password Reset State
    const [resetPasswordData, setResetPasswordData] = React.useState<{ name: string, password: string } | null>(null);

    // Notification Modal State
    const [notifModal, setNotifModal] = React.useState<{
        isOpen: boolean;
        type: 'student' | 'class' | 'all';
        targetId?: string | number;
        targetName?: string;
    }>({
        isOpen: false,
        type: 'class'
    });

    // Fetch Data
    const fetchData = async () => {
        try {
            setLoading(true);
            setIsNotFound(false);

            // 1. Fetch Class Data first
            try {
                const classRes = await api.get(`/classes/${classId}`);
                setClassData(classRes.data.data);
            } catch (e: any) {
                if (e.response?.status === 404) {
                    setIsNotFound(true);
                } else {
                    console.error("Error fetching class:", e);
                }
                setLoading(false);
                return; // Stop if class is not found
            }

            // 2. Fetch Students Data (don't block if it fails)
            try {
                const studentsRes = await api.get(`/classes/${classId}/students`);
                setStudents(studentsRes.data.data || []);
            } catch (e) {
                console.error("Error fetching students:", e);
                // Don't set isNotFound if only students fail
            }
        } catch (e) {
            console.error("Overall fetch error:", e);
        } finally {
            setLoading(false);
        }
    };

    React.useEffect(() => {
        if (classId) fetchData();
    }, [classId]);

    // Invite Logic
    const openInviteModal = async (tab: 'invite_link' | 'manual_add' = 'invite_link') => {
        setIsInviteModalOpen(true);
        setModalTab(tab);

        if (tab === 'invite_link') {
            setFetchingInvite(true);
            setInviteData(null);
            try {
                const res = await api.get(`/classes/${classId}/invites/active`);
                if (res.data.data) {
                    setInviteData(res.data.data);
                    setUsageLimit(res.data.data.usage_limit);
                } else {
                    // AUTO GENERATE with 30 limit if not found
                    const genRes = await api.post(`/classes/${classId}/invites`, {
                        usage_limit: 30,
                        expires_in_hours: 168 // 1 week
                    });
                    setInviteData(genRes.data.data);
                    setUsageLimit(30);
                }
            } catch (e) {
                console.error(e);
            } finally {
                setFetchingInvite(false);
            }
        }
    };

    const handleGenerateInvite = async () => {
        setGeneratingInvite(true);
        try {
            const res = await api.post(`/classes/${classId}/invites`, {
                usage_limit: Number(usageLimit),
                expires_in_hours: 168 // 1 week
            });
            setInviteData(res.data.data);
            setInviteCopied(false);
        } catch (e) {
            console.error(e);
            alert(tCommon('generic'));
        } finally {
            setGeneratingInvite(false);
        }
    };

    const copyInviteCode = () => {
        if (!inviteData) return;
        navigator.clipboard.writeText(`${window.location.origin}/join/${inviteData.code}`);
        setInviteCopied(true);
        setTimeout(() => setInviteCopied(false), 2000);
    };

    // Remove Student
    const handleRemoveStudent = async (studentId: string) => {
        if (!confirm(t('remove_student_confirm'))) return;
        try {
            await api.delete(`/classes/${classId}/students/${studentId}`);
            fetchData();
        } catch (e) {
            console.error(e);
        }
    };

    // Reset Password
    const handleResetPassword = async (student: any) => {
        if (!confirm(t('reset_password_confirm', { name: student.first_name }))) return;
        try {
            const res = await api.put(`/teacher/students/${student.id}/reset-password`);
            setResetPasswordData({
                name: `${student.first_name} ${student.last_name}`,
                password: res.data.data.password
            });
        } catch (e) {
            console.error(e);
        }
    };

    // Delete Class
    const handleDeleteClass = async () => {
        if (!confirm(t('delete_confirm'))) return;
        try {
            await api.delete(`/classes/${classId}`);
            router.push('/teacher/classes');
        } catch (e) {
            console.error(e);
        }
    };

    if (loading) {
        return <div className="p-10 text-center text-foreground/40">{tCommon('loading')}</div>;
    }

    if (isNotFound) {
        return (
            <div className="min-h-[60vh] flex flex-col items-center justify-center text-center space-y-6">
                <div className="w-24 h-24 bg-red-500/10 rounded-full flex items-center justify-center text-red-500">
                    <AlertTriangle size={48} />
                </div>
                <div>
                    <h1 className="text-3xl font-bold font-outfit">{t('class_not_found')}</h1>
                    <p className="text-foreground/60 mt-2">{t('class_not_found_desc')}</p>
                </div>
                <button
                    onClick={() => router.push('/teacher/classes')}
                    className="px-8 py-3 bg-primary text-white rounded-2xl font-bold hover:bg-primary-hover transition-all shadow-lg shadow-primary/20 flex items-center gap-2"
                >
                    <ChevronLeft size={20} />
                    {t('back_to_classes')}
                </button>
            </div>
        );
    }

    if (!classData) {
        return (
            <div className="p-10 text-center text-red-500">
                <p className="text-xl font-bold">{tCommon('generic')}</p>
                <button onClick={() => fetchData()} className="mt-4 text-primary underline">{tCommon('retry')}</button>
            </div>
        );
    }

    return (
        <div className="space-y-8">
            {/* Header */}
            <div className="flex items-center gap-2 md:gap-4">
                <button
                    onClick={() => router.push('/teacher/classes')}
                    className="p-2 bg-slate-100 dark:bg-white/5 rounded-xl hover:bg-slate-200 dark:hover:bg-white/10 transition-all text-foreground/60 active:scale-95"
                >
                    <ChevronLeft size={18} />
                </button>
                <div className="min-w-0 flex-1">
                    <h1 className="text-xl md:text-3xl font-bold font-outfit truncate tracking-tight">{classData.name}</h1>
                    <p className="text-[10px] md:text-sm text-foreground/40 font-medium uppercase tracking-wider">{t('created_at')}: <span suppressHydrationWarning>{new Date(classData.created_at).toLocaleDateString()}</span></p>
                </div>
            </div>

            {/* Tabs */}
            <div className="flex overflow-x-auto scrollbar-hide gap-0 md:gap-2 border-b border-slate-300 dark:border-white/10 -mx-4 md:mx-0 px-4 md:px-0">
                <button
                    onClick={() => setActiveTab('students')}
                    className={`flex-none px-3 md:px-6 py-3 font-medium transition-all relative ${activeTab === 'students' ? 'text-primary' : 'text-foreground/60 hover:text-white'}`}
                >
                    <div className="flex items-center justify-center gap-1.5 md:gap-2 text-[11px] md:text-base">
                        <Users size={14} className="md:w-[18px] md:h-[18px]" />
                        <span className="whitespace-nowrap">{t('students_tab')} ({students.length})</span>
                    </div>
                    {activeTab === 'students' && (
                        <motion.div layoutId="tab" className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary" />
                    )}
                </button>
                <button
                    onClick={() => setActiveTab('settings')}
                    className={`flex-none px-3 md:px-6 py-3 font-medium transition-all relative ${activeTab === 'settings' ? 'text-primary' : 'text-foreground/60 hover:text-white'}`}
                >
                    <div className="flex items-center justify-center gap-1.5 md:gap-2 text-[11px] md:text-base">
                        <Settings size={14} className="md:w-[18px] md:h-[18px]" />
                        <span className="whitespace-nowrap">{t('settings_tab')}</span>
                    </div>
                    {activeTab === 'settings' && (
                        <motion.div layoutId="tab" className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary" />
                    )}
                </button>
                <button
                    onClick={() => setActiveTab('history')}
                    className={`flex-none px-3 md:px-6 py-3 font-medium transition-all relative ${activeTab === 'history' ? 'text-primary' : 'text-foreground/60 hover:text-white'}`}
                >
                    <div className="flex items-center justify-center gap-1.5 md:gap-2 text-[11px] md:text-base">
                        <Clock size={14} className="md:w-[18px] md:h-[18px]" />
                        <span className="whitespace-nowrap">{t('history_tab')}</span>
                    </div>
                    {activeTab === 'history' && (
                        <motion.div layoutId="tab" className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary" />
                    )}
                </button>
            </div>

            {/* Content */}
            {activeTab === 'students' && (
                <div className="glass-card p-4 md:p-6 rounded-3xl">
                    <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-6">
                        <h2 className="text-xl font-bold font-outfit">{t('student_list')}</h2>
                        <div className="flex flex-col sm:grid sm:grid-cols-2 lg:flex lg:flex-row gap-2 md:gap-4 w-full md:w-auto">
                            <button
                                onClick={() => setNotifModal({ isOpen: true, type: 'class', targetId: classId, targetName: classData.name })}
                                className="sm:col-span-2 lg:col-auto flex items-center justify-center gap-2 px-4 py-2.5 md:py-2 bg-primary/10 text-primary rounded-xl text-[11px] md:text-sm font-bold hover:bg-primary/20 transition-all border border-primary/20 active:scale-[0.98]"
                            >
                                <Bell size={14} className="md:w-4 md:h-4" />
                                <span className="truncate">{t('notify_class')}</span>
                            </button>
                            <button
                                onClick={() => router.push(`/teacher/whiteboard/${params.id}`)}
                                className="flex items-center justify-center gap-2 px-4 py-2.5 md:py-2 bg-primary/10 text-primary rounded-xl text-[11px] md:text-sm font-bold hover:bg-primary/20 transition-all border border-primary/20 active:scale-[0.98]"
                                title={tTeacher('tools.whiteboard')}
                            >
                                <LayoutPanelLeft size={14} className="md:w-4 md:h-4" />
                                <span className="truncate">{tTeacher('tools.whiteboard')}</span>
                            </button>
                            <button
                                onClick={() => openInviteModal('manual_add')}
                                className="flex items-center justify-center gap-2 px-3 md:px-4 py-2.5 md:py-2 bg-slate-100 dark:bg-white/5 text-foreground rounded-xl text-[11px] md:text-sm font-bold hover:bg-slate-200 dark:hover:bg-white/10 transition-all border border-slate-300 dark:border-white/10 active:scale-[0.98]"
                            >
                                <Plus size={14} className="md:w-4 md:h-4" />
                                <span className="truncate">{t('add_manually')}</span>
                            </button>
                            <button
                                onClick={() => openInviteModal('invite_link')}
                                className="flex items-center justify-center gap-2 px-3 md:px-4 py-2.5 md:py-2 bg-primary text-white rounded-xl text-[11px] md:text-sm font-bold hover:bg-primary-hover transition-all shadow-md active:scale-95"
                            >
                                <UserPlus size={14} className="md:w-4 md:h-4" />
                                <span className="truncate">{t('invite_link')}</span>
                            </button>
                        </div>
                    </div>

                    {students.length === 0 ? (
                        <p className="text-center py-10 text-foreground/40">{t('no_classes')}</p>
                    ) : (
                        <>
                            {/* Mobile Student Cards - Professional list for small screens */}
                            <div className="md:hidden space-y-4">
                                {students.map((student) => (
                                    <div key={student.id} className="bg-slate-50 dark:bg-white/5 border border-slate-300 dark:border-white/10 rounded-2xl p-3.5 space-y-4 hover:border-primary/30 transition-all shadow-sm">
                                        <div className="flex justify-between items-start gap-4">
                                            <div className="min-w-0 flex-1">
                                                <p className="font-bold text-base text-foreground truncate ">{student.first_name} {student.last_name}</p>
                                                <div className="flex items-center gap-2 mt-1">
                                                    <span className="text-[9px] uppercase font-bold text-foreground/30 leading-none tracking-tight">{t('iin_col')}</span>
                                                    <p className="text-[11px] text-foreground/60 font-mono tracking-tight">{student.iin}</p>
                                                </div>
                                            </div>
                                            <div className="flex flex-shrink-0 gap-1.5 pt-0.5">
                                                <button
                                                    onClick={() => setNotifModal({ isOpen: true, type: 'student', targetId: student.id, targetName: `${student.first_name} ${student.last_name}` })}
                                                    className="p-2 bg-blue-500/10 text-blue-500 rounded-xl active:scale-95 transition-all"
                                                >
                                                    <Bell size={16} />
                                                </button>
                                                <button
                                                    onClick={() => handleResetPassword(student)}
                                                    className="p-2 bg-primary/10 text-primary rounded-xl active:scale-95 transition-all"
                                                >
                                                    <Key size={16} />
                                                </button>
                                                <button
                                                    onClick={() => handleRemoveStudent(student.id)}
                                                    className="p-2 bg-red-500/10 text-red-500 rounded-xl active:scale-95 transition-all"
                                                >
                                                    <Trash2 size={16} />
                                                </button>
                                            </div>
                                        </div>
                                        <div className="pt-3 border-t border-slate-300 dark:border-white/5 flex justify-between items-center text-[11px]">
                                            <span className="text-foreground/40 font-medium uppercase tracking-wider">{t('joined_col')}</span>
                                            <span className="text-foreground/60" suppressHydrationWarning>
                                                {new Date(student.created_at).toLocaleDateString()}
                                            </span>
                                        </div>
                                    </div>
                                ))}
                            </div>

                            {/* Desktop Table - Hidden on mobile, optimized for desktop */}
                            <div className="hidden md:block overflow-x-auto">
                                <table className="w-full text-left border-collapse">
                                    <thead>
                                        <tr className="text-xs text-foreground/40 border-b border-slate-300 dark:border-white/5 uppercase tracking-widest font-bold">
                                            <th className="py-4 pl-4">{t('name_col')}</th>
                                            <th className="py-4">{t('iin_col')}</th>
                                            <th className="py-4">{t('joined_col')}</th>
                                            <th className="py-4 pr-4 text-right">{t('actions_col')}</th>
                                        </tr>
                                    </thead>
                                    <tbody className="divide-y divide-slate-300 dark:divide-white/5">
                                        {students.map((student) => (
                                            <tr key={student.id} className="group hover:bg-slate-50 dark:hover:bg-white/5 transition-colors text-sm">
                                                <td className="py-4 pl-4">
                                                    <div className="font-semibold text-foreground group-hover:text-primary transition-colors">
                                                        {student.first_name} {student.last_name}
                                                    </div>
                                                </td>
                                                <td className="py-4">
                                                    <span className="text-foreground/60 font-mono tracking-wider bg-slate-100 dark:bg-white/5 px-2 py-1 rounded-md text-xs">
                                                        {student.iin}
                                                    </span>
                                                </td>
                                                <td className="py-4 text-foreground/60" suppressHydrationWarning>
                                                    {new Date(student.created_at).toLocaleDateString()}
                                                </td>
                                                <td className="py-4 pr-4 text-right">
                                                    <div className="flex items-center justify-end gap-2 opacity-60 group-hover:opacity-100 transition-opacity">
                                                        <button
                                                            onClick={() => setNotifModal({ isOpen: true, type: 'student', targetId: student.id, targetName: `${student.first_name} ${student.last_name}` })}
                                                            className="p-2 bg-blue-500/10 text-blue-500 rounded-lg hover:bg-blue-500 hover:text-white transition-all"
                                                            title={t('title')}
                                                        >
                                                            <Bell size={16} />
                                                        </button>
                                                        <button
                                                            onClick={() => handleResetPassword(student)}
                                                            className="p-2 bg-primary/10 text-primary rounded-lg hover:bg-primary hover:text-white transition-all"
                                                            title={t('password_reset_title')}
                                                        >
                                                            <Key size={16} />
                                                        </button>
                                                        <button
                                                            onClick={() => handleRemoveStudent(student.id)}
                                                            className="p-2 bg-red-500/10 text-red-500 rounded-lg hover:bg-red-500 hover:text-white transition-all"
                                                            title={t('actions_col')}
                                                        >
                                                            <Trash2 size={16} />
                                                        </button>
                                                    </div>
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        </>
                    )}
                </div>
            )
            }

            {activeTab === 'history' && (
                <ClassHistoryTab classId={classId} />
            )}

            {
                activeTab === 'settings' && (
                    <div className="space-y-6">
                        <div className="glass-card p-6 rounded-3xl border border-red-500/20">
                            <h3 className="text-xl font-bold font-outfit text-red-500 mb-2">{t('danger_zone')}</h3>
                            <p className="text-foreground/60 text-sm mb-4">{t('delete_warning')}</p>
                            <button onClick={handleDeleteClass} className="px-6 py-3 bg-red-500/10 text-red-500 border border-red-500/50 rounded-xl font-bold hover:bg-red-500 hover:text-white transition-all flex items-center gap-2"><Trash2 size={18} /> {tCommon('delete')} {t('edit_class')}</button>
                        </div>
                    </div>
                )
            }

            {/* Invite & Add Modal */}
            {
                isInviteModalOpen && (
                    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
                        <motion.div initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }} className="glass-card p-5 md:p-8 rounded-3xl w-full max-w-2xl border-slate-300 dark:border-white/10 bg-background/95 max-h-[90vh] overflow-y-auto">
                            <div className="flex justify-between items-center mb-6">
                                <h3 className="text-xl md:text-2xl font-bold font-outfit">{t('add_students_modal_title')}</h3>
                                <button onClick={() => setIsInviteModalOpen(false)} className="p-2 text-foreground/40 hover:text-foreground hover:bg-slate-100 dark:hover:bg-white/5 rounded-full transition-all"><X size={24} /></button>
                            </div>

                            <div className="flex gap-4 border-b border-slate-300 dark:border-white/10 mb-6">
                                <button onClick={() => setModalTab('invite_link')} className={`px-4 py-2 font-medium transition-all ${modalTab === 'invite_link' ? 'text-primary border-b-2 border-primary' : 'text-foreground/60'}`}>{t('invite_link')}</button>
                                <button onClick={() => setModalTab('manual_add')} className={`px-4 py-2 font-medium transition-all ${modalTab === 'manual_add' ? 'text-primary border-b-2 border-primary' : 'text-foreground/60'}`}>{t('add_manually')}</button>
                            </div>

                            {modalTab === 'invite_link' && (
                                <div className="space-y-6">
                                    {fetchingInvite ? (
                                        <div className="flex flex-col items-center justify-center py-20 gap-4">
                                            <RefreshCw size={32} className="animate-spin text-primary" />
                                            <p className="text-foreground/40 animate-pulse">{t('checking_invite')}</p>
                                        </div>
                                    ) : (
                                        <>
                                            <div className="bg-slate-50 dark:bg-white/5 rounded-2xl p-6 text-center border border-slate-300 dark:border-white/5">
                                                {inviteData ? (
                                                    <div className="space-y-6">
                                                        <div className="flex flex-col items-center gap-4">
                                                            <div className="p-4 bg-white rounded-2xl">
                                                                <img src={`https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=${encodeURIComponent(`${window.location.origin}/join/${inviteData.code}`)}`} alt="QR Code" className="w-32 h-32" />
                                                            </div>
                                                            <div className="flex items-center gap-4 text-sm">
                                                                <div className="flex items-center gap-2 px-3 py-1 bg-primary/10 text-primary rounded-full border border-primary/20">
                                                                    <UserCheck size={14} />
                                                                    <span>{t('joined_stat', { count: inviteData.usage_count })}</span>
                                                                </div>
                                                                <div className="text-foreground/40">/</div>
                                                                <div className="px-3 py-1 bg-slate-200 dark:bg-white/5 text-foreground/60 rounded-full border border-slate-400 dark:border-white/10">
                                                                    {t('limit_stat', { limit: inviteData.usage_limit })}
                                                                </div>
                                                            </div>
                                                        </div>

                                                        <div onClick={copyInviteCode} className="flex items-center justify-between bg-slate-100 dark:bg-black/40 border border-slate-400 dark:border-white/10 p-4 rounded-2xl cursor-pointer hover:border-primary/50 transition-all group overflow-hidden">
                                                            <div className="text-left truncate flex-1 min-w-0">
                                                                <p className="text-[10px] text-foreground/40 uppercase tracking-widest mb-1 font-bold">{t('shareable_link')}</p>
                                                                <p className="font-mono text-primary truncate text-sm">{`${window.location.origin}/join/${inviteData.code}`}</p>
                                                            </div>
                                                            <div className="p-3 bg-white/50 dark:bg-white/5 rounded-xl text-foreground/60 group-hover:text-white group-hover:bg-primary/20 transition-all shrink-0 ml-4">
                                                                {inviteCopied ? <Check size={20} className="text-green-500" /> : <Copy size={18} />}
                                                            </div>
                                                        </div>
                                                    </div>
                                                ) : (
                                                    <div className="flex flex-col items-center justify-center py-10 text-foreground/40 gap-3">
                                                        <div className="w-16 h-16 bg-slate-100 dark:bg-white/5 rounded-full flex items-center justify-center text-foreground/20 dark:text-white/20"><UserPlus size={32} /></div>
                                                        <p className="font-medium">{t('no_invite_found')}</p>
                                                        <p className="text-xs">{t('create_invite_desc')}</p>
                                                    </div>
                                                )}
                                            </div>

                                            <div className="space-y-4 pt-6 mt-4 border-t border-slate-200 dark:border-white/5">
                                                <div className="flex items-center justify-between gap-6 px-2">
                                                    <div>
                                                        <p className="text-sm font-bold truncate">{t('max_students_limit')}</p>
                                                        <p className="text-[10px] text-foreground/40">{t('max_students_desc')}</p>
                                                    </div>
                                                    <div className="relative">
                                                        <input type="number" value={usageLimit} onChange={e => setUsageLimit(parseInt(e.target.value) || 0)} className="bg-slate-100 dark:bg-white/5 border border-slate-400 dark:border-white/10 rounded-xl px-4 py-3 w-32 text-center outline-none focus:border-primary/50 text-xl font-bold font-mono" min="1" max="1000" />
                                                    </div>
                                                </div>

                                                {inviteData && (
                                                    <div className="flex items-start gap-3 p-4 bg-yellow-500/5 border border-yellow-500/20 rounded-2xl text-yellow-600 dark:text-yellow-500 text-xs leading-relaxed">
                                                        <AlertTriangle size={18} className="shrink-0 mt-0.5" />
                                                        <p>{t('deactivate_warning')}</p>
                                                    </div>
                                                )}

                                                <button onClick={handleGenerateInvite} disabled={generatingInvite || !usageLimit} className={`w-full py-4 rounded-2xl font-bold transition-all flex items-center justify-center gap-3 text-lg ${inviteData ? 'bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-foreground dark:text-white border border-slate-400 dark:border-white/10' : 'bg-primary hover:bg-primary-hover text-white shadow-lg shadow-primary/20'}`}>
                                                    {generatingInvite ? <RefreshCw className="animate-spin" size={24} /> : <RefreshCw size={24} />}
                                                    {inviteData ? t('deactivate_and_generate') : t('generate_invite')}
                                                </button>
                                            </div>
                                        </>
                                    )}
                                </div>
                            )}

                            {modalTab === 'manual_add' && <AddStudentForm classId={classId} onSuccess={() => { setIsInviteModalOpen(false); fetchData(); }} />}
                        </motion.div>
                    </div>
                )
            }

            {/* Password Reset Modal */}
            {
                resetPasswordData && (
                    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-md">
                        <motion.div initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }} className="glass-card p-10 rounded-[40px] w-full max-w-sm border-primary/30 bg-background/95 text-center shadow-2xl shadow-primary/10">
                            <div className="w-20 h-20 bg-primary/10 rounded-full flex items-center justify-center mx-auto text-primary mb-8 border border-primary/20"><Lock size={40} /></div>
                            <h3 className="text-2xl font-bold mb-3 font-outfit">{t('password_reset_title')}</h3>
                            <p className="text-foreground/60 mb-8 px-4 leading-relaxed">{t('new_temp_password', { name: resetPasswordData.name })}</p>
                            <div className="bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl p-6 mb-8 flex items-center justify-between group hover:border-primary/50 transition-all">
                                <span className="font-mono font-bold text-3xl tracking-wider text-primary">{resetPasswordData.password}</span>
                                <button onClick={() => { navigator.clipboard.writeText(resetPasswordData.password); alert('Copied!'); }} className="p-3 hover:bg-primary/20 rounded-xl text-primary transition-all"><Copy size={24} /></button>
                            </div>
                            <button onClick={() => setResetPasswordData(null)} className="w-full py-4 bg-primary text-white hover:bg-primary-hover rounded-2xl font-bold text-lg shadow-lg shadow-primary/20 transition-all">{t('got_it')}</button>
                        </motion.div>
                    </div>
                )
            }

            <SendNotificationModal
                isOpen={notifModal.isOpen}
                onClose={() => setNotifModal({ ...notifModal, isOpen: false })}
                targetType={notifModal.type}
                targetId={notifModal.targetId}
                targetName={notifModal.targetName}
            />
        </div >
    );
}

function AddStudentForm({ classId, onSuccess }: { classId: string, onSuccess: () => void }) {
    const t = useTranslations('classes.add_form');
    const tParent = useTranslations('classes');
    const tReg = useTranslations('register');
    const tCommon = useTranslations('common');
    const defaultForm = { first_name: '', last_name: '', iin: '', phone: '', gender: 'male', birth_year: new Date().getFullYear() - 10 };
    const [formData, setFormData] = React.useState(defaultForm);
    const [loading, setLoading] = React.useState(false);
    const [error, setError] = React.useState('');
    const [successData, setSuccessData] = React.useState<{ student: any, password?: string } | null>(null);

    const handleSubmit = async () => {
        setError('');
        if (formData.iin.length !== 12) { setError('IIN must be 12 digits'); return; }
        setLoading(true);
        try {
            const res = await api.post(`/classes/${classId}/students`, formData);
            setSuccessData(res.data.data);
        } catch (e: any) {
            if (e.response?.data?.error?.code === "LIMIT_REACHED") {
                setError(tParent('limit_reached_error'));
            } else {
                setError(e.response?.data?.error?.message || tCommon('generic'));
            }
        } finally { setLoading(false); }
    };

    if (successData) {
        return (
            <div className="space-y-8 text-center py-4">
                <div className="w-20 h-20 bg-green-500/10 rounded-full flex items-center justify-center mx-auto text-green-500 border border-green-500/20"><Check size={40} /></div>
                <div><h3 className="text-2xl font-bold mb-2">{t('success_title')}</h3><p className="text-foreground/60">{t('success_desc', { name: successData.student.first_name + ' ' + successData.student.last_name })}</p></div>
                {successData.password && (
                    <div className="bg-white/5 border border-white/10 rounded-2xl p-6 text-left space-y-4">
                        <p className="text-[10px] text-foreground/40 uppercase tracking-widest font-bold">{t('temp_creds')}</p>
                        <div className="flex justify-between items-center bg-black/20 p-3 rounded-xl border border-white/5">
                            <div><span className="text-[10px] text-foreground/40 block">{t('login_hint')}</span><span className="font-mono font-bold tracking-wider">{successData.student.iin}</span></div>
                        </div>
                        <div className="flex justify-between items-center bg-black/20 p-3 rounded-xl border border-white/5 font-bold">
                            <div><span className="text-[10px] text-foreground/40 block">{t('pass_hint')}</span><span className="font-mono text-primary text-2xl">{successData.password}</span></div>
                            <button onClick={() => { navigator.clipboard.writeText(successData.password!); alert('Copied!'); }} className="p-2 hover:bg-primary/20 rounded-lg text-primary transition-all"><Copy size={22} /></button>
                        </div>
                        <p className="text-[10px] text-red-500 font-bold bg-red-500/10 p-2 rounded-lg text-center underline italic">{t('save_pass_warning')}</p>
                    </div>
                )}
                <div className="flex gap-4"><button onClick={() => { setSuccessData(null); setFormData(defaultForm); }} className="flex-1 py-4 bg-white/5 rounded-2xl font-bold hover:bg-white/10 transition-all border border-white/10">{t('add_another')}</button><button onClick={onSuccess} className="flex-1 py-4 bg-primary text-white rounded-2xl font-bold hover:bg-primary-hover shadow-lg shadow-primary/20 transition-all">{t('done')}</button></div>
            </div>
        );
    }

    return (
        <div className="space-y-5">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div><label className="text-[10px] text-foreground/40 uppercase tracking-widest mb-1.5 block font-bold">{t('first_name')}</label><input className="w-full bg-slate-100 dark:bg-white/5 border border-slate-400 dark:border-white/10 rounded-xl px-4 py-3 outline-none focus:border-primary/50 transition-all" value={formData.first_name} onChange={e => setFormData({ ...formData, first_name: e.target.value })} placeholder="First Name" /></div>
                <div><label className="text-[10px] text-foreground/40 uppercase tracking-widest mb-1.5 block font-bold">{t('last_name')}</label><input className="w-full bg-slate-100 dark:bg-white/5 border border-slate-400 dark:border-white/10 rounded-xl px-4 py-3 outline-none focus:border-primary/50 transition-all" value={formData.last_name} onChange={e => setFormData({ ...formData, last_name: e.target.value })} placeholder="Last Name" /></div>
            </div>
            <div><label className="text-[10px] text-foreground/40 uppercase tracking-widest mb-1.5 block font-bold">{t('iin_label')}</label><input className="w-full bg-slate-100 dark:bg-white/5 border border-slate-400 dark:border-white/10 rounded-xl px-4 py-3 outline-none focus:border-primary/50 transition-all font-mono tracking-widest" value={formData.iin} onChange={e => setFormData({ ...formData, iin: e.target.value.replace(/\D/g, '') })} maxLength={12} placeholder="000000000000" /></div>
            <div><label className="text-[10px] text-foreground/40 uppercase tracking-widest mb-1.5 block font-bold">{t('phone_label')}</label><input className="w-full bg-slate-100 dark:bg-white/5 border border-slate-400 dark:border-white/10 rounded-xl px-4 py-3 outline-none focus:border-primary/50 transition-all" value={formData.phone} onChange={e => setFormData({ ...formData, phone: formatPhoneInput(e.target.value) })} placeholder="+7 (___) ___ __ __" /></div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                    <label className="text-[10px] text-foreground/40 uppercase tracking-widest mb-1.5 block font-bold">{t('gender')}</label>
                    <select
                        className="w-full bg-slate-100 dark:bg-white/5 border border-slate-400 dark:border-white/10 rounded-xl px-4 py-3 outline-none focus:border-primary/50 bg-background transition-all"
                        value={formData.gender}
                        onChange={e => setFormData({ ...formData, gender: e.target.value })}
                    >
                        <option value="male" className="bg-background text-foreground">{tReg('male')}</option>
                        <option value="female" className="bg-background text-foreground">{tReg('female')}</option>
                    </select>
                </div>
                <div>
                    <label className="text-[10px] text-foreground/40 uppercase tracking-widest mb-1.5 block font-bold">{t('birth_year')}</label>
                    <input
                        type="number"
                        className="w-full bg-slate-100 dark:bg-white/5 border border-slate-400 dark:border-white/10 rounded-xl px-4 py-3 outline-none focus:border-primary/50 transition-all"
                        value={formData.birth_year}
                        onChange={e => setFormData({ ...formData, birth_year: parseInt(e.target.value) })}
                    />
                </div>
            </div>
            {error && <p className="text-red-500 text-xs bg-red-500/10 p-3 rounded-xl border border-red-500/20 text-center font-bold">{error}</p>}
            <button onClick={handleSubmit} disabled={loading || !formData.first_name || !formData.iin} className="w-full py-4 bg-primary text-white rounded-2xl font-bold hover:bg-primary-hover transition-all disabled:opacity-50 mt-4 shadow-lg shadow-primary/20 text-lg">{loading ? tCommon('loading') : t('register_btn')}</button>
        </div>
    );
}

function ClassHistoryTab({ classId }: { classId: string }) {
    const t = useTranslations('classes');
    const [sessions, setSessions] = React.useState<any[]>([]);
    const [loading, setLoading] = React.useState(true);
    const router = useRouter();

    React.useEffect(() => {
        const fetchHistory = async () => {
            try {
                const res = await api.get(`/whiteboard/history/${classId}`);
                setSessions(res.data.data || []);
            } catch (err) {
                console.error("Failed to fetch history:", err);
            } finally {
                setLoading(false);
            }
        };
        fetchHistory();
    }, [classId]);

    if (loading) return <div className="py-20 text-center text-foreground/40 font-outfit">{t('history_loading')}</div>;

    if (sessions.length === 0) {
        return (
            <div className="py-20 text-center glass-card rounded-3xl border-dashed border-white/10">
                <Clock className="w-12 h-12 text-white/20 mx-auto mb-4" />
                <p className="text-white/40 font-outfit">{t('history_no_sessions')}</p>
            </div>
        );
    }

    return (
        <div className="grid gap-4">
            {sessions.map((session, i) => (
                <motion.div
                    key={session.id}
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: i * 0.05 }}
                    className="p-6 glass-card rounded-3xl flex flex-col md:flex-row items-start md:items-center justify-between hover:bg-white/5 transition-all gap-4 border border-white/5"
                >
                    <div className="flex items-center gap-4">
                        <div className="w-12 h-12 rounded-2xl bg-primary/10 flex items-center justify-center text-primary">
                            <Clock size={24} />
                        </div>
                        <div>
                            <p className="font-bold text-lg font-outfit">
                                {session.title || `${t('lesson_label')} ${new Date(session.created_at).toLocaleDateString()}`}
                            </p>
                            <p className="text-sm text-foreground/40 font-outfit">
                                {new Date(session.created_at).toLocaleDateString()} • {new Date(session.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                {session.status === 'completed' && session.ended_at && (
                                    <> • {t('duration_label')}: {Math.round((new Date(session.ended_at).getTime() - new Date(session.created_at).getTime()) / 60000)} min</>
                                )}
                            </p>
                        </div>
                    </div>
                    <div className="flex gap-2 w-full md:w-auto">
                        <button
                            onClick={() => router.push(`/teacher/whiteboard/summary/${session.id}`)}
                            className="flex-1 md:flex-none px-6 py-2 bg-primary/10 text-primary rounded-xl text-sm font-bold hover:bg-primary hover:text-white transition-all font-outfit shadow-sm"
                        >
                            {t('view_summary_btn')}
                        </button>
                    </div>
                </motion.div>
            ))}
        </div>
    );
}
