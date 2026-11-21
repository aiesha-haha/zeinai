export interface WhiteboardAction {
    type: "start" | "move" | "end" | "clear" | "page_change";
    x?: number;
    y?: number;
    color?: string;
    lineWidth?: number;
    isEraser?: boolean;
    page_index?: number;
}

export interface WhiteboardSession {
    id: string;
    teacher_id: string;
    class_id: number;
    status: "active" | "completed";
    snapshot?: any;
    title?: string;
    created_at: string;
    ended_at?: string;
}

export interface WhiteboardSummary {
    id: string;
    session_id: string;
    session_title: string;
    summary_text: string; // RU
    summary_kk: string;
    summary_en: string;
    explanation: string;  // RU
    explanation_kk: string;
    explanation_en: string;
    created_at: string;
}
