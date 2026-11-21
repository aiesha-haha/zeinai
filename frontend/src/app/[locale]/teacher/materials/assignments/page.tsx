'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { TestTube, GraduationCap, ClipboardList } from 'lucide-react';
import ToolGrid from '@/components/materials/ToolGrid';

export default function AssignmentsPage() {
    const t = useTranslations('education_materials');
    const d = useTranslations('tool_desc');
    const tShell = useTranslations('shell');
    return (
        <ToolGrid
            title={t('sections.assignments')}
            tools={[
                { href: '/teacher/materials/tests', name: t('tools.test_generation'), desc: d('tests'), icon: <TestTube size={24} /> },
                { href: '/teacher/materials/pisa', name: t('tools.pisa_tests'), desc: d('pisa'), icon: <GraduationCap size={24} /> },
                { href: '/teacher/materials/unt', name: t('tools.unt_tests'), desc: d('unt'), icon: <GraduationCap size={24} /> },
                { href: '/teacher/tests/assignments', name: tShell('assigned_results'), desc: d('results'), icon: <ClipboardList size={24} /> },
            ]}
        />
    );
}
