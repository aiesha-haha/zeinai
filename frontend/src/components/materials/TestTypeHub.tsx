'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { Link } from '@/i18n/routing';
import { Plus, Loader2, AlertTriangle, ChevronRight, Send, Trash2 } from 'lucide-react';
import api from '@/lib/api';
import PageHeader from '@/components/ui/PageHeader';
import { EmptyState } from '@/components/dashboard/Charts';
import AssignTestModal from '@/components/materials/AssignTestModal';

type TestRow = {
    id: string;
    topic: string;
    subject: string;
    grade: number;
    test_type: string;
    status: string;
    error_message?: string;
    created_at: string;
};

/**
 * Lists the teacher's tests of the given types and lets them create, open,
 * assign or delete them. Used by PISA, ҰБТ, Olympiad and Games pages.
 */
export default function TestTypeHub({ types, title, description, backHref, selfHref, createType, createLabel, emptyText, icon, extra }: {
    types: string[];
    title: string;
    description: string;
    backHref: string;
    /** this page, where the create form returns to */
    selfHref: string;
    createType: string;
    createLabel: string;
    emptyText: string;
    icon: React.ReactNode;
    extra?: React.ReactNode;
}) {
    const tTypes = useTranslations('test_types');
    const tTests = useTranslations('tests');
    const tCommon = useTranslations('common');
    const [rows, setRows] = React.useState<TestRow[]>([]);
    const [loading, setLoading] = React.useState(true);
    const [error, setError] = React.useState<string | null>(null);
    const [assign, setAssign] = React.useState<TestRow | null>(null);

    const typesKey = types.join(',');
    const load = React.useCallback(async () => {
        try {
            const res = await api.get('/tests', { params: { page_size: 100 } });
            const all: TestRow[] = res.data.data?.tests || [];
            setRows(all.filter((r) => typesKey.split(',').includes(r.test_type)));
            setError(null);
        } catch (e: any) {
            setError(e.response?.data?.error?.message || tCommon('error_occurred'));
        } finally {
            setLoading(false);
        }
    }, [typesKey, tCommon]);

    React.useEffect(() => { load(); }, [load]);

    // Refresh while something is generating
    const generating = rows.some((r) => r.status === 'generating' || r.status === 'draft');
    React.useEffect(() => {
        if (!generating) return;
        const id = setInterval(load, 5000);
        return () => clearInterval(id);
    }, [generating, load]);

    const remove = async (r: TestRow) => {
        if (!confirm(tCommon('confirm_delete'))) return;
        try {
            await api.delete(`/tests/${r.id}`);
            setRows((prev) => prev.filter((x) => x.id !== r.id));
        } catch (e: any) {
            alert(e.response?.data?.error?.message || tCommon('error_occurred'));
        }
    };

    const createHref = `/teacher/materials/tests/create?type=${createType}&back=${encodeURIComponent(selfHref)}`;

    return (
        <div>
            <PageHeader title={title} description={description} backHref={backHref}
                actions={<Link href={createHref} className="btn-primary"><Plus size={18} /> {createLabel}</Link>} />

            {extra}

            <section className="glass-card rounded-3xl p-6">
                {loading ? (
                    <p className="text-sm text-muted" role="status">{tCommon('loading')}</p>
                ) : error ? (
                    <p role="alert" className="text-sm text-red-700">{error}</p>
                ) : rows.length === 0 ? (
                    <EmptyState icon={icon} text={emptyText}
                        action={<Link href={createHref} className="btn-primary"><Plus size={18} /> {createLabel}</Link>} />
                ) : (
                    <ul className="divide-y divide-line">
                        {rows.map((r) => (
                            <li key={r.id} className="flex flex-wrap items-center gap-3 py-4">
                                <span className="w-11 h-11 rounded-xl bg-primary/10 text-primary flex items-center justify-center shrink-0">{icon}</span>
                                <div className="flex-1 min-w-[180px]">
                                    <p className="font-bold">{r.topic}</p>
                                    <p className="text-xs text-muted">
                                        {tTypes(`${r.test_type}.title`)} · {r.subject} · {r.grade} {tCommon('grade_suffix')} · {new Date(r.created_at).toLocaleDateString()}
                                    </p>
                                </div>
                                {r.status === 'completed' ? (
                                    <>
                                        <button className="btn-ghost text-sm !py-2" onClick={() => setAssign(r)}><Send size={16} /> {tTests('assign')}</button>
                                        <Link href={`/teacher/materials/tests/${r.id}`} className="btn-primary text-sm !py-2">{tTests('open')} <ChevronRight size={16} /></Link>
                                    </>
                                ) : r.status === 'failed' ? (
                                    <span className="flex items-center gap-1.5 text-sm text-red-700" title={r.error_message}>
                                        <AlertTriangle size={16} /> {tTests('status_failed')}
                                    </span>
                                ) : (
                                    <span className="flex items-center gap-1.5 text-sm text-primary" role="status">
                                        <Loader2 size={16} className="animate-spin" /> {tTests('status_generating')}
                                    </span>
                                )}
                                <button onClick={() => remove(r)} className="p-2 rounded-lg text-red-600 hover:bg-red-500/10" aria-label={`${tCommon('delete')}: ${r.topic}`}>
                                    <Trash2 size={16} />
                                </button>
                            </li>
                        ))}
                    </ul>
                )}
            </section>

            {assign && (
                <AssignTestModal isOpen={!!assign} onClose={() => setAssign(null)} testId={assign.id} testTopic={assign.topic} />
            )}
        </div>
    );
}
