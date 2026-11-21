'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { Play, Users, Swords, Camera, Gamepad2 } from 'lucide-react';
import ToolGrid from '@/components/materials/ToolGrid';

export default function InteractivePage() {
    const t = useTranslations('education_materials');
    const d = useTranslations('tool_desc');
    const tShell = useTranslations('shell');
    return (
        <ToolGrid
            title={t('sections.interactive')}
            tools={[
                { href: '/teacher/materials/whiteboard', name: t('tools.whiteboard'), desc: d('whiteboard'), icon: <Play size={24} /> },
                { href: '/teacher/materials/group-tasks', name: t('tools.group_tasks'), desc: d('group_tasks'), icon: <Users size={24} /> },
                { href: '/teacher/materials/tug-of-war', name: t('tools.tug_of_war'), desc: d('tug_of_war'), icon: <Swords size={24} /> },
                { href: '/teacher/materials/games', name: tShell('games'), desc: d('games'), icon: <Gamepad2 size={24} /> },
                { href: '/teacher/materials/face-picker', name: t('tools.face_picker'), desc: d('face_picker'), icon: <Camera size={24} /> },
            ]}
        />
    );
}
