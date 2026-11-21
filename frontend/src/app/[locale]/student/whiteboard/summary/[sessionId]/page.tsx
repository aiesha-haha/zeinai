"use client";
import React, { use } from 'react';
import SummaryView from "@/components/Whiteboard/SummaryView";

interface SummaryPageProps {
    params: Promise<{
        sessionId: string;
        locale: string;
    }>;
}

export default function WhiteboardSummaryPage({ params }: SummaryPageProps) {
    const resolvedParams = use(params);

    return (
        <div className="py-6">
            <SummaryView sessionId={resolvedParams.sessionId} />
        </div>
    );
}
