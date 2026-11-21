'use client';

import React from 'react';
import { motion } from 'framer-motion';
import { Shield, Building, Users, AlertTriangle, Search, Filter, LogOut, CreditCard, X, Save, Calendar, Plus, Pencil, KeyRound, Trash2, Eye, EyeOff } from 'lucide-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { logout } from '@/lib/auth';
import api from '@/lib/api';

import { useTranslations } from 'next-intl';

const PLAN_DEFAULTS = {
    free: { students_limit: 10, tokens_limit: 100 },
    basic: { students_limit: 200, tokens_limit: 1000 },
    ultra: { students_limit: -1, tokens_limit: -1 },
};

export default function AdminDashboard() {
    const router = useRouter();
    const tNav = useTranslations('nav');

    const [activeTab, setActiveTab] = React.useState('Statistics');
    const [blocks, setBlocks] = React.useState<any[]>([]);
    const [loadingBlocks, setLoadingBlocks] = React.useState(false);

    const [teachers, setTeachers] = React.useState<any[]>([]);
    const [loadingTeachers, setLoadingTeachers] = React.useState(false);

    // --- Subscription Modal ---
    const [isSubModalOpen, setIsSubModalOpen] = React.useState(false);
    const [selectedTeacher, setSelectedTeacher] = React.useState<any>(null);
    const [subData, setSubData] = React.useState({ plan: 'free', students_limit: 10, tokens_limit: 100, duration_months: 1 });
    const [savingSub, setSavingSub] = React.useState(false);

    // --- Create Modal ---
    const [isCreateOpen, setIsCreateOpen] = React.useState(false);
    const [createData, setCreateData] = React.useState({ first_name: '', last_name: '', email: '', phone: '', password: '', birth_year: new Date().getFullYear() - 30, gender: 'male' });
    const [createLoading, setCreateLoading] = React.useState(false);
    const [createError, setCreateError] = React.useState('');
    const [showCreatePass, setShowCreatePass] = React.useState(false);

    // --- Edit Modal ---
    const [isEditOpen, setIsEditOpen] = React.useState(false);
    const [editData, setEditData] = React.useState({ first_name: '', last_name: '', email: '', phone: '', birth_year: 1990, gender: 'male' });
    const [editLoading, setEditLoading] = React.useState(false);
    const [editError, setEditError] = React.useState('');

    // --- Password Modal ---
    const [isPassOpen, setIsPassOpen] = React.useState(false);
    const [passData, setPassData] = React.useState({ new_password: '', confirm: '' });
    const [passLoading, setPassLoading] = React.useState(false);
    const [passError, setPassError] = React.useState('');
    const [showNewPass, setShowNewPass] = React.useState(false);

    // --- Delete Confirm ---
    const [deleteId, setDeleteId] = React.useState<string | null>(null);
    const [deleteLoading, setDeleteLoading] = React.useState(false);

    React.useEffect(() => {
        if (typeof window !== 'undefined') {
            const token = localStorage.getItem('access_token');
            const role = localStorage.getItem('user_role');

            if (!token || role !== 'admin') {
                router.push('/login');
            }
        }
    }, [router]);

    const fetchBlocks = async () => {
        setLoadingBlocks(true);
        try {
            const res = await api.get('/admin/verification/blocks');
            setBlocks(res.data.data || []);
        } catch (err) {
            console.error("Failed to fetch blocks", err);
        } finally {
            setLoadingBlocks(false);
        }
    };

    const fetchTeachers = async () => {
        setLoadingTeachers(true);
        try {
            const res = await api.get('/admin/teachers');
            setTeachers(res.data.data);
        } catch (err) {
            console.error("Failed to fetch teachers", err);
        } finally {
            setLoadingTeachers(false);
        }
    };

    React.useEffect(() => {
        if (activeTab === 'Bans/Logs') {
            fetchBlocks();
        } else if (activeTab === 'User Management') {
            fetchTeachers();
        }
    }, [activeTab]);

    const handleUnblock = async (identifier: string) => {
        try {
            await api.delete(`/admin/verification/unblock?identifier=${identifier}`);
            fetchBlocks();
        } catch (err) {
            console.error("Failed to unblock", err);
        }
    };

    const openSubModal = (teacher: any) => {
        setSelectedTeacher(teacher);
        setSubData({
            plan: teacher.subscription_plan || 'free',
            students_limit: teacher.students_limit || 10,
            tokens_limit: teacher.tokens_limit || 100,
            duration_months: 1
        });
        setIsSubModalOpen(true);
    };

    const handlePlanChange = (plan: string) => {
        const defaults = PLAN_DEFAULTS[plan as keyof typeof PLAN_DEFAULTS] || PLAN_DEFAULTS.free;
        setSubData({
            ...subData,
            plan,
            students_limit: defaults.students_limit,
            tokens_limit: defaults.tokens_limit
        });
    };

    const handleSaveSub = async () => {
        setSavingSub(true);
        try {
            await api.put(`/admin/teachers/${selectedTeacher.id}/subscription`, subData);
            setIsSubModalOpen(false);
            fetchTeachers();
        } catch (err) {
            console.error("Failed to update subscription", err);
            alert("Failed to update subscription");
        } finally {
            setSavingSub(false);
        }
    };
    // --- Create Handler ---
    const handleCreateTeacher = async () => {
        setCreateLoading(true); setCreateError('');
        try {
            await api.post('/admin/teachers', { ...createData, birth_year: Number(createData.birth_year) });
            setIsCreateOpen(false);
            setCreateData({ first_name: '', last_name: '', email: '', phone: '', password: '', birth_year: new Date().getFullYear() - 30, gender: 'male' });
            fetchTeachers();
        } catch (err: any) {
            setCreateError(err?.response?.data?.error?.message || 'Failed to create teacher');
        } finally { setCreateLoading(false); }
    };

    // --- Edit Handler ---
    const openEditModal = (teacher: any) => {
        setSelectedTeacher(teacher);
        setEditData({ first_name: teacher.first_name || '', last_name: teacher.last_name || '', email: teacher.email || '', phone: teacher.phone || '', birth_year: teacher.birth_year || 1990, gender: teacher.gender || 'male' });
        setEditError('');
        setIsEditOpen(true);
    };
    const handleEditTeacher = async () => {
        setEditLoading(true); setEditError('');
        try {
            await api.put(`/admin/teachers/${selectedTeacher.id}`, editData);
            setIsEditOpen(false);
            fetchTeachers();
        } catch (err: any) {
            setEditError(err?.response?.data?.error?.message || 'Failed to update teacher');
        } finally { setEditLoading(false); }
    };

    // --- Password Handler ---
    const openPassModal = (teacher: any) => {
        setSelectedTeacher(teacher);
        setPassData({ new_password: '', confirm: '' });
        setPassError('');
        setIsPassOpen(true);
    };
    const handleSetPassword = async () => {
        if (passData.new_password !== passData.confirm) { setPassError('Passwords do not match'); return; }
        setPassLoading(true); setPassError('');
        try {
            await api.post(`/admin/teachers/${selectedTeacher.id}/set-password`, { new_password: passData.new_password });
            setIsPassOpen(false);
        } catch (err: any) {
            setPassError(err?.response?.data?.error?.message || 'Failed to set password');
        } finally { setPassLoading(false); }
    };

    // --- Delete Handler ---
    const handleDeleteTeacher = async () => {
        if (!deleteId) return;
        setDeleteLoading(true);
        try {
            await api.delete(`/admin/teachers/${deleteId}`);
            setDeleteId(null);
            fetchTeachers();
        } catch (err) {
            console.error('Failed to delete teacher', err);
        } finally { setDeleteLoading(false); }
    };

    const stats = [
        { label: 'Total Schools', value: '42', icon: <Building /> },
        { label: 'Total Teachers', value: '1,280', icon: <Users /> },
        { label: 'Active Students', value: '34,500', icon: <Users /> },
        { label: 'System Alerts', value: '2', icon: <AlertTriangle className="text-red-500" /> },
    ];

    return (
        <div>
            <header className="flex justify-between items-center mb-8">
                <h1 className="text-4xl font-bold font-outfit">System Overview</h1>
                <div className="flex items-center gap-4">
                    <div className="relative">
                        <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-foreground/40" />
                        <input
                            type="text"
                            placeholder="Search resources..."
                            className="bg-white/5 border border-white/10 rounded-xl py-2 pl-12 pr-4 text-sm outline-none focus:border-red-500/50"
                            suppressHydrationWarning
                        />
                    </div>
                </div>
            </header>

            {/* Dashboard Sub-Navigation */}
            <div className="flex gap-2 overflow-x-auto pb-4 mb-8">
                {['Statistics', 'Schools', 'User Management', 'AI Configs', 'Bans/Logs', 'System Health'].map((item) => (
                    <button
                        key={item}
                        onClick={() => setActiveTab(item)}
                        className={`px-5 py-2.5 rounded-xl transition-all font-medium whitespace-nowrap text-sm border ${activeTab === item
                            ? 'bg-red-500/10 text-red-500 border-red-500/20'
                            : 'bg-white/5 text-foreground/60 border-transparent hover:bg-white/10 hover:text-white'
                            }`}
                    >
                        {item}
                    </button>
                ))}
            </div>

            {activeTab === 'Statistics' && (
                <>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-12">
                        {stats.map((stat, i) => (
                            <motion.div
                                key={i}
                                initial={{ opacity: 0, scale: 0.9 }}
                                animate={{ opacity: 1, scale: 1 }}
                                transition={{ delay: i * 0.1 }}
                                className="glass-card p-6 rounded-3xl border-white/5 hover:border-white/10 transition-all"
                            >
                                <div className="flex justify-between items-center mb-4">
                                    <div className="p-3 bg-white/5 rounded-xl text-foreground/60">{stat.icon}</div>
                                </div>
                                <h3 className="text-sm font-medium text-foreground/50 mb-1">{stat.label}</h3>
                                <p className="text-3xl font-bold font-outfit">{stat.value}</p>
                            </motion.div>
                        ))}
                    </div>

                    <div className="glass-card p-10 rounded-3xl">
                        <div className="flex justify-between items-center mb-8">
                            <h2 className="text-2xl font-bold font-outfit">Recent School Registrations</h2>
                            <button className="flex items-center gap-2 px-4 py-2 bg-white/5 rounded-xl text-sm border border-white/10 hover:bg-white/10 transition-all">
                                <Filter size={16} />
                                Filters
                            </button>
                        </div>
                        <table className="w-full">
                            <thead>
                                <tr className="text-left text-foreground/40 text-sm border-b border-white/5">
                                    <th className="pb-4 font-medium">School Name</th>
                                    <th className="pb-4 font-medium">Region</th>
                                    <th className="pb-4 font-medium">Admin</th>
                                    <th className="pb-4 font-medium">Status</th>
                                    <th className="pb-4 text-right font-medium">Actions</th>
                                </tr>
                            </thead>
                            <tbody className="text-sm">
                                {[
                                    { name: 'Innovation Academy', region: 'Almaty', admin: 'Serik K.', status: 'Active' },
                                    { name: 'Future High', region: 'Astana', admin: 'Dinara A.', status: 'Pending' },
                                    { name: 'Tech Lyceum', region: 'Shymkent', admin: 'Marat O.', status: 'Active' },
                                ].map((row, i) => (
                                    <tr key={i} className="border-b border-white/5 last:border-0">
                                        <td className="py-6 font-bold">{row.name}</td>
                                        <td className="py-6 text-foreground/60">{row.region}</td>
                                        <td className="py-6 text-foreground/60">{row.admin}</td>
                                        <td className="py-6">
                                            <span className={`px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider ${row.status === 'Active' ? 'bg-green-500/10 text-green-500' : 'bg-yellow-500/10 text-yellow-500'
                                                }`}>
                                                {row.status}
                                            </span>
                                        </td>
                                        <td className="py-6 text-right">
                                            <button className="text-primary font-bold hover:underline">Manage</button>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </>
            )}

            {activeTab === 'User Management' && (
                <div className="glass-card p-10 rounded-3xl">
                    <div className="flex justify-between items-center mb-8">
                        <h2 className="text-2xl font-bold font-outfit">Teachers Management</h2>
                        <div className="flex items-center gap-3">
                            <button onClick={fetchTeachers} className="px-4 py-2 bg-white/5 rounded-xl text-sm hover:bg-white/10 transition-all">Refresh</button>
                            <button
                                onClick={() => { setCreateError(''); setIsCreateOpen(true); }}
                                className="flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-xl text-sm font-bold hover:bg-red-600 transition-all shadow-lg shadow-red-500/20"
                            >
                                <Plus size={16} /> Add Teacher
                            </button>
                        </div>
                    </div>
                    {loadingTeachers ? (
                        <p className="text-center py-10 text-foreground/40">Loading teachers...</p>
                    ) : teachers.length === 0 ? (
                        <p className="text-center py-10 text-foreground/40">No teachers found.</p>
                    ) : (
                        <div className="overflow-x-auto">
                            <table className="w-full min-w-[900px]">
                                <thead>
                                    <tr className="text-left text-foreground/40 text-sm border-b border-white/5">
                                        <th className="pb-4 font-medium">Teacher</th>
                                        <th className="pb-4 font-medium">Plan</th>
                                        <th className="pb-4 font-medium">Tokens</th>
                                        <th className="pb-4 font-medium">Students</th>
                                        <th className="pb-4 font-medium">Expires</th>
                                        <th className="pb-4 text-right font-medium">Actions</th>
                                    </tr>
                                </thead>
                                <tbody className="text-sm">
                                    {teachers.map((teacher, i) => (
                                        <tr key={i} className="border-b border-white/5 last:border-0 hover:bg-white/[0.02] transition-colors">
                                            <td className="py-4">
                                                <div className="flex flex-col">
                                                    <span className="font-bold">{teacher.first_name} {teacher.last_name}</span>
                                                    <span className="text-xs text-foreground/40">{teacher.phone || teacher.email || '—'}</span>
                                                </div>
                                            </td>
                                            <td className="py-4">
                                                <span className={`px-2 py-1 rounded-lg text-[10px] font-bold uppercase tracking-wider ${teacher.subscription_plan === 'ultra' ? 'bg-purple-500/20 text-purple-400' :
                                                        teacher.subscription_plan === 'basic' ? 'bg-blue-500/20 text-blue-400' :
                                                            'bg-slate-500/20 text-slate-400'
                                                    }`}>{teacher.subscription_plan}</span>
                                            </td>
                                            <td className="py-4 text-foreground/60">{teacher.tokens_used} / {teacher.tokens_limit === -1 ? '∞' : teacher.tokens_limit}</td>
                                            <td className="py-4 text-foreground/60">{teacher.students_count} / {teacher.students_limit === -1 ? '∞' : teacher.students_limit}</td>
                                            <td className="py-4 text-foreground/60 text-xs">{teacher.subscription_expires_at ? new Date(teacher.subscription_expires_at).toLocaleDateString() : 'Never'}</td>
                                            <td className="py-4 text-right">
                                                <div className="flex items-center justify-end gap-1">
                                                    <button onClick={() => openEditModal(teacher)} className="p-2 hover:bg-white/5 rounded-lg text-blue-400 transition-colors" title="Edit Profile"><Pencil size={16} /></button>
                                                    <button onClick={() => openPassModal(teacher)} className="p-2 hover:bg-white/5 rounded-lg text-yellow-400 transition-colors" title="Set Password"><KeyRound size={16} /></button>
                                                    <button onClick={() => openSubModal(teacher)} className="p-2 hover:bg-white/5 rounded-lg text-primary transition-colors" title="Manage Subscription"><CreditCard size={16} /></button>
                                                    {deleteId === teacher.id ? (
                                                        <div className="flex items-center gap-1 ml-1">
                                                            <button onClick={handleDeleteTeacher} disabled={deleteLoading} className="px-2 py-1 rounded-lg bg-red-500/20 text-red-400 text-xs font-bold hover:bg-red-500/30 transition-colors">{deleteLoading ? '...' : 'Confirm'}</button>
                                                            <button onClick={() => setDeleteId(null)} className="px-2 py-1 rounded-lg bg-white/5 text-foreground/40 text-xs hover:bg-white/10 transition-colors">Cancel</button>
                                                        </div>
                                                    ) : (
                                                        <button onClick={() => setDeleteId(teacher.id)} className="p-2 hover:bg-white/5 rounded-lg text-red-400 transition-colors" title="Delete"><Trash2 size={16} /></button>
                                                    )}
                                                </div>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    )}
                </div>
            )}

            {activeTab === 'Bans/Logs' && (
                <div className="glass-card p-10 rounded-3xl">
                    <div className="flex justify-between items-center mb-8">
                        <h2 className="text-2xl font-bold font-outfit">Verification Blocks</h2>
                        <button
                            onClick={fetchBlocks}
                            className="px-4 py-2 bg-white/5 rounded-xl text-sm hover:bg-white/10 transition-all"
                        >
                            Refresh
                        </button>
                    </div>
                    {loadingBlocks ? (
                        <p className="text-center py-10 text-foreground/40">Loading blocks...</p>
                    ) : !blocks || blocks.length === 0 ? (
                        <p className="text-center py-10 text-foreground/40">No active blocks found.</p>
                    ) : (
                        <table className="w-full">
                            <thead>
                                <tr className="text-left text-foreground/40 text-sm border-b border-white/5">
                                    <th className="pb-4 font-medium">Identifier</th>
                                    <th className="pb-4 font-medium">Type</th>
                                    <th className="pb-4 font-medium">Reason</th>
                                    <th className="pb-4 font-medium">Expires At</th>
                                    <th className="pb-4 text-right font-medium">Actions</th>
                                </tr>
                            </thead>
                            <tbody className="text-sm">
                                {blocks.map((block, i) => (
                                    <tr key={i} className="border-b border-white/5 last:border-0">
                                        <td className="py-6 font-bold">{block.identifier}</td>
                                        <td className="py-6 uppercase">{block.type}</td>
                                        <td className="py-6 text-foreground/60">{block.reason}</td>
                                        <td className="py-6 text-foreground/60">
                                            {block.is_permanently_blocked ? (
                                                <span className="text-red-500 font-bold">PERMANENT</span>
                                            ) : new Date(block.blocked_until).toLocaleString()}
                                        </td>
                                        <td className="py-6 text-right">
                                            <button
                                                onClick={() => handleUnblock(block.identifier)}
                                                className="text-primary font-bold hover:underline"
                                            >
                                                Unblock
                                            </button>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    )}
                </div>
            )}

            {activeTab !== 'Statistics' && activeTab !== 'Bans/Logs' && activeTab !== 'User Management' && (
                <div className="flex flex-col items-center justify-center py-20 text-foreground/40">
                    <Shield className="w-16 h-16 mb-4 opacity-10" />
                    <p className="text-lg">Content for {activeTab} is being developed.</p>
                </div>
            )}

            {/* Subscription Management Modal */}
            {isSubModalOpen && selectedTeacher && (
                <div className="fixed inset-0 z-[100] flex items-center justify-center p-6 bg-black/80 backdrop-blur-sm">
                    <motion.div initial={{ opacity: 0, scale: 0.95, y: 20 }} animate={{ opacity: 1, scale: 1, y: 0 }} className="relative w-full max-w-lg bg-background border border-white/10 rounded-[32px] p-8 shadow-2xl overflow-hidden">
                        <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-primary to-purple-500" />
                        <button onClick={() => setIsSubModalOpen(false)} className="absolute top-6 right-6 p-2 rounded-xl hover:bg-white/5 text-foreground/40"><X size={20} /></button>
                        <div className="mb-8"><h3 className="text-2xl font-bold mb-1">Manage Subscription</h3><p className="text-sm text-foreground/40">Updating limits for <span className="text-primary font-semibold">{selectedTeacher.first_name} {selectedTeacher.last_name}</span></p></div>
                        <div className="space-y-6">
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2"><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest ml-1">Plan</label><select value={subData.plan} onChange={(e) => handlePlanChange(e.target.value)} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 outline-none focus:border-primary/50 text-sm appearance-none"><option value="free">Free</option><option value="basic">Basic</option><option value="ultra">Ultra</option></select></div>
                                <div className="space-y-2"><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest ml-1">Duration (Months)</label><div className="relative"><Calendar className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-foreground/40 pointer-events-none" /><input type="number" value={subData.duration_months} onChange={(e) => setSubData({ ...subData, duration_months: parseInt(e.target.value) || 0 })} className="w-full bg-white/5 border border-white/10 rounded-2xl pl-12 pr-4 py-3 outline-none focus:border-primary/50 text-sm" min="0" max="12" /></div></div>
                            </div>
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2"><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest ml-1">Tokens Limit (-1 = ∞)</label><input type="number" value={subData.tokens_limit} onChange={(e) => setSubData({ ...subData, tokens_limit: parseInt(e.target.value) || 0 })} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 outline-none focus:border-primary/50 text-sm font-mono" /></div>
                                <div className="space-y-2"><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest ml-1">Students Limit (-1 = ∞)</label><input type="number" value={subData.students_limit} onChange={(e) => setSubData({ ...subData, students_limit: parseInt(e.target.value) || 0 })} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 outline-none focus:border-primary/50 text-sm font-mono" /></div>
                            </div>
                            <div className="pt-4"><button onClick={handleSaveSub} disabled={savingSub} className="w-full bg-primary text-white py-4 rounded-2xl font-bold flex items-center justify-center gap-2 hover:bg-red-600 transition-all shadow-xl shadow-red-500/20 disabled:opacity-50">{savingSub ? 'Processing...' : <><Save size={20} /> Save Changes</>}</button></div>
                        </div>
                    </motion.div>
                </div>
            )}

            {/* Create Teacher Modal */}
            {isCreateOpen && (
                <div className="fixed inset-0 z-[100] flex items-center justify-center p-6 bg-black/80 backdrop-blur-sm">
                    <motion.div initial={{ opacity: 0, scale: 0.95, y: 20 }} animate={{ opacity: 1, scale: 1, y: 0 }} className="relative w-full max-w-lg bg-background border border-white/10 rounded-[32px] p-8 shadow-2xl overflow-hidden">
                        <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-green-400 to-emerald-600" />
                        <button onClick={() => setIsCreateOpen(false)} className="absolute top-6 right-6 p-2 rounded-xl hover:bg-white/5 text-foreground/40"><X size={20} /></button>
                        <div className="mb-6"><h3 className="text-2xl font-bold mb-1">Add Teacher</h3><p className="text-sm text-foreground/40">Create a new teacher account</p></div>
                        <div className="space-y-4">
                            <div className="grid grid-cols-2 gap-3">
                                <div><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest mb-1 block">First Name *</label><input value={createData.first_name} onChange={e => setCreateData({ ...createData, first_name: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 outline-none focus:border-green-400/50 text-sm" placeholder="Алихан" /></div>
                                <div><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest mb-1 block">Last Name *</label><input value={createData.last_name} onChange={e => setCreateData({ ...createData, last_name: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 outline-none focus:border-green-400/50 text-sm" placeholder="Сейткали" /></div>
                            </div>
                            <div className="grid grid-cols-2 gap-3">
                                <div><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest mb-1 block">Email</label><input type="email" value={createData.email} onChange={e => setCreateData({ ...createData, email: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 outline-none focus:border-green-400/50 text-sm" placeholder="teacher@school.kz" /></div>
                                <div><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest mb-1 block">Phone</label><input value={createData.phone} onChange={e => setCreateData({ ...createData, phone: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 outline-none focus:border-green-400/50 text-sm" placeholder="+77001234567" /></div>
                            </div>
                            <div className="grid grid-cols-2 gap-3">
                                <div><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest mb-1 block">Birth Year *</label><input type="number" value={createData.birth_year} onChange={e => setCreateData({ ...createData, birth_year: Number(e.target.value) })} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 outline-none focus:border-green-400/50 text-sm" min="1950" max="2010" /></div>
                                <div><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest mb-1 block">Gender *</label><select value={createData.gender} onChange={e => setCreateData({ ...createData, gender: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 outline-none focus:border-green-400/50 text-sm appearance-none"><option value="male">Male</option><option value="female">Female</option></select></div>
                            </div>
                            <div><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest mb-1 block">Password *</label><div className="relative"><input type={showCreatePass ? 'text' : 'password'} value={createData.password} onChange={e => setCreateData({ ...createData, password: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 pr-12 outline-none focus:border-green-400/50 text-sm" placeholder="Min 8 characters" /><button type="button" onClick={() => setShowCreatePass(v => !v)} className="absolute right-4 top-1/2 -translate-y-1/2 text-foreground/40 hover:text-foreground transition-colors">{showCreatePass ? <EyeOff size={16} /> : <Eye size={16} />}</button></div></div>
                            {createError && <p className="text-red-400 text-xs font-medium bg-red-500/10 px-4 py-3 rounded-xl">{createError}</p>}
                            <div className="pt-2"><button onClick={handleCreateTeacher} disabled={createLoading || !createData.first_name || !createData.last_name || !createData.password} className="w-full bg-green-500 text-white py-4 rounded-2xl font-bold flex items-center justify-center gap-2 hover:bg-green-600 transition-all shadow-xl shadow-green-500/20 disabled:opacity-50">{createLoading ? 'Creating...' : <><Plus size={20} /> Create Teacher</>}</button></div>
                        </div>
                    </motion.div>
                </div>
            )}

            {/* Edit Teacher Modal */}
            {isEditOpen && selectedTeacher && (
                <div className="fixed inset-0 z-[100] flex items-center justify-center p-6 bg-black/80 backdrop-blur-sm">
                    <motion.div initial={{ opacity: 0, scale: 0.95, y: 20 }} animate={{ opacity: 1, scale: 1, y: 0 }} className="relative w-full max-w-lg bg-background border border-white/10 rounded-[32px] p-8 shadow-2xl overflow-hidden">
                        <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-blue-400 to-blue-600" />
                        <button onClick={() => setIsEditOpen(false)} className="absolute top-6 right-6 p-2 rounded-xl hover:bg-white/5 text-foreground/40"><X size={20} /></button>
                        <div className="mb-6"><h3 className="text-2xl font-bold mb-1">Edit Teacher</h3><p className="text-sm text-foreground/40">Editing <span className="text-blue-400 font-semibold">{selectedTeacher.first_name} {selectedTeacher.last_name}</span></p></div>
                        <div className="space-y-4">
                            <div className="grid grid-cols-2 gap-3">
                                <div><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest mb-1 block">First Name</label><input value={editData.first_name} onChange={e => setEditData({ ...editData, first_name: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 outline-none focus:border-blue-400/50 text-sm" /></div>
                                <div><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest mb-1 block">Last Name</label><input value={editData.last_name} onChange={e => setEditData({ ...editData, last_name: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 outline-none focus:border-blue-400/50 text-sm" /></div>
                            </div>
                            <div className="grid grid-cols-2 gap-3">
                                <div><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest mb-1 block">Email</label><input type="email" value={editData.email} onChange={e => setEditData({ ...editData, email: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 outline-none focus:border-blue-400/50 text-sm" /></div>
                                <div><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest mb-1 block">Phone</label><input value={editData.phone} onChange={e => setEditData({ ...editData, phone: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 outline-none focus:border-blue-400/50 text-sm" /></div>
                            </div>
                            <div className="grid grid-cols-2 gap-3">
                                <div><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest mb-1 block">Birth Year</label><input type="number" value={editData.birth_year} onChange={e => setEditData({ ...editData, birth_year: Number(e.target.value) })} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 outline-none focus:border-blue-400/50 text-sm" min="1950" max="2010" /></div>
                                <div><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest mb-1 block">Gender</label><select value={editData.gender} onChange={e => setEditData({ ...editData, gender: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 outline-none focus:border-blue-400/50 text-sm appearance-none"><option value="male">Male</option><option value="female">Female</option></select></div>
                            </div>
                            {editError && <p className="text-red-400 text-xs font-medium bg-red-500/10 px-4 py-3 rounded-xl">{editError}</p>}
                            <div className="pt-2"><button onClick={handleEditTeacher} disabled={editLoading} className="w-full bg-blue-500 text-white py-4 rounded-2xl font-bold flex items-center justify-center gap-2 hover:bg-blue-600 transition-all shadow-xl shadow-blue-500/20 disabled:opacity-50">{editLoading ? 'Saving...' : <><Save size={20} /> Save Changes</>}</button></div>
                        </div>
                    </motion.div>
                </div>
            )}

            {/* Set Password Modal */}
            {isPassOpen && selectedTeacher && (
                <div className="fixed inset-0 z-[100] flex items-center justify-center p-6 bg-black/80 backdrop-blur-sm">
                    <motion.div initial={{ opacity: 0, scale: 0.95, y: 20 }} animate={{ opacity: 1, scale: 1, y: 0 }} className="relative w-full max-w-md bg-background border border-white/10 rounded-[32px] p-8 shadow-2xl overflow-hidden">
                        <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-yellow-400 to-orange-500" />
                        <button onClick={() => setIsPassOpen(false)} className="absolute top-6 right-6 p-2 rounded-xl hover:bg-white/5 text-foreground/40"><X size={20} /></button>
                        <div className="mb-6"><h3 className="text-2xl font-bold mb-1">Set Password</h3><p className="text-sm text-foreground/40">New password for <span className="text-yellow-400 font-semibold">{selectedTeacher.first_name} {selectedTeacher.last_name}</span></p></div>
                        <div className="space-y-4">
                            <div><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest mb-1 block">New Password</label><div className="relative"><input type={showNewPass ? 'text' : 'password'} value={passData.new_password} onChange={e => setPassData({ ...passData, new_password: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 pr-12 outline-none focus:border-yellow-400/50 text-sm" placeholder="Min 8 characters" /><button type="button" onClick={() => setShowNewPass(v => !v)} className="absolute right-4 top-1/2 -translate-y-1/2 text-foreground/40 hover:text-foreground">{showNewPass ? <EyeOff size={16} /> : <Eye size={16} />}</button></div></div>
                            <div><label className="text-xs font-bold text-foreground/40 uppercase tracking-widest mb-1 block">Confirm Password</label><input type="password" value={passData.confirm} onChange={e => setPassData({ ...passData, confirm: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-2xl px-4 py-3 outline-none focus:border-yellow-400/50 text-sm" placeholder="Repeat password" /></div>
                            {passError && <p className="text-red-400 text-xs font-medium bg-red-500/10 px-4 py-3 rounded-xl">{passError}</p>}
                            <div className="pt-2"><button onClick={handleSetPassword} disabled={passLoading || !passData.new_password || !passData.confirm} className="w-full bg-yellow-500 text-black py-4 rounded-2xl font-bold flex items-center justify-center gap-2 hover:bg-yellow-400 transition-all shadow-xl shadow-yellow-500/20 disabled:opacity-50">{passLoading ? 'Setting...' : <><KeyRound size={20} /> Set Password</>}</button></div>
                        </div>
                    </motion.div>
                </div>
            )}
        </div>
    );
}
