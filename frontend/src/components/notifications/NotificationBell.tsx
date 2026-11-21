'use client';

import React, { useState, useRef, useEffect } from 'react';
import { Bell, Check, BellOff, Info, AlertTriangle, AlertCircle, Calendar, Sparkles, ExternalLink } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { useNotifications } from '@/hooks/useNotifications';
import { useTranslations } from 'next-intl';
import { NotificationType } from '@/types/notification';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { Link } from '@/i18n/routing';

function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

const getNotificationIcon = (type: NotificationType) => {
    switch (type) {
        case 'token_limit_warning':
        case 'student_limit_warning':
            return <AlertTriangle className="w-5 h-5 text-amber-500" />;
        case 'token_limit_reached':
        case 'subscription_expired':
            return <AlertCircle className="w-5 h-5 text-red-500" />;
        case 'subscription_expiring':
            return <Info className="w-5 h-5 text-blue-500" />;
        case 'monthly_reset':
            return <Calendar className="w-5 h-5 text-green-500" />;
        case 'material_ready':
        case 'whiteboard_summary_ready':
            return <Sparkles className="w-5 h-5 text-purple-500" />;
        case 'whiteboard_started':
            return <Bell className="w-5 h-5 text-primary animate-pulse" />;
        default:
            return <Bell className="w-5 h-5 text-primary" />;
    }
};

export default function NotificationBell() {
    const [isOpen, setIsOpen] = useState(false);
    const { notifications, unreadCount, markAsRead, markAllAsRead, isLoading } = useNotifications();
    const t = useTranslations('notifications');
    const dropdownRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    return (
        <div className="relative" ref={dropdownRef}>
            <button
                onClick={() => setIsOpen(!isOpen)}
                className="relative p-2 rounded-xl hover:bg-slate-100 dark:hover:bg-white/5 transition-all active:scale-95"
                aria-label="Notifications"
            >
                <Bell className={cn("w-5 h-5 transition-colors", unreadCount > 0 ? "text-primary" : "text-foreground/70")} />
                {unreadCount > 0 && (
                    <motion.span
                        initial={{ scale: 0 }}
                        animate={{ scale: 1 }}
                        className="absolute top-1 right-1 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-[10px] font-bold text-white"
                    >
                        {unreadCount > 9 ? '9+' : unreadCount}
                    </motion.span>
                )}
            </button>

            <AnimatePresence>
                {isOpen && (
                    <motion.div
                        initial={{ opacity: 0, y: 10, scale: 0.95 }}
                        animate={{ opacity: 1, y: 0, scale: 1 }}
                        exit={{ opacity: 0, y: 10, scale: 0.95 }}
                        className="absolute right-0 mt-3 w-80 md:w-96 glass-card overflow-hidden shadow-2xl z-[100] rounded-2xl"
                    >
                        <div className="p-4 border-b border-slate-200 dark:border-white/10 flex items-center justify-between bg-slate-50 dark:bg-slate-900/50 backdrop-blur-sm">
                            <h3 className="font-semibold">{t('title')}</h3>
                            {unreadCount > 0 && (
                                <button
                                    onClick={() => markAllAsRead()}
                                    className="text-xs text-primary hover:underline flex items-center gap-1"
                                >
                                    <Check className="w-3 h-3" />
                                    {t('mark_all_read')}
                                </button>
                            )}
                        </div>

                        <div className="max-h-[400px] overflow-y-auto scrollbar-hide bg-white dark:bg-[#0f172a]">
                            {notifications.length === 0 ? (
                                <div className="p-10 text-center flex flex-col items-center gap-3">
                                    <div className="w-12 h-12 rounded-full bg-slate-100 dark:bg-white/5 flex items-center justify-center">
                                        <BellOff className="w-6 h-6 text-foreground/30" />
                                    </div>
                                    <p className="text-sm text-foreground/50">{t('empty')}</p>
                                </div>
                            ) : (
                                <div className="flex flex-col">
                                    {notifications.slice(0, 6).map((notification) => (
                                        <div
                                            key={notification.id}
                                            className={cn(
                                                "p-4 border-b border-slate-200 dark:border-white/5 hover:bg-slate-50 dark:hover:bg-white/5 transition-colors cursor-pointer relative",
                                                !notification.is_read && "bg-primary/5"
                                            )}
                                            onClick={() => !notification.is_read && markAsRead(notification.id)}
                                        >
                                            <div className="flex gap-4">
                                                <div className="flex-shrink-0 mt-1">
                                                    {getNotificationIcon(notification.type)}
                                                </div>
                                                <div className="flex-1 space-y-1">
                                                    {notification.sender_name && (
                                                        <p className="text-[10px] font-bold text-primary uppercase tracking-widest leading-none mb-1">
                                                            {notification.sender_name}
                                                        </p>
                                                    )}
                                                    <p className={cn("text-sm font-medium leading-none", !notification.is_read ? "text-foreground" : "text-foreground/70")}>
                                                        {notification.title.startsWith('notif_') ? t(notification.title) : notification.title}
                                                    </p>
                                                    <p className="text-xs text-foreground/60 line-clamp-2">
                                                        {notification.message.startsWith('notif_') ? t(notification.message) : notification.message}
                                                    </p>
                                                    <p className="text-[10px] text-foreground/40">
                                                        {new Date(notification.created_at).toLocaleString()}
                                                    </p>
                                                    {notification.link && (
                                                        <div className="pt-2">
                                                            <Link
                                                                href={notification.link}
                                                                onClick={() => {
                                                                    setIsOpen(false);
                                                                    if (!notification.is_read) markAsRead(notification.id);
                                                                }}
                                                                className="inline-flex items-center gap-2 text-xs font-semibold text-primary hover:text-primary/80 transition-colors bg-primary/10 px-3 py-1.5 rounded-lg"
                                                            >
                                                                <span>{t('open_material')}</span>
                                                                <ExternalLink className="w-3 h-3" />
                                                            </Link>
                                                        </div>
                                                    )}
                                                </div>
                                                {!notification.is_read && (
                                                    <div className="w-2 h-2 rounded-full bg-primary flex-shrink-0 mt-1.5" />
                                                )}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>

                        <div className="p-3 border-t border-slate-200 dark:border-white/10 bg-slate-50 dark:bg-white/5">
                            <Link
                                href="/notifications"
                                onClick={() => setIsOpen(false)}
                                className="block w-full text-center text-xs font-medium text-foreground/60 hover:text-primary transition-colors py-2"
                            >
                                {t('view_all')}
                            </Link>
                        </div>
                    </motion.div>
                )}
            </AnimatePresence>
        </div>
    );
}
