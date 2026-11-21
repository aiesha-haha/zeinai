'use client';

import React, { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { useTranslations } from 'next-intl';
import {
    ChevronLeft, Loader2, Calendar,
    BookOpen, FileText, Download, FileDown
} from 'lucide-react';
import api from '@/lib/api';
import { motion } from 'framer-motion';
import { jsPDF } from 'jspdf';
import { robotoBase64 } from '@/lib/pdfFonts';
import { Document, Packer, Paragraph, TextRun, HeadingLevel, AlignmentType } from 'docx';
import { saveAs } from 'file-saver';

interface Article {
    id: string;
    subject: string;
    grade: number;
    topic: string;
    language: string;
    style: string;
    complexity: string;
    title: string;
    content: string;
    created_at: string;
}

export default function ArticleDetail() {
    const params = useParams();
    const router = useRouter();
    const t = useTranslations('articles');
    const [loading, setLoading] = useState(true);
    const [article, setArticle] = useState<Article | null>(null);

    useEffect(() => {
        if (params.id) {
            fetchArticle();
        }
    }, [params.id]);

    const fetchArticle = async () => {
        setLoading(true);
        try {
            const res = await api.get(`/articles/${params.id}`);
            setArticle(res.data.data);
        } catch (err) {
            console.error('Failed to fetch article:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleExportPDF = async () => {
        if (!article) return;

        const doc = new jsPDF();

        // Add Cyrillic font
        doc.addFileToVFS('Roboto-Regular.ttf', robotoBase64);
        doc.addFont('Roboto-Regular.ttf', 'Roboto', 'normal');
        doc.setFont('Roboto');

        const margin = 20;
        const pageWidth = doc.internal.pageSize.getWidth();
        const lineLimit = pageWidth - margin * 2;

        // Title
        doc.setFontSize(22);
        const titleLines = doc.splitTextToSize(article.title, lineLimit);
        doc.text(titleLines, margin, 30);

        let cursorY = 30 + (titleLines.length * 10) + 10;

        // Meta info
        doc.setFontSize(10);
        doc.setTextColor(100);
        doc.text(`${article.subject} · ${article.grade}-класс · ${article.language.toUpperCase()}`, margin, cursorY);
        cursorY += 15;

        // Content
        doc.setFontSize(12);
        doc.setTextColor(0);
        const contentLines = doc.splitTextToSize(article.content, lineLimit);

        // Handle pages
        const pageHeight = doc.internal.pageSize.getHeight();
        contentLines.forEach((line: string) => {
            if (cursorY > pageHeight - 20) {
                doc.addPage();
                doc.setFont('Roboto'); // Font needs to be reset on new page
                cursorY = 20;
            }
            doc.text(line, margin, cursorY);
            cursorY += 7;
        });

        doc.save(`${article.title.substring(0, 30)}.pdf`);
    };

    const handleExportWord = async () => {
        if (!article) return;

        const doc = new Document({
            sections: [{
                properties: {},
                children: [
                    new Paragraph({
                        text: article.title,
                        heading: HeadingLevel.HEADING_1,
                        alignment: AlignmentType.CENTER,
                        spacing: { after: 400 },
                    }),
                    new Paragraph({
                        children: [
                            new TextRun({
                                text: `${article.subject} · ${article.grade}-класс · ${article.language.toUpperCase()}`,
                                color: "666666",
                                size: 20,
                            }),
                        ],
                        spacing: { after: 400 },
                    }),
                    ...article.content.split('\n').map(line =>
                        new Paragraph({
                            children: [new TextRun({ text: line, size: 24 })],
                            spacing: { after: 200 },
                        })
                    ),
                ],
            }],
        });

        const blob = await Packer.toBlob(doc);
        saveAs(blob, `${article.title.substring(0, 30)}.docx`);
    };

    if (loading) {
        return (
            <div className="flex flex-col items-center justify-center py-20">
                <Loader2 className="animate-spin text-primary mb-4" size={40} />
                <p className="text-foreground/60">{t('loading')}</p>
            </div>
        );
    }

    if (!article) {
        return (
            <div className="text-center py-20">
                <h1 className="text-2xl font-bold mb-4">{t('not_found')}</h1>
                <button
                    onClick={() => router.back()}
                    className="text-primary font-semibold hover:underline"
                >
                    {t('form.cancel')}
                </button>
            </div>
        );
    }

    return (
        <div className="max-w-4xl mx-auto space-y-6 md:space-y-8 px-4 md:px-0 pb-20">
            {/* Header / Nav */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <button
                    onClick={() => router.push('/teacher/materials/article')}
                    className="p-2 glass-card rounded-xl hover:bg-slate-200 dark:hover:bg-white/10 transition-all text-foreground/60 flex items-center gap-2 pr-4 self-start"
                >
                    <ChevronLeft size={20} />
                    <span className="font-medium">{t('form.cancel')}</span>
                </button>

                <div className="flex items-center gap-2 w-full sm:w-auto">
                    <button
                        onClick={handleExportWord}
                        className="flex-1 sm:flex-initial px-4 md:px-6 py-3 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 text-foreground rounded-xl font-bold hover:bg-slate-50 dark:hover:bg-white/10 transition-all flex items-center justify-center gap-2 shadow-sm"
                    >
                        <FileDown size={20} className="text-blue-500" />
                        <span className="hidden xs:inline">{t('export_word')}</span>
                        <span className="xs:hidden">DOCX</span>
                    </button>
                    <button
                        onClick={handleExportPDF}
                        className="flex-1 sm:flex-initial px-4 md:px-6 py-3 bg-primary text-white rounded-xl font-bold hover:bg-primary-hover transition-all flex items-center justify-center gap-2 shadow-lg shadow-primary/25"
                    >
                        <Download size={20} />
                        <span className="hidden xs:inline">{t('export_pdf')}</span>
                        <span className="xs:hidden">PDF</span>
                    </button>
                </div>
            </div>

            {/* Article Card */}
            <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                className="glass-card p-6 md:p-16 rounded-[32px] md:rounded-[40px] shadow-2xl space-y-6 md:space-y-8 min-h-[60vh] bg-white dark:bg-slate-900/50"
            >
                {/* Meta */}
                <div className="flex flex-wrap items-center gap-3 md:gap-6 text-sm text-foreground/50 border-b border-slate-200 dark:border-white/10 pb-6">
                    <div className="flex items-center gap-2 bg-primary/5 px-4 py-2 rounded-full text-primary font-medium">
                        <BookOpen size={16} />
                        {article.subject}
                    </div>
                    <div className="flex items-center gap-2">
                        <FileText size={16} />
                        {article.grade}{t('form.grade_suffix')} · {article.language.toUpperCase()}
                    </div>
                    <div className="flex flex-wrap items-center gap-2">
                        <span className="opacity-50 text-[10px] md:text-xs px-2 py-0.5 border border-current rounded-md uppercase whitespace-nowrap">{article.style}</span>
                        <span className="opacity-50 text-[10px] md:text-xs px-2 py-0.5 border border-current rounded-md uppercase whitespace-nowrap">{article.complexity}</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <Calendar size={16} />
                        {new Date(article.created_at).toLocaleDateString()}
                    </div>
                </div>

                {/* Title */}
                <h1 className="text-3xl md:text-5xl font-bold font-outfit leading-tight text-slate-900 dark:text-white">
                    {article.title}
                </h1>

                {/* Content */}
                <div className="prose prose-lg dark:prose-invert max-w-none prose-p:leading-relaxed prose-p:text-slate-600 dark:prose-p:text-slate-300 prose-headings:font-outfit prose-headings:text-slate-900 dark:prose-headings:text-white whitespace-pre-wrap">
                    {article.content}
                </div>
            </motion.div>
        </div>
    );
}
