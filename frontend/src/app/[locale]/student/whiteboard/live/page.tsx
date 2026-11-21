"use client";
import React from 'react';
import { Suspense, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import api from "@/lib/api";
import { useRouter } from "@/i18n/routing";
import { useSearchParams, useParams } from "next/navigation";
import { Loader2, AlertCircle } from "lucide-react";

function LiveSessionContent() {
    const t = useTranslations('whiteboard_locator');
    const router = useRouter();
    const searchParams = useSearchParams();
    const params = useParams();
    const classId = searchParams.get("classId");
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!classId) {
            setError(t('class_id_missing'));
            return;
        }
        findActiveSession();
    }, [classId]);

    const findActiveSession = async () => {
        try {
            const res = await api.get("/whiteboard/active");

            // If the active session is found.
            if (res.data && res.data.data && res.data.data.id) {
                router.push(`/student/whiteboard/${res.data.data.id}`);
            } else {
                setError(t('no_session'));
            }
        } catch (err) {
            setError(t('no_lesson'));
        }
    };

    return (
        <div className="flex flex-col items-center justify-center min-h-[60vh] p-4 text-center">
            {error ? (
                <div className="space-y-4">
                    <AlertCircle className="w-12 h-12 text-zinc-400 mx-auto" />
                    <h2 className="text-xl font-bold dark:text-white">{error}</h2>
                    <button
                        onClick={() => window.history.back()}
                        className="text-blue-500 hover:underline"
                    >
                        {t('go_back')}
                    </button>
                </div>
            ) : (
                <div className="space-y-4">
                    <Loader2 className="w-12 h-12 text-primary animate-spin mx-auto" />
                    <h2 className="text-xl font-bold dark:text-white">{t('locating')}</h2>
                </div>
            )}
        </div>
    );
}

export default function LiveSessionLocator() {
    return (
        <Suspense fallback={
            <div className="flex items-center justify-center min-h-[60vh]">
                <Loader2 className="w-12 h-12 text-primary animate-spin" />
            </div>
        }>
            <LiveSessionContent />
        </Suspense>
    );
}
