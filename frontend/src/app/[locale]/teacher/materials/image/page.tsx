'use client';

import React, { Suspense } from 'react';
import { useTranslations } from 'next-intl';
import StudioTool from '@/components/studio/StudioTool';
import { MediaResult } from '@/components/studio/Results';

function ImageTool() {
    const t = useTranslations('studio');
    const tM = useTranslations('education_materials');
    return (
        <StudioTool
            kind="image"
            title={tM('tools.image')}
            description={t('image_desc')}
            backHref="/teacher/materials/generation"
            submitLabel={t('image_submit')}
            busyHint={t('image_busy')}
            fields={[
                { name: 'prompt', label: t('image_prompt'), type: 'textarea', required: true, placeholder: t('image_prompt_ph') },
                {
                    name: 'style', label: t('image_style'), type: 'select', defaultValue: '',
                    options: [
                        { value: '', label: t('style_any') },
                        { value: 'flat educational illustration', label: t('style_illustration') },
                        { value: 'realistic photo', label: t('style_photo') },
                        { value: 'simple diagram with labels', label: t('style_diagram') },
                        { value: 'cartoon for kids', label: t('style_cartoon') },
                    ],
                },
                {
                    name: 'size', label: t('image_size'), type: 'select', defaultValue: '1024x1024',
                    options: [
                        { value: '1024x1024', label: t('size_square') },
                        { value: '1536x1024', label: t('size_landscape') },
                        { value: '1024x1536', label: t('size_portrait') },
                    ],
                },
            ]}
            renderResult={(item, h) => <MediaResult item={item} h={h} />}
        />
    );
}

export default function Page() {
    return <Suspense><ImageTool /></Suspense>;
}
