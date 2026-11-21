'use client';

import React from 'react';
import { motion } from 'framer-motion';
import { BookOpen, Star, Zap, Clock, ChevronRight, PlayCircle } from 'lucide-react';

export default function StudentDashboard() {
    const currentLearning = [
        { title: 'Newtonian Mechanics', subject: 'Physics', progress: 65, color: 'bg-primary' },
        { title: 'Organic Chemistry', subject: 'Chemistry', progress: 30, color: 'bg-secondary' },
        { title: 'Linear Algebra', subject: 'Mathematics', progress: 85, color: 'bg-accent' },
    ];

    return (
        <div className="space-y-10">
            {/* Daily Goal Card */}
            <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                className="glass-card p-8 rounded-3xl bg-gradient-to-r from-primary/10 to-transparent border-primary/20 flex flex-col md:flex-row justify-between items-center gap-8"
            >
                <div className="flex-1">
                    <div className="flex items-center gap-2 text-primary font-bold text-sm uppercase mb-2">
                        <Zap size={16} />
                        Keep the Streak
                    </div>
                    <h2 className="text-3xl font-bold font-outfit mb-4">Daily Goal: 30 Minutes</h2>
                    <p className="text-foreground/60 mb-6 max-w-md">You've completed 20 minutes today. Just 10 more to maintain your 5-day streak!</p>
                    <div className="h-3 w-full max-w-md bg-white/5 rounded-full overflow-hidden">
                        <div className="h-full bg-primary w-2/3 shadow-[0_0_20px_rgba(139,92,246,0.5)]" />
                    </div>
                </div>
                <button className="px-10 py-5 bg-white text-black font-bold rounded-2xl hover:bg-white/90 transition-all shadow-xl">
                    Resume Lesson
                </button>
            </motion.div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-10">
                <div className="lg:col-span-2 space-y-6">
                    <h3 className="text-2xl font-bold font-outfit">My Current Subjects</h3>
                    <div className="grid gap-6">
                        {currentLearning.map((item, i) => (
                            <motion.div
                                key={i}
                                initial={{ opacity: 0, x: -20 }}
                                animate={{ opacity: 1, x: 0 }}
                                transition={{ delay: i * 0.1 }}
                                whileHover={{ x: 5 }}
                                className="glass-card p-6 rounded-3xl flex items-center justify-between group cursor-pointer"
                            >
                                <div className="flex items-center gap-6">
                                    <div className={`w-14 h-14 rounded-2xl ${item.color}/20 flex items-center justify-center text-white`}>
                                        <PlayCircle size={32} className={`text-${item.color.split('-')[1]}`} />
                                    </div>
                                    <div>
                                        <h4 className="text-lg font-bold mb-1">{item.title}</h4>
                                        <p className="text-sm text-foreground/50">{item.subject}</p>
                                    </div>
                                </div>
                                <div className="flex items-center gap-6">
                                    <div className="text-right hidden sm:block">
                                        <p className="text-sm font-bold">{item.progress}%</p>
                                        <p className="text-xs text-foreground/40">Progress</p>
                                    </div>
                                    <ChevronRight className="text-foreground/20 group-hover:text-primary transition-colors" />
                                </div>
                            </motion.div>
                        ))}
                    </div>
                </div>

                {/* Sidebar content */}
                <div className="space-y-8">
                    <div className="glass-card p-8 rounded-3xl">
                        <h3 className="text-xl font-bold mb-6 font-outfit flex items-center gap-2">
                            <Star className="text-yellow-400" />
                            Achievements
                        </h3>
                        <div className="space-y-4">
                            {[
                                { title: 'Early Bird', icon: '🌅', desc: 'Login before 8 AM' },
                                { title: 'Problem Solver', icon: '🧩', desc: '100% score on quiz' },
                                { title: 'Fast Learner', icon: '⚡', desc: 'Finish 2 topics a day' },
                            ].map((badge, i) => (
                                <div key={i} className="flex items-center gap-4 p-3 bg-white/5 rounded-2xl border border-white/5 opacity-50 grayscale hover:opacity-100 hover:grayscale-0 cursor-default transition-all">
                                    <div className="text-2xl">{badge.icon}</div>
                                    <div>
                                        <p className="font-bold text-sm">{badge.title}</p>
                                        <p className="text-[10px] text-foreground/40">{badge.desc}</p>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
