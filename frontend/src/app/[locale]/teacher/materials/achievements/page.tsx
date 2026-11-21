'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { Link, useRouter } from '@/i18n/routing';
import { Award, Medal, BadgeCheck, Loader2 } from 'lucide-react';
import api from '@/lib/api';
import PageHeader from '@/components/ui/PageHeader';
import { EmptyState } from '@/components/dashboard/Charts';

type Row = { id: string; name: string; classes: Set<string>; score: number; max: number; done: number; best: number };

/** Student achievements built from real test / game results of the teacher's assignments */
export default function AchievementsPage() {
    const t = useTranslations('achievements_page');
    const tM = useTranslations('education_materials');
    const tCommon = useTranslations('common');
    const router = useRouter();
    const [rows, setRows] = React.useState<Row[]>([]);
    const [classes, setClasses] = React.useState<string[]>([]);
    const [filter, setFilter] = React.useState('');
    const [loading, setLoading] = React.useState(true);
    const [error, setError] = React.useState<string | null>(null);

    React.useEffect(() => {
        (async () => {
            try {
                const asg = (await api.get('/tests/assignments/all')).data.data || [];
                const byStudent = new Map<string, Row>();
                const classNames = new Set<string>();
                const results = await Promise.allSettled(asg.map((a: any) => api.get(`/tests/assignments/${a.id}/results`)));
                results.forEach((r, i) => {
                    if (r.status !== 'fulfilled') return;
                    const a = asg[i];
                    classNames.add(a.class_name);
                    for (const s of r.value.data.data || []) {
                        if (s.status !== 'completed' || !s.max_score) continue;
                        const row = byStudent.get(s.student_id) || { id: s.student_id, name: s.student_name, classes: new Set<string>(), score: 0, max: 0, done: 0, best: 0 };
                        row.classes.add(a.class_name);
                        row.score += s.score; row.max += s.max_score; row.done += 1;
                        row.best = Math.max(row.best, Math.round((s.score / s.max_score) * 100));
                        byStudent.set(s.student_id, row);
                    }
                });
                setRows([...byStudent.values()].sort((x, y) => (y.score / y.max) - (x.score / x.max) || y.done - x.done));
                setClasses([...classNames]);
            } catch (e: any) {
                setError(e.response?.data?.error?.message || tCommon('error_occurred'));
            } finally {
                setLoading(false);
            }
        })();
    }, [tCommon]);

    const visible = filter ? rows.filter((r) => r.classes.has(filter)) : rows;
    const badge = (r: Row) => {
        const p = r.max ? (r.score / r.max) * 100 : 0;
        if (p >= 90 && r.done >= 3) return t('badge_master');
        if (p >= 75) return t('badge_strong');
        if (r.done >= 5) return t('badge_active');
        return null;
    };

    return (
        <div>
            <PageHeader title={tM('tools.achievements')} description={t('desc')} backHref="/teacher/materials/achievements-category"
                actions={classes.length > 1 && (
                    <div>
                        <label htmlFor="ach-class" className="sr-only">{t('filter_class')}</label>
                        <select id="ach-class" className="field !py-2.5" value={filter} onChange={(e) => setFilter(e.target.value)}>
                            <option value="">{t('all_classes')}</option>
                            {classes.map((c) => <option key={c} value={c}>{c}</option>)}
                        </select>
                    </div>
                )} />

            <section className="glass-card rounded-3xl p-6">
                {loading ? (
                    <p role="status" className="flex items-center gap-2 text-muted"><Loader2 className="animate-spin" size={18} /> {tCommon('loading')}</p>
                ) : error ? (
                    <p role="alert" className="text-red-700">{error}</p>
                ) : visible.length === 0 ? (
                    <EmptyState icon={<Award />} text={t('empty')}
                        action={<Link href="/teacher/materials/games" className="btn-primary">{t('give_game')}</Link>} />
                ) : (
                    <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                            <caption className="sr-only">{t('table_caption')}</caption>
                            <thead>
                                <tr className="text-left text-muted border-b border-line">
                                    <th scope="col" className="py-3 pr-3 w-12">#</th>
                                    <th scope="col" className="py-3 pr-3">{t('student')}</th>
                                    <th scope="col" className="py-3 pr-3">{t('done')}</th>
                                    <th scope="col" className="py-3 pr-3">{t('avg')}</th>
                                    <th scope="col" className="py-3 pr-3">{t('best')}</th>
                                    <th scope="col" className="py-3 pr-3">{t('badge')}</th>
                                    <th scope="col" className="py-3"><span className="sr-only">{tCommon('actions')}</span></th>
                                </tr>
                            </thead>
                            <tbody>
                                {visible.map((r, i) => {
                                    const avg = r.max ? Math.round((r.score / r.max) * 100) : 0;
                                    const b = badge(r);
                                    return (
                                        <tr key={r.id} className="border-b border-line last:border-0">
                                            <td className="py-3 pr-3 font-bold">
                                                {i < 3 ? <Medal size={18} className={i === 0 ? 'text-amber-500' : i === 1 ? 'text-slate-400' : 'text-orange-700'} aria-label={`${i + 1}`} /> : i + 1}
                                            </td>
                                            <td className="py-3 pr-3">
                                                <span className="font-semibold">{r.name}</span>
                                                <span className="block text-xs text-muted">{[...r.classes].join(', ')}</span>
                                            </td>
                                            <td className="py-3 pr-3 tabular-nums">{r.done}</td>
                                            <td className="py-3 pr-3 tabular-nums font-semibold">{avg}%</td>
                                            <td className="py-3 pr-3 tabular-nums">{r.best}%</td>
                                            <td className="py-3 pr-3">{b ? <span className="px-2 py-1 rounded-full bg-primary/10 text-primary text-xs font-bold">{b}</span> : '—'}</td>
                                            <td className="py-3 text-right">
                                                <button className="btn-ghost text-xs !py-1.5 !px-3"
                                                    onClick={() => {
                                                        // hand the name over in-memory (not in the URL)
                                                        try { sessionStorage.setItem('studio_prefill_certificate', JSON.stringify({ student_name: r.name, achievement: t('cert_reason', { avg }) })); } catch { }
                                                        router.push('/teacher/materials/certificates');
                                                    }}>
                                                    <BadgeCheck size={14} /> {t('give_certificate')}
                                                </button>
                                            </td>
                                        </tr>
                                    );
                                })}
                            </tbody>
                        </table>
                    </div>
                )}
            </section>
        </div>
    );
}
