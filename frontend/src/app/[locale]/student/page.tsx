'use client';

import React from 'react';
import { Link } from '@/i18n/routing';
import { useTranslations } from 'next-intl';
import { BookOpen, ChevronRight, Clock, Loader2, PlayCircle, Gamepad2, ClipboardList, Radio } from 'lucide-react';
import api from '@/lib/api';
import { BarChart, Donut, EmptyState } from '@/components/dashboard/Charts';
import RightRail from '@/components/shell/RightRail';
import { isGameType } from '@/lib/testTypes';

type Task = {
    assignment_id: string;
    topic: string;
    subject: string;
    grade: number;
    test_type?: string;
    status: string;
    assigned_at: string;
    score?: number;
    max_score?: number;
};

const pct = (t: Task) => (t.max_score ? ((t.score || 0) / t.max_score) * 100 : 0);

export default function StudentDashboard() {
    const t = useTranslations('student');
    const tD = useTranslations('dashboard');
    const tCommon = useTranslations('common');
    const [profile, setProfile] = React.useState<any>(null);
    const [data, setData] = React.useState<any>(null);
    const [history, setHistory] = React.useState<any[]>([]);
    const [tasks, setTasks] = React.useState<Task[]>([]);
    const [loading, setLoading] = React.useState(true);

    React.useEffect(() => {
        Promise.allSettled([
            api.get('/users/me'),
            api.get('/student/dashboard'),
            api.get('/whiteboard/student/history'),
            api.get('/student/tests'),
        ]).then(([me, dash, hist, tests]) => {
            if (me.status === 'fulfilled') setProfile(me.value.data.data);
            if (dash.status === 'fulfilled') setData(dash.value.data.data);
            if (hist.status === 'fulfilled') setHistory(hist.value.data.data || []);
            if (tests.status === 'fulfilled') setTasks(tests.value.data.data || []);
            setLoading(false);
        });
    }, []);

    if (loading) {
        return (
            <div className="flex items-center justify-center py-20" role="status">
                <Loader2 className="w-10 h-10 text-primary animate-spin" />
                <span className="sr-only">{tCommon('loading')}</span>
            </div>
        );
    }

    const active = tasks.filter((x) => x.status !== 'completed');
    const done = tasks.filter((x) => x.status === 'completed' && x.max_score);
    const avg = done.length ? done.reduce((s, x) => s + pct(x), 0) / done.length : 0;

    // Average per subject for the rings
    const bySubject = Object.values(done.reduce<Record<string, { s: string; sum: number; n: number }>>((acc, x) => {
        const k = x.subject || '—';
        acc[k] = acc[k] || { s: k, sum: 0, n: 0 };
        acc[k].sum += pct(x); acc[k].n += 1;
        return acc;
    }, {})).slice(0, 6);

    const firstActive = active[0];

    return (
        <div className="space-y-6">
            {/* Greeting */}
            <section className="glass-card rounded-3xl overflow-hidden flex flex-col sm:flex-row items-stretch">
                <div className="flex-1 p-6 md:p-8">
                    <h1 className="text-2xl md:text-3xl font-bold font-outfit mb-2">
                        {profile?.first_name ? tD('hello', { name: profile.first_name }) : tD('hello_plain')}
                    </h1>
                    <p className="text-muted mb-5 max-w-md">
                        {active.length > 0 ? tD('student_has_tasks', { count: active.length }) : tD('student_no_tasks')}
                    </p>
                    {firstActive ? (
                        <Link href={`/student/tests/${firstActive.assignment_id}`} className="btn-primary">
                            {isGameType(firstActive.test_type) ? tD('play_next') : tD('open_next_task')}
                            <ChevronRight size={18} />
                        </Link>
                    ) : (
                        <Link href="/student/tests" className="btn-ghost">{tD('see_my_tasks')}</Link>
                    )}
                </div>
                <div className="sm:w-56 md:w-64 h-40 sm:h-auto bg-primary/5 relative" aria-hidden>
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src="/hero-ai.png" alt="" className="absolute inset-0 w-full h-full object-cover object-center" />
                </div>
            </section>

            <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
                {/* Results */}
                <section className="glass-card rounded-3xl p-6 lg:col-span-3">
                    <div className="flex items-center justify-between mb-4">
                        <h2 className="text-lg font-bold font-outfit">{tD('my_results')}</h2>
                        <Link href="/student/tests" className="text-sm font-semibold text-primary hover:underline">{tCommon('view_all')}</Link>
                    </div>
                    {done.length === 0 ? (
                        <EmptyState icon={<ClipboardList />} text={tD('results_empty_student')}
                            action={firstActive && <Link className="btn-ghost text-sm" href={`/student/tests/${firstActive.assignment_id}`}>{tD('open_next_task')}</Link>} />
                    ) : (
                        <>
                            <div className="flex items-baseline gap-3 mb-5">
                                <span className="text-4xl font-bold font-outfit tabular-nums">{avg.toFixed(1)}%</span>
                                <span className="text-sm text-muted">{tD('average_score', { count: done.length })}</span>
                            </div>
                            <BarChart
                                ariaLabel={tD('my_results')}
                                bars={done.slice(0, 6).reverse().map((x) => ({ label: x.topic, value: pct(x), hint: `${x.score}/${x.max_score}` }))}
                            />
                        </>
                    )}
                </section>

                {/* By subject */}
                <section className="glass-card rounded-3xl p-6 lg:col-span-2">
                    <h2 className="text-lg font-bold font-outfit mb-4">{tD('by_subject')}</h2>
                    {bySubject.length === 0 ? (
                        <EmptyState text={tD('by_subject_empty')} />
                    ) : (
                        <div className="grid grid-cols-2 gap-5 justify-items-center">
                            {bySubject.map((s) => <Donut key={s.s} value={s.sum / s.n} label={s.s} />)}
                        </div>
                    )}
                </section>
            </div>

            {/* Active tasks & games */}
            <section className="glass-card rounded-3xl p-6">
                <div className="flex items-center justify-between mb-4">
                    <h2 className="text-lg font-bold font-outfit">{t('assigned_tests')}</h2>
                    <Link href="/student/tests" className="text-sm font-semibold text-primary hover:underline">{tCommon('view_all')}</Link>
                </div>
                {active.length === 0 ? (
                    <EmptyState icon={<Gamepad2 />} text={tD('student_no_tasks')} />
                ) : (
                    <ul className="grid grid-cols-1 md:grid-cols-2 gap-3">
                        {active.slice(0, 4).map((x) => (
                            <li key={x.assignment_id}>
                                <Link href={`/student/tests/${x.assignment_id}`}
                                    className="flex items-center gap-4 p-4 rounded-2xl border border-line hover:border-primary/40 hover:bg-primary/[0.04] transition-colors">
                                    <span className="w-11 h-11 rounded-xl bg-primary/10 text-primary flex items-center justify-center shrink-0">
                                        {isGameType(x.test_type) ? <Gamepad2 size={20} /> : <ClipboardList size={20} />}
                                    </span>
                                    <span className="min-w-0 flex-1">
                                        <span className="block font-bold truncate">{x.topic}</span>
                                        <span className="block text-xs text-muted">{x.subject} · {x.grade} {tCommon('grade_suffix')}</span>
                                    </span>
                                    <span className="text-sm font-bold text-primary whitespace-nowrap">
                                        {isGameType(x.test_type) ? tD('play') : t('start_test')}
                                    </span>
                                </Link>
                            </li>
                        ))}
                    </ul>
                )}
            </section>

            {/* Classes / teachers */}
            <section className="glass-card rounded-3xl p-6">
                <h2 className="text-lg font-bold font-outfit mb-4">{t('my_classes')}</h2>
                {data?.classes?.length > 0 ? (
                    <ul className="space-y-3">
                        {data.classes.map((item: any) => (
                            <li key={item.id} className="flex flex-wrap items-center gap-4 p-4 rounded-2xl bg-primary/[0.04]">
                                <span className="w-11 h-11 rounded-xl bg-primary/10 text-primary flex items-center justify-center"><PlayCircle size={22} /></span>
                                <span className="flex-1 min-w-[140px]">
                                    <span className="block font-bold">{item.name}</span>
                                    <span className="block text-sm text-muted">{item.teacher_name}</span>
                                </span>
                                {item.has_active_session ? (
                                    <Link href={`/student/whiteboard/live?classId=${item.id}`} className="btn-primary text-sm !py-2">
                                        <Radio size={16} /> {t('live_board_btn')}
                                    </Link>
                                ) : (
                                    <span className="text-xs text-muted">{tD('no_live_lesson')}</span>
                                )}
                            </li>
                        ))}
                    </ul>
                ) : (
                    <EmptyState icon={<BookOpen />} text={tD('student_no_classes')} />
                )}
            </section>

            {/* Past lessons */}
            {history.length > 0 && (
                <section className="glass-card rounded-3xl p-6">
                    <div className="flex items-center justify-between mb-4">
                        <h2 className="text-lg font-bold font-outfit">{t('past_lessons')}</h2>
                        <Link href="/student/history" className="text-sm font-semibold text-primary hover:underline">{tCommon('view_all')}</Link>
                    </div>
                    <ul className="space-y-2">
                        {history.slice(0, 3).map((session: any) => (
                            <li key={session.id}>
                                <Link href={`/student/whiteboard/summary/${session.id}`}
                                    className="flex items-center gap-4 p-3 rounded-2xl hover:bg-primary/[0.05]">
                                    <span className="w-10 h-10 rounded-xl bg-primary/10 text-primary flex items-center justify-center"><Clock size={18} /></span>
                                    <span className="flex-1 min-w-0">
                                        <span className="block font-semibold text-sm truncate">
                                            {session.title || `${t('lesson_label')} ${new Date(session.created_at).toLocaleDateString()}`}
                                        </span>
                                        <span className="block text-xs text-muted">{new Date(session.created_at).toLocaleString()}</span>
                                    </span>
                                    <span className="text-sm font-semibold text-primary">{t('view_summary')}</span>
                                </Link>
                            </li>
                        ))}
                    </ul>
                </section>
            )}

            {/* Calendar inline on smaller screens */}
            <div className="xl:hidden space-y-6"><RightRail role="student" /></div>
        </div>
    );
}
