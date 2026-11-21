'use client';

import React, { Suspense } from 'react';
import { useTranslations } from 'next-intl';
import StudioTool from '@/components/studio/StudioTool';
import { SectionsEditor } from '@/components/studio/Results';
import { languageField } from '@/components/studio/fields';

function NewsTool() {
    const t = useTranslations('studio');
    const tM = useTranslations('education_materials');
    return (
        <StudioTool
            kind="news"
            title={tM('tools.news')}
            description={t('news_desc')}
            backHref="/teacher/materials/news-category"
            submitLabel={t('news_submit')}
            fields={[
                { name: 'topic', label: t('news_topic'), type: 'text', required: true, placeholder: t('news_topic_ph') },
                { name: 'details', label: t('news_details'), type: 'textarea', placeholder: t('news_details_ph'), help: t('news_details_help') },
                {
                    name: 'audience', label: t('news_audience'), type: 'select', defaultValue: 'parents',
                    options: [
                        { value: 'parents', label: t('aud_parents') },
                        { value: 'students', label: t('aud_students') },
                        { value: 'school website', label: t('aud_site') },
                    ],
                },
                languageField((k) => t(k)),
            ]}
            renderResult={(item, h) => <SectionsEditor item={item} h={h} />}
        />
    );
}

export default function Page() {
    return <Suspense><NewsTool /></Suspense>;
}
