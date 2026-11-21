'use client';

import React from 'react';
import { useLocale, useTranslations } from 'next-intl';
import { Link } from '@/i18n/routing';
import { ChevronLeft, ChevronRight, CalendarDays, Clock, FileText, Gamepad2, ClipboardList } from 'lucide-react';
import api from '@/lib/api';
import { isGameType } from '@/lib/testTypes';

type Event = {
    id: string;
    date: Date;
    title: string;
    subtitle?: string;
    href: string;
    kind: 'lesson' | 'task' | 'game';
};

const sameDay = (a: Date, b: Date) =>
    a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate();

function toLocalDate(iso: string) {
    // lesson_date comes as "2025-11-21T00:00:00Z": keep the calendar day, ignore the timezone
    const d = new Date(iso);
    return new Date(d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate());
}

/** Loads the real upcoming items for the current role */
export function useUpcoming(role: 'teacher' | 'student') {
    const [events, setEvents] = React.useState<Event[]>([]);
    const [loading, setLoading] = React.useState(true);

    React.useEffect(() => {
        let alive = true;
        const load = async () => {
            const list: Event[] = [];
            try {
                if (role === 'teacher') {
                    const res = await api.get('/kmj', { params: { page_size: 100 } });
                    const plans = res.data.data?.plans || [];
                    for (const p of plans) {
                        if (!p.lesson_date) continue;
                        list.push({
                            id: `kmj-${p.id}`,
                            date: toLocalDate(p.lesson_date),
                            title: p.lesson_topic,
                            subtitle: `${p.subject} · ${p.grade}`,
                            href: `/teacher/materials/short-term-plan/${p.id}`,
                            kind: 'lesson',
                        });
                    }
                } else {
                    const res = await api.get('/student/tests');
                    const tasks = res.data.data || [];
                    for (const t of tasks) {
                        if (t.status === 'completed') continue;
                        list.push({
                            id: `task-${t.assignment_id}`,
                            date: new Date(t.assigned_at),
                            title: t.topic,
                            subtitle: t.subject,
                            href: `/student/tests/${t.assignment_id}`,
                            kind: isGameType(t.test_type) ? 'game' : 'task',
                        });
                    }
                }
            } catch {
                /* empty state is shown */
            }
            if (alive) {
                setEvents(list);
                setLoading(false);
            }
        };
        load();
        return () => { alive = false; };
    }, [role]);

    return { events, loading };
}

function Calendar({ events }: { events: Event[] }) {
    const locale = useLocale();
    const t = useTranslations('shell');
    const today = new Date();
    const [cursor, setCursor] = React.useState(new Date(today.getFullYear(), today.getMonth(), 1));
    const [selected, setSelected] = React.useState<Date>(today);

    const intlLocale = locale === 'kk' ? 'kk-KZ' : locale === 'ru' ? 'ru-RU' : 'en-US';
    const monthLabel = cursor.toLocaleDateString(intlLocale, { month: 'long', year: 'numeric' });
    const weekdays = Array.from({ length: 7 }, (_, i) =>
        new Date(2024, 0, 1 + i).toLocaleDateString(intlLocale, { weekday: 'narrow' })); // Monday first

    const firstWeekday = (cursor.getDay() + 6) % 7;
    const daysInMonth = new Date(cursor.getFullYear(), cursor.getMonth() + 1, 0).getDate();
    const cells: (Date | null)[] = [
        ...Array.from({ length: firstWeekday }, () => null),
        ...Array.from({ length: daysInMonth }, (_, i) => new Date(cursor.getFullYear(), cursor.getMonth(), i + 1)),
    ];

    const dayEvents = events.filter((e) => sameDay(e.date, selected));

    return (
        <section className="glass-card rounded-3xl p-5">
            <div className="flex items-center justify-between mb-4">
                <h2 className="font-bold text-lg font-outfit flex items-center gap-2">
                    <CalendarDays size={18} className="text-primary" /> {t('calendar')}
                </h2>
                <button
                    onClick={() => { setCursor(new Date(today.getFullYear(), today.getMonth(), 1)); setSelected(today); }}
                    className="text-xs font-semibold px-3 py-1.5 rounded-full bg-primary/10 text-primary"
                >
                    {t('today')}
                </button>
            </div>

            <div className="flex items-center justify-between mb-2">
                <button onClick={() => setCursor(new Date(cursor.getFullYear(), cursor.getMonth() - 1, 1))}
                    className="p-1.5 rounded-lg hover:bg-primary/10" aria-label={t('prev_month')}>
                    <ChevronLeft size={18} />
                </button>
                <span className="text-sm font-semibold capitalize" aria-live="polite">{monthLabel}</span>
                <button onClick={() => setCursor(new Date(cursor.getFullYear(), cursor.getMonth() + 1, 1))}
                    className="p-1.5 rounded-lg hover:bg-primary/10" aria-label={t('next_month')}>
                    <ChevronRight size={18} />
                </button>
            </div>

            <div className="grid grid-cols-7 gap-1 text-center" role="grid">
                {weekdays.map((w, i) => (
                    <span key={i} className="text-[11px] font-semibold text-muted py-1 uppercase">{w}</span>
                ))}
                {cells.map((d, i) => {
                    if (!d) return <span key={`e${i}`} />;
                    const has = events.some((e) => sameDay(e.date, d));
                    const isToday = sameDay(d, today);
                    const isSel = sameDay(d, selected);
                    return (
                        <button
                            key={i}
                            onClick={() => setSelected(d)}
                            aria-pressed={isSel}
                            aria-label={d.toLocaleDateString(intlLocale, { day: 'numeric', month: 'long' }) + (has ? ` — ${t('has_events')}` : '')}
                            className={`relative h-9 rounded-xl text-sm font-medium transition-colors ${isSel
                                ? 'bg-primary text-white'
                                : isToday ? 'bg-primary/10 text-primary font-bold' : 'hover:bg-primary/10'}`}
                        >
                            {d.getDate()}
                            {has && <span className={`absolute bottom-1 left-1/2 -translate-x-1/2 w-1 h-1 rounded-full ${isSel ? 'bg-white' : 'bg-primary'}`} />}
                        </button>
                    );
                })}
            </div>

            <div className="mt-4 border-t border-line pt-4 space-y-2">
                {dayEvents.length === 0 ? (
                    <p className="text-sm text-muted">{t('no_events_day')}</p>
                ) : dayEvents.map((e) => <EventRow key={e.id} e={e} />)}
            </div>
        </section>
    );
}

function EventRow({ e }: { e: Event }) {
    const icon = e.kind === 'lesson' ? <FileText size={16} /> : e.kind === 'game' ? <Gamepad2 size={16} /> : <ClipboardList size={16} />;
    return (
        <Link href={e.href} className="flex items-center gap-3 p-3 rounded-2xl bg-primary/[0.06] hover:bg-primary/10 transition-colors">
            <span className="w-9 h-9 rounded-xl bg-primary text-white flex items-center justify-center shrink-0">{icon}</span>
            <span className="min-w-0">
                <span className="block text-sm font-bold truncate">{e.title}</span>
                {e.subtitle && <span className="block text-xs text-muted truncate">{e.subtitle}</span>}
            </span>
        </Link>
    );
}

/** Calendar + "upcoming" list. Shown on the right on wide screens, inline on phones. */
export default function RightRail({ role }: { role: 'teacher' | 'student' }) {
    const t = useTranslations('shell');
    const locale = useLocale();
    const { events, loading } = useUpcoming(role);
    const intlLocale = locale === 'kk' ? 'kk-KZ' : locale === 'ru' ? 'ru-RU' : 'en-US';

    const startOfToday = new Date(); startOfToday.setHours(0, 0, 0, 0);
    const upcoming = role === 'teacher'
        ? events.filter((e) => e.date >= startOfToday).sort((a, b) => +a.date - +b.date).slice(0, 4)
        : events.slice(0, 4);

    return (
        <>
            <Calendar events={events} />
            <section className="glass-card rounded-3xl p-5">
                <div className="flex items-center justify-between mb-3">
                    <h2 className="font-bold text-lg font-outfit">
                        {role === 'teacher' ? t('upcoming_lessons') : t('my_tasks')}
                    </h2>
                    <Link href={role === 'teacher' ? '/teacher/materials/short-term-plan' : '/student/tests'}
                        className="text-xs font-semibold text-primary hover:underline">
                        {t('see_all')}
                    </Link>
                </div>
                {loading ? (
                    <p className="text-sm text-muted">{t('loading')}</p>
                ) : upcoming.length === 0 ? (
                    <div className="text-sm text-muted space-y-3">
                        <p>{role === 'teacher' ? t('no_upcoming_teacher') : t('no_upcoming_student')}</p>
                        {role === 'teacher' && (
                            <Link href="/teacher/materials/short-term-plan" className="btn-ghost text-sm w-full">
                                {t('create_kmj')}
                            </Link>
                        )}
                    </div>
                ) : (
                    <ul className="space-y-2">
                        {upcoming.map((e) => (
                            <li key={e.id} className="flex gap-3 items-start">
                                <span className="w-14 shrink-0 text-xs text-muted pt-3 flex items-center gap-1">
                                    <Clock size={12} />
                                    {e.date.toLocaleDateString(intlLocale, { day: '2-digit', month: '2-digit' })}
                                </span>
                                <div className="flex-1 min-w-0"><EventRow e={e} /></div>
                            </li>
                        ))}
                    </ul>
                )}
            </section>
        </>
    );
}
