'use client';

import React from 'react';
import { motion } from 'framer-motion';
import { useTranslations } from 'next-intl';
import { useRouter } from 'next/navigation';
import { Home, AlertCircle, ChevronLeft } from 'lucide-react';

export default function NotFound() {
    const t = useTranslations('page_not_found');
    const router = useRouter();

    return (
        <div className="min-h-screen flex items-center justify-center p-6 bg-gradient-premium relative overflow-hidden">
            {/* Background Decorations */}
            <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-primary/20 blur-[120px] rounded-full animate-pulse" />
            <div className="absolute bottom-1/4 right-1/4 w-[500px] h-[500px] bg-secondary/20 blur-[160px] rounded-full delay-1000 animate-pulse" />

            <div className="relative z-10 max-w-2xl w-full text-center space-y-12">
                {/* 404 Visual */}
                <motion.div
                    initial={{ opacity: 0, scale: 0.8 }}
                    animate={{ opacity: 1, scale: 1 }}
                    transition={{ duration: 0.6 }}
                    className="relative inline-block"
                >
                    <h1 className="text-[150px] md:text-[220px] font-black font-outfit leading-none tracking-tighter opacity-10 select-none">
                        404
                    </h1>
                    <div className="absolute inset-0 flex items-center justify-center">
                        <motion.div
                            animate={{
                                y: [0, -20, 0],
                            }}
                            transition={{
                                duration: 4,
                                repeat: Infinity,
                                ease: "easeInOut"
                            }}
                            className="w-24 h-24 md:w-32 md:h-32 bg-primary/10 rounded-3xl flex items-center justify-center border border-primary/20 backdrop-blur-xl shadow-2xl shadow-primary/20"
                        >
                            <AlertCircle size={48} className="text-primary md:hidden" />
                            <AlertCircle size={64} className="text-primary hidden md:block" />
                        </motion.div>
                    </div>
                </motion.div>

                {/* Content */}
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: 0.3, duration: 0.6 }}
                    className="space-y-4"
                >
                    <h2 className="text-4xl md:text-5xl font-bold font-outfit">{t('title')}</h2>
                    <p className="text-lg md:text-xl text-foreground/60 max-w-md mx-auto leading-relaxed">
                        {t('desc')}
                    </p>
                </motion.div>

                {/* Actions */}
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: 0.5, duration: 0.6 }}
                    className="flex flex-col md:flex-row items-center justify-center gap-4"
                >
                    <button
                        onClick={() => router.back()}
                        className="w-full md:w-auto px-8 py-4 glass-card border-white/10 rounded-2xl font-bold hover:bg-white/5 transition-all flex items-center justify-center gap-2 group"
                    >
                        <ChevronLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
                        <span>Go Back</span>
                    </button>

                    <button
                        onClick={() => router.push('/')}
                        className="w-full md:w-auto px-8 py-4 bg-primary text-white rounded-2xl font-bold hover:bg-primary-hover transition-all flex items-center justify-center gap-2 shadow-lg shadow-primary/30 group"
                    >
                        <Home className="w-5 h-5 group-hover:scale-110 transition-transform" />
                        <span>{t('btn')}</span>
                    </button>
                </motion.div>

                {/* Branding */}
                <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 0.4 }}
                    transition={{ delay: 0.8 }}
                    className="pt-12"
                >
                    <p className="text-sm font-outfit uppercase tracking-widest font-bold">ZeinAI Platform</p>
                </motion.div>
            </div>
        </div>
    );
}
