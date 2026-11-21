"use client";

import React, { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { useRouter } from "@/i18n/routing";
import api from "@/lib/api";
import { motion } from "framer-motion";
import { ChevronLeft, GraduationCap, ChevronRight, Loader2 } from "lucide-react";

export default function WhiteboardClassSelector() {
    const t = useTranslations("classes");
    const tWhiteboard = useTranslations("education_materials");
    const router = useRouter();
    const [classes, setClasses] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetchClasses();
    }, []);

    const fetchClasses = async () => {
        try {
            const res = await api.get("/classes");
            setClasses(res.data.data || []);
        } catch (err) {
            console.error("Failed to fetch classes", err);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="space-y-8 p-4 md:p-8 max-w-5xl mx-auto">
            <div className="flex items-center gap-4">
                <button
                    onClick={() => router.back()}
                    className="p-2 bg-white/5 rounded-xl hover:bg-white/10 transition-all text-foreground/60"
                >
                    <ChevronLeft size={20} />
                </button>
                <div>
                    <h1 className="text-3xl font-bold font-outfit">
                        {tWhiteboard("tools.whiteboard")}
                    </h1>
                    <p className="text-foreground/40 text-sm mt-1">
                        Выберите класс, чтобы начать онлайн-урок
                    </p>
                </div>
            </div>

            {loading ? (
                <div className="flex justify-center p-12">
                    <Loader2 className="w-8 h-8 text-primary animate-spin" />
                </div>
            ) : classes.length === 0 ? (
                <div className="text-center p-12 glass-card rounded-3xl">
                    <p className="text-foreground/40">У вас пока нет активных классов.</p>
                    <button
                        onClick={() => router.push("/teacher")}
                        className="mt-4 text-primary font-bold hover:underline"
                    >
                        Создать класс в кабинете
                    </button>
                </div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {classes.map((cls, idx) => (
                        <motion.div
                            key={cls.id}
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ delay: idx * 0.05 }}
                            onClick={() => router.push(`/teacher/whiteboard/${cls.id}`)}
                            className="glass-card hover:border-primary/50 transition-all cursor-pointer group p-5 rounded-2xl flex items-center justify-between"
                        >
                            <div className="flex items-center gap-4">
                                <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center text-primary group-hover:bg-primary group-hover:text-white transition-all">
                                    <GraduationCap size={24} />
                                </div>
                                <div>
                                    <h3 className="font-bold text-lg">{cls.name}</h3>
                                    <p className="text-xs text-foreground/40">
                                        {cls.students_count || 0} учеников
                                    </p>
                                </div>
                            </div>
                            <ChevronRight className="text-foreground/20 group-hover:text-primary transition-all" size={20} />
                        </motion.div>
                    ))}
                </div>
            )}
        </div>
    );
}
