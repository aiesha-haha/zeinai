'use client';

import React, { Suspense } from 'react';
import { useTranslations } from 'next-intl';
import StudioTool from '@/components/studio/StudioTool';
import { MediaResult } from '@/components/studio/Results';

function VideoTool() {
    const t = useTranslations('studio');
    const tM = useTranslations('education_materials');
    return (
        <StudioTool
            kind="video"
            title={tM('tools.video')}
            description={t('video_desc')}
            backHref="/teacher/materials/generation"
            submitLabel={t('video_submit')}
            busyHint={t('video_busy')}
            fields={[
                { name: 'title', label: t('video_title'), type: 'text', required: true, placeholder: t('video_title_ph') },
                { name: 'description', label: t('video_description'), type: 'textarea', required: true, placeholder: t('video_description_ph'), help: t('video_description_help') },
                {
                    name: 'seconds', label: t('video_length'), type: 'select', defaultValue: '8',
                    options: [
                        { value: '4', label: t('seconds', { n: 4 }) },
                        { value: '8', label: t('seconds', { n: 8 }) },
                        { value: '12', label: t('seconds', { n: 12 }) },
                    ],
                },
                {
                    name: 'orientation', label: t('video_orientation'), type: 'select', defaultValue: 'landscape',
                    options: [
                        { value: 'landscape', label: t('size_landscape') },
                        { value: 'portrait', label: t('size_portrait') },
                    ],
                },
            ]}
            renderResult={(item, h) => <MediaResult item={item} h={h} />}
        />
    );
}

export default function Page() {
    return <Suspense><VideoTool /></Suspense>;
}
