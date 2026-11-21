'use client';

import React, { Suspense } from 'react';
import { useTranslations } from 'next-intl';
import StudioTool from '@/components/studio/StudioTool';
import { FeedbackEditor } from '@/components/studio/Results';
import { gradeOptions, languageField } from '@/components/studio/fields';

const PURPOSES = ['exit_ticket', 'self_assessment', 'peer_assessment', 'two_stars_wish', 'traffic_light', 'reflection'];

function FeedbackTool() {
    const t = useTranslations('studio');
    const tM = useTranslations('education_materials');
    const tCommon = useTranslations('common');
    return (
        <StudioTool
            kind="feedback"
            title={tM('tools.feedback')}
            description={t('feedback_desc')}
            backHref="/teacher/materials/generation"
            submitLabel={t('feedback_submit')}
            fields={[
                { name: 'class_id', label: t('class'), type: 'class' },
                { name: 'grade', label: t('grade'), type: 'select', defaultValue: '7', options: gradeOptions(tCommon('grade_suffix')) },
                { name: 'topic', label: t('lesson_topic'), type: 'text', required: true, placeholder: t('lesson_topic_ph') },
                {
                    name: 'purpose', label: t('feedback_purpose'), type: 'select', required: true, defaultValue: 'exit_ticket',
                    options: PURPOSES.map((v) => ({ value: v, label: t(`purpose_${v}`) })),
                    help: t('feedback_purpose_help'),
                },
                { name: 'question_count', label: t('question_count'), type: 'number', defaultValue: '4', min: 2, max: 8, help: t('question_count_help') },
                languageField((k) => t(k)),
            ]}
            renderResult={(item, h) => <FeedbackEditor item={item} h={h} />}
        />
    );
}

export default function Page() {
    return <Suspense><FeedbackTool /></Suspense>;
}
