"use client";
import React, { use } from 'react';
import SummaryView from "@/components/Whiteboard/SummaryView";

interface TeacherSummaryPageProps {
    params: Promise<{
        sessionId: string;
        locale: string;
    }>;
}

export default function TeacherWhiteboardSummaryPage({ params }: TeacherSummaryPageProps) {
    const resolvedParams = use(params);

    return (
        <div className="py-6">
            <SummaryView sessionId={resolvedParams.sessionId} />
        </div>
    );
}
