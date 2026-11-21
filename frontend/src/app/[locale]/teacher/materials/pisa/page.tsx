'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { GraduationCap } from 'lucide-react';
import TestTypeHub from '@/components/materials/TestTypeHub';

export default function PisaPage() {
    const tM = useTranslations('education_materials');
    const tTypes = useTranslations('test_types');
    return (
        <TestTypeHub
            types={['pisa']}
            title={tM('tools.pisa_tests')}
            description={tTypes('pisa.desc')}
            backHref="/teacher/materials/assignments"
            selfHref="/teacher/materials/pisa"
            createType="pisa"
            createLabel={tTypes('create_pisa')}
            emptyText={tTypes('empty_pisa')}
            icon={<GraduationCap size={20} />}
        />
    );
}
