"use client";
import React from 'react';
import { useState, useEffect } from "react";
import { useTranslations, useLocale } from "next-intl";
import api from "@/lib/api";
import { Loader2, FileText, Lightbulb, Clock, Calendar } from "lucide-react";
import { motion } from "framer-motion";
import { WhiteboardSummary } from "@/types/whiteboard";
import { usePathname } from "next/navigation";
import Link from "next/link";

interface SummaryViewProps {
    sessionId: string;
}

export default function SummaryView({ sessionId }: SummaryViewProps) {
    const t = useTranslations('whiteboard_summary');
    const locale = useLocale();
    const pathname = usePathname();
    const [summary, setSummary] = useState<WhiteboardSummary | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        let interval: NodeJS.Timeout;

        const load = async () => {
            const success = await fetchSummary();
            if (!success) {
                // Poll every 5 seconds if not found
                interval = setInterval(async () => {
                    const found = await fetchSummary();
                    if (found) clearInterval(interval);
                }, 5000);
            }
        };

        load();
        return () => {
            if (interval) clearInterval(interval);
        };
    }, [sessionId]);

    const fetchSummary = async () => {
        try {
            const res = await api.get(`/whiteboard/summary/${sessionId}`);
            if (res.data && res.data.data) {
                setSummary(res.data.data);
                return true;
            }
            return false;
        } catch (err: any) {
            if (err.response?.status !== 404) {
                console.error("Failed to fetch summary", err);
            }
            return false;
        } finally {
            setLoading(false);
        }
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[60vh]">
                <Loader2 className="w-8 h-8 animate-spin text-zinc-500" />
            </div>
        );
    }

    if (!summary) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[60vh] text-center p-4">
                <Loader2 className="w-12 h-12 text-blue-500 animate-pulse mb-4" />
                <h2 className="text-xl font-bold dark:text-white">{t('generating')}</h2>
                <p className="text-zinc-500">{t('generating_desc')}</p>
            </div>
        );
    }

    // Determine which summary/explanation to show based on locale
    const localizedSummary = locale === 'kk' ? summary.summary_kk : (locale === 'en' ? summary.summary_en : summary.summary_text);
    const localizedExplanation = locale === 'kk' ? summary.explanation_kk : (locale === 'en' ? summary.explanation_en : summary.explanation);

    // Fallback if AI failed some languages
    const finalSummary = localizedSummary || summary.summary_text;
    const finalExplanation = localizedExplanation || summary.explanation;

    return (
        <div className="max-w-4xl mx-auto p-6 space-y-8">
            <div className="space-y-2">
                <h1 className="text-3xl font-extrabold dark:text-white">
                    {summary.session_title || t('title')}
                </h1>
                <div className="flex items-center gap-4 text-zinc-500 text-sm">
                    <div className="flex items-center gap-1">
                        <Calendar className="w-4 h-4" />
                        {new Date(summary.created_at).toLocaleDateString()}
                    </div>
                    <div className="flex items-center gap-1">
                        <Clock className="w-4 h-4" />
                        {new Date(summary.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                    </div>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    className="md:col-span-2 space-y-4"
                >
                    <div className="p-6 bg-white dark:bg-zinc-900 rounded-2xl shadow-sm border border-zinc-200 dark:border-zinc-800">
                        <div className="flex items-center gap-2 mb-4 text-blue-600 dark:text-blue-400">
                            <FileText className="w-5 h-5" />
                            <h2 className="text-xl font-bold">{t('notes_title')}</h2>
                        </div>
                        <div className="prose prose-zinc dark:prose-invert max-w-none">
                            {finalSummary ? finalSummary.split('\n').map((line: string, i: number) => (
                                <p key={i}>{line}</p>
                            )) : <p className="italic text-zinc-500">{t('no_summary_available')}</p>}
                        </div>
                    </div>
                </motion.div>

                <motion.div
                    initial={{ opacity: 0, x: 20 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: 0.1 }}
                    className="space-y-6"
                >
                    <div className="p-6 bg-amber-50 dark:bg-amber-900/10 rounded-2xl border border-amber-200 dark:border-amber-900/30">
                        <div className="flex items-center gap-2 mb-3 text-amber-700 dark:text-amber-400">
                            <Lightbulb className="w-5 h-5" />
                            <h3 className="font-bold">{t('explanation_title')}</h3>
                        </div>
                        <p className="text-sm text-amber-900 dark:text-amber-200 leading-relaxed">
                            {finalExplanation || t('no_explanation_available')}
                        </p>
                    </div>

                    <div className="p-6 bg-zinc-100 dark:bg-zinc-800/50 rounded-2xl text-center">
                        <p className="text-sm text-zinc-500 mb-2">{t('view_board_hint')}</p>
                        <Link
                            href={`${pathname.includes('/teacher') ? '/teacher' : '/student'}/whiteboard/archive/${sessionId}`}
                            className="text-sm font-medium text-blue-600 hover:underline inline-block"
                        >
                            {t('view_board_btn')}
                        </Link>
                    </div>
                </motion.div>
            </div>
        </div>
    );
}
