'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { FileText, Presentation, Video, Image as ImageIcon, MessageSquare, PenTool } from 'lucide-react';
import ToolGrid from '@/components/materials/ToolGrid';

export default function GenerationPage() {
    const t = useTranslations('education_materials');
    const d = useTranslations('tool_desc');
    return (
        <ToolGrid
            title={t('sections.generation')}
            tools={[
                { href: '/teacher/materials/short-term-plan', name: t('tools.short_term_plan'), desc: d('kmj'), icon: <FileText size={24} /> },
                { href: '/teacher/materials/slides', name: t('tools.slides'), desc: d('slides'), icon: <Presentation size={24} /> },
                { href: '/teacher/materials/article', name: t('tools.article'), desc: d('article'), icon: <PenTool size={24} /> },
                { href: '/teacher/materials/video', name: t('tools.video'), desc: d('video'), icon: <Video size={24} /> },
                { href: '/teacher/materials/image', name: t('tools.image'), desc: d('image'), icon: <ImageIcon size={24} /> },
                { href: '/teacher/materials/feedback', name: t('tools.feedback'), desc: d('feedback'), icon: <MessageSquare size={24} /> },
            ]}
        />
    );
}
