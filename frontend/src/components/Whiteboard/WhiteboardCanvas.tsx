"use client";

import React, { useRef, useEffect, useState, useCallback, forwardRef, useImperativeHandle } from "react";
import { WhiteboardAction } from "@/types/whiteboard";
import { Eraser, Pen, Hand, ZoomIn, ZoomOut, Maximize } from "lucide-react";
import { useTranslations } from "next-intl";

type Tool = "pen" | "eraser" | "hand";

export interface WhiteboardCanvasProps {
    onDraw?: (data: WhiteboardAction) => void;
    isReadOnly?: boolean;
    initialData?: WhiteboardAction[];
    lastAction?: WhiteboardAction | null;
    classId: number;
    currentPage?: number;
}

export interface WhiteboardCanvasHandle {
    clearCanvas: () => void;
    getImageData: () => ImageData | null;
    restoreImageData: (data: ImageData) => void;
}

export const WhiteboardCanvas = forwardRef<WhiteboardCanvasHandle, WhiteboardCanvasProps>((props: WhiteboardCanvasProps, ref: React.ForwardedRef<WhiteboardCanvasHandle>) => {
    const t = useTranslations('whiteboard_teacher');
    const {
        onDraw,
        isReadOnly = false,
        initialData = [],
        lastAction = null,
        currentPage = 0,
    } = props;

    const canvasRef = useRef<HTMLCanvasElement>(null);
    const contextRef = useRef<CanvasRenderingContext2D | null>(null);
    const [isDrawing, setIsDrawing] = useState(false);
    const [tool, setTool] = useState<Tool>(isReadOnly ? "hand" : "pen");
    const [color, setColor] = useState("#000000");
    const [lineWidth, setLineWidth] = useState(3);

    // Panning and Zooming state
    const [offset, setOffset] = useState({ x: 0, y: 0 });
    const [scale, setScale] = useState(1);
    const [isPanning, setIsPanning] = useState(false);
    const lastPanPos = useRef({ x: 0, y: 0 });
    const lastPinchDist = useRef<number | null>(null);
    const lastPinchMid = useRef<{ x: number, y: number } | null>(null);

    const applyToolToContext = useCallback((ctx: CanvasRenderingContext2D, isEraser: boolean, c?: string, lw?: number) => {
        if (isEraser) {
            ctx.globalCompositeOperation = "destination-out";
            ctx.strokeStyle = "rgba(0,0,0,1)";
            ctx.lineWidth = 24;
        } else {
            ctx.globalCompositeOperation = "source-over";
            ctx.strokeStyle = c || "#000000";
            ctx.lineWidth = lw || 3;
        }
    }, []);

    const drawAction = useCallback((ctx: CanvasRenderingContext2D, data: WhiteboardAction, currentOffset: { x: number, y: number }, currentScale: number) => {
        switch (data.type) {
            case "start":
                ctx.beginPath();
                applyToolToContext(ctx, !!data.isEraser, data.color, (data.lineWidth || 3) * currentScale);
                if (data.x !== undefined && data.y !== undefined) {
                    ctx.moveTo(data.x * currentScale + currentOffset.x, data.y * currentScale + currentOffset.y);
                }
                break;
            case "move":
                if (data.x !== undefined && data.y !== undefined) {
                    ctx.lineTo(data.x * currentScale + currentOffset.x, data.y * currentScale + currentOffset.y);
                    ctx.stroke();
                }
                break;
            case "end":
                ctx.closePath();
                break;
            case "clear":
                const canvas = canvasRef.current;
                if (canvas) ctx.clearRect(0, 0, canvas.width, canvas.height);
                break;
        }
    }, [applyToolToContext]);

    const redrawAll = useCallback(() => {
        const canvas = canvasRef.current;
        const ctx = contextRef.current;
        if (!canvas || !ctx) return;

        ctx.clearRect(0, 0, canvas.width, canvas.height);

        // Filter actions for current page and redraw
        const pageActions = (initialData as WhiteboardAction[]).filter(a => (a.page_index ?? 0) === currentPage);
        pageActions.forEach(action => drawAction(ctx, action, offset, scale));

        // Restore currently selected tool settings
        applyToolToContext(ctx, tool === "eraser", color, lineWidth * scale);
    }, [initialData, currentPage, offset, scale, tool, color, lineWidth, drawAction, applyToolToContext]);

    // Setup canvas
    useEffect(() => {
        const canvas = canvasRef.current;
        if (!canvas) return;

        const resizeCanvas = () => {
            const parent = canvas.parentElement;
            if (parent) {
                canvas.width = parent.clientWidth;
                canvas.height = parent.clientHeight;
                const context = canvas.getContext("2d");
                if (context) {
                    context.lineCap = "round";
                    context.lineJoin = "round";
                    contextRef.current = context;
                    redrawAll();
                }
            }
        };

        resizeCanvas();
        window.addEventListener("resize", resizeCanvas);
        return () => window.removeEventListener("resize", resizeCanvas);
    }, [redrawAll]);

    const getCoordinates = useCallback((e: React.MouseEvent | React.TouchEvent | TouchEvent | MouseEvent) => {
        const canvas = canvasRef.current;
        if (!canvas) return { x: 0, y: 0 };
        const rect = canvas.getBoundingClientRect();

        let clientX, clientY;

        if ("touches" in e) {
            if (e.touches.length > 0) {
                clientX = e.touches[0].clientX;
                clientY = e.touches[0].clientY;
            } else if ("changedTouches" in e && e.changedTouches.length > 0) {
                clientX = e.changedTouches[0].clientX;
                clientY = e.changedTouches[0].clientY;
            } else {
                return { x: 0, y: 0 };
            }
        } else {
            clientX = (e as MouseEvent).clientX;
            clientY = (e as MouseEvent).clientY;
        }

        return {
            x: clientX - rect.left,
            y: clientY - rect.top,
        };
    }, []);

    const startDrawingInternal = useCallback((e: React.MouseEvent | React.TouchEvent | TouchEvent | MouseEvent) => {
        const { x, y } = getCoordinates(e);

        if (tool === "hand") {
            setIsPanning(true);
            lastPanPos.current = { x, y };
            return;
        }

        if (isReadOnly) return;

        const ctx = contextRef.current;
        if (!ctx) return;

        // Convert to world coordinates for storage
        const worldX = (x - offset.x) / scale;
        const worldY = (y - offset.y) / scale;

        ctx.beginPath();
        ctx.moveTo(x, y);
        setIsDrawing(true);
        onDraw?.({
            type: "start",
            x: worldX,
            y: worldY,
            color,
            lineWidth,
            isEraser: tool === "eraser",
            page_index: currentPage
        });
    }, [isReadOnly, getCoordinates, tool, offset, scale, color, lineWidth, currentPage, onDraw]);

    const drawInternal = useCallback((e: React.MouseEvent | React.TouchEvent | TouchEvent | MouseEvent) => {
        const { x, y } = getCoordinates(e);

        if (isPanning && tool === "hand") {
            const dx = x - lastPanPos.current.x;
            const dy = y - lastPanPos.current.y;
            setOffset((prev: { x: number, y: number }) => ({ x: prev.x + dx, y: prev.y + dy }));
            lastPanPos.current = { x, y };
            return;
        }

        if (isReadOnly || !isDrawing) return;

        const ctx = contextRef.current;
        if (!ctx) return;

        const worldX = (x - offset.x) / scale;
        const worldY = (y - offset.y) / scale;

        ctx.lineTo(x, y);
        ctx.stroke();
        onDraw?.({ type: "move", x: worldX, y: worldY, page_index: currentPage });
    }, [isReadOnly, getCoordinates, isPanning, tool, isDrawing, offset, scale, currentPage, onDraw]);

    const stopDrawingInternal = useCallback(() => {
        if (tool === "hand") {
            setIsPanning(false);
            return;
        }

        if (!isReadOnly && isDrawing) {
            contextRef.current?.closePath();
            setIsDrawing(false);
            onDraw?.({ type: "end", page_index: currentPage });
        }
    }, [isReadOnly, tool, isDrawing, onDraw, currentPage]);

    const handleZoom = useCallback((factor: number, px: number, py: number) => {
        setScale((prevScale: number) => {
            const newScale = Math.min(Math.max(prevScale * factor, 0.1), 10);
            const actualFactor = newScale / prevScale;

            setOffset((prevOffset: { x: number, y: number }) => ({
                x: px - (px - prevOffset.x) * actualFactor,
                y: py - (py - prevOffset.y) * actualFactor,
            }));

            return newScale;
        });
    }, []);

    // Manual event listeners for touch and wheel
    useEffect(() => {
        const canvas = canvasRef.current;
        if (!canvas) return;

        const handleTouchStart = (e: TouchEvent) => {
            if (e.touches.length === 2) {
                const dist = Math.hypot(
                    e.touches[0].clientX - e.touches[1].clientX,
                    e.touches[0].clientY - e.touches[1].clientY
                );
                const rect = canvas.getBoundingClientRect();
                const midX = (e.touches[0].clientX + e.touches[1].clientX) / 2 - rect.left;
                const midY = (e.touches[0].clientY + e.touches[1].clientY) / 2 - rect.top;
                lastPinchDist.current = dist;
                lastPinchMid.current = { x: midX, y: midY };
                return;
            }
            if (e.cancelable) e.preventDefault();
            startDrawingInternal(e);
        };

        const handleTouchMove = (e: TouchEvent) => {
            if (e.touches.length === 2 && lastPinchDist.current !== null && lastPinchMid.current !== null) {
                const dist = Math.hypot(
                    e.touches[0].clientX - e.touches[1].clientX,
                    e.touches[0].clientY - e.touches[1].clientY
                );
                const factor = dist / lastPinchDist.current;
                handleZoom(factor, lastPinchMid.current.x, lastPinchMid.current.y);
                lastPinchDist.current = dist;
                return;
            }
            if (e.cancelable) e.preventDefault();
            drawInternal(e);
        };

        const handleTouchEnd = (e: TouchEvent) => {
            if (e.touches.length < 2) {
                lastPinchDist.current = null;
                lastPinchMid.current = null;
            }
            stopDrawingInternal();
        };

        const handleWheel = (e: WheelEvent) => {
            if (e.ctrlKey || e.metaKey) {
                e.preventDefault();
                const rect = canvas.getBoundingClientRect();
                const px = e.clientX - rect.left;
                const py = e.clientY - rect.top;
                const factor = e.deltaY > 0 ? 0.9 : 1.1;
                handleZoom(factor, px, py);
            }
        };

        canvas.addEventListener("touchstart", handleTouchStart, { passive: false });
        canvas.addEventListener("touchmove", handleTouchMove, { passive: false });
        canvas.addEventListener("touchend", handleTouchEnd, { passive: false });
        canvas.addEventListener("wheel", handleWheel, { passive: false });

        return () => {
            canvas.removeEventListener("touchstart", handleTouchStart);
            canvas.removeEventListener("touchmove", handleTouchMove);
            canvas.removeEventListener("touchend", handleTouchEnd);
            canvas.removeEventListener("wheel", handleWheel);
        };
    }, [isReadOnly, startDrawingInternal, drawInternal, stopDrawingInternal, handleZoom]);

    const clearCanvas = useCallback(() => {
        const canvas = canvasRef.current;
        const ctx = contextRef.current;
        if (canvas && ctx) {
            ctx.clearRect(0, 0, canvas.width, canvas.height);
            onDraw?.({ type: "clear", page_index: currentPage });
        }
    }, [onDraw, currentPage]);

    useImperativeHandle(ref, () => ({
        clearCanvas,
        getImageData: () => {
            const canvas = canvasRef.current;
            const ctx = contextRef.current;
            if (!canvas || !ctx) return null;
            return ctx.getImageData(0, 0, canvas.width, canvas.height);
        },
        restoreImageData: (data: ImageData) => {
            const ctx = contextRef.current;
            if (ctx) ctx.putImageData(data, 0, 0);
        },
    }));

    // Trigger redraw when initialData, offset or scale changes
    useEffect(() => {
        redrawAll();
    }, [initialData, offset, scale, currentPage, redrawAll]);

    // Handle lastAction from other users
    useEffect(() => {
        if (lastAction && isReadOnly) {
            const ctx = contextRef.current;
            if (ctx) drawAction(ctx, lastAction, offset, scale);
        }
    }, [lastAction, isReadOnly, offset, scale, drawAction]);

    const toolBtn = (t: Tool, icon: React.ReactNode, label: string) => (
        <button
            type="button"
            title={label}
            onClick={() => setTool(t)}
            className={`p-2 rounded-lg transition-all ${tool === t
                ? "bg-primary text-white shadow"
                : "hover:bg-zinc-100 dark:hover:bg-zinc-800 text-zinc-600 dark:text-zinc-300"
                }`}
        >
            {icon}
        </button>
    );

    return (
        <div className="relative w-full h-full bg-white dark:bg-zinc-900 rounded-lg shadow-inner overflow-hidden border-2 border-zinc-200 dark:border-zinc-800 touch-none">
            {!isReadOnly && (
                <div className="absolute top-4 left-4 z-10 flex flex-col items-center gap-2 p-2 bg-zinc-100/90 dark:bg-zinc-800/90 backdrop-blur rounded-xl shadow border border-zinc-200 dark:border-zinc-700">
                    {toolBtn("pen", <Pen className="w-4 h-4" />, t('tool_pen'))}
                    {toolBtn("eraser", <Eraser className="w-4 h-4" />, t('tool_eraser'))}
                    {toolBtn("hand", <Hand className="w-4 h-4" />, t('tool_hand'))}

                    <div className="h-px w-6 bg-zinc-300 dark:bg-zinc-600 my-1" />

                    <input
                        type="color"
                        value={color}
                        onChange={(e: React.ChangeEvent<HTMLInputElement>) => { setColor(e.target.value); setTool("pen"); }}
                        disabled={tool === "eraser" || tool === "hand"}
                        title={t('tool_color')}
                        className="w-8 h-8 rounded cursor-pointer border-none bg-transparent disabled:opacity-40"
                    />

                    <select
                        value={lineWidth}
                        onChange={(e: React.ChangeEvent<HTMLSelectElement>) => setLineWidth(Number(e.target.value))}
                        disabled={tool === "eraser" || tool === "hand"}
                        className="text-[10px] w-10 bg-white/50 dark:bg-zinc-900/50 dark:text-white border-none focus:ring-0 disabled:opacity-40 cursor-pointer rounded-lg px-0.5 py-1 text-center"
                    >
                        {[1, 2, 3, 5, 8, 12].map((w) => (
                            <option
                                key={w}
                                value={w}
                                className="bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100"
                            >
                                {w}
                            </option>
                        ))}
                    </select>

                    <div className="h-px w-6 bg-zinc-300 dark:bg-zinc-600 my-1" />

                    <div className="flex flex-col items-center gap-1">
                        <button
                            type="button"
                            title={t('tool_zoom_in') || "Zoom In"}
                            onClick={() => handleZoom(1.1, (canvasRef.current?.width || 0) / 2, (canvasRef.current?.height || 0) / 2)}
                            className="p-1.5 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-700 text-zinc-600 dark:text-zinc-300"
                        >
                            <ZoomIn className="w-4 h-4" />
                        </button>
                        <span className="text-[10px] font-medium min-h-[20px] flex items-center justify-center text-zinc-500">
                            {Math.round(scale * 100)}%
                        </span>
                        <button
                            type="button"
                            title={t('tool_zoom_out') || "Zoom Out"}
                            onClick={() => handleZoom(0.9, (canvasRef.current?.width || 0) / 2, (canvasRef.current?.height || 0) / 2)}
                            className="p-1.5 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-700 text-zinc-600 dark:text-zinc-300"
                        >
                            <ZoomOut className="w-4 h-4" />
                        </button>
                    </div>

                    <div className="h-px w-6 bg-zinc-300 dark:bg-zinc-600 my-1" />

                    <button
                        type="button"
                        onClick={clearCanvas}
                        title={t('tool_clear')}
                        className="p-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition"
                    >
                        <Eraser className="w-4 h-4" /> {/* Using icon for clear in vertical layout */}
                    </button>

                    {(offset.x !== 0 || offset.y !== 0 || scale !== 1) && (
                        <button
                            type="button"
                            onClick={() => { setOffset({ x: 0, y: 0 }); setScale(1); }}
                            className="p-1.5 rounded-lg bg-zinc-500 text-white hover:bg-zinc-600 transition"
                            title={t('tool_reset_view') || "Reset View"}
                        >
                            <Maximize className="w-4 h-4" />
                        </button>
                    )}
                </div>
            )}

            {/* Read-only navigation toolbar for students */}
            {isReadOnly && (
                <div className="absolute top-4 left-4 z-10 flex flex-col items-center gap-2 p-2 bg-zinc-100/90 dark:bg-zinc-800/90 backdrop-blur rounded-xl shadow border border-zinc-200 dark:border-zinc-700">
                    {toolBtn("hand", <Hand className="w-4 h-4" />, t('tool_hand'))}

                    <div className="h-px w-6 bg-zinc-300 dark:bg-zinc-600 my-1" />

                    <div className="flex flex-col items-center gap-1">
                        <button
                            type="button"
                            title={t('tool_zoom_in') || "Zoom In"}
                            onClick={() => handleZoom(1.1, (canvasRef.current?.width || 0) / 2, (canvasRef.current?.height || 0) / 2)}
                            className="p-1.5 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-700 text-zinc-600 dark:text-zinc-300"
                        >
                            <ZoomIn className="w-4 h-4" />
                        </button>
                        <span className="text-[10px] font-medium min-h-[20px] flex items-center justify-center text-zinc-500">
                            {Math.round(scale * 100)}%
                        </span>
                        <button
                            type="button"
                            title={t('tool_zoom_out') || "Zoom Out"}
                            onClick={() => handleZoom(0.9, (canvasRef.current?.width || 0) / 2, (canvasRef.current?.height || 0) / 2)}
                            className="p-1.5 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-700 text-zinc-600 dark:text-zinc-300"
                        >
                            <ZoomOut className="w-4 h-4" />
                        </button>
                    </div>

                    {(offset.x !== 0 || offset.y !== 0 || scale !== 1) && (
                        <button
                            type="button"
                            onClick={() => { setOffset({ x: 0, y: 0 }); setScale(1); }}
                            className="p-1.5 rounded-lg bg-zinc-500 text-white hover:bg-zinc-600 transition"
                            title={t('tool_reset_view') || "Reset View"}
                        >
                            <Maximize className="w-4 h-4" />
                        </button>
                    )}
                </div>
            )}

            <canvas
                ref={canvasRef}
                onMouseDown={(e) => startDrawingInternal(e)}
                onMouseMove={(e) => drawInternal(e)}
                onMouseUp={() => stopDrawingInternal()}
                onMouseLeave={() => stopDrawingInternal()}
                className={`w-full h-full touch-none ${tool === "hand"
                        ? isPanning ? "cursor-grabbing" : "cursor-grab"
                        : isReadOnly
                            ? "cursor-default"
                            : tool === "eraser"
                                ? "cursor-cell"
                                : "cursor-crosshair"
                    }`}
            />
        </div>
    );
});

WhiteboardCanvas.displayName = "WhiteboardCanvas";
