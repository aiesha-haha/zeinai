'use client';

import React from 'react';
import { useParams } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { useRouter } from '@/i18n/routing';
import { Loader2, AlertCircle } from 'lucide-react';
import api from '@/lib/api';
import GamePlayer, { ResultsReview } from '@/components/game/GamePlayer';

export default function StudentTestDetailPage() {
    const params = useParams();
    const id = params.id as string;
    const t = useTranslations('game');
    const tCommon = useTranslations('common');
    const router = useRouter();
    const [data, setData] = React.useState<any>(null);
    const [loading, setLoading] = React.useState(true);
    const [failed, setFailed] = React.useState(false);

    const load = React.useCallback(async () => {
        try {
            const res = await api.get(`/student/tests/${id}`);
            setData(res.data.data);
            setFailed(false);
        } catch {
            setFailed(true);
        } finally {
            setLoading(false);
        }
    }, [id]);

    React.useEffect(() => { if (id) load(); }, [id, load]);

    if (loading) {
        return (
            <div className="flex items-center justify-center py-20" role="status">
                <Loader2 className="w-10 h-10 text-primary animate-spin" />
                <span className="sr-only">{tCommon('loading')}</span>
            </div>
        );
    }

    if (failed || !data) {
        return (
            <div className="glass-card rounded-3xl p-10 text-center space-y-4 max-w-lg mx-auto">
                <AlertCircle className="mx-auto text-red-600" size={36} />
                <p>{t('load_failed')}</p>
                <button className="btn-primary" onClick={() => router.push('/student/tests')}>{t('back_to_tasks')}</button>
            </div>
        );
    }

    const test = data.test;
    const questions = (data.questions || []).map((q: any) => ({
        id: q.id, question: q.question, options: q.options || [], answer: q.answer || undefined, explanation: q.explanation,
    }));

    if (data.status === 'completed') {
        return (
            <ResultsReview
                questions={questions}
                answers={data.student_answers || {}}
                score={data.score ?? 0}
                maxScore={data.max_score ?? questions.length}
                testType={test.test_type}
                onExit={() => router.push('/student/tests')}
                exitLabel={t('back_to_tasks')}
            />
        );
    }

    return (
        <GamePlayer
            mode="student"
            title={test.topic}
            subtitle={`${test.subject} · ${test.grade} ${tCommon('grade_suffix')}`}
            testType={test.test_type}
            questions={questions}
            exitLabel={t('back_to_tasks')}
            onExit={() => router.push('/student/tests')}
            onSubmit={async (answers) => {
                await api.post(`/student/tests/${id}/submit`, { answers });
                setLoading(true);
                await load(); // now completed: the server returns the answer key for review
            }}
        />
    );
}
