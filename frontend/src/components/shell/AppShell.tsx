'use client';

import React from 'react';
import { Link, usePathname, useRouter } from '@/i18n/routing';
import { useTranslations } from 'next-intl';
import { LogOut, Menu, Search, X, User as UserIcon, GraduationCap } from 'lucide-react';
import NotificationBell from '@/components/notifications/NotificationBell';
import LanguageSwitcher from '@/components/LanguageSwitcher';
import { ThemeToggle } from '@/components/ThemeToggle';
import { logout } from '@/lib/auth';
import api from '@/lib/api';

export type NavItem = {
    icon: React.ReactNode;
    label: string;
    href: string;
    /** extra path prefixes that should mark this item active */
    match?: string[];
};

export type NavGroup = { title?: string; items: NavItem[] };

/** Extra (non-menu) pages that the top search can open */
export type SearchEntry = { label: string; href: string };

type Props = {
    role: 'teacher' | 'student';
    groups: NavGroup[];
    searchEntries?: SearchEntry[];
    rightRail?: React.ReactNode;
    children: React.ReactNode;
};

export function avatarSrc(url?: string | null) {
    if (!url) return null;
    return url.startsWith('http') ? url : `${process.env.NEXT_PUBLIC_API_URL || ''}${url}`;
}

/**
 * Layout used by the teacher and student cabinets:
 * dark-blue sidebar (left), search + profile bar (top), content (centre),
 * optional calendar rail (right). On phones the sidebar becomes a drawer.
 */
export default function AppShell({ role, groups, searchEntries = [], rightRail, children }: Props) {
    const pathname = usePathname();
    const router = useRouter();
    const t = useTranslations('shell');
    const tAuth = useTranslations('auth');
    const tProfile = useTranslations('profile');

    const [profile, setProfile] = React.useState<any>(null);
    const [drawerOpen, setDrawerOpen] = React.useState(false);
    const [query, setQuery] = React.useState('');
    const [searchOpen, setSearchOpen] = React.useState(false);
    const searchRef = React.useRef<HTMLDivElement>(null);

    // Auth + role guard (the server also checks the role on every API call)
    React.useEffect(() => {
        const token = localStorage.getItem('access_token');
        const refresh = localStorage.getItem('refresh_token');
        const storedRole = localStorage.getItem('user_role');
        if (!token && !refresh) {
            router.push('/login');
            return;
        }
        const allowed = role === 'teacher' ? ['teacher', 'admin'] : ['student'];
        if (storedRole && !allowed.includes(storedRole)) {
            router.push(storedRole === 'student' ? '/student' : storedRole === 'admin' ? '/admin' : '/teacher');
            return;
        }
        api.get('/users/me')
            .then((res) => {
                const p = res.data.data;
                setProfile(p);
                if (p?.role && p.role !== storedRole) localStorage.setItem('user_role', p.role);
                if (p?.role && !allowed.includes(p.role)) {
                    router.push(p.role === 'student' ? '/student' : '/teacher');
                }
            })
            .catch(() => { /* 401 is handled by the api interceptor */ });
    }, [role, router]);

    // Close drawer on navigation, close on Escape
    React.useEffect(() => { setDrawerOpen(false); }, [pathname]);
    React.useEffect(() => {
        const onKey = (e: KeyboardEvent) => {
            if (e.key === 'Escape') { setDrawerOpen(false); setSearchOpen(false); }
        };
        const onClick = (e: MouseEvent) => {
            if (searchRef.current && !searchRef.current.contains(e.target as Node)) setSearchOpen(false);
        };
        window.addEventListener('keydown', onKey);
        document.addEventListener('mousedown', onClick);
        return () => { window.removeEventListener('keydown', onKey); document.removeEventListener('mousedown', onClick); };
    }, []);

    const allItems = groups.flatMap((g) => g.items);
    const isActive = (item: NavItem) => {
        const home = role === 'teacher' ? '/teacher' : '/student';
        if (item.href === home) return pathname === home;
        return pathname === item.href || pathname.startsWith(item.href + '/') ||
            (item.match || []).some((m) => pathname === m || pathname.startsWith(m + '/'));
    };

    const searchPool: SearchEntry[] = [...allItems.map((i) => ({ label: i.label, href: i.href })), ...searchEntries];
    const results = query.trim()
        ? searchPool.filter((e, idx, arr) =>
            e.label.toLowerCase().includes(query.trim().toLowerCase()) &&
            arr.findIndex((x) => x.href === e.href) === idx).slice(0, 7)
        : [];

    const roleLabel = profile?.role === 'admin' ? tProfile('role_admin')
        : profile?.role === 'student' ? tProfile('role_student') : tProfile('role_teacher');
    const fullName = profile ? `${profile.first_name || ''} ${profile.last_name || ''}`.trim() : '';
    const avatar = avatarSrc(profile?.avatar_url);
    const profileHref = role === 'teacher' ? '/teacher/profile' : '/student/profile';

    const sidebar = (
        <div className="flex h-full flex-col text-white"
            style={{ background: 'linear-gradient(to right, var(--sidebar-deep) 0, var(--sidebar-deep) 64px, var(--sidebar) 64px)' }}>
            <div className="flex items-center gap-3 h-20 pl-4 pr-4">
                <div className="w-9 h-9 rounded-xl bg-white/15 flex items-center justify-center shrink-0" aria-hidden>
                    <GraduationCap size={20} />
                </div>
                <Link href={role === 'teacher' ? '/teacher' : '/student'} className="pl-3 text-xl font-bold tracking-tight font-outfit">
                    ZeinAI
                </Link>
            </div>

            <nav aria-label={t('main_menu')} className="flex-1 overflow-y-auto overflow-x-hidden pb-4 [scrollbar-width:thin]">
                {groups.map((group, gi) => (
                    <div key={gi} className="mb-2">
                        {group.title && (
                            <p className="pl-[76px] pr-4 pt-4 pb-1 text-[11px] font-semibold uppercase tracking-wider text-white/60">
                                {group.title}
                            </p>
                        )}
                        <ul className="space-y-1">
                            {group.items.map((item) => {
                                const active = isActive(item);
                                return (
                                    <li key={item.href} className="relative">
                                        <Link
                                            href={item.href}
                                            aria-current={active ? 'page' : undefined}
                                            className={`group flex items-center h-11 focus-visible:outline-white ${active ? '' : 'hover:bg-white/10'}`}
                                        >
                                            <span className={`w-16 shrink-0 flex justify-center ${active ? 'text-white' : 'text-white/75 group-hover:text-white'}`}>
                                                {item.icon}
                                            </span>
                                            <span className={`ml-3 flex-1 h-full flex items-center pl-4 pr-3 text-sm font-semibold rounded-l-2xl transition-colors ${active
                                                ? 'bg-[var(--background)] text-primary dark:text-white'
                                                : 'text-white/85 group-hover:text-white'}`}>
                                                <span className="truncate">{item.label}</span>
                                            </span>
                                        </Link>
                                        {active && <span className="absolute left-0 top-2 bottom-2 w-1 rounded-r bg-white" aria-hidden />}
                                    </li>
                                );
                            })}
                        </ul>
                    </div>
                ))}
            </nav>

            <div className="border-t border-white/15 py-3">
                <button
                    onClick={() => logout()}
                    className="flex items-center h-11 w-full text-white/85 hover:text-white hover:bg-white/10"
                >
                    <span className="w-16 flex justify-center"><LogOut size={20} /></span>
                    <span className="ml-3 pl-4 text-sm font-semibold">{tAuth('logout')}</span>
                </button>
            </div>
        </div>
    );

    return (
        <div className="min-h-screen bg-background">
            <a href="#main" className="sr-only focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-[100] btn-primary">
                {t('skip_to_content')}
            </a>

            {/* Desktop sidebar */}
            <aside className="hidden lg:block fixed inset-y-0 left-0 w-64 z-40 rounded-r-[28px] overflow-hidden shadow-xl shadow-primary/10">
                {sidebar}
            </aside>

            {/* Mobile drawer */}
            {drawerOpen && (
                <div className="lg:hidden fixed inset-0 z-[70]" role="dialog" aria-modal="true" aria-label={t('main_menu')}>
                    <div className="absolute inset-0 bg-black/40" onClick={() => setDrawerOpen(false)} />
                    <div className="absolute inset-y-0 left-0 w-72 max-w-[85vw] rounded-r-[28px] overflow-hidden shadow-2xl">
                        <button
                            onClick={() => setDrawerOpen(false)}
                            className="absolute top-5 right-4 z-10 p-2 rounded-xl text-white bg-white/10 hover:bg-white/20"
                            aria-label={t('close_menu')}
                        >
                            <X size={18} />
                        </button>
                        {sidebar}
                    </div>
                </div>
            )}

            <div className="lg:pl-64">
                {/* Top bar */}
                <header className="sticky top-0 z-30 bg-background/90 backdrop-blur border-b border-line/60">
                    <div className="flex items-center gap-3 px-4 md:px-8 h-16 md:h-20">
                        <button
                            onClick={() => setDrawerOpen(true)}
                            className="lg:hidden p-2.5 rounded-xl bg-surface border border-line"
                            aria-label={t('open_menu')}
                        >
                            <Menu size={20} />
                        </button>

                        {/* Search (pages & tools) */}
                        <div ref={searchRef} className="relative flex-1 max-w-md">
                            <label htmlFor="shell-search" className="sr-only">{t('search_label')}</label>
                            <Search size={18} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-muted pointer-events-none" />
                            <input
                                id="shell-search"
                                type="search"
                                value={query}
                                onChange={(e) => { setQuery(e.target.value); setSearchOpen(true); }}
                                onFocus={() => setSearchOpen(true)}
                                onKeyDown={(e) => {
                                    if (e.key === 'Enter' && results[0]) {
                                        router.push(results[0].href);
                                        setQuery(''); setSearchOpen(false);
                                    }
                                }}
                                placeholder={t('search_placeholder')}
                                className="field !py-2.5 !pl-10 !rounded-full"
                                autoComplete="off"
                            />
                            {searchOpen && query.trim() && (
                                <ul className="absolute mt-2 w-full glass-card rounded-2xl p-1.5 z-50" role="listbox">
                                    {results.length === 0 ? (
                                        <li className="px-3 py-2.5 text-sm text-muted">{t('search_empty')}</li>
                                    ) : results.map((r) => (
                                        <li key={r.href}>
                                            <Link
                                                href={r.href}
                                                onClick={() => { setQuery(''); setSearchOpen(false); }}
                                                className="block px-3 py-2.5 rounded-xl text-sm font-medium hover:bg-primary/10"
                                            >
                                                {r.label}
                                            </Link>
                                        </li>
                                    ))}
                                </ul>
                            )}
                        </div>

                        <div className="ml-auto flex items-center gap-1.5 md:gap-3">
                            <div className="hidden sm:flex items-center gap-1.5"><LanguageSwitcher /><ThemeToggle /></div>
                            <NotificationBell />
                            <Link href={profileHref} className="flex items-center gap-3 pl-2 md:pl-4 md:border-l border-line rounded-xl">
                                <span className="w-10 h-10 rounded-full overflow-hidden bg-primary/15 text-primary flex items-center justify-center font-bold">
                                    {avatar ? (
                                        // eslint-disable-next-line @next/next/no-img-element
                                        <img src={avatar} alt="" className="w-full h-full object-cover" />
                                    ) : profile ? (
                                        <span>{(profile.first_name?.[0] || '')}{(profile.last_name?.[0] || '')}</span>
                                    ) : <UserIcon size={18} />}
                                </span>
                                <span className="hidden md:block text-left leading-tight">
                                    <span className="block text-sm font-bold max-w-[160px] truncate">{fullName || t('loading')}</span>
                                    <span className="block text-xs text-muted">{roleLabel}</span>
                                </span>
                            </Link>
                        </div>
                    </div>
                </header>

                <div className="flex gap-6 px-4 md:px-8 py-6 md:py-8">
                    <main id="main" className="flex-1 min-w-0 pb-10">
                        {children}
                    </main>
                    {rightRail && (
                        <aside className="hidden xl:block w-[320px] shrink-0" aria-label={t('calendar')}>
                            <div className="sticky top-28 space-y-6">{rightRail}</div>
                        </aside>
                    )}
                </div>
            </div>
        </div>
    );
}
