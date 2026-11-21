'use client';

import React from 'react';
import { Link, useRouter } from '@/i18n/routing';
import { Shield, AlertTriangle, LogOut, ChevronLeft } from 'lucide-react';
import { logout } from '@/lib/auth';
import { useTranslations } from 'next-intl';
import ProfileContent from '@/components/profile/ProfileContent';

export default function AdminProfilePage() {
    const router = useRouter();
    const tNav = useTranslations('nav');

    React.useEffect(() => {
        if (typeof window !== 'undefined') {
            const token = localStorage.getItem('access_token');
            const role = localStorage.getItem('user_role');

            if (!token || role !== 'admin') {
                router.push('/login');
            }
        }
    }, [router]);

    return (
        <div className="bg-[#020617]">
            <ProfileContent />
        </div>
    );
}
