"use client";

import React, { use } from "react";
import ArchiveView from "@/components/Whiteboard/ArchiveView";

interface TeacherArchivePageProps {
    params: Promise<{
        sessionId: string;
        locale: string;
    }>;
}

export default function TeacherArchivePage({ params }: TeacherArchivePageProps) {
    const resolvedParams = use(params);

    return (
        <div className="py-6">
            <ArchiveView sessionId={resolvedParams.sessionId} />
        </div>
    );
}
