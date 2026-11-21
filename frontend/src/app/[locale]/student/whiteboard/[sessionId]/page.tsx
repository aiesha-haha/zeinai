"use client";
import React from 'react';
import { useState, useEffect, use, useRef } from "react";
import { useTranslations } from "next-intl";
import api from "@/lib/api";
import { Loader2, Info } from "lucide-react";
import { WhiteboardAction, WhiteboardSession } from "@/types/whiteboard";
import { WhiteboardCanvas } from "@/components/Whiteboard/WhiteboardCanvas";

interface StudentWhiteboardPageProps {
    params: Promise<{
        sessionId: string;
        locale: string;
    }>;
}

export default function StudentWhiteboardPage({ params }: StudentWhiteboardPageProps) {
    const t = useTranslations('whiteboard_live');
    const resolvedParams = use(params);
    const [session, setSession] = useState<WhiteboardSession | null>(null);
    const [loading, setLoading] = useState(true);
    const ws = useRef<WebSocket | null>(null);
    const [incoming, setIncoming] = useState<WhiteboardAction | null>(null);
    const [currentPage, setCurrentPage] = useState(0);
    const [pageHistory, setPageHistory] = useState<WhiteboardAction[]>([]);

    useEffect(() => {
        fetchSession();
        return () => { ws.current?.close(); };
    }, []);

    const fetchSession = async () => {
        try {
            const sessionId = resolvedParams.sessionId;
            let sessionData = null;

            if (sessionId && sessionId !== 'live') {
                // Try to get the specific session by ID
                try {
                    const res = await api.get(`/whiteboard/session/${sessionId}`);
                    if (res.data?.data) {
                        sessionData = res.data.data;
                        // If session already ended, redirect to summary
                        if (sessionData.ended_at || sessionData.status === 'ended') {
                            window.location.href = `/student/whiteboard/summary/${sessionId}`;
                            return;
                        }
                    }
                } catch (err: any) {
                    if (err.response?.status === 404) {
                        // Session ended or doesn't exist — go to summary
                        window.location.href = `/student/whiteboard/summary/${sessionId}`;
                        return;
                    }
                }
            } else {
                // Fallback: get the currently active session
                try {
                    const res = await api.get("/whiteboard/active");
                    if (res.data?.data) {
                        sessionData = res.data.data;
                    }
                } catch {
                    // No active session — will show "no active session" UI
                }
            }

            if (sessionData) {
                setSession(sessionData);
                initWebSocket(sessionData.class_id);
            }
        } catch (err) {
            console.error("Failed to fetch session", err);
        } finally {
            setLoading(false);
        }
    };

    const initWebSocket = (cid: number) => {
        const wsUrl = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8000";
        const socket = new WebSocket(`${wsUrl}/api/v1/ws/whiteboard/${cid}`);

        socket.onopen = () => {
            // Request full sync when connected
            socket.send(JSON.stringify({ type: "request_sync" }));
        };

        socket.onmessage = (event) => {
            const data = JSON.parse(event.data);

            if (data.type === "SESSION_ENDED") {
                alert(t('session_ended'));
                window.location.href = `/student/whiteboard/summary/${session?.id || resolvedParams.sessionId}`;
                return;
            }

            if (data.type === "page_sync") {
                setPageHistory(data.actions || []);
                setCurrentPage(data.page_index ?? 0);
                return;
            }

            if (data.type === "page_change") {
                setCurrentPage(data.page_index ?? 0);
                return;
            }

            setIncoming(data);
        };

        ws.current = socket;
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[60vh]">
                <div className="flex flex-col items-center gap-4">
                    <Loader2 className="w-8 h-8 animate-spin text-zinc-500" />
                    <p className="text-zinc-500 text-sm animate-pulse">{t('loading')}</p>
                </div>
            </div>
        );
    }

    if (!session) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[60vh] text-center p-4">
                <Info className="w-12 h-12 text-zinc-400 mb-4" />
                <h2 className="text-xl font-bold dark:text-white">{t('no_active_session')}</h2>
                <p className="text-zinc-500">{t('wait_for_teacher')}</p>
            </div>
        );
    }

    return (
        <div className="flex flex-col h-[calc(100dvh-120px)] p-4">
            <div className="flex items-center justify-between mb-4">
                <div>
                    <h1 className="text-2xl font-bold dark:text-white">{t('title')}</h1>
                    <p className="text-zinc-500 text-sm">{t('viewing_desc')}</p>
                </div>
                <div className="px-3 py-1 bg-zinc-100 dark:bg-zinc-800 rounded-full text-xs text-zinc-500 font-medium">
                    Page {currentPage + 1}
                </div>
            </div>

            <div className="flex-1 min-h-0 bg-white dark:bg-zinc-900 rounded-2xl shadow-lg border border-zinc-200 dark:border-zinc-800">
                <WhiteboardCanvas
                    key={`canvas-${currentPage}`}
                    classId={session.class_id}
                    isReadOnly={true}
                    initialData={pageHistory}
                    lastAction={incoming}
                    currentPage={currentPage}
                />
            </div>
        </div>
    );
}
