'use client';

import React, { useState, useEffect } from 'react';
import { useNotifications } from '@/hooks/useNotifications';
import { useTranslations } from 'next-intl';
import { motion } from 'framer-motion';
import {
    Bell,
    Search,
    Trash2,
    CheckCircle2,
    Clock,
    Filter,
    AlertTriangle,
    AlertCircle,
    Info,
    Calendar,
    ChevronLeft,
    Sparkles
} from 'lucide-react';
import { useRouter, Link } from '@/i18n/routing';
import { ExternalLink } from 'lucide-react';
import { NotificationType } from '@/types/notification';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

const getNotificationIcon = (type: NotificationType) => {
    switch (type) {
        case 'token_limit_warning':
        case 'student_limit_warning':
            return <AlertTriangle className="w-6 h-6 text-amber-500" />;
        case 'token_limit_reached':
        case 'subscription_expired':
            return <AlertCircle className="w-6 h-6 text-red-500" />;
        case 'subscription_expiring':
            return <Info className="w-6 h-6 text-blue-500" />;
        case 'monthly_reset':
            return <Calendar className="w-6 h-6 text-green-500" />;
        case 'material_ready':
        case 'whiteboard_summary_ready':
            return <Sparkles className="w-6 h-6 text-purple-500" />;
        case 'whiteboard_started':
            return <Bell className="w-6 h-6 text-primary animate-pulse" />;
        default:
            return <Bell className="w-6 h-6 text-primary" />;
    }
};

export default function NotificationsPage() {
    const t = useTranslations('notifications');
    const tCommon = useTranslations('common');
    const router = useRouter();
    const [searchQuery, setSearchQuery] = useState('');
    const {
        notifications,
        unreadCount,
        totalCount,
        page,
        unreadOnly,
        pageCount,
        setPage,
        setUnreadOnly,
        markAsRead,
        markAllAsRead,
        deleteNotification,
        isLoading,
        refresh
    } = useNotifications();

    const filteredNotifications = notifications.filter(n => {
        const matchesSearch = n.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
            n.message.toLowerCase().includes(searchQuery.toLowerCase());
        return matchesSearch;
    });

    return (
        <div className="min-h-screen bg-gradient-premium pt-24 pb-12 px-4 md:px-8">
            <div className="max-w-4xl mx-auto space-y-8">
                {/* Header */}
                <div className="flex flex-col md:flex-row justify-between items-start md:items-end gap-6">
                    <div className="space-y-2">
                        <button
                            onClick={() => router.back()}
                            className="text-primary hover:text-primary-hover flex items-center gap-2 text-sm font-medium transition-colors mb-4 group"
                        >
                            <ChevronLeft size={16} className="group-hover:-translate-x-1 transition-transform" />
                            {tCommon('back')}
                        </button>
                        <h1 className="text-4xl font-bold font-outfit">{t('title')}</h1>
                        <p className="text-foreground/50">
                            {unreadCount > 0
                                ? `${unreadCount} ${t('unread_count') || 'unread messages'}`
                                : t('all_read') || 'All caught up!'}
                        </p>
                    </div>

                    <div className="flex gap-3 w-full md:w-auto">
                        {unreadCount > 0 && (
                            <button
                                onClick={markAllAsRead}
                                className="flex-1 md:flex-none flex items-center justify-center gap-2 px-6 py-3 bg-white/5 hover:bg-white/10 text-white rounded-2xl font-bold transition-all border border-white/10"
                            >
                                <CheckCircle2 size={18} />
                                {t('mark_all_read')}
                            </button>
                        )}
                    </div>
                </div>

                {/* Filters & Search */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="md:col-span-2 relative">
                        <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-foreground/40" size={20} />
                        <input
                            type="text"
                            placeholder={t('search_placeholder') || 'Search notifications...'}
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="w-full bg-white/5 border border-white/10 rounded-2xl pl-12 pr-4 py-4 outline-none focus:border-primary/50 transition-all"
                        />
                    </div>
                    <div className="flex bg-white/5 p-1 rounded-2xl border border-white/10">
                        <button
                            onClick={() => setUnreadOnly(false)}
                            className={cn(
                                "flex-1 py-3 rounded-xl text-sm font-bold transition-all",
                                !unreadOnly ? "bg-primary text-white shadow-lg shadow-primary/20" : "text-foreground/40 hover:text-white"
                            )}
                        >
                            {t('filter_all') || 'All'}
                        </button>
                        <button
                            onClick={() => setUnreadOnly(true)}
                            className={cn(
                                "flex-1 py-3 rounded-xl text-sm font-bold transition-all",
                                unreadOnly ? "bg-primary text-white shadow-lg shadow-primary/20" : "text-foreground/40 hover:text-white"
                            )}
                        >
                            {t('filter_unread') || 'Unread'}
                        </button>
                    </div>
                </div>

                {pageCount > 1 && (
                    <div className="flex items-center justify-between bg-white/5 p-2 rounded-2xl border border-white/10">
                        <button
                            onClick={() => setPage(Math.max(1, page - 1))}
                            disabled={page === 1}
                            className="p-3 text-foreground/40 hover:text-white disabled:opacity-20 transition-all"
                        >
                            <ChevronLeft size={20} />
                        </button>

                        <div className="flex items-center gap-2 overflow-x-auto no-scrollbar py-1">
                            {Array.from({ length: pageCount }, (_, i) => i + 1).map((p) => (
                                <button
                                    key={p}
                                    onClick={() => setPage(p)}
                                    className={cn(
                                        "w-10 h-10 flex-shrink-0 rounded-xl text-sm font-bold transition-all",
                                        page === p ? "bg-primary text-white shadow-lg shadow-primary/20" : "text-foreground/40 hover:text-white hover:bg-white/5"
                                    )}
                                >
                                    {p}
                                </button>
                            ))}
                        </div>

                        <button
                            onClick={() => setPage(Math.min(pageCount, page + 1))}
                            disabled={page === pageCount}
                            className="p-3 text-foreground/40 hover:text-white disabled:opacity-20 transition-all"
                        >
                            <ChevronLeft size={20} className="rotate-180" />
                        </button>
                    </div>
                )}

                {/* List */}
                <div className="space-y-4">
                    {isLoading ? (
                        <div className="py-20 text-center animate-pulse text-foreground/40">
                            {tCommon('loading')}
                        </div>
                    ) : filteredNotifications.length === 0 ? (
                        <div className="py-32 text-center glass-card rounded-[32px] border-dashed border-white/10">
                            <div className="w-20 h-20 bg-white/5 rounded-full flex items-center justify-center mx-auto mb-6">
                                <Bell className="w-10 h-10 text-foreground/20" />
                            </div>
                            <h3 className="text-xl font-bold mb-2">{t('empty')}</h3>
                            <p className="text-foreground/40">{t('empty_desc') || 'We will notify you when something important happens.'}</p>
                        </div>
                    ) : (
                        <div className="space-y-4">
                            {filteredNotifications.map((notif, i) => (
                                <motion.div
                                    key={notif.id}
                                    initial={{ opacity: 0, y: 20 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    transition={{ delay: i * 0.05 }}
                                    onClick={() => !notif.is_read && markAsRead(notif.id)}
                                    className={cn(
                                        "glass-card p-6 rounded-[32px] group cursor-pointer transition-all border border-white/5 hover:border-white/20 relative overflow-hidden",
                                        !notif.is_read ? "bg-primary/5" : "bg-white/5 opacity-80"
                                    )}
                                >
                                    <div className="flex gap-6">
                                        <div className="flex-shrink-0 mt-1">
                                            <div className="w-14 h-14 rounded-2xl bg-white/5 flex items-center justify-center border border-white/5 group-hover:scale-110 transition-transform">
                                                {getNotificationIcon(notif.type)}
                                            </div>
                                        </div>
                                        <div className="flex-1 space-y-2">
                                            <div className="flex justify-between items-start">
                                                <div className="space-y-1">
                                                    {notif.sender_name && (
                                                        <p className="text-[10px] font-bold text-primary uppercase tracking-widest leading-none mb-1">
                                                            {notif.sender_name}
                                                        </p>
                                                    )}
                                                    <h3 className={cn("text-xl font-bold font-outfit", !notif.is_read ? "text-white" : "text-white/60")}>
                                                        {notif.title.startsWith('notif_') ? t(notif.title) : notif.title}
                                                    </h3>
                                                </div>
                                                <div className="flex items-center gap-2 text-foreground/30 text-xs font-medium">
                                                    <Clock size={14} />
                                                    <span suppressHydrationWarning>{new Date(notif.created_at).toLocaleString()}</span>
                                                </div>
                                            </div>
                                            <p className={cn("text-lg leading-relaxed", !notif.is_read ? "text-foreground/80" : "text-foreground/40")}>
                                                {notif.message.startsWith('notif_') ? t(notif.message) : notif.message}
                                            </p>

                                            {notif.link && (
                                                <div className="pt-2">
                                                    <Link
                                                        href={notif.link}
                                                        onClick={(e) => {
                                                            e.stopPropagation();
                                                            if (!notif.is_read) markAsRead(notif.id);
                                                        }}
                                                        className="inline-flex items-center gap-2 text-sm font-semibold text-primary hover:text-primary/80 transition-colors bg-primary/10 px-4 py-2 rounded-xl"
                                                    >
                                                        <span>{t('open_material')}</span>
                                                        <ExternalLink className="w-4 h-4" />
                                                    </Link>
                                                </div>
                                            )}
                                        </div>

                                        <div className="flex flex-col items-center gap-4">
                                            {!notif.is_read && (
                                                <div className="w-3 h-3 rounded-full bg-primary shadow-[0_0_15px_rgba(139,92,246,0.5)]" />
                                            )}
                                            <button
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    deleteNotification(notif.id);
                                                }}
                                                className="p-3 text-foreground/20 hover:text-red-500 hover:bg-red-500/10 rounded-2xl transition-all opacity-0 group-hover:opacity-100"
                                                title={tCommon('delete')}
                                            >
                                                <Trash2 size={20} />
                                            </button>
                                        </div>
                                    </div>
                                </motion.div>
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
