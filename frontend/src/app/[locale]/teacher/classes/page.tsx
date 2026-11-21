'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { useRouter } from '@/i18n/routing';
import api from '@/lib/api';
import { motion } from 'framer-motion';
import {
    Users, Plus, Search, ChevronLeft, MoreVertical,
    Edit2, Trash2, Clock, Bell, LayoutPanelLeft
} from 'lucide-react';
import SendNotificationModal from '@/components/notifications/SendNotificationModal';

export default function ClassesListPage() {
    const t = useTranslations('classes');
    const tCommon = useTranslations('common');
    const tProfile = useTranslations('profile');
    const tTeacher = useTranslations('education_materials');
    const router = useRouter();
    const [classes, setClasses] = React.useState<any[]>([]);
    const [loading, setLoading] = React.useState(true);
    const [searchQuery, setSearchQuery] = React.useState('');
    const [subscriptionInfo, setSubscriptionInfo] = React.useState<{ limit: number, used: number } | null>(null);

    // Modal States
    const [isCreateModalOpen, setIsCreateModalOpen] = React.useState(false);
    const [isEditModalOpen, setIsEditModalOpen] = React.useState(false);
    const [selectedClass, setSelectedClass] = React.useState<any>(null);
    const [classNameInput, setClassNameInput] = React.useState('');

    // Notification Modal State
    const [notifModal, setNotifModal] = React.useState<{
        isOpen: boolean;
        type: 'student' | 'class' | 'all';
    }>({
        isOpen: false,
        type: 'all'
    });

    const fetchClasses = async () => {
        try {
            const res = await api.get('/classes');
            setClasses(res.data.data || []);
        } catch (e) {
            console.error(e);
        } finally {
            setLoading(false);
        }
    };

    const fetchProfile = async () => {
        try {
            const res = await api.get('/users/me');
            if (res.data.data?.subscription) {
                setSubscriptionInfo({
                    limit: res.data.data.subscription.students_limit,
                    used: res.data.data.subscription.students_count
                });
            }
        } catch (e) {
            console.error(e);
        }
    };

    React.useEffect(() => {
        fetchClasses();
        fetchProfile();
    }, []);

    const handleCreate = async () => {
        if (!classNameInput.trim()) return;
        try {
            await api.post('/classes', { name: classNameInput });
            setIsCreateModalOpen(false);
            setClassNameInput('');
            fetchClasses();
            fetchProfile(); // Update limit after creating class (though creating class doesn't consume student limit directly, adding students does. But good to refresh)
        } catch (e) {
            console.error(e);
        }
    };

    const handleUpdate = async () => {
        if (!selectedClass || !classNameInput.trim()) return;
        try {
            await api.put(`/classes/${selectedClass.id}`, { name: classNameInput });
            setIsEditModalOpen(false);
            setSelectedClass(null);
            setClassNameInput('');
            fetchClasses();
        } catch (e) {
            console.error(e);
        }
    };

    const handleDelete = async (e: React.MouseEvent, id: string) => {
        e.stopPropagation();
        if (!confirm(t('delete_confirm'))) return;
        try {
            await api.delete(`/classes/${id}`);
            fetchClasses();
            fetchProfile();
        } catch (e) {
            console.error(e);
        }
    };

    const openEditModal = (e: React.MouseEvent, cls: any) => {
        e.stopPropagation();
        setSelectedClass(cls);
        setClassNameInput(cls.name);
        setIsEditModalOpen(true);
    };

    const filteredClasses = classes.filter(c =>
        c.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <div className="space-y-8">
            {/* Header */}
            <div className="flex flex-col gap-6">
                <div className="flex items-center gap-3 md:gap-4">
                    <button
                        onClick={() => router.push('/teacher')}
                        className="p-2 glass-card rounded-xl hover:bg-slate-100 dark:hover:bg-white/10 transition-all text-foreground/60"
                    >
                        <ChevronLeft size={20} />
                    </button>
                    <div className="min-w-0">
                        <h1 className="text-2xl md:text-3xl font-bold font-outfit truncate">{t('title')}</h1>
                        <p className="text-foreground/40 text-xs md:text-sm">{t('description')}</p>
                    </div>
                </div>

                <div className="flex flex-col gap-4">
                    <div className="flex flex-wrap items-center gap-3">
                        <button
                            onClick={() => setNotifModal({ isOpen: true, type: 'all' })}
                            className="flex items-center gap-2 px-4 py-2 bg-primary/10 text-primary border border-primary/20 rounded-xl text-sm font-bold hover:bg-primary/20 transition-all font-outfit"
                        >
                            <Bell size={18} />
                            <span className="text-xs md:text-sm">{tCommon('send_all')}</span>
                        </button>

                        <button
                            onClick={() => {
                                setClassNameInput('');
                                setIsCreateModalOpen(true);
                            }}
                            className="flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-xl text-sm font-bold hover:bg-primary-hover transition-all"
                        >
                            <Plus size={18} />
                            <span>{tCommon('add')}</span>
                        </button>

                        {subscriptionInfo && (
                            <div className="flex items-center gap-3 glass-card pl-3 pr-1 py-1 rounded-xl">
                                <div className="hidden lg:block text-[9px] uppercase font-bold text-foreground/40 tracking-wider">
                                    {tProfile('students_limit')}
                                </div>
                                <div className="flex items-center gap-1.5 bg-white dark:bg-black/20 px-2.5 py-1 rounded-lg border border-slate-300 dark:border-white/5">
                                    <Users size={12} className="text-primary" />
                                    <span className="font-mono font-bold text-xs">
                                        <span className={subscriptionInfo.used >= (subscriptionInfo.limit === -1 ? 999999 : subscriptionInfo.limit) ? 'text-red-500' : 'text-foreground'}>
                                            {subscriptionInfo.used}
                                        </span>
                                        <span className="text-foreground/40 mx-0.5">/</span>
                                        <span>{subscriptionInfo.limit === -1 ? '∞' : subscriptionInfo.limit}</span>
                                    </span>
                                </div>
                            </div>
                        )}

                        <div className="relative flex-1 min-w-[200px]">
                            <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-foreground/40" size={18} />
                            <input
                                type="text"
                                placeholder={t('search_placeholder')}
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                className="w-full glass-card rounded-xl pl-12 pr-4 py-2 outline-none focus:border-primary/50 transition-all text-sm"
                            />
                        </div>
                    </div>
                </div>
            </div>

            {/* Grid */}
            {loading ? (
                <div className="text-center py-20 text-foreground/40">{tCommon('loading')}</div>
            ) : filteredClasses.length === 0 ? (
                <div className="text-center py-20 bg-slate-50 dark:bg-white/5 rounded-3xl border border-dashed border-slate-400 dark:border-white/10">
                    <div className="w-16 h-16 bg-slate-100 dark:bg-white/5 rounded-full flex items-center justify-center mx-auto mb-4 text-foreground/40">
                        <Users size={32} />
                    </div>
                    <p className="text-lg font-medium text-foreground/60 mb-4">{t('no_classes')}</p>
                    {searchQuery && (
                        <button
                            onClick={() => setSearchQuery('')}
                            className="text-primary hover:underline"
                        >
                            {t('clear_search')}
                        </button>
                    )}
                </div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {filteredClasses.map((cl, i) => (
                        <motion.div
                            key={cl.id}
                            initial={{ opacity: 0, y: 20 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ delay: i * 0.05 }}
                            onClick={() => router.push(`/teacher/classes/${cl.id}`)}
                            className="glass-card p-6 rounded-3xl group cursor-pointer hover:border-primary/30 transition-all relative overflow-hidden"
                        >
                            <div className="absolute top-0 right-0 p-4 flex gap-2">
                                <button
                                    onClick={(e) => openEditModal(e, cl)}
                                    className="p-2 bg-primary/10 rounded-lg hover:bg-primary text-primary hover:text-white transition-all"
                                    title={tCommon('edit')}
                                >
                                    <Edit2 size={16} />
                                </button>
                                <button
                                    onClick={(e) => { e.stopPropagation(); router.push(`/teacher/whiteboard/${cl.id}`); }}
                                    className="p-2 bg-primary/10 rounded-lg hover:bg-primary text-primary hover:text-white transition-all"
                                    title={tTeacher('tools.whiteboard')}
                                >
                                    <LayoutPanelLeft size={16} />
                                </button>
                                <button
                                    onClick={(e) => handleDelete(e, cl.id)}
                                    className="p-2 bg-red-500/10 rounded-lg hover:bg-red-500 text-red-500 hover:text-white transition-colors"
                                    title={tCommon('delete')}
                                >
                                    <Trash2 size={16} />
                                </button>
                            </div>

                            <div className="w-12 h-12 bg-primary/20 rounded-2xl flex items-center justify-center mb-4 text-primary">
                                <Users size={24} />
                            </div>

                            <h3 className="text-xl font-bold font-outfit mb-2 pr-16">{cl.name}</h3>

                            <div className="flex items-center gap-4 text-sm text-foreground/50 mt-4 pt-4 border-t border-slate-200 dark:border-white/5">
                                <div className="flex items-center gap-1.5">
                                    <Clock size={14} />
                                    <span suppressHydrationWarning>{new Date(cl.created_at).toLocaleDateString()}</span>
                                </div>
                                <div className="flex items-center gap-1.5 ml-auto">
                                    <Users size={14} />
                                    <span>{cl.students_count} {t('students_count_label')}</span>
                                </div>
                            </div>
                        </motion.div>
                    ))}
                </div>
            )}

            {/* Create/Edit Modal */}
            {(isCreateModalOpen || isEditModalOpen) && (
                <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
                    <motion.div
                        initial={{ opacity: 0, scale: 0.95 }}
                        animate={{ opacity: 1, scale: 1 }}
                        className="glass-card p-8 rounded-3xl w-full max-w-md bg-background"
                    >
                        <h3 className="text-2xl font-bold font-outfit mb-6">
                            {isEditModalOpen ? t('edit_class') : t('create_btn')}
                        </h3>
                        <div className="space-y-4 mb-8">
                            <div>
                                <label className="block text-sm font-medium text-foreground/60 mb-2">{t('new_class_name')}</label>
                                <input
                                    type="text"
                                    value={classNameInput}
                                    onChange={(e) => setClassNameInput(e.target.value)}
                                    className="w-full glass-card rounded-xl px-4 py-3 outline-none focus:border-primary/50 transition-all font-outfit"
                                    placeholder={t('placeholder')}
                                    autoFocus
                                />
                            </div>
                        </div>
                        <div className="flex gap-4">
                            <button
                                onClick={() => {
                                    setIsCreateModalOpen(false);
                                    setIsEditModalOpen(false);
                                    setClassNameInput('');
                                }}
                                className="flex-1 py-3 rounded-xl font-bold hover:bg-slate-100 dark:hover:bg-white/5 transition-all text-foreground/60 hover:text-foreground"
                            >
                                {tCommon('cancel')}
                            </button>
                            <button
                                onClick={isEditModalOpen ? handleUpdate : handleCreate}
                                disabled={!classNameInput.trim()}
                                className="flex-1 py-3 bg-primary text-white rounded-xl font-bold hover:bg-primary-hover transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                {isEditModalOpen ? tCommon('save') : tCommon('confirm')}
                            </button>
                        </div>
                    </motion.div>
                </div>
            )}

            <SendNotificationModal
                isOpen={notifModal.isOpen}
                onClose={() => setNotifModal({ ...notifModal, isOpen: false })}
                targetType={notifModal.type}
            />
        </div>
    );
}
