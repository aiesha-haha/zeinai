'use client';

import { useEffect } from 'react';
import { useRouter } from '@/i18n/routing';

export default function MaterialsPage() {
    const router = useRouter();

    useEffect(() => {
        router.replace('/teacher/materials/generation');
    }, [router]);

    return null;
}
