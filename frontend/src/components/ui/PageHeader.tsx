'use client';

import React from 'react';
import { ChevronLeft } from 'lucide-react';
import { Link } from '@/i18n/routing';
import { useTranslations } from 'next-intl';

export default function PageHeader({ title, description, backHref, actions }: {
    title: string;
    description?: string;
    backHref?: string;
    actions?: React.ReactNode;
}) {
    const tCommon = useTranslations('common');
    return (
        <div className="flex flex-wrap items-start gap-4 mb-6">
            {backHref && (
                <Link href={backHref} className="p-2.5 rounded-xl bg-surface border border-line hover:border-primary/40 text-foreground"
                    aria-label={tCommon('back')}>
                    <ChevronLeft size={20} />
                </Link>
            )}
            <div className="flex-1 min-w-[200px]">
                <h1 className="text-2xl md:text-3xl font-bold font-outfit">{title}</h1>
                {description && <p className="text-muted mt-1 max-w-2xl">{description}</p>}
            </div>
            {actions && <div className="flex flex-wrap gap-2">{actions}</div>}
        </div>
    );
}
