'use client';

import React, { useState } from 'react';
import { Send, X, AlertCircle, CheckCircle2 } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { useTranslations } from 'next-intl';
import api from '@/lib/api';

interface SendNotificationModalProps {
    isOpen: boolean;
    onClose: () => void;
    targetType: 'student' | 'class' | 'all';
    targetId?: string | number;
    targetName?: string;
}

export default function SendNotificationModal({
    isOpen,
    onClose,
    targetType,
    targetId,
    targetName,
}: SendNotificationModalProps) {
    const t = useTranslations('notifications.send');
    const tCommon = useTranslations('common');

    const [title, setTitle] = useState('');
    const [message, setMessage] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState(false);

    const handleSend = async () => {
        if (!title || !message) return;

        setIsLoading(true);
        setError(null);

        try {
            let endpoint = '';
            let payload: any = { title, message };

            switch (targetType) {
                case 'student':
                    endpoint = '/teacher/notify-student';
                    payload.student_id = targetId;
                    break;
                case 'class':
                    endpoint = '/teacher/notify-class';
                    payload.class_id = Number(targetId);
                    break;
                case 'all':
                    endpoint = '/teacher/notify-all-students';
                    break;
            }

            await api.post(endpoint, payload);
            setSuccess(true);
            setTimeout(() => {
                setSuccess(false);
                setTitle('');
                setMessage('');
                onClose();
            }, 2000);
        } catch (err: any) {
            setError(err.response?.data?.error?.message || tCommon('generic'));
        } finally {
            setIsLoading(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
            <motion.div
                initial={{ opacity: 0, scale: 0.95, y: 20 }}
                animate={{ opacity: 1, scale: 1, y: 0 }}
                className="glass-card p-5 md:p-8 rounded-[32px] w-full max-w-lg border-white/10 bg-white dark:bg-[#0f172a] shadow-2xl relative max-h-[90vh] overflow-y-auto"
            >
                <div className="flex justify-between items-center mb-6">
                    <h3 className="text-2xl font-bold font-outfit">
                        {t('title')}
                    </h3>
                    <button
                        onClick={onClose}
                        className="p-2 text-foreground/40 hover:text-white hover:bg-white/5 rounded-full transition-all"
                    >
                        <X size={24} />
                    </button>
                </div>

                {success ? (
                    <div className="py-12 flex flex-col items-center justify-center text-center space-y-4">
                        <div className="w-20 h-20 bg-green-500/10 rounded-full flex items-center justify-center text-green-500 border border-green-500/20">
                            <CheckCircle2 size={40} />
                        </div>
                        <div>
                            <h4 className="text-xl font-bold text-white">{t('success_title')}</h4>
                            <p className="text-foreground/60">{t('success_msg')}</p>
                        </div>
                    </div>
                ) : (
                    <div className="space-y-6">
                        <div className="px-4 py-3 bg-white/5 rounded-2xl border border-white/5 flex items-center gap-3">
                            <div className="w-2 h-2 rounded-full bg-primary" />
                            <p className="text-sm font-medium">
                                {t('targeting')}: <span className="text-primary">{targetType === 'all' ? t('all_students') : targetName}</span>
                            </p>
                        </div>

                        <div className="space-y-2">
                            <label className="text-[10px] text-foreground/40 uppercase tracking-widest font-bold ml-1">
                                {t('subject_label')}
                            </label>
                            <input
                                type="text"
                                value={title}
                                onChange={(e) => setTitle(e.target.value)}
                                placeholder={t('subject_placeholder')}
                                className="w-full bg-white/5 border border-white/10 rounded-2xl px-5 py-4 outline-none focus:border-primary/50 transition-all font-medium"
                            />
                        </div>

                        <div className="space-y-2">
                            <label className="text-[10px] text-foreground/40 uppercase tracking-widest font-bold ml-1">
                                {t('message_label')}
                            </label>
                            <textarea
                                value={message}
                                onChange={(e) => setMessage(e.target.value)}
                                placeholder={t('message_placeholder')}
                                rows={4}
                                className="w-full bg-white/5 border border-white/10 rounded-2xl px-5 py-4 outline-none focus:border-primary/50 transition-all font-medium resize-none"
                            />
                        </div>

                        {error && (
                            <div className="flex items-center gap-2 p-4 bg-red-500/10 border border-red-500/20 rounded-2xl text-red-500 text-sm">
                                <AlertCircle size={18} />
                                <p>{error}</p>
                            </div>
                        )}

                        <button
                            onClick={handleSend}
                            disabled={isLoading || !title || !message}
                            className="w-full py-4 bg-primary text-white hover:bg-primary-hover rounded-2xl font-bold text-lg shadow-lg shadow-primary/20 transition-all disabled:opacity-50 flex items-center justify-center gap-3"
                        >
                            {isLoading ? (
                                <div className="w-6 h-6 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                            ) : (
                                <>
                                    <Send size={20} />
                                    {t('send_button')}
                                </>
                            )}
                        </button>
                    </div>
                )}
            </motion.div>
        </div>
    );
}
