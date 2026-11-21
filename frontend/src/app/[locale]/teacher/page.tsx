'use client';

import React from 'react';
import { motion } from 'framer-motion';
import { useTranslations } from 'next-intl';
import { Link, useRouter } from '@/i18n/routing';
import {
    Users, Plus, Edit2, Trash2, Bell, LayoutPanelLeft, FileText, Sparkles, ClipboardCheck,
    GraduationCap, ChevronRight, Gamepad2, ArrowRight, Image as ImageIcon, Video, MessageSquare,
} from 'lucide-react';
import api from '@/lib/api';
import SendNotificationModal from '@/components/notifications/SendNotificationModal';
import { BarChart, EmptyState, StatCard } from '@/components/dashboard/Charts';
import RightRail from '@/components/shell/RightRail';

type Recent = { id: string; title: string; type: string; href: string; date: string };

const studioHref: Record<string, string> = {
    image: '/teacher/materials/image', video: '/teacher/materials/video', feedback: '/teacher/materials/feedback',
    lab: '/teacher/materials/lab', news: '/teacher/materials/news', certificate: '/teacher/materials/certificates',
};

export default function TeacherDashboard() {
    const tProfile = useTranslations('profile');
    const tCommon = useTranslations('common');
    const tM = useTranslations('education_materials');
    const tD = useTranslations('dashboard');
    const router = useRouter();

    const [profile, setProfile] = React.useState<any>(null);
    const [classes, setClasses] = React.useState<any[]>([]);
    const [assignments, setAssignments] = React.useState<any[]>([]);
    const [recent, setRecent] = React.useState<Recent[]>([]);
    const [loading, setLoading] = React.useState(true);

    const [isCreateModalOpen, setIsCreateModalOpen] = React.useState(false);
    const [isEditModalOpen, setIsEditModalOpen] = React.useState(false);
    const [classNameInput, setClassNameInput] = React.useState('');
    const [selectedClass, setSelectedClass] = React.useState<any>(null);
    const [saving, setSaving] = React.useState(false);
    const [notifOpen, setNotifOpen] = React.useState(false);

    const fetchClasses = React.useCallback(async () => {
        try {
            const res = await api.get('/classes');
            setClasses(res.data.data || []);
        } catch { /* shown as empty */ }
    }, []);

    const fetchProfile = React.useCallback(async () => {
        try {
            const res = await api.get('/users/me');
            setProfile(res.data.data);
        } catch { /* handled by interceptor */ }
    }, []);

    React.useEffect(() => {
        Promise.allSettled([
            fetchProfile(),
            fetchClasses(),
            api.get('/tests/assignments/all'),
            api.get('/kmj', { params: { page_size: 5 } }),
            api.get('/ai-studio'),
        ]).then(([, , asg, kmj, studio]) => {
            if (asg.status === 'fulfilled') setAssignments(asg.value.data.data || []);
            const items: Recent[] = [];
            if (kmj.status === 'fulfilled') {
                for (const p of kmj.value.data.data?.plans || []) {
                    items.push({ id: p.id, title: p.lesson_topic, type: 'kmj', href: `/teacher/materials/short-term-plan/${p.id}`, date: p.created_at });
                }
            }
            if (studio.status === 'fulfilled') {
                for (const s of (studio.value.data.data || []).slice(0, 5)) {
                    items.push({ id: s.id, title: s.title, type: s.kind, href: `${studioHref[s.kind] || '/teacher'}?id=${s.id}`, date: s.created_at });
                }
            }
            items.sort((a, b) => +new Date(b.date) - +new Date(a.date));
            setRecent(items.slice(0, 5));
            setLoading(false);
        });
    }, [fetchClasses, fetchProfile]);

    const closeModal = () => { setIsCreateModalOpen(false); setIsEditModalOpen(false); setClassNameInput(''); setSelectedClass(null); };

    const handleSaveClass = async () => {
        if (!classNameInput.trim() || saving) return;
        setSaving(true);
        try {
            if (isEditModalOpen && selectedClass) await api.put(`/classes/${selectedClass.id}`, { name: classNameInput });
            else await api.post('/classes', { name: classNameInput });
            closeModal();
            await Promise.all([fetchClasses(), fetchProfile()]);
        } catch (e: any) {
            alert(e.response?.data?.error?.message || tCommon('error_occurred'));
        } finally {
            setSaving(false);
        }
    };

    const handleDelete = async (id: string) => {
        if (!confirm(tCommon('confirm_delete'))) return;
        try {
            await api.delete(`/classes/${id}`);
            fetchClasses(); fetchProfile();
        } catch (e: any) {
            alert(e.response?.data?.error?.message || tCommon('error_occurred'));
        }
    };

    const totalStudents = classes.reduce((s, c) => s + (c.students_count || 0), 0);
    const sub = profile?.subscription;
    const bars = assignments.slice(0, 6).reverse().map((a) => ({
        label: a.test_title,
        value: a.student_count ? (a.finished_count / a.student_count) * 100 : 0,
        hint: `${a.class_name}: ${a.finished_count}/${a.student_count}`,
    }));
    const avgCompletion = bars.length ? bars.reduce((s, b) => s + b.value, 0) / bars.length : 0;

    const quick = [
        { icon: <FileText size={20} />, label: tM('tools.short_term_plan'), href: '/teacher/materials/short-term-plan' },
        { icon: <Gamepad2 size={20} />, label: tD('quick_game'), href: '/teacher/materials/games' },
        { icon: <MessageSquare size={20} />, label: tM('tools.feedback'), href: '/teacher/materials/feedback' },
        { icon: <ImageIcon size={20} />, label: tM('tools.image'), href: '/teacher/materials/image' },
        { icon: <Video size={20} />, label: tM('tools.video'), href: '/teacher/materials/video' },
        { icon: <ClipboardCheck size={20} />, label: tM('tools.test_generation'), href: '/teacher/materials/tests/create' },
    ];

    const recentIcon = (type: string) => type === 'kmj' ? <FileText size={18} /> : type === 'image' ? <ImageIcon size={18} />
        : type === 'video' ? <Video size={18} /> : type === 'feedback' ? <MessageSquare size={18} /> : <Sparkles size={18} />;

    return (
        <div className="space-y-6">
            {/* Greeting */}
            <section className="glass-card rounded-3xl overflow-hidden flex flex-col sm:flex-row">
                <div className="flex-1 p-6 md:p-8">
                    <h1 className="text-2xl md:text-3xl font-bold font-outfit mb-2">
                        {profile?.first_name ? tD('hello', { name: profile.first_name }) : tD('hello_plain')}
                    </h1>
                    <p className="text-muted mb-5 max-w-lg">
                        {classes.length === 0 ? tD('teacher_start_hint') : tD('teacher_summary', { classes: classes.length, students: totalStudents })}
                    </p>
                    <div className="flex flex-wrap gap-3">
                        <Link href="/teacher/materials/short-term-plan" className="btn-primary"><FileText size={18} /> {tD('create_kmj')}</Link>
                        <Link href="/teacher/materials/generation" className="btn-ghost"><Sparkles size={18} /> {tD('create_material')}</Link>
                    </div>
                </div>
                <div className="sm:w-56 md:w-72 h-40 sm:h-auto relative bg-primary/5" aria-hidden>
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src="/hero-ai.png" alt="" className="absolute inset-0 w-full h-full object-cover" />
                </div>
            </section>

            {/* Stats */}
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                <StatCard icon={<GraduationCap size={22} />} label={tD('stat_classes')} value={classes.length} />
                <StatCard icon={<Users size={22} />} label={tD('stat_students')} value={totalStudents}
                    hint={sub ? `${tProfile('students_limit')}: ${sub.students_limit === -1 ? '∞' : sub.students_limit}` : undefined} />
                <StatCard icon={<ClipboardCheck size={22} />} label={tD('stat_assignments')} value={assignments.length} />
                <StatCard icon={<Sparkles size={22} />} label={tD('stat_tokens')}
                    value={sub ? (sub.tokens_limit === -1 ? '∞' : Math.max(0, sub.tokens_remaining ?? 0)) : '—'} />
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
                {/* Assignment completion chart */}
                <section className="glass-card rounded-3xl p-6 lg:col-span-3">
                    <div className="flex items-center justify-between mb-4">
                        <h2 className="text-lg font-bold font-outfit">{tD('completion_title')}</h2>
                        <Link href="/teacher/tests/assignments" className="text-sm font-semibold text-primary hover:underline">{tCommon('view_all')}</Link>
                    </div>
                    {loading ? <p className="text-sm text-muted">{tCommon('loading')}</p> : bars.length === 0 ? (
                        <EmptyState icon={<ClipboardCheck />} text={tD('completion_empty')}
                            action={<Link href="/teacher/materials/games" className="btn-ghost text-sm">{tD('give_first_task')}</Link>} />
                    ) : (
                        <>
                            <div className="flex items-baseline gap-3 mb-5">
                                <span className="text-4xl font-bold font-outfit tabular-nums">{avgCompletion.toFixed(0)}%</span>
                                <span className="text-sm text-muted">{tD('completion_avg')}</span>
                            </div>
                            <BarChart bars={bars} ariaLabel={tD('completion_title')} />
                        </>
                    )}
                </section>

                {/* Quick tools */}
                <section className="glass-card rounded-3xl p-6 lg:col-span-2">
                    <h2 className="text-lg font-bold font-outfit mb-4">{tD('quick_tools')}</h2>
                    <ul className="grid grid-cols-2 gap-3">
                        {quick.map((q) => (
                            <li key={q.href}>
                                <Link href={q.href} className="flex flex-col gap-2 p-4 h-full rounded-2xl bg-primary/[0.05] hover:bg-primary/10 transition-colors">
                                    <span className="text-primary">{q.icon}</span>
                                    <span className="text-sm font-semibold leading-snug">{q.label}</span>
                                </Link>
                            </li>
                        ))}
                    </ul>
                </section>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
                {/* Classes */}
                <section className="glass-card rounded-3xl p-6 lg:col-span-3">
                    <div className="flex flex-wrap items-center justify-between gap-3 mb-4">
                        <h2 className="text-lg font-bold font-outfit">{tProfile('my_classes')}</h2>
                        <div className="flex gap-2">
                            <button onClick={() => setNotifOpen(true)} className="btn-ghost text-sm !py-2" disabled={classes.length === 0}>
                                <Bell size={16} /> {tCommon('send_all')}
                            </button>
                            <button onClick={() => { setClassNameInput(''); setIsCreateModalOpen(true); }} className="btn-primary text-sm !py-2">
                                <Plus size={16} /> {tCommon('add')}
                            </button>
                        </div>
                    </div>
                    {classes.length === 0 ? (
                        <EmptyState icon={<Users />} text={tProfile('no_classes_yet')}
                            action={<button className="btn-primary text-sm" onClick={() => setIsCreateModalOpen(true)}>{tProfile('create_first_class')}</button>} />
                    ) : (
                        <ul className="space-y-2">
                            {classes.slice(0, 5).map((cl) => (
                                <li key={cl.id} className="flex items-center gap-3 p-3 rounded-2xl hover:bg-primary/[0.04]">
                                    <span className="w-11 h-11 rounded-xl bg-primary/10 text-primary flex items-center justify-center font-bold shrink-0">
                                        {String(cl.name).slice(0, 2).toUpperCase()}
                                    </span>
                                    <Link href={`/teacher/classes/${cl.id}`} className="flex-1 min-w-0">
                                        <span className="block font-bold truncate">{cl.name}</span>
                                        <span className="block text-xs text-muted">{tProfile('students_count_val', { count: cl.students_count })}</span>
                                    </Link>
                                    <button onClick={() => router.push(`/teacher/whiteboard/${cl.id}`)} className="p-2 rounded-lg text-primary hover:bg-primary/10"
                                        aria-label={`${tM('tools.whiteboard')}: ${cl.name}`} title={tM('tools.whiteboard')}>
                                        <LayoutPanelLeft size={18} />
                                    </button>
                                    <button onClick={() => { setSelectedClass(cl); setClassNameInput(cl.name); setIsEditModalOpen(true); }}
                                        className="p-2 rounded-lg text-primary hover:bg-primary/10" aria-label={`${tCommon('edit')}: ${cl.name}`} title={tCommon('edit')}>
                                        <Edit2 size={18} />
                                    </button>
                                    <button onClick={() => handleDelete(cl.id)} className="p-2 rounded-lg text-red-600 hover:bg-red-500/10"
                                        aria-label={`${tCommon('delete')}: ${cl.name}`} title={tCommon('delete')}>
                                        <Trash2 size={18} />
                                    </button>
                                </li>
                            ))}
                            {classes.length > 5 && (
                                <li>
                                    <Link href="/teacher/classes" className="flex items-center justify-center gap-2 py-2.5 text-sm font-semibold text-primary">
                                        {tProfile('view_all_classes')} <ArrowRight size={16} />
                                    </Link>
                                </li>
                            )}
                        </ul>
                    )}
                </section>

                {/* Recent materials */}
                <section className="glass-card rounded-3xl p-6 lg:col-span-2">
                    <h2 className="text-lg font-bold font-outfit mb-4">{tD('recent_materials')}</h2>
                    {loading ? <p className="text-sm text-muted">{tCommon('loading')}</p> : recent.length === 0 ? (
                        <EmptyState icon={<Sparkles />} text={tD('recent_empty')}
                            action={<Link href="/teacher/materials/generation" className="btn-primary text-sm"><Plus size={16} /> {tD('create_first_material')}</Link>} />
                    ) : (
                        <ul className="space-y-2">
                            {recent.map((r) => (
                                <li key={`${r.type}-${r.id}`}>
                                    <Link href={r.href} className="flex items-center gap-3 p-3 rounded-2xl hover:bg-primary/[0.05]">
                                        <span className="w-10 h-10 rounded-xl bg-primary/10 text-primary flex items-center justify-center shrink-0">{recentIcon(r.type)}</span>
                                        <span className="flex-1 min-w-0">
                                            <span className="block text-sm font-semibold truncate">{r.title}</span>
                                            <span className="block text-xs text-muted">{new Date(r.date).toLocaleDateString()}</span>
                                        </span>
                                        <ChevronRight size={16} className="text-muted" />
                                    </Link>
                                </li>
                            ))}
                        </ul>
                    )}
                </section>
            </div>

            <div className="xl:hidden space-y-6"><RightRail role="teacher" /></div>

            {(isCreateModalOpen || isEditModalOpen) && (
                <div className="fixed inset-0 z-[80] flex items-center justify-center p-4 bg-black/40" role="dialog" aria-modal="true" aria-labelledby="class-modal-title"
                    onKeyDown={(e) => e.key === 'Escape' && closeModal()}>
                    <motion.form
                        initial={{ opacity: 0, scale: 0.97 }} animate={{ opacity: 1, scale: 1 }}
                        onSubmit={(e) => { e.preventDefault(); handleSaveClass(); }}
                        className="glass-card p-6 md:p-8 rounded-3xl w-full max-w-md"
                    >
                        <h3 id="class-modal-title" className="text-xl font-bold font-outfit mb-5">
                            {isEditModalOpen ? tProfile('edit_class') : tProfile('create_class')}
                        </h3>
                        <label htmlFor="class-name" className="field-label">{tProfile('new_class_name')}</label>
                        <input id="class-name" type="text" value={classNameInput} onChange={(e) => setClassNameInput(e.target.value)}
                            className="field mb-6" placeholder={tD('class_name_placeholder')} autoFocus />
                        <div className="flex gap-3">
                            <button type="button" onClick={closeModal} className="btn-ghost flex-1">{tCommon('cancel')}</button>
                            <button type="submit" disabled={!classNameInput.trim() || saving} className="btn-primary flex-1">
                                {isEditModalOpen ? tProfile('save_changes') : tProfile('create')}
                            </button>
                        </div>
                    </motion.form>
                </div>
            )}

            <SendNotificationModal isOpen={notifOpen} onClose={() => setNotifOpen(false)} targetType="all" />
        </div>
    );
}
