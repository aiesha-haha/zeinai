'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { GraduationCap } from 'lucide-react';
import TestTypeHub from '@/components/materials/TestTypeHub';

export default function UntPage() {
    const tM = useTranslations('education_materials');
    const tTypes = useTranslations('test_types');
    return (
        <TestTypeHub
            types={['unt']}
            title={tM('tools.unt_tests')}
            description={tTypes('unt.desc')}
            backHref="/teacher/materials/assignments"
            selfHref="/teacher/materials/unt"
            createType="unt"
            createLabel={tTypes('create_unt')}
            emptyText={tTypes('empty_unt')}
            icon={<GraduationCap size={20} />}
        />
    );
}
