// @ts-nocheck
/* eslint-disable */
import { Document, Packer, Paragraph, Table, TableCell, TableRow, WidthType, BorderStyle, AlignmentType, TextRun, HeadingLevel } from 'docx';
import { saveAs } from 'file-saver';
import { kmjStageTimes } from './kmj';

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

const BORDER_CONFIG = {
    top: { style: BorderStyle.SINGLE, size: 1, color: '000000' },
    bottom: { style: BorderStyle.SINGLE, size: 1, color: '000000' },
    left: { style: BorderStyle.SINGLE, size: 1, color: '000000' },
    right: { style: BorderStyle.SINGLE, size: 1, color: '000000' },
};

// Helper to create bold paragraph
const boldPara = (text: string) => new Paragraph({
    children: [new TextRun({ text, bold: true })],
});

// Helper to create regular paragraph
const regularPara = (text: string) => new Paragraph({ children: [new TextRun({ text })] });

// Helper to create centered bold paragraph
const centeredBoldPara = (text: string) => new Paragraph({
    children: [new TextRun({ text, bold: true })],
    alignment: AlignmentType.CENTER,
});

export async function exportKMJToWord(data: KMJPlanWithCells, locale: string = 'kk'): Promise<void> {
    const { plan, cells } = data;

    const translations = {
        kk: {
            title: 'ҚЫСҚА МЕРЗІМДІ ЖОСПАР (ҚМЖ)',
            section: 'Бөлім',
            teacher: 'Педагогтың аты-жөні',
            date: 'Күні',
            grade: 'Сынып',
            participants: 'Қатысушылар саны',
            absent: 'Қатыспағандар саны',
            topic: 'Сабақтың тақырыбы:',
            objectives: 'Оқу бағдарламасына сәйкес оқыту мақсаттары',
            purpose: 'Сабақтың мақсаты',
            unity: 'Бірлік және ынтымақ',
            progress: 'САБАҚТЫҢ БАРЫСЫ',
            stage_time: 'Сабақтың кезеңі / уақыт',
            teacher_action: 'Педагогтың әрекеті',
            student_action: 'Оқушының әрекеті',
            assessment_label: 'Бағалау',
            resources: 'Ресурстар',
            start: 'Сабақтың басы (10 мин)',
            middle: 'Сабақтың ортасы (30 мин)',
            end: 'Сабақтың соңы (5 мин)',
            homework: 'Үй тапсырмасы',
            filename_suffix: 'сынып'
        },
        ru: {
            title: 'КРАТКОСРОЧНЫЙ ПЛАН (КСП)',
            section: 'Раздел',
            teacher: 'ФИО педагога',
            date: 'Дата',
            grade: 'Класс',
            participants: 'Количество участников',
            absent: 'Количество отсутствующих',
            topic: 'Тема урока:',
            objectives: 'Цели обучения в соответствии с учебной программой',
            purpose: 'Цель урока',
            unity: 'Единство и сплоченность',
            progress: 'ХОД УРОКА',
            stage_time: 'Этап урока / время',
            teacher_action: 'Действия педагога',
            student_action: 'Действия ученика',
            assessment_label: 'Оценивание',
            resources: 'Ресурсы',
            start: 'Начало урока (10 мин)',
            middle: 'Середина урока (30 мин)',
            end: 'Конец урока (5 мин)',
            homework: 'Домашнее задание',
            filename_suffix: 'класс'
        },
        en: {
            title: 'SHORT-TERM PLAN (STP)',
            section: 'Section',
            teacher: 'Teacher Name',
            date: 'Date',
            grade: 'Grade',
            participants: 'Number of participants',
            absent: 'Number of absent',
            topic: 'Lesson Topic:',
            objectives: 'Learning objectives in accordance with the curriculum',
            purpose: 'Lesson objective',
            unity: 'Unity and solidarity',
            progress: 'LESSON PROGRESS',
            stage_time: 'Lesson stage / time',
            teacher_action: 'Teacher activity',
            student_action: 'Student activity',
            assessment_label: 'Assessment',
            resources: 'Resources',
            start: 'Start of the lesson (10 min)',
            middle: 'Middle of the lesson (30 min)',
            end: 'End of the lesson (5 min)',
            homework: 'Homework',
            filename_suffix: 'grade'
        }
    };

    const baseT = translations[locale as keyof typeof translations] || translations.kk;
    // Replace the fixed "(10 мин)" etc. with the real stage times of this plan
    const times = kmjStageTimes(data.cells || [], (data.plan as any)?.duration_minutes, locale === 'en' ? 'min' : 'мин');
    const strip = (x: string) => x.replace(/\s*\([^)]*\)\s*$/, '');
    const t = {
        ...baseT,
        start: `${strip(baseT.start)} (${times.lesson_start.text})`,
        middle: `${strip(baseT.middle)} (${times.lesson_middle.text})`,
        end: `${strip(baseT.end)} (${times.lesson_end.text})`,
    };

    // Helper function to get localized national value name
    const getNationalValueName = () => {
        if (!data.national_value) return t.unity;
        const langSuffix = locale === 'kk' ? 'kaz' : locale === 'ru' ? 'rus' : 'eng';
        return data.national_value[`name_${langSuffix}`] || data.national_value.name_kaz || t.unity;
    };

    // Helper function to get cell value
    const getCellValue = (section: string, key: string, stageColumn?: string | null): string => {
        const cell = cells.find(
            c => c.section === section &&
                (stageColumn ? c.stage_column === stageColumn : c.cell_key === key)
        );
        return cell?.cell_value || '';
    };

    // Createdocument with sections
    const doc = new Document({
        sections: [{
            properties: {
                page: {
                    margin: {
                        top: 720,    // 0.5 inch
                        right: 720,
                        bottom: 720,
                        left: 720,
                    },
                },
            },
            children: [
                // Title
                new Paragraph({
                    text: t.title,
                    heading: HeadingLevel.HEADING_1,
                    alignment: AlignmentType.CENTER,
                    spacing: { after: 200 },
                }),

                // Header Table
                new Table({
                    width: { size: 100, type: WidthType.PERCENTAGE },
                    rows: [
                        new TableRow({
                            children: [
                                new TableCell({
                                    children: [boldPara(t.section)],
                                    borders: BORDER_CONFIG,
                                    width: { size: 33, type: WidthType.PERCENTAGE },
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('header', 'bolem'))],
                                    borders: BORDER_CONFIG,
                                    width: { size: 67, type: WidthType.PERCENTAGE },
                                }),
                            ],
                        }),
                        new TableRow({
                            children: [
                                new TableCell({
                                    children: [boldPara(t.teacher)],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('header', 'pedagog_name'))],
                                    borders: BORDER_CONFIG,
                                }),
                            ],
                        }),
                        new TableRow({
                            children: [
                                new TableCell({
                                    children: [boldPara(t.date)],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('header', 'date'))],
                                    borders: BORDER_CONFIG,
                                }),
                            ],
                        }),
                        new TableRow({
                            children: [
                                new TableCell({
                                    children: [boldPara(`${t.grade} ${plan.grade}`)],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [
                                        new Table({
                                            width: { size: 100, type: WidthType.PERCENTAGE },
                                            rows: [
                                                new TableRow({
                                                    children: [
                                                        new TableCell({
                                                            children: [boldPara(t.participants)],
                                                            borders: BORDER_CONFIG,
                                                            width: { size: 25, type: WidthType.PERCENTAGE },
                                                        }),
                                                        new TableCell({
                                                            children: [regularPara(getCellValue('header', 'participants_count'))],
                                                            borders: BORDER_CONFIG,
                                                            width: { size: 25, type: WidthType.PERCENTAGE },
                                                        }),
                                                        new TableCell({
                                                            children: [boldPara(t.absent)],
                                                            borders: BORDER_CONFIG,
                                                            width: { size: 25, type: WidthType.PERCENTAGE },
                                                        }),
                                                        new TableCell({
                                                            children: [regularPara(getCellValue('header', 'absent_count'))],
                                                            borders: BORDER_CONFIG,
                                                            width: { size: 25, type: WidthType.PERCENTAGE },
                                                        }),
                                                    ],
                                                }),
                                            ],
                                        }),
                                    ],
                                    borders: BORDER_CONFIG,
                                }),
                            ],
                        }),
                        new TableRow({
                            children: [
                                new TableCell({
                                    children: [boldPara(t.topic)],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('header', 'topic'))],
                                    borders: BORDER_CONFIG,
                                }),
                            ],
                        }),
                        new TableRow({
                            children: [
                                new TableCell({
                                    children: [boldPara(t.objectives)],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('header', 'learning_objectives'))],
                                    borders: BORDER_CONFIG,
                                }),
                            ],
                        }),
                        new TableRow({
                            children: [
                                new TableCell({
                                    children: [boldPara(t.purpose)],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('header', 'lesson_objectives'))],
                                    borders: BORDER_CONFIG,
                                }),
                            ],
                        }),
                        new TableRow({
                            children: [
                                new TableCell({
                                    children: [boldPara(getNationalValueName())],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('header', 'unity_cooperation'))],
                                    borders: BORDER_CONFIG,
                                }),
                            ],
                        }),
                    ],
                }),

                // Spacer
                new Paragraph({ text: '' }),

                // Section title
                centeredBoldPara(t.progress),

                new Paragraph({ text: '' }),

                // Main process table
                new Table({
                    width: { size: 100, type: WidthType.PERCENTAGE },
                    rows: [
                        // Header row
                        new TableRow({
                            children: [
                                new TableCell({
                                    children: [centeredBoldPara(t.stage_time)],
                                    borders: BORDER_CONFIG,
                                    width: { size: 15, type: WidthType.PERCENTAGE },
                                }),
                                new TableCell({
                                    children: [centeredBoldPara(t.teacher_action)],
                                    borders: BORDER_CONFIG,
                                    width: { size: 25, type: WidthType.PERCENTAGE },
                                }),
                                new TableCell({
                                    children: [centeredBoldPara(t.student_action)],
                                    borders: BORDER_CONFIG,
                                    width: { size: 25, type: WidthType.PERCENTAGE },
                                }),
                                new TableCell({
                                    children: [centeredBoldPara(t.assessment_label)],
                                    borders: BORDER_CONFIG,
                                    width: { size: 15, type: WidthType.PERCENTAGE },
                                }),
                                new TableCell({
                                    children: [centeredBoldPara(t.resources)],
                                    borders: BORDER_CONFIG,
                                    width: { size: 20, type: WidthType.PERCENTAGE },
                                }),
                            ],
                        }),
                        // Lesson start
                        new TableRow({
                            children: [
                                new TableCell({
                                    children: [boldPara(t.start)],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('lesson_start', '', 'pedagog_action'))],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('lesson_start', '', 'student_action'))],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('lesson_start', '', 'assessment'))],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('lesson_start', '', 'resources'))],
                                    borders: BORDER_CONFIG,
                                }),
                            ],
                        }),
                        // Lesson middle
                        new TableRow({
                            children: [
                                new TableCell({
                                    children: [boldPara(t.middle)],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('lesson_middle', '', 'pedagog_action'))],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('lesson_middle', '', 'student_action'))],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('lesson_middle', '', 'assessment'))],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('lesson_middle', '', 'resources'))],
                                    borders: BORDER_CONFIG,
                                }),
                            ],
                        }),
                        // Lesson end
                        new TableRow({
                            children: [
                                new TableCell({
                                    children: [boldPara(t.end)],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('lesson_end', '', 'pedagog_action'))],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('lesson_end', '', 'student_action'))],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('lesson_end', '', 'assessment'))],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('lesson_end', '', 'resources'))],
                                    borders: BORDER_CONFIG,
                                }),
                            ],
                        }),
                        // Assessment row
                        new TableRow({
                            children: [
                                new TableCell({
                                    children: [boldPara(t.assessment_label)],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('footer', '', 'assessment_total'))],
                                    borders: BORDER_CONFIG,
                                    columnSpan: 4,
                                }),
                            ],
                        }),
                        // Homework row
                        new TableRow({
                            children: [
                                new TableCell({
                                    children: [boldPara(t.homework)],
                                    borders: BORDER_CONFIG,
                                }),
                                new TableCell({
                                    children: [regularPara(getCellValue('footer', '', 'homework'))],
                                    borders: BORDER_CONFIG,
                                    columnSpan: 4,
                                }),
                            ],
                        }),
                    ],
                }),
            ],
        }],
    });

    // Generate and save the document
    const blob = await Packer.toBlob(doc);
    const fileName = `KMJ_${plan.subject}_${plan.grade}${t.filename_suffix}_${plan.lesson_topic}.docx`;
    saveAs(blob, fileName);
}
