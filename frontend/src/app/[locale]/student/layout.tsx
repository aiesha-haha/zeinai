'use client';

import React from 'react';
import { usePathname } from '@/i18n/routing';
import { useTranslations } from 'next-intl';
import { Home, Trophy, History as HistoryIcon, User } from 'lucide-react';
import AppShell, { NavGroup } from '@/components/shell/AppShell';
import RightRail from '@/components/shell/RightRail';

export default function StudentLayout({ children }: { children: React.ReactNode }) {
    const pathname = usePathname();
    const tNav = useTranslations('nav');
    const tShell = useTranslations('shell');

    const groups: NavGroup[] = [
        {
            items: [
                { icon: <Home size={20} />, label: tNav('home'), href: '/student', match: ['/student/dashboard'] },
                { icon: <Trophy size={20} />, label: tShell('tasks_and_games'), href: '/student/tests' },
                { icon: <HistoryIcon size={20} />, label: tNav('history'), href: '/student/history', match: ['/student/whiteboard'] },
                { icon: <User size={20} />, label: tNav('profile'), href: '/student/profile' },
            ],
        },
    ];

    return (
        <AppShell
            role="student"
            groups={groups}
            rightRail={pathname === '/student' ? <RightRail role="student" /> : undefined}
        >
            {children}
        </AppShell>
    );
}
