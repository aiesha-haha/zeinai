'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { Link } from '@/i18n/routing';
import { ChevronRight } from 'lucide-react';
import PageHeader from '@/components/ui/PageHeader';

export type Tool = { href: string; name: string; desc?: string; icon: React.ReactNode };

/** A section page that lists tools as keyboard-accessible cards */
export default function ToolGrid({ title, description, backHref = '/teacher', tools }: {
    title: string;
    description?: string;
    backHref?: string;
    tools: Tool[];
}) {
    const t = useTranslations('education_materials');
    return (
        <div>
            <PageHeader title={title} description={description} backHref={backHref} />
            <ul className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
                {tools.map((tool) => (
                    <li key={tool.href}>
                        <Link href={tool.href}
                            className="glass-card rounded-3xl p-5 h-full flex items-center gap-4 hover:border-primary/40 transition-colors group">
                            <span className="w-14 h-14 rounded-2xl bg-primary text-white flex items-center justify-center shrink-0 group-hover:scale-105 transition-transform">
                                {tool.icon}
                            </span>
                            <span className="flex-1 min-w-0">
                                <span className="block font-bold text-base group-hover:text-primary">{tool.name}</span>
                                {tool.desc && <span className="block text-sm text-muted mt-0.5">{tool.desc}</span>}
                            </span>
                            <ChevronRight size={18} className="text-muted shrink-0" aria-hidden />
                            <span className="sr-only">{t('open_tool')}</span>
                        </Link>
                    </li>
                ))}
            </ul>
        </div>
    );
}
