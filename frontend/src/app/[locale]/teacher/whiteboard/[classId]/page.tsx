"use client";

import React, { useState, useEffect, useRef, use, useCallback } from "react";
import { WhiteboardCanvas, WhiteboardCanvasHandle } from "@/components/Whiteboard/WhiteboardCanvas";
import api from "@/lib/api";
import { motion, AnimatePresence } from "framer-motion";
import { Play, Square, Loader2, Users, Plus, ChevronLeft, ChevronRight } from "lucide-react";
import { WhiteboardAction, WhiteboardSession } from "@/types/whiteboard";
import { useTranslations } from "next-intl";

interface TeacherWhiteboardPageProps {
    params: Promise<{
        classId: string;
        locale: string;
    }>;
}

export default function TeacherWhiteboardPage({ params }: TeacherWhiteboardPageProps) {
    const t = useTranslations('whiteboard_teacher');
    const resolvedParams = use(params);
    const classId = parseInt(resolvedParams.classId);
    const [activeSession, setActiveSession] = useState<WhiteboardSession | null>(null);
    const [loading, setLoading] = useState(true);
    const [starting, setStarting] = useState(false);
    const ws = useRef<WebSocket | null>(null);
    const canvasRef = useRef<WhiteboardCanvasHandle>(null);

    // Multi-page state: each page stores its actions
    const [pages, setPages] = useState<WhiteboardAction[][]>([[]]);
    const [currentPage, setCurrentPage] = useState(0);

    const [sessionTitle, setSessionTitle] = useState("");

    useEffect(() => {
        const checkActiveSession = async () => {
            try {
                const res = await api.get("/whiteboard/active");
                if (res.data && res.data.data) {
                    setActiveSession(res.data.data);
                    initWebSocket(res.data.data.class_id);
                }
            } catch {
                // No active session
            } finally {
                setLoading(false);
            }
        };

        checkActiveSession();
        return () => { ws.current?.close(); };
    }, []);

    const initWebSocket = (cid: number) => {
        const wsUrl = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8000";
        const socket = new WebSocket(`${wsUrl}/api/v1/ws/whiteboard/${cid}`);

        socket.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                if (data.type === "request_sync") {
                    // Student joined or requested sync, send current page history
                    socket.send(JSON.stringify({
                        type: "page_sync",
                        page_index: currentPage,
                        actions: pages[currentPage] || []
                    }));
                }
            } catch (err) {
                console.error("WS teacher error:", err);
            }
        };

        ws.current = socket;
    };

    const startSession = async () => {
        setStarting(true);
        try {
            const res = await api.post("/whiteboard/start", { class_id: classId, title: sessionTitle });
            if (res.data && res.data.data) {
                setActiveSession(res.data.data);
                initWebSocket(classId);
            }
        } catch {
            alert(t('start_failed'));
        } finally {
            setStarting(false);
        }
    };

    const endSession = async () => {
        if (!activeSession) return;
        if (!confirm(t('end_confirm'))) return;
        try {
            await api.post("/whiteboard/end", {
                session_id: activeSession.id,
                snapshot: { pages: pages, version: 2 }
            });
            setActiveSession(null);
            setPages([[]]);
            setCurrentPage(0);
            ws.current?.close();
            alert(t('end_success'));
        } catch {
            alert(t('end_failed'));
        }
    };

    const handleDraw = useCallback((data: WhiteboardAction) => {
        if (ws.current?.readyState === WebSocket.OPEN) {
            ws.current.send(JSON.stringify({ ...data, page_index: currentPage }));
        }
        setPages((prev: WhiteboardAction[][]) => {
            const updated = [...prev];
            const currentPageActions = updated[currentPage] || [];
            updated[currentPage] = [...currentPageActions, { ...data, page_index: currentPage }];
            return updated;
        });
    }, [currentPage]);

    const switchPage = useCallback((targetPage: number) => {
        if (targetPage === currentPage) return;
        setCurrentPage(targetPage);
        // Broadcast full page history to sync students immediately
        if (ws.current?.readyState === WebSocket.OPEN) {
            ws.current.send(JSON.stringify({
                type: "page_sync",
                page_index: targetPage,
                actions: pages[targetPage] || []
            }));
        }
    }, [currentPage, pages]);

    const addPage = useCallback(() => {
        const newIdx = pages.length;
        setPages((prev: WhiteboardAction[][]) => [...prev, []]);
        setCurrentPage(newIdx);
        // Broadcast new empty page
        if (ws.current?.readyState === WebSocket.OPEN) {
            ws.current.send(JSON.stringify({
                type: "page_sync",
                page_index: newIdx,
                actions: []
            }));
        }
    }, [pages.length]);

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[60vh]">
                <Loader2 className="w-8 h-8 animate-spin text-zinc-500" />
            </div>
        );
    }

    return (
        <div className="flex flex-col h-[calc(100dvh-120px)] p-4 gap-4">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold dark:text-white">
                        {activeSession?.title || t('title')}
                    </h1>
                    <p className="text-zinc-500 text-sm">{t('subtitle')}</p>
                </div>

                <div className="flex items-center gap-3">
                    {activeSession ? (
                        <>
                            <div className="flex items-center gap-2 px-3 py-1 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300 rounded-full text-sm font-medium animate-pulse">
                                <Users className="w-4 h-4" /> {t('live_session')}
                            </div>
                            <button
                                onClick={endSession}
                                className="flex items-center gap-2 px-4 py-2 bg-red-500 hover:bg-red-600 text-white rounded-lg transition shadow-sm"
                            >
                                <Square className="w-4 h-4" /> {t('end_lesson')}
                            </button>
                        </>
                    ) : (
                        <button
                            onClick={startSession}
                            disabled={starting}
                            className="flex items-center gap-2 px-6 py-2 bg-zinc-900 dark:bg-white text-white dark:text-zinc-900 rounded-lg hover:opacity-90 transition shadow-md disabled:opacity-50"
                        >
                            {starting ? <Loader2 className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
                            {t('start_lesson')}
                        </button>
                    )}
                </div>
            </div>

            <div className="flex flex-1 min-h-0 gap-3">
                {/* Canvas area */}
                <div className="flex-1 min-w-0 relative">
                    <AnimatePresence mode="wait">
                        {activeSession ? (
                            <motion.div
                                key={`canvas-${currentPage}`}
                                initial={{ opacity: 0, scale: 0.98 }}
                                animate={{ opacity: 1, scale: 1 }}
                                exit={{ opacity: 0 }}
                                className="w-full h-full"
                            >
                                <WhiteboardCanvas
                                    ref={canvasRef}
                                    classId={classId}
                                    onDraw={handleDraw}
                                    currentPage={currentPage}
                                    initialData={pages[currentPage]}
                                />
                            </motion.div>
                        ) : (
                            <motion.div
                                key="placeholder"
                                initial={{ opacity: 0 }}
                                animate={{ opacity: 1 }}
                                className="w-full h-full flex flex-col items-center justify-center border-2 border-dashed border-zinc-200 dark:border-zinc-800 rounded-2xl bg-zinc-50 dark:bg-zinc-900/50"
                            >
                                <div className="text-center space-y-4">
                                    <div className="p-4 bg-zinc-100 dark:bg-zinc-800 rounded-full inline-block">
                                        <Play className="w-8 h-8 text-zinc-400" />
                                    </div>
                                    <div>
                                        <h3 className="text-lg font-semibold dark:text-white">{t('ready_title')}</h3>
                                        <p className="text-zinc-500 mb-6">{t('ready_desc')}</p>
                                        <div className="max-w-xs mx-auto mb-6">
                                            <input
                                                type="text"
                                                value={sessionTitle}
                                                onChange={(e) => setSessionTitle(e.target.value)}
                                                placeholder={t('topic_placeholder')}
                                                className="w-full px-4 py-3 rounded-xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 outline-none focus:ring-2 focus:ring-primary/20 transition-all text-center"
                                            />
                                        </div>
                                    </div>
                                </div>
                            </motion.div>
                        )
                        }
                    </AnimatePresence >
                </div >

                {/* Pages sidebar — only when session is active */}
                {
                    activeSession && (
                        <div className="flex flex-col gap-2 w-20">
                            <p className="text-xs text-zinc-500 text-center font-medium">{t('pages')}</p>
                            <div className="flex flex-col gap-2 overflow-y-auto flex-1">
                                {pages.map((_, idx) => (
                                    <button
                                        key={idx}
                                        onClick={() => switchPage(idx)}
                                        className={`w-full aspect-[3/4] rounded-lg border-2 transition-all flex items-center justify-center text-sm font-bold ${idx === currentPage
                                            ? "border-primary bg-primary/10 text-primary"
                                            : "border-zinc-200 dark:border-zinc-700 text-zinc-500 hover:border-zinc-400"
                                            }`}
                                    >
                                        {idx + 1}
                                    </button>
                                ))}
                                <button
                                    onClick={addPage}
                                    className="w-full aspect-[3/4] rounded-lg border-2 border-dashed border-zinc-300 dark:border-zinc-600 hover:border-primary hover:text-primary text-zinc-400 transition-all flex items-center justify-center"
                                >
                                    <Plus className="w-5 h-5" />
                                </button>
                            </div>

                            {/* Prev / Next */}
                            <div className="flex gap-1 mt-1">
                                <button
                                    disabled={currentPage === 0}
                                    onClick={() => switchPage(currentPage - 1)}
                                    className="flex-1 p-1 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 disabled:opacity-30 transition"
                                >
                                    <ChevronLeft className="w-4 h-4 mx-auto text-zinc-500" />
                                </button>
                                <button
                                    disabled={currentPage === pages.length - 1}
                                    onClick={() => switchPage(currentPage + 1)}
                                    className="flex-1 p-1 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 disabled:opacity-30 transition"
                                >
                                    <ChevronRight className="w-4 h-4 mx-auto text-zinc-500" />
                                </button>
                            </div>
                        </div>
                    )
                }
            </div >
        </div>
    );
}
