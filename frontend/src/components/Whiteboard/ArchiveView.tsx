"use client";

import React, { useState, useEffect } from "react";
import api from "@/lib/api";
import { Loader2, FileX, Calendar, Clock, ArrowLeft, ChevronLeft, ChevronRight } from "lucide-react";
import { WhiteboardAction, WhiteboardSession } from "@/types/whiteboard";
import { WhiteboardCanvas } from "@/components/Whiteboard/WhiteboardCanvas";
import { useRouter } from "next/navigation";

interface ArchiveViewProps {
    sessionId: string;
}

export default function ArchiveView({ sessionId }: ArchiveViewProps) {
    const router = useRouter();
    const [session, setSession] = useState<WhiteboardSession | null>(null);
    const [loading, setLoading] = useState(true);
    const [currentPage, setCurrentPage] = useState(0);
    const [pages, setPages] = useState<WhiteboardAction[][]>([[]]);

    useEffect(() => {
        const fetchSession = async () => {
            try {
                const res = await api.get(`/whiteboard/session/${sessionId}`);
                if (res.data && res.data.data) {
                    const s: WhiteboardSession = res.data.data;
                    setSession(s);

                    // Parse snapshot: support both new {pages: [...]} and old flat array
                    const snap = s.snapshot as any;
                    if (snap && snap.version === 2 && Array.isArray(snap.pages)) {
                        setPages(snap.pages);
                    } else if (Array.isArray(snap)) {
                        // Legacy: single page
                        setPages([snap]);
                    } else {
                        setPages([[]]);
                    }
                }
            } catch (err) {
                console.error("Failed to fetch session", err);
            } finally {
                setLoading(false);
            }
        };
        fetchSession();
    }, [sessionId]);

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[60vh]">
                <Loader2 className="w-8 h-8 animate-spin text-zinc-500" />
            </div>
        );
    }

    if (!session) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[60vh] text-center p-4">
                <FileX className="w-12 h-12 text-zinc-400 mb-4" />
                <h2 className="text-xl font-bold dark:text-white">Session Not Found</h2>
                <p className="text-zinc-500">We couldn&apos;t find the board you&apos;re looking for.</p>
            </div>
        );
    }

    const currentPageActions = pages[currentPage] ?? [];

    return (
        <div className="flex flex-col h-[calc(100vh-120px)] max-w-6xl mx-auto p-4 gap-4">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                    <button
                        onClick={() => router.back()}
                        className="p-2 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded-full transition-colors"
                    >
                        <ArrowLeft className="w-5 h-5 dark:text-white" />
                    </button>
                    <div>
                        <h1 className="text-2xl font-bold dark:text-white">
                            {session.title || "Untitled Lesson"}
                        </h1>
                        <div className="flex items-center gap-4 text-zinc-500 text-sm">
                            <div className="flex items-center gap-1">
                                <Calendar className="w-4 h-4" />
                                {new Date(session.created_at).toLocaleDateString()}
                            </div>
                            <div className="flex items-center gap-1">
                                <Clock className="w-4 h-4" />
                                {new Date(session.created_at).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
                            </div>
                        </div>
                    </div>
                </div>

                {/* Page navigator */}
                {pages.length > 1 && (
                    <div className="flex items-center gap-2">
                        <button
                            disabled={currentPage === 0}
                            onClick={() => setCurrentPage(p => p - 1)}
                            className="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 disabled:opacity-30 transition"
                        >
                            <ChevronLeft className="w-4 h-4" />
                        </button>
                        <span className="text-sm font-medium dark:text-white">
                            Page {currentPage + 1} / {pages.length}
                        </span>
                        <button
                            disabled={currentPage === pages.length - 1}
                            onClick={() => setCurrentPage(p => p + 1)}
                            className="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 disabled:opacity-30 transition"
                        >
                            <ChevronRight className="w-4 h-4" />
                        </button>
                    </div>
                )}
            </div>

            {/* Canvas — re-mounts per page so initialData rerenders correctly */}
            <div className="flex-1 min-h-0 bg-white dark:bg-zinc-900 rounded-2xl shadow-lg border border-zinc-200 dark:border-zinc-800">
                <WhiteboardCanvas
                    key={`archive-${sessionId}-page-${currentPage}`}
                    classId={session.class_id}
                    isReadOnly={true}
                    initialData={currentPageActions}
                    currentPage={currentPage}
                    lastAction={null}
                />
            </div>
        </div>
    );
}
