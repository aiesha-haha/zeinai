'use client';

import { useState, useEffect } from 'react';
import { kmjStageTimes } from '@/lib/kmj';
import { useRouter } from '@/i18n/routing';
import { useParams } from 'next/navigation';
import { useTranslations, useLocale } from 'next-intl';
import { ChevronLeft, Loader2, CheckCircle, XCircle, Plus, Download } from 'lucide-react';
import api from '@/lib/api';
import { exportKMJToWord } from '@/lib/exportToWord';

interface KMJCell {
    id: string;
    section: string;
    cell_key: string;
    cell_value: string;
    stage_column: string | null;
    is_ai_generated: boolean;
    is_edited: boolean;
}

interface KMJPlan {
    id: string;
    subject: string;
    grade: number;
    lesson_number: number;
    lesson_topic: string;
    teacher_name: string;
    lesson_date: string | null;
    content_lang: string;
    status: string;
    national_value_id: string | null;
    created_at: string;
    learning_objective?: string;
    duration_minutes?: number;
}

interface NationalValue {
    id: string;
    month_number: number;
    name_kaz: string;
    name_rus: string;
    name_eng: string;
}

interface KMJPlanWithCells {
    plan: KMJPlan;
    cells: KMJCell[];
    national_value?: NationalValue;
}

export default function KMJViewPage() {
    const params = useParams();
    console.log('KMJ params:', params); // Debug log
    const router = useRouter();
    const t = useTranslations('education_materials');
    const tCommon = useTranslations('common');
    const tKmj = useTranslations('kmj');
    const locale = useLocale();

    const getLocalizedField = (obj: any, field: string) => {
        if (!obj) return null;
        const langSuffix = locale === 'kk' ? 'kaz' : locale === 'ru' ? 'rus' : 'eng';
        return obj[`${field}_${langSuffix}`] || obj[`${field}_kaz`];
    };

    const [data, setData] = useState<KMJPlanWithCells | null>(null);
    const [loading, setLoading] = useState(true);
    const [editingCell, setEditingCell] = useState<string | null>(null);
    const [editValue, setEditValue] = useState('');

    useEffect(() => {
        const fetchPlan = async () => {
            try {
                const response = await api.get(`/kmj/${params.id}`);
                setData(response.data.data);
            } catch (error) {
                console.error('Failed to fetch plan:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchPlan();

        // Poll if generating
        const interval = setInterval(() => {
            if (data?.plan.status === 'generating') {
                fetchPlan();
            }
        }, 3000);

        return () => clearInterval(interval);
    }, [params.id, data?.plan.status]);

    const handleEditCell = async (cellId: string, value: string) => {
        try {
            await api.put(`/kmj/cells/${cellId}`, { cell_value: value });

            if (data) {
                const updatedCells = data.cells.map(cell =>
                    cell.id === cellId ? { ...cell, cell_value: value, is_edited: true } : cell
                );
                setData({ ...data, cells: updatedCells });
            }

            setEditingCell(null);
        } catch (error) {
            console.error('Failed to update cell:', error);
            alert(tKmj('update_failed'));
        }
    };

    const handleExportToWord = async () => {
        if (!data) return;

        try {
            await exportKMJToWord(data, locale);
        } catch (error) {
            console.error('Export failed:', error);
            alert(tKmj('export_failed'));
        }
    };

    const getCellValue = (section: string, key: string, stageColumn?: string | null) => {
        const cell = data?.cells.find(
            c => c.section === section &&
                (stageColumn ? c.stage_column === stageColumn : c.cell_key === key)
        );
        return cell;
    };

    // Stage name + editable time (minutes follow the lesson duration)
    const renderStage = (section: 'lesson_start' | 'lesson_middle' | 'lesson_end', label: string) => {
        const times = kmjStageTimes(data?.cells || [], (data as any)?.plan?.duration_minutes);
        const { text, cell } = times[section];
        const isEditing = cell && editingCell === cell.id;
        return (
            <td className="border border-black p-2 font-bold align-top">
                {label}
                <div className="mt-1 font-normal">
                    {isEditing && cell ? (
                        <span className="flex gap-1">
                            <input aria-label={`${label}: ${tKmj('table.stage_time')}`} value={editValue} autoFocus
                                onChange={(e) => setEditValue(e.target.value)}
                                onKeyDown={(e) => { if (e.key === 'Enter') handleEditCell(cell.id, editValue); if (e.key === 'Escape') setEditingCell(null); }}
                                className="w-20 px-1 border border-primary rounded text-black" />
                            <button onClick={() => handleEditCell(cell.id, editValue)} className="px-2 bg-primary text-white rounded text-xs">{tKmj('save')}</button>
                        </span>
                    ) : cell ? (
                        <button onClick={() => { setEditingCell(cell.id); setEditValue(cell.cell_value); }}
                            className="underline decoration-dotted text-left" title={tKmj('table.stage_time')}>({text})</button>
                    ) : (
                        <span>({text})</span>
                    )}
                </div>
            </td>
        );
    };

    const renderCell = (section: string, key: string, stageColumn?: string | null) => {
        const cell = getCellValue(section, key, stageColumn);
        if (!cell) return <td className="border border-black p-2 bg-gray-50 italic text-gray-400">-</td>;

        const isEditing = editingCell === cell.id;

        return (
            <td className="border border-black p-0 relative align-top min-w-[150px]">
                {isEditing ? (
                    <div className="p-2 bg-blue-50/50">
                        <textarea
                            value={editValue}
                            onChange={(e) => setEditValue(e.target.value)}
                            className="w-full p-2 bg-white border border-primary rounded text-sm text-black focus:outline-none shadow-sm"
                            rows={4}
                            autoFocus
                        />
                        <div className="flex gap-2 mt-2">
                            <button
                                onClick={() => handleEditCell(cell.id, editValue)}
                                className="px-2 py-1 bg-primary text-white text-[10px] rounded font-bold uppercase"
                            >
                                {tKmj('save')}
                            </button>
                            <button
                                onClick={() => setEditingCell(null)}
                                className="px-2 py-1 bg-gray-200 text-gray-700 text-[10px] rounded font-bold uppercase"
                            >
                                {tKmj('cancel_edit')}
                            </button>
                        </div>
                    </div>
                ) : (
                    <div
                        role="button"
                        tabIndex={0}
                        aria-label={tKmj('edit_cell')}
                        onKeyDown={(e) => {
                            if (e.key === 'Enter' || e.key === ' ') {
                                e.preventDefault();
                                setEditingCell(cell.id);
                                setEditValue(cell.cell_value);
                            }
                        }}
                        onClick={() => {
                            setEditingCell(cell.id);
                            setEditValue(cell.cell_value);
                        }}
                        className="p-2 cursor-text hover:bg-gray-50 min-h-[40px] h-full transition-colors group relative"
                    >
                        <div className="text-sm whitespace-pre-wrap text-black">
                            {cell.cell_value || <span className="text-gray-300 italic">{tKmj('table.fill_placeholder')}</span>}
                        </div>
                        <div className="absolute top-1 right-1 opacity-0 group-hover:opacity-100 flex gap-1 transition-opacity">
                            {cell.is_ai_generated && (
                                <span className="text-[8px] bg-blue-50 text-blue-500 px-1 rounded border border-blue-100 uppercase font-bold">
                                    AI
                                </span>
                            )}
                            <span className="text-[10px] text-gray-400">✎</span>
                        </div>
                    </div>
                )}
            </td>
        );
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <div className="text-center space-y-4">
                    <Loader2 className="w-12 h-12 animate-spin text-primary mx-auto" />
                    <p className="text-foreground/60">{tKmj('loading')}</p>
                </div>
            </div>
        );
    }

    if (!data) {
        return (
            <div className="space-y-8">
                <div className="flex items-center gap-4">
                    <button
                        onClick={() => router.push('/teacher/materials/short-term-plan')}
                        className="p-2 bg-white/5 rounded-xl hover:bg-white/10 transition-all"
                    >
                        <ChevronLeft size={20} />
                    </button>
                    <h1 className="text-3xl font-bold">{tKmj('not_found')}</h1>
                </div>
            </div>
        );
    }

    const { plan, cells } = data;

    if (plan.status === 'generating') {
        return (
            <div className="space-y-8">
                <div className="flex items-center gap-4">
                    <button
                        onClick={() => router.push('/teacher/materials/short-term-plan')}
                        className="p-2 bg-white/5 rounded-xl hover:bg-white/10 transition-all"
                    >
                        <ChevronLeft size={20} />
                    </button>
                    <h1 className="text-3xl font-bold">ҚМЖ</h1>
                </div>

                <div className="glass-card p-12 rounded-[32px] text-center space-y-6">
                    <div className="w-20 h-20 bg-primary/20 rounded-full mx-auto flex items-center justify-center animate-pulse">
                        <span className="text-4xl">🤖</span>
                    </div>
                    <div>
                        <h2 className="text-2xl font-bold mb-3">{tKmj('ai_generating')}</h2>
                        <p className="text-foreground/60">{tKmj('generation_time')}</p>
                    </div>
                    <div className="w-full bg-white/10 rounded-full h-2 overflow-hidden">
                        <div className="bg-primary h-2 rounded-full animate-pulse" style={{ width: '60%' }}></div>
                    </div>
                </div>
            </div>
        );
    }

    if (plan.status === 'failed') {
        return (
            <div className="space-y-8">
                <div className="flex items-center gap-4">
                    <button
                        onClick={() => router.push('/teacher/materials/short-term-plan')}
                        className="p-2 bg-white/5 rounded-xl hover:bg-white/10 transition-all"
                    >
                        <ChevronLeft size={20} />
                    </button>
                    <h1 className="text-3xl font-bold">ҚМЖ</h1>
                </div>

                <div className="glass-card p-12 rounded-[32px] text-center space-y-6">
                    <XCircle className="w-20 h-20 text-red-500 mx-auto" />
                    <div>
                        <h2 className="text-2xl font-bold mb-3">{tKmj('generation_failed')}</h2>
                        <p className="text-foreground/60">{tKmj('generation_failed_desc')}</p>
                    </div>
                    <button
                        onClick={() => router.push('/teacher/materials/short-term-plan')}
                        className="px-6 py-3 bg-primary text-white rounded-xl hover:bg-primary-hover font-semibold"
                    >
                        {tKmj('create_new')}
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="space-y-8 pb-20 print:p-0 print:m-0">
            {/* Header */}
            <div className="flex items-center justify-between print:hidden">
                <div className="flex items-center gap-4">
                    <button
                        onClick={() => router.push('/teacher/materials/short-term-plan')}
                        className="p-2 bg-white/5 rounded-xl hover:bg-white/10 transition-all"
                    >
                        <ChevronLeft size={20} />
                    </button>
                    <div>
                        <div className="flex items-center gap-3 mb-2">
                            <h1 className="text-3xl font-bold font-outfit">{plan.lesson_topic}</h1>
                            {plan.status === 'completed' && (
                                <span className="px-3 py-1 bg-green-500/20 text-green-400 rounded-full text-sm font-semibold flex items-center gap-1">
                                    <CheckCircle size={16} />
                                    {tKmj('status_completed')}
                                </span>
                            )}
                        </div>
                    </div>
                </div>
                <div className="flex gap-4">
                    <button
                        onClick={handleExportToWord}
                        className="px-6 py-3 bg-primary hover:bg-primary-hover rounded-xl font-semibold transition-all flex items-center gap-2"
                    >
                        <Download size={20} />
                        {tKmj('export_word')}
                    </button>
                </div>
            </div>

            {/* KMJ Table Layout - Matching Image */}
            <div className="bg-white text-black p-8 rounded-sm shadow-xl print:shadow-none print:p-0 font-serif min-h-[1000px]">
                {/* Header Section */}
                <table className="w-full border-2 border-black border-collapse mb-4">
                    <tbody>
                        <tr>
                            <td className="border border-black p-2 font-bold w-1/3">{tKmj('table.section')}</td>
                            {renderCell('header', 'bolem')}
                        </tr>
                        <tr>
                            <td className="border border-black p-2 font-bold">{tKmj('table.pedagog_name')}</td>
                            {renderCell('header', 'pedagog_name')}
                        </tr>
                        <tr>
                            <td className="border border-black p-2 font-bold">{tKmj('table.date')}</td>
                            {renderCell('header', 'date')}
                        </tr>
                        <tr>
                            <td className="border border-black p-2 font-bold">{tKmj('table.grade')} {plan.grade}</td>
                            <td className="border border-black p-0">
                                <table className="w-full border-collapse">
                                    <tbody>
                                        <tr>
                                            <td className="border-r border-black p-2 font-bold w-1/4">{tKmj('table.participants_count')}</td>
                                            {renderCell('header', 'participants_count')}
                                            <td className="border-r border-black p-2 font-bold w-1/4">{tKmj('table.absent_count')}</td>
                                            {renderCell('header', 'absent_count')}
                                        </tr>
                                    </tbody>
                                </table>
                            </td>
                        </tr>
                        <tr>
                            <td className="border border-black p-2 font-bold">{tKmj('table.topic')}</td>
                            {renderCell('header', 'topic')}
                        </tr>
                        <tr>
                            <td className="border border-black p-2 font-bold text-sm">{tKmj('table.learning_objectives')}</td>
                            {renderCell('header', 'learning_objectives')}
                        </tr>
                        <tr>
                            <td className="border border-black p-2 font-bold text-sm">{tKmj('table.lesson_objectives')}</td>
                            {renderCell('header', 'lesson_objectives')}
                        </tr>
                        <tr>
                            <td className="border border-black p-2 font-bold text-sm">
                                {getLocalizedField(data?.national_value, 'name') || tKmj('table.default_value')}
                            </td>
                            {renderCell('header', 'unity_cooperation')}
                        </tr>
                    </tbody>
                </table>

                <div className="text-center font-bold mb-2 uppercase tracking-wide">{tKmj('table.lesson_progress')}</div>

                {/* Main Process Table */}
                <table className="w-full border-2 border-black border-collapse">
                    <thead>
                        <tr className="font-bold text-center text-sm">
                            <td className="border border-black p-2 w-[15%]">{tKmj('table.stage_time')}</td>
                            <td className="border border-black p-2 w-[25%]">{tKmj('table.pedagog_action')}</td>
                            <td className="border border-black p-2 w-[25%]">{tKmj('table.student_action')}</td>
                            <td className="border border-black p-2 w-[15%]">{tKmj('table.assessment')}</td>
                            <td className="border border-black p-2 w-[20%]">{tKmj('table.resources')}</td>
                        </tr>
                    </thead>
                    <tbody>
                        {/* Start */}
                        <tr className="text-sm align-top">
                            {renderStage('lesson_start', tKmj('table.lesson_start'))}
                            {renderCell('lesson_start', '', 'pedagog_action')}
                            {renderCell('lesson_start', '', 'student_action')}
                            {renderCell('lesson_start', '', 'assessment')}
                            {renderCell('lesson_start', '', 'resources')}
                        </tr>
                        {/* Middle */}
                        <tr className="text-sm align-top min-h-[300px]">
                            {renderStage('lesson_middle', tKmj('table.lesson_middle'))}
                            {renderCell('lesson_middle', '', 'pedagog_action')}
                            {renderCell('lesson_middle', '', 'student_action')}
                            {renderCell('lesson_middle', '', 'assessment')}
                            {renderCell('lesson_middle', '', 'resources')}
                        </tr>
                        {/* End */}
                        <tr className="text-sm align-top">
                            {renderStage('lesson_end', tKmj('table.lesson_end'))}
                            {renderCell('lesson_end', '', 'pedagog_action')}
                            {renderCell('lesson_end', '', 'student_action')}
                            {renderCell('lesson_end', '', 'assessment')}
                            {renderCell('lesson_end', '', 'resources')}
                        </tr>

                        {/* Footer Rows */}
                        <tr className="text-sm align-top">
                            <td className="border border-black p-2 font-bold">{tKmj('table.assessment')}</td>
                            <td colSpan={4} className="border border-black p-0">
                                <table className="w-full h-full">
                                    <tbody>
                                        <tr>
                                            {renderCell('footer', '', 'assessment_total')}
                                        </tr>
                                    </tbody>
                                </table>
                            </td>
                        </tr>
                        <tr className="text-sm align-top">
                            <td className="border border-black p-2 font-bold">{tKmj('table.homework')}</td>
                            <td colSpan={4} className="border border-black p-0">
                                <table className="w-full h-full">
                                    <tbody>
                                        <tr>
                                            {renderCell('footer', '', 'homework')}
                                        </tr>
                                    </tbody>
                                </table>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>

            {/* Actions (Floating Print Button for Mobile/Scroll) */}
            <div className="fixed bottom-8 right-8 print:hidden flex gap-4">
                <button
                    onClick={() => router.push('/teacher/materials/short-term-plan')}
                    className="p-4 bg-white/10 backdrop-blur-md rounded-full shadow-2xl hover:bg-white/20 transition-all border border-white/10"
                >
                    <Plus size={24} className="text-primary" />
                </button>
            </div>
        </div>
    );
}
