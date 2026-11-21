'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { Award, BadgeCheck } from 'lucide-react';
import ToolGrid from '@/components/materials/ToolGrid';

export default function AchievementsCategoryPage() {
    const t = useTranslations('education_materials');
    const d = useTranslations('tool_desc');
    return (
        <ToolGrid
            title={t('sections.achievements')}
            tools={[
                { href: '/teacher/materials/achievements', name: t('tools.achievements'), desc: d('achievements'), icon: <Award size={24} /> },
                { href: '/teacher/materials/certificates', name: t('tools.certificates'), desc: d('certificates'), icon: <BadgeCheck size={24} /> },
            ]}
        />
    );
}
