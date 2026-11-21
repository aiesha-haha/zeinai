'use client';

import { useState, useEffect } from 'react';
import { useRouter } from '@/i18n/routing';
import { useParams } from 'next/navigation';
import { useTranslations } from 'next-intl';
import {
    ChevronLeft, ChevronRight, Loader2, XCircle, BookOpen,
    Layers, FileText, CheckCircle, Image as ImageIcon,
    Maximize, Minimize
} from 'lucide-react';
import api from '@/lib/api';

// ── Types ────────────────────────────────────────────────────────────────────

interface Slide {
    id: string;
    slide_number: number;
    title: string;
    content: string;
    notes: string;
    layout: string;
    image_prompt?: string;
    image_url?: string;
}

interface Presentation {
    id: string;
    subject: string;
    grade: number;
    topic: string;
    language: string;
    slide_count: number;
    description: string;
    style: string;
    status: 'pending' | 'processing' | 'completed' | 'failed';
    error_message?: string;
    created_at: string;
}

interface PresentationWithSlides {
    presentation: Presentation;
    slides: Slide[];
}

// ── Helpers ──────────────────────────────────────────────────────────────────

const LAYOUT_EMOJI: Record<string, string> = {
    title_only: '📌',
    bullets: '📋',
    two_column: '📊',
    image_text: '🖼️',
};

const getSlideBackground = (style: string = 'creative', layout: string): string => {
    switch (style) {
        case 'minimal':
            return 'bg-white text-gray-900 border border-t-8 border-t-primary shadow-xl';
        case 'professional':
            return 'bg-gradient-to-br from-slate-800 via-slate-900 to-black text-white shadow-2xl';
        case 'cartoon':
            return 'bg-gradient-to-br from-yellow-100 via-orange-100 to-pink-100 text-slate-800 border-4 border-dashed border-primary/50 font-comic';
        case 'creative':
        default:
            const colors: Record<string, string> = {
                title_only: 'from-purple-600 to-indigo-600',
                bullets: 'from-blue-600 to-cyan-500',
                two_column: 'from-emerald-600 to-teal-500',
                image_text: 'from-orange-500 to-pink-600',
            };
            return `bg-gradient-to-br ${colors[layout] || 'from-primary to-purple-600'} text-white shadow-2xl`;
    }
};

const getTextColor = (style: string) => {
    return style === 'minimal' || style === 'cartoon' ? 'text-slate-900' : 'text-white';
};

const Bullets = ({ lines }: { lines: string[] }) => (
    <ul className="space-y-4 text-left w-full max-w-4xl mx-auto">
        {lines.map((line, i) => (
            <li key={i} className="flex gap-4 items-start animate-fadeIn" style={{ animationDelay: `${i * 100}ms` }}>
                <span className="shrink-0 w-3 h-3 rounded-full bg-current mt-2.5 opacity-60" />
                <span className="text-xl md:text-2xl leading-relaxed font-medium">{line.replace(/^[•\-]\s*/, '')}</span>
            </li>
        ))}
    </ul>
);

// ── Component ────────────────────────────────────────────────────────────────

export default function PresentationViewPage() {
    const params = useParams();
    const router = useRouter();
    const t = useTranslations('slides');

    const [data, setData] = useState<PresentationWithSlides | null>(null);
    const [loading, setLoading] = useState(true);
    const [currentSlide, setCurrentSlide] = useState(0);
    const [isFullscreen, setIsFullscreen] = useState(false);
    const [containerRef, setContainerRef] = useState<HTMLDivElement | null>(null);

    const toggleFullscreen = () => {
        if (!containerRef) return;

        if (!document.fullscreenElement) {
            containerRef.requestFullscreen().catch(err => {
                console.error(`Error attempting to enable full-screen mode: ${err.message}`);
            });
        } else {
            document.exitFullscreen();
        }
    };

    useEffect(() => {
        const handler = () => {
            setIsFullscreen(!!document.fullscreenElement);
        };
        document.addEventListener('fullscreenchange', handler);
        return () => document.removeEventListener('fullscreenchange', handler);
    }, []);

    const fetchPresentation = async () => {
        try {
            const res = await api.get(`/presentations/${params.id}`);
            setData(res.data.data);
        } catch (err) {
            console.error('Failed to fetch presentation:', err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchPresentation();
    }, [params.id]);

    // Poll while not finished
    useEffect(() => {
        let interval: NodeJS.Timeout;
        if (data && (data.presentation.status === 'pending' || data.presentation.status === 'processing')) {
            interval = setInterval(fetchPresentation, 4000);
        }
        return () => clearInterval(interval);
    }, [data?.presentation.status]);

    // Keyboard navigation
    useEffect(() => {
        const onKey = (e: KeyboardEvent) => {
            if (!data?.slides?.length) return;
            if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
                setCurrentSlide((prev) => Math.min(prev + 1, data.slides.length - 1));
            }
            if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
                setCurrentSlide((prev) => Math.max(prev - 1, 0));
            }
        };
        window.addEventListener('keydown', onKey);
        return () => window.removeEventListener('keydown', onKey);
    }, [data?.slides?.length]);

    // ── Loading state ──────────────────────────────────────────────────────
    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <div className="text-center space-y-4">
                    <Loader2 className="w-14 h-14 animate-spin text-primary mx-auto" />
                    <p className="text-foreground/60">{t('loading')}</p>
                </div>
            </div>
        );
    }

    if (!data) {
        return (
            <div className="space-y-8">
                <div className="flex items-center gap-4">
                    <button
                        onClick={() => router.push('/teacher/materials/slides')}
                        className="p-2 glass-card rounded-xl hover:bg-white/10 transition-all"
                    >
                        <ChevronLeft size={20} />
                    </button>
                    <h1 className="text-3xl font-bold">{t('not_found')}</h1>
                </div>
            </div>
        );
    }

    const { presentation: p, slides = [] } = data;

    // ── Generating state ───────────────────────────────────────────────────
    if (p.status === 'pending' || p.status === 'processing') {
        return (
            <div className="space-y-8">
                <div className="flex items-center gap-4">
                    <button
                        onClick={() => router.push('/teacher/materials/slides')}
                        className="p-2 glass-card rounded-xl hover:bg-white/10 transition-all"
                    >
                        <ChevronLeft size={20} />
                    </button>
                    <h1 className="text-3xl font-bold font-outfit">{p.topic}</h1>
                </div>

                <div className="glass-card p-16 rounded-[32px] text-center space-y-8">
                    <div className="w-24 h-24 bg-primary/20 rounded-full mx-auto flex items-center justify-center animate-pulse">
                        <span className="text-5xl">🪄</span>
                    </div>
                    <div>
                        <h2 className="text-2xl font-bold mb-3">{t('ai_generating')}</h2>
                        <p className="text-foreground/60">{t('generation_time')}</p>
                    </div>
                    <div className="w-full max-w-md mx-auto bg-slate-200 dark:bg-white/10 rounded-full h-2.5 overflow-hidden">
                        <div className="bg-gradient-to-r from-primary to-purple-500 h-2.5 rounded-full animate-[pulse_2s_ease-in-out_infinite]" style={{ width: '65%' }} />
                    </div>
                    <p className="text-xs text-foreground/40">
                        {t('slide')} 0 / {p.slide_count}
                    </p>
                </div>
            </div>
        );
    }

    // ── Failed state ───────────────────────────────────────────────────────
    if (p.status === 'failed') {
        return (
            <div className="space-y-8">
                <div className="flex items-center gap-4">
                    <button onClick={() => router.push('/teacher/materials/slides')}
                        className="p-2 glass-card rounded-xl hover:bg-white/10 transition-all">
                        <ChevronLeft size={20} />
                    </button>
                    <h1 className="text-3xl font-bold">{p.topic}</h1>
                </div>
                <div className="glass-card p-12 rounded-[32px] text-center space-y-6">
                    <XCircle className="w-20 h-20 text-red-500 mx-auto" />
                    <div>
                        <h2 className="text-2xl font-bold mb-3">{t('generation_failed')}</h2>
                        <p className="text-foreground/60">{p.error_message || t('generation_failed_desc')}</p>
                    </div>
                    <button
                        onClick={() => router.push('/teacher/materials/slides')}
                        className="px-6 py-3 bg-primary text-white rounded-xl font-semibold hover:bg-primary-hover"
                    >
                        {t('create_new')}
                    </button>
                </div>
            </div>
        );
    }

    // ── Render Completed Slide ─────────────────────────────────────────────
    const slide = slides[currentSlide];
    const bgClass = getSlideBackground(p.style, slide?.layout);
    // const textClass = getTextColor(p.style); // Text color handled by bgClass logic mostly, or internal

    // Content Parsing
    const lines = slide?.content ? slide.content.split('\n').filter(Boolean) : [];

    // Render specific layout content
    const renderLayoutContent = () => {
        if (!slide) return null;

        if (slide.layout === 'image_text' || (slide.image_prompt && slide.image_prompt.length > 5)) {
            return (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-8 items-center h-full w-full">
                    <div className="order-2 md:order-1 flex flex-col justify-center text-left space-y-4">
                        <Bullets lines={lines} />
                    </div>
                    <div className="order-1 md:order-2 flex justify-center items-center h-full w-full relative">
                        {slide.image_url ? (
                            <div className="relative w-full aspect-video rounded-[24px] overflow-hidden shadow-lg border-2 border-white/20 group">
                                <img
                                    src={slide.image_url}
                                    alt={slide.image_prompt || "Slide image"}
                                    className="w-full h-full object-cover transition-transform duration-700 group-hover:scale-105"
                                    loading="lazy"
                                />
                                <div className="absolute inset-0 bg-gradient-to-t from-black/20 to-transparent pointer-events-none" />
                            </div>
                        ) : (
                            <div className="aspect-square w-full max-w-sm rounded-[24px] bg-black/20 backdrop-blur-sm border-2 border-white/30 flex flex-col items-center justify-center p-6 text-center space-y-4 shadow-lg">
                                <ImageIcon size={48} className="opacity-70" />
                                <p className="text-xs opacity-70 italic line-clamp-4">"{slide.image_prompt}"</p>
                                <p className="text-[10px] uppercase font-bold opacity-50 bg-black/20 px-2 py-1 rounded">AI Image Placeholder</p>
                            </div>
                        )}
                    </div>
                </div>
            );
        }

        if (slide.layout === 'two_column') {
            const midpoint = Math.ceil(lines.length / 2);
            const col1 = lines.slice(0, midpoint);
            const col2 = lines.slice(midpoint);
            return (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-8 w-full max-w-5xl">
                    <div><Bullets lines={col1} /></div>
                    <div><Bullets lines={col2} /></div>
                </div>
            );
        }

        // Default or bullets
        return <Bullets lines={lines} />;
    };

    return (
        <div className="md:h-[calc(100vh-7.5rem)] flex flex-col gap-4 overflow-hidden pb-4">
            {/* Header */}
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 shrink-0 px-2 lg:px-0">
                <div className="flex items-center gap-4">
                    <button
                        onClick={() => router.push('/teacher/materials/slides')}
                        className="p-2 glass-card rounded-xl hover:bg-white/10 transition-all"
                    >
                        <ChevronLeft size={20} />
                    </button>
                    <div>
                        <div className="flex items-center gap-3">
                            <h1 className="text-2xl md:text-3xl font-bold font-outfit line-clamp-1">{p.topic}</h1>
                            <span className="hidden md:flex items-center gap-1 px-3 py-1 bg-green-500/20 text-green-400 rounded-full text-sm font-semibold">
                                <CheckCircle size={14} />
                                {t('status_completed')}
                            </span>
                        </div>
                        <div className="flex items-center gap-4 mt-1 text-sm text-foreground/60">
                            <span className="flex items-center gap-1"><BookOpen size={14} /> {p.subject} · {p.grade}-сынып</span>
                            <span className="flex items-center gap-1"><Layers size={14} /> {slides.length} {t('slide')}</span>
                            <span className="opacity-50 text-xs px-2 py-0.5 border border-current rounded-md uppercase">{p.style || 'creative'}</span>
                        </div>
                    </div>
                </div>
                <div className="flex items-center gap-3">
                    <button
                        onClick={toggleFullscreen}
                        className="flex items-center gap-2 px-4 py-2 bg-primary/10 text-primary hover:bg-primary/20 rounded-xl font-semibold transition-all"
                        title="Полноэкранный режим"
                    >
                        {isFullscreen ? <Minimize size={20} /> : <Maximize size={20} />}
                        <span className="hidden sm:inline">{isFullscreen ? 'Свернуть' : 'Во весь экран'}</span>
                    </button>
                </div>
            </div>

            {/* Slide viewer */}
            {slides.length === 0 ? (
                <div className="glass-card p-12 rounded-[32px] text-center flex-1 flex items-center justify-center">
                    <p className="text-foreground/60">{t('no_slides')}</p>
                </div>
            ) : (
                <div className="flex-1 min-h-0 flex flex-col gap-4">
                    {/* Main slide / preview container container */}
                    <div className={`flex-1 min-h-0 flex items-center justify-center w-full ${isFullscreen ? '' : 'px-2 lg:px-0'}`}>
                        <div
                            ref={setContainerRef}
                            className={`
                                relative w-full overflow-hidden transition-all duration-500 ${bgClass}
                                ${isFullscreen
                                    ? 'fixed inset-0 z-[100] h-screen w-screen flex flex-col items-center justify-center rounded-0'
                                    : 'h-full max-w-full aspect-[16/9] rounded-[32px] shadow-2xl mx-auto'
                                }
                            `}
                        >

                            {/* Fullscreen Toggle / Exit Button (Inside) */}
                            <div className="absolute top-6 right-16 z-[110]">
                                {isFullscreen && (
                                    <button
                                        onClick={toggleFullscreen}
                                        className="p-3 bg-black/20 hover:bg-black/40 text-white rounded-full transition-all backdrop-blur-md border border-white/30 shadow-xl"
                                        title="Выйти из полноэкранного режима"
                                    >
                                        <Minimize size={24} />
                                    </button>
                                )}
                            </div>

                            {/* Slide number badge */}
                            <div className={`absolute bottom-6 right-6 font-bold opacity-50 ${isFullscreen ? 'text-2xl' : 'text-lg'}`}>
                                {currentSlide + 1}
                            </div>

                            {/* Top Bar (Title) */}
                            <div className={`absolute top-0 left-0 right-0 px-12 md:px-20 pt-12 pb-0 flex items-start justify-between z-10 ${isFullscreen ? 'mt-4' : ''}`}>
                                <div className="max-w-[85%]">
                                    <h2 className={`
                                    font-bold leading-tight drop-shadow-sm
                                    ${isFullscreen ? 'text-4xl md:text-6xl lg:text-7xl' : 'text-3xl md:text-5xl'}
                                    ${slide.layout === 'title_only' ? 'text-center w-full mt-20' : ''}
                                `}>
                                        {slide.title}
                                    </h2>
                                </div>
                                {/* Layout Icon */}
                                <div className={`${isFullscreen ? 'text-5xl' : 'text-2xl'} opacity-40`} title={slide.layout}>
                                    {LAYOUT_EMOJI[slide.layout]}
                                </div>
                            </div>

                            {/* Content Area */}
                            <div className={`
                            absolute inset-0 flex flex-col items-center justify-center overflow-y-auto custom-scrollbar
                            ${isFullscreen ? 'px-20 md:px-32 top-32 md:top-40' : 'px-8 md:px-12 top-24 md:top-32 pb-8'}
                        `}>
                                {slide.layout === 'title_only' ? null : (
                                    <div className={`w-full h-full flex flex-col justify-center ${isFullscreen ? 'text-xl md:text-2xl' : ''}`}>
                                        {renderLayoutContent()}
                                    </div>
                                )}
                            </div>

                            {/* Prev / Next buttons (overlay) */}
                            {currentSlide > 0 && (
                                <button
                                    onClick={() => setCurrentSlide(currentSlide - 1)}
                                    className="absolute left-4 top-1/2 -translate-y-1/2 p-3 rounded-full hover:bg-black/10 dark:hover:bg-white/10 transition-all opacity-0 hover:opacity-100 group-hover:opacity-100"
                                >
                                    <ChevronLeft size={32} className="opacity-70" />
                                </button>
                            )}
                            {currentSlide < slides.length - 1 && (
                                <button
                                    onClick={() => setCurrentSlide(currentSlide + 1)}
                                    className="absolute right-4 top-1/2 -translate-y-1/2 p-3 rounded-full hover:bg-black/10 dark:hover:bg-white/10 transition-all opacity-0 hover:opacity-100 group-hover:opacity-100"
                                >
                                    <ChevronRight size={32} className="opacity-70" />
                                </button>
                            )}
                        </div>
                    </div>

                    {/* Bottom area */}
                    <div className={`grid grid-cols-1 lg:grid-cols-3 gap-6 shrink-0 ${isFullscreen ? 'hidden' : 'h-40 md:h-48 px-2 lg:px-0'}`}>
                        {/* Thumbnails */}
                        <div className="lg:col-span-2 flex flex-col min-w-0">
                            <h3 className="font-bold text-sm px-2 mb-2 shrink-0">{t('all_slides')}</h3>
                            <div className="flex gap-4 overflow-x-auto pb-4 px-2 scrollbar-hide snap-x flex-1 items-center">
                                {slides.map((s, i) => {
                                    const isActive = i === currentSlide;
                                    return (
                                        <button
                                            key={s.id}
                                            onClick={() => setCurrentSlide(i)}
                                            className={`
                                                snap-center shrink-0 w-32 md:w-36 aspect-video rounded-xl transition-all border-2
                                                flex flex-col items-center justify-center p-2 gap-1 relative overflow-hidden
                                                ${isActive
                                                    ? 'border-primary ring-2 ring-primary/20 scale-105 shadow-lg bg-white dark:bg-slate-800'
                                                    : 'border-transparent bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 opacity-70 hover:opacity-100'
                                                }
                                            `}
                                        >
                                            <span className="text-[10px] uppercase font-bold opacity-40 w-full text-left">{i + 1}</span>
                                            <p className="text-[10px] md:text-xs font-bold text-center line-clamp-2 leading-tight w-full">
                                                {s.title}
                                            </p>
                                            <span className="absolute bottom-1 right-1 text-xs opacity-30 grayscale filter">
                                                {LAYOUT_EMOJI[s.layout]}
                                            </span>
                                        </button>
                                    );
                                })}
                            </div>
                        </div>

                        {/* Speaker notes */}
                        <div className="glass-card p-4 rounded-2xl flex flex-col min-h-0 overflow-hidden shadow-sm border-white/10">
                            <div className="flex items-center gap-2 mb-2 text-[10px] font-bold text-primary uppercase tracking-wider shrink-0">
                                <FileText size={14} />
                                {t('notes')}
                            </div>
                            <div className="flex-1 overflow-y-auto custom-scrollbar pr-1">
                                {slide.notes ? (
                                    <p className="text-[11px] md:text-xs leading-relaxed text-foreground/80 whitespace-pre-wrap">
                                        {slide.notes}
                                    </p>
                                ) : (
                                    <p className="text-xs text-foreground/40 italic">Нет заметок для этого слайда</p>
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
