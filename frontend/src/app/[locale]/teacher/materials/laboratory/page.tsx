'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { Microscope } from 'lucide-react';
import ToolGrid from '@/components/materials/ToolGrid';

export default function LaboratoryPage() {
    const t = useTranslations('education_materials');
    const d = useTranslations('tool_desc');
    return (
        <ToolGrid
            title={t('sections.laboratory')}
            tools={[{ href: '/teacher/materials/lab', name: t('tools.lab_work'), desc: d('lab'), icon: <Microscope size={24} /> }]}
        />
    );
}
