// Stage timing helpers for the short-term plan (ҚМЖ).
// Must match domain.KMJStageMinutes on the backend.

export function kmjStageMinutes(duration?: number | null): [number, number, number] {
    let d = duration && duration > 0 ? duration : 45;
    let start = Math.floor((d * 2 + 4) / 9);
    let end = Math.floor((d + 4) / 9);
    if (start < 3) start = 3;
    if (end < 3) end = 3;
    let middle = d - start - end;
    if (middle < 1) {
        middle = 1;
        start = Math.floor((d - middle) / 2);
        end = d - middle - start;
    }
    return [start, middle, end];
}

type Cell = { section: string; stage_column: string | null; cell_value: string; id: string };

export const KMJ_STAGES = ['lesson_start', 'lesson_middle', 'lesson_end'] as const;

/** Time text for each stage: the editable "time" cell when present, otherwise computed from the duration */
export function kmjStageTimes(cells: Cell[], duration?: number | null, unit = 'мин'): Record<string, { text: string; cell?: Cell }> {
    const mins = kmjStageMinutes(duration);
    const out: Record<string, { text: string; cell?: Cell }> = {};
    KMJ_STAGES.forEach((s, i) => {
        const cell = cells.find((c) => c.section === s && c.stage_column === 'time');
        out[s] = { text: cell?.cell_value || `${mins[i]} ${unit}`, cell };
    });
    return out;
}
