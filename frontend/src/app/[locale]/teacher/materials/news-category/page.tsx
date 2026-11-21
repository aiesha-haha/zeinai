'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { Newspaper } from 'lucide-react';
import ToolGrid from '@/components/materials/ToolGrid';

export default function NewsCategoryPage() {
    const t = useTranslations('education_materials');
    const d = useTranslations('tool_desc');
    return (
        <ToolGrid
            title={t('sections.news')}
            tools={[{ href: '/teacher/materials/news', name: t('tools.news'), desc: d('news'), icon: <Newspaper size={24} /> }]}
        />
    );
}
