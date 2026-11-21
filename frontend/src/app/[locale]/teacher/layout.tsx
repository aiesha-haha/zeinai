'use client';

import React from 'react';
import { usePathname } from '@/i18n/routing';
import { useTranslations } from 'next-intl';
import {
    LayoutDashboard, Users, FileText, PenTool, Gamepad2, Layers, ClipboardCheck,
    Microscope, Trophy, Newspaper, Award, User,
} from 'lucide-react';
import AppShell, { NavGroup } from '@/components/shell/AppShell';
import RightRail from '@/components/shell/RightRail';

export default function TeacherLayout({ children }: { children: React.ReactNode }) {
    const pathname = usePathname();
    const tNav = useTranslations('nav');
    const tM = useTranslations('education_materials');
    const tShell = useTranslations('shell');

    const groups: NavGroup[] = [
        {
            items: [
                { icon: <LayoutDashboard size={20} />, label: tNav('dashboard'), href: '/teacher' },
                { icon: <Users size={20} />, label: tNav('my_classes'), href: '/teacher/classes', match: ['/teacher/whiteboard'] },
                { icon: <FileText size={20} />, label: tShell('kmj'), href: '/teacher/materials/short-term-plan' },
            ],
        },
        {
            title: tNav('materials'),
            items: [
                {
                    icon: <PenTool size={20} />, label: tM('sections.generation'), href: '/teacher/materials/generation',
                    match: ['/teacher/materials/slides', '/teacher/materials/article', '/teacher/materials/image', '/teacher/materials/video', '/teacher/materials/feedback'],
                },
                { icon: <Gamepad2 size={20} />, label: tShell('games'), href: '/teacher/materials/games', match: ['/teacher/materials/tug-of-war'] },
                {
                    icon: <Layers size={20} />, label: tM('sections.interactive'), href: '/teacher/materials/interactive',
                    match: ['/teacher/materials/whiteboard', '/teacher/materials/group-tasks', '/teacher/materials/face-picker'],
                },
                {
                    icon: <ClipboardCheck size={20} />, label: tM('sections.assignments'), href: '/teacher/materials/assignments',
                    match: ['/teacher/materials/tests', '/teacher/materials/pisa', '/teacher/materials/unt', '/teacher/tests'],
                },
                { icon: <Microscope size={20} />, label: tM('sections.laboratory'), href: '/teacher/materials/laboratory', match: ['/teacher/materials/lab'] },
                { icon: <Trophy size={20} />, label: tM('sections.olympiad'), href: '/teacher/materials/olympiad' },
                { icon: <Newspaper size={20} />, label: tM('sections.news'), href: '/teacher/materials/news-category', match: ['/teacher/materials/news'] },
                {
                    icon: <Award size={20} />, label: tM('sections.achievements'), href: '/teacher/materials/achievements-category',
                    match: ['/teacher/materials/achievements', '/teacher/materials/certificates'],
                },
            ],
        },
        { items: [{ icon: <User size={20} />, label: tNav('profile'), href: '/teacher/profile' }] },
    ];

    const searchEntries = [
        { label: tM('tools.short_term_plan'), href: '/teacher/materials/short-term-plan' },
        { label: tM('tools.slides'), href: '/teacher/materials/slides' },
        { label: tM('tools.article'), href: '/teacher/materials/article' },
        { label: tM('tools.video'), href: '/teacher/materials/video' },
        { label: tM('tools.image'), href: '/teacher/materials/image' },
        { label: tM('tools.feedback'), href: '/teacher/materials/feedback' },
        { label: tM('tools.test_generation'), href: '/teacher/materials/tests' },
        { label: tM('tools.pisa_tests'), href: '/teacher/materials/pisa' },
        { label: tM('tools.unt_tests'), href: '/teacher/materials/unt' },
        { label: tM('tools.lab_work'), href: '/teacher/materials/lab' },
        { label: tM('tools.olympiad_tasks'), href: '/teacher/materials/olympiad' },
        { label: tM('tools.news'), href: '/teacher/materials/news' },
        { label: tM('tools.certificates'), href: '/teacher/materials/certificates' },
        { label: tM('tools.achievements'), href: '/teacher/materials/achievements' },
        { label: tM('tools.tug_of_war'), href: '/teacher/materials/tug-of-war' },
        { label: tM('tools.group_tasks'), href: '/teacher/materials/group-tasks' },
        { label: tM('tools.whiteboard'), href: '/teacher/materials/whiteboard' },
        { label: tM('tools.face_picker'), href: '/teacher/materials/face-picker' },
        { label: tShell('assigned_results'), href: '/teacher/tests/assignments' },
    ];

    return (
        <AppShell
            role="teacher"
            groups={groups}
            searchEntries={searchEntries}
            rightRail={pathname === '/teacher' ? <RightRail role="teacher" /> : undefined}
        >
            {children}
        </AppShell>
    );
}
