'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { Trophy } from 'lucide-react';
import TestTypeHub from '@/components/materials/TestTypeHub';

export default function OlympiadPage() {
    const tM = useTranslations('education_materials');
    const tTypes = useTranslations('test_types');
    return (
        <TestTypeHub
            types={['olympiad']}
            title={tM('tools.olympiad_tasks')}
            description={tTypes('olympiad.desc')}
            backHref="/teacher"
            selfHref="/teacher/materials/olympiad"
            createType="olympiad"
            createLabel={tTypes('create_olympiad')}
            emptyText={tTypes('empty_olympiad')}
            icon={<Trophy size={20} />}
        />
    );
}
