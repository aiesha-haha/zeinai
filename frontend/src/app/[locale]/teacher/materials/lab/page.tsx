'use client';

import React, { Suspense } from 'react';
import { useTranslations } from 'next-intl';
import StudioTool from '@/components/studio/StudioTool';
import { SectionsEditor } from '@/components/studio/Results';
import { gradeOptions, languageField } from '@/components/studio/fields';

function LabTool() {
    const t = useTranslations('studio');
    const tM = useTranslations('education_materials');
    const tCommon = useTranslations('common');
    return (
        <StudioTool
            kind="lab"
            title={tM('tools.lab_work')}
            description={t('lab_desc')}
            backHref="/teacher/materials/laboratory"
            submitLabel={t('lab_submit')}
            fields={[
                { name: 'subject', label: t('subject'), type: 'text', required: true, placeholder: t('subject_ph') },
                { name: 'grade', label: t('grade'), type: 'select', defaultValue: '8', options: gradeOptions(tCommon('grade_suffix')) },
                { name: 'topic', label: t('lesson_topic'), type: 'text', required: true, placeholder: t('lab_topic_ph') },
                { name: 'details', label: t('extra_wishes'), type: 'textarea', placeholder: t('lab_details_ph') },
                languageField((k) => t(k)),
            ]}
            renderResult={(item, h) => <SectionsEditor item={item} h={h} />}
        />
    );
}

export default function Page() {
    return <Suspense><LabTool /></Suspense>;
}
