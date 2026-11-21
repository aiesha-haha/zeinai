"use client";

import React, { use } from "react";
import ArchiveView from "@/components/Whiteboard/ArchiveView";

interface StudentArchivePageProps {
    params: Promise<{
        sessionId: string;
        locale: string;
    }>;
}

export default function StudentArchivePage({ params }: StudentArchivePageProps) {
    const resolvedParams = use(params);

    return (
        <div className="py-6">
            <ArchiveView sessionId={resolvedParams.sessionId} />
        </div>
    );
}
