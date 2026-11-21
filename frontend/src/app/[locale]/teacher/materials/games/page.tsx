'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { Link } from '@/i18n/routing';
import { Gamepad2, Swords, ToggleLeft, Users, ListChecks } from 'lucide-react';
import TestTypeHub from '@/components/materials/TestTypeHub';

/** Educational games: AI quiz, true/false, tug of war (solo for students) and group tasks (in class). */
export default function GamesPage() {
    const t = useTranslations('games_hub');
    const tTypes = useTranslations('test_types');
    const tM = useTranslations('education_materials');

    const kinds = [
        { href: '/teacher/materials/tests/create?type=game_quiz', icon: <ListChecks size={22} />, title: tTypes('game_quiz.title'), desc: tTypes('game_quiz.desc') },
        { href: '/teacher/materials/tests/create?type=game_true_false', icon: <ToggleLeft size={22} />, title: tTypes('game_true_false.title'), desc: tTypes('game_true_false.desc') },
        { href: '/teacher/materials/tug-of-war', icon: <Swords size={22} />, title: tM('tools.tug_of_war'), desc: t('tug_desc') },
        { href: '/teacher/materials/group-tasks', icon: <Users size={22} />, title: tM('tools.group_tasks'), desc: t('group_desc') },
    ];

    return (
        <TestTypeHub
            types={['game_quiz', 'game_true_false', 'tug_of_war']}
            title={t('title')}
            description={t('desc')}
            backHref="/teacher"
            selfHref="/teacher/materials/games"
            createType="game_quiz"
            createLabel={t('create')}
            emptyText={t('empty')}
            icon={<Gamepad2 size={20} />}
            extra={
                <section className="mb-6">
                    <h2 className="text-lg font-bold font-outfit mb-3">{t('choose_type')}</h2>
                    <ul className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
                        {kinds.map((k) => (
                            <li key={k.href}>
                                <Link href={k.href} className="glass-card rounded-3xl p-5 h-full flex flex-col gap-3 hover:border-primary/40 transition-colors">
                                    <span className="w-11 h-11 rounded-xl bg-primary/10 text-primary flex items-center justify-center">{k.icon}</span>
                                    <span className="font-bold">{k.title}</span>
                                    <span className="text-sm text-muted">{k.desc}</span>
                                </Link>
                            </li>
                        ))}
                    </ul>
                    <h2 className="text-lg font-bold font-outfit mt-8 mb-1">{t('my_games')}</h2>
                    <p className="text-sm text-muted mb-3">{t('my_games_hint')}</p>
                </section>
            }
        />
    );
}
