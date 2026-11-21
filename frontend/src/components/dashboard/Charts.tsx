'use client';

import React from 'react';

export type Bar = { label: string; value: number; hint?: string };

/** Vertical bars 0..100 (%). Text values are always shown so colour is never the only cue. */
export function BarChart({ bars, ariaLabel }: { bars: Bar[]; ariaLabel: string }) {
    return (
        <div role="img" aria-label={ariaLabel} className="w-full">
            <div className="flex items-end gap-3 sm:gap-4 h-44">
                {bars.map((b, i) => {
                    const v = Math.max(0, Math.min(100, Math.round(b.value)));
                    return (
                        <div key={i} className="flex-1 min-w-0 flex flex-col items-center gap-1.5 h-full justify-end" title={b.hint || b.label}>
                            <span className="text-xs font-bold tabular-nums">{v}%</span>
                            <div className="w-full max-w-[34px] h-full bg-primary/10 rounded-xl overflow-hidden flex items-end">
                                <div
                                    className="w-full rounded-xl bg-primary transition-[height] duration-700"
                                    style={{ height: `${Math.max(v, 3)}%` }}
                                />
                            </div>
                        </div>
                    );
                })}
            </div>
            <div className="flex gap-3 sm:gap-4 mt-2">
                {bars.map((b, i) => (
                    <span key={i} className="flex-1 min-w-0 text-[11px] leading-tight text-center text-muted line-clamp-2 break-words">
                        {b.label}
                    </span>
                ))}
            </div>
        </div>
    );
}

/** Ring with a percentage in the middle */
export function Donut({ value, label, size = 84 }: { value: number; label: string; size?: number }) {
    const v = Math.max(0, Math.min(100, Math.round(value)));
    const stroke = 9;
    const r = (size - stroke) / 2;
    const c = 2 * Math.PI * r;
    return (
        <figure className="flex flex-col items-center gap-2 text-center">
            <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`} role="img" aria-label={`${label}: ${v}%`}>
                <circle cx={size / 2} cy={size / 2} r={r} fill="none" stroke="currentColor" strokeWidth={stroke}
                    className="text-primary/15" />
                <circle cx={size / 2} cy={size / 2} r={r} fill="none" stroke="currentColor" strokeWidth={stroke}
                    strokeLinecap="round" strokeDasharray={c} strokeDashoffset={c - (v / 100) * c}
                    transform={`rotate(-90 ${size / 2} ${size / 2})`} className="text-primary transition-all duration-700" />
                <text x="50%" y="50%" dominantBaseline="central" textAnchor="middle"
                    className="fill-current text-[15px] font-bold">{v}%</text>
            </svg>
            <figcaption className="text-xs font-medium text-muted max-w-[96px] line-clamp-2">{label}</figcaption>
        </figure>
    );
}

export function StatCard({ icon, label, value, hint }: { icon: React.ReactNode; label: string; value: React.ReactNode; hint?: string }) {
    return (
        <div className="glass-card rounded-3xl p-5 flex items-center gap-4">
            <span className="w-12 h-12 rounded-2xl bg-primary/10 text-primary flex items-center justify-center shrink-0">{icon}</span>
            <span className="min-w-0">
                <span className="block text-2xl font-bold font-outfit tabular-nums leading-none">{value}</span>
                <span className="block text-sm text-muted mt-1 leading-snug">{label}</span>
                {hint && <span className="block text-xs text-muted/80 truncate">{hint}</span>}
            </span>
        </div>
    );
}

export function EmptyState({ icon, text, action }: { icon?: React.ReactNode; text: string; action?: React.ReactNode }) {
    return (
        <div className="flex flex-col items-center justify-center text-center gap-3 py-8 px-4">
            {icon && <span className="w-14 h-14 rounded-2xl bg-primary/10 text-primary flex items-center justify-center">{icon}</span>}
            <p className="text-sm text-muted max-w-xs">{text}</p>
            {action}
        </div>
    );
}
