'use client';

import React, { Suspense } from 'react';
import { useTranslations } from 'next-intl';
import StudioTool from '@/components/studio/StudioTool';
import { CertificateView } from '@/components/studio/Results';
import { languageField } from '@/components/studio/fields';

function CertificateTool() {
    const t = useTranslations('studio');
    const tM = useTranslations('education_materials');
    return (
        <StudioTool
            kind="certificate"
            title={tM('tools.certificates')}
            description={t('cert_desc')}
            backHref="/teacher/materials/achievements-category"
            submitLabel={t('cert_submit')}
            fields={[
                { name: 'student_name', label: t('cert_student'), type: 'text', required: true, placeholder: t('cert_student_ph') },
                { name: 'achievement', label: t('cert_achievement'), type: 'textarea', required: true, placeholder: t('cert_achievement_ph') },
                { name: 'teacher_name', label: t('cert_teacher'), type: 'text' },
                { name: 'school', label: t('cert_school'), type: 'text' },
                languageField((k) => t(k)),
            ]}
            renderResult={(item, h) => <CertificateView item={item} h={h} />}
        />
    );
}

export default function Page() {
    return <Suspense><CertificateTool /></Suspense>;
}
