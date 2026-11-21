'use client';

import React, { useState } from 'react';
import { useTranslations } from 'next-intl';
import { useRouter } from '@/i18n/routing';
import { ChevronLeft } from 'lucide-react';
import FaceSelectorModal from '@/components/Whiteboard/FaceSelectorModal';
import { motion } from 'framer-motion';
import { Camera } from 'lucide-react';

export default function FacePickerPage() {
    const t = useTranslations('education_materials');
    const tw = useTranslations('whiteboard_teacher');
    const router = useRouter();
    const [open, setOpen] = useState(false);

    return (
        <div className="space-y-8">
            <div className="flex items-center gap-4">
                <button
                    onClick={() => router.push('/teacher/materials/interactive')}
                    className="p-2 bg-white/5 rounded-xl hover:bg-white/10 transition-all text-foreground/60"
                >
                    <ChevronLeft size={20} />
                </button>
                <h1 className="text-3xl font-bold font-outfit">{t('tools.face_picker')}</h1>
            </div>

            <motion.div
                initial={{ opacity: 0, y: 24 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.4 }}
                className="flex flex-col items-center justify-center gap-8 py-16"
            >
                <div className="w-28 h-28 rounded-3xl bg-violet-600 flex items-center justify-center shadow-2xl shadow-violet-500/30">
                    <Camera size={52} className="text-white" />
                </div>

                <div className="text-center space-y-2 max-w-md">
                    <h2 className="text-2xl font-bold">{tw('face_picker_title')}</h2>
                    <p className="text-foreground/50 text-sm leading-relaxed">
                        Камера арқылы оқушылардың беттерін анықтап, кездейсоқ оқушыны таңдайды. Слот-машина анимациясы және конфетти 🎉
                    </p>
                </div>

                <button
                    onClick={() => setOpen(true)}
                    className="flex items-center gap-3 px-8 py-4 bg-violet-600 hover:bg-violet-700 text-white rounded-2xl text-lg font-semibold transition-all shadow-lg shadow-violet-500/30 hover:scale-105 active:scale-95"
                >
                    <Camera size={22} />
                    {tw('face_picker_start')}
                </button>
            </motion.div>

            <FaceSelectorModal isOpen={open} onClose={() => setOpen(false)} />
        </div>
    );
}
