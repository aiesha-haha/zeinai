'use client';

import React from 'react';
import { Shield, LayoutDashboard, User, LogOut } from 'lucide-react';
import { Link, usePathname, useRouter } from '@/i18n/routing';
import { logout } from '@/lib/auth';
import NotificationBell from '@/components/notifications/NotificationBell';
import LanguageSwitcher from '@/components/LanguageSwitcher';
import { ThemeToggle } from '@/components/ThemeToggle';

export default function AdminLayout({ children }: { children: React.ReactNode }) {
    const pathname = usePathname();
    const router = useRouter();

    React.useEffect(() => {
        if (typeof window !== 'undefined') {
            const token = localStorage.getItem('access_token');
            const refreshToken = localStorage.getItem('refresh_token');
            const role = localStorage.getItem('user_role');

            if (!token && !refreshToken) {
                router.push('/login');
            } else if (token && role !== 'admin') {
                router.push('/login');
            }
        }
    }, [router]);

    const menuItems = [
        { icon: <LayoutDashboard size={20} />, label: 'Dashboard', href: '/admin' },
        { icon: <User size={20} />, label: 'Profile', href: '/admin/profile' },
    ];

    return (
        <div className="flex min-h-screen bg-background">
            {/* Admin Sidebar */}
            <aside className="w-72 glass-card border-none rounded-none border-r border-slate-200 dark:border-white/5 p-6 flex flex-col fixed h-screen z-50">
                <Link href="/admin" className="text-2xl font-bold tracking-tighter mb-12 flex items-center gap-2">
                    <Shield className="text-red-500" />
                    <span className="text-gradient">Admin Panel</span>
                </Link>

                <nav className="flex-1 space-y-2">
                    <div className="px-4 py-3 bg-red-500/10 text-red-500 rounded-xl font-bold text-sm mb-4">
                        Authorized Personnel Only
                    </div>
                    {menuItems.map((item) => {
                        const isActive = pathname === item.href;
                        return (
                            <Link
                                key={item.href}
                                href={item.href}
                                className={`flex items-center gap-4 px-4 py-3.5 rounded-2xl transition-all font-medium ${isActive
                                    ? 'bg-red-500/10 text-red-500 shadow-[0_0_20px_-5px_rgba(239,68,68,0.3)]'
                                    : 'text-foreground/60 hover:text-foreground hover:bg-slate-100 dark:hover:bg-white/5'
                                    }`}
                            >
                                {item.icon}
                                {item.label}
                            </Link>
                        );
                    })}
                </nav>

                <div className="mt-auto pt-6 border-t border-slate-200 dark:border-white/5">
                    <button
                        onClick={() => logout()}
                        className="flex items-center gap-4 px-4 py-3.5 w-full rounded-2xl text-foreground/60 hover:bg-red-500/10 hover:text-red-500 transition-all font-semibold"
                    >
                        <LogOut size={20} />
                        Logout
                    </button>
                </div>
            </aside>

            {/* Main Content */}
            <main className="flex-1 ml-72 p-10 min-h-screen">
                <header className="flex justify-between items-center mb-8">
                    <h1 className="text-2xl font-bold">Admin Panel</h1>
                    <div className="flex items-center gap-4">
                        <LanguageSwitcher />
                        <ThemeToggle />
                        <NotificationBell />
                    </div>
                </header>
                {children}
            </main>
        </div>
    );
}
