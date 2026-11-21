"use client";

import React, { useEffect, useRef, useState, useCallback } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { X, Camera, Shuffle, RotateCcw, Settings2, RefreshCcw } from "lucide-react";
import { useTranslations } from "next-intl";

declare global {
    interface Window {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        faceapi: any;
    }
}

type SelectorState = "loading" | "camera" | "selecting" | "result";

interface FaceSelectorModalProps {
    isOpen: boolean;
    onClose: () => void;
}

const FACE_API_CDN = "https://cdn.jsdelivr.net/npm/face-api.js@0.22.2/dist/face-api.min.js";
const MODELS_URL = "https://cdn.jsdelivr.net/gh/justadudewhohacks/face-api.js@0.22.2/weights";

export default function FaceSelectorModal({ isOpen, onClose }: FaceSelectorModalProps) {
    const t = useTranslations("whiteboard_teacher");

    const videoRef = useRef<HTMLVideoElement>(null);
    const overlayRef = useRef<HTMLCanvasElement>(null);
    const resultCanvasRef = useRef<HTMLCanvasElement>(null);
    const streamRef = useRef<MediaStream | null>(null);
    const detectionLoopRef = useRef<number | null>(null);
    const detectedFacesRef = useRef<any[]>([]);
    const animFrameRef = useRef<number | null>(null);
    const usedIndicesRef = useRef<Set<number>>(new Set()); // no-repeat tracking

    const [state, setState] = useState<SelectorState>("loading");
    const [videoDevices, setVideoDevices] = useState<MediaDeviceInfo[]>([]);
    const [selectedDeviceId, setSelectedDeviceId] = useState<string | null>(null);
    const [faceCount, setFaceCount] = useState(0);
    const [error, setError] = useState<string | null>(null);
    const [showSettings, setShowSettings] = useState(false);
    const [animSpeed, setAnimSpeed] = useState(80);
    const [animIterations, setAnimIterations] = useState(25);
    const [zoomLevel, setZoomLevel] = useState(2.2);
    const [showConfetti, setShowConfetti] = useState(false);
    const [resultImageUrl, setResultImageUrl] = useState<string | null>(null);
    const [cycleComplete, setCycleComplete] = useState(false); // all faces picked this round
    const [highAccuracy, setHighAccuracy] = useState(true);

    // ─── Load face-api script ────────────────────────────────────────────────
    const loadFaceApi = useCallback(() => {
        if (window.faceapi) return Promise.resolve();
        return new Promise<void>((resolve, reject) => {
            const existing = document.getElementById("face-api-script");
            if (existing) { resolve(); return; }
            const script = document.createElement("script");
            script.id = "face-api-script";
            script.src = FACE_API_CDN;
            script.onload = () => resolve();
            script.onerror = () => reject(new Error("Failed to load face-api.js"));
            document.head.appendChild(script);
        });
    }, []);

    // ─── Init: load models then start camera ────────────────────────────────
    useEffect(() => {
        if (!isOpen) return;
        setState("loading");
        setError(null);
        setFaceCount(0);

        let cancelled = false;

        const init = async () => {
            try {
                await loadFaceApi();
                if (cancelled) return;

                if (!window.faceapi.nets.tinyFaceDetector.isLoaded) {
                    await window.faceapi.nets.tinyFaceDetector.loadFromUri(MODELS_URL);
                }
                if (!window.faceapi.nets.ssdMobilenetv1.isLoaded) {
                    await window.faceapi.nets.ssdMobilenetv1.loadFromUri(MODELS_URL);
                }
                if (cancelled) return;

                // Enumerate devices
                const devices = await navigator.mediaDevices.enumerateDevices();
                const vDevices = devices.filter(d => d.kind === "videoinput");
                setVideoDevices(vDevices);

                // Use existing selection or default to first
                const deviceId = selectedDeviceId || vDevices[0]?.deviceId;
                if (deviceId && !selectedDeviceId) setSelectedDeviceId(deviceId);

                const stream = await navigator.mediaDevices.getUserMedia({
                    video: {
                        deviceId: deviceId ? { exact: deviceId } : undefined,
                        width: { ideal: 1280 },
                        height: { ideal: 720 },
                        facingMode: deviceId ? undefined : "user"
                    },
                    audio: false,
                });
                if (cancelled) { stream.getTracks().forEach(t => t.stop()); return; }

                // RE-ENUMERATE devices after permission granted
                const devicesAfter = await navigator.mediaDevices.enumerateDevices();
                const vDevicesAfter = devicesAfter.filter(d => d.kind === "videoinput");
                setVideoDevices(vDevicesAfter);

                streamRef.current = stream;
                setState("camera");
            } catch (e: any) {
                if (!cancelled) setError(e?.message || "Камераға қол жетімділік жоқ");
            }
        };

        init();

        return () => {
            cancelled = true;
            stopAll();
        };
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [isOpen, selectedDeviceId]);

    const stopAll = () => {
        if (detectionLoopRef.current) cancelAnimationFrame(detectionLoopRef.current);
        if (animFrameRef.current) clearTimeout(animFrameRef.current);
        streamRef.current?.getTracks().forEach(t => t.stop());
        streamRef.current = null;
    };

    // ─── Attach stream to video element after it mounts ──────────────────────
    // The video element is conditionally rendered, so we must set srcObject
    // in a useEffect triggered by state change, not in the async init above.
    useEffect(() => {
        if (state === "camera" && videoRef.current && streamRef.current) {
            videoRef.current.srcObject = streamRef.current;
            videoRef.current.play().catch(() => { });
        }
    }, [state]);

    // ─── Detection loop ──────────────────────────────────────────────────────
    useEffect(() => {
        if (state !== "camera") { return; }

        const runDetection = async () => {
            const video = videoRef.current;
            const canvas = overlayRef.current;
            if (!video || !canvas || !window.faceapi) return;

            // Wait until the video has actual frame data (readyState >= 2 = HAVE_CURRENT_DATA)
            // and real dimensions — otherwise faceapi processes empty/black frames
            if (video.readyState < 2 || video.videoWidth === 0 || video.videoHeight === 0 || video.paused || video.ended) {
                detectionLoopRef.current = requestAnimationFrame(runDetection);
                return;
            }

            const detections = highAccuracy
                ? await window.faceapi.detectAllFaces(
                    video,
                    new window.faceapi.SsdMobilenetv1Options({
                        minConfidence: 0.35,
                        maxResults: 100
                    })
                )
                : await window.faceapi.detectAllFaces(
                    video,
                    new window.faceapi.TinyFaceDetectorOptions({
                        inputSize: 608,
                        scoreThreshold: 0.25
                    })
                );

            detectedFacesRef.current = detections;
            setFaceCount(detections.length);

            const dims = window.faceapi.matchDimensions(canvas, video, true);
            if (dims.width === 0 || dims.height === 0) {
                detectionLoopRef.current = requestAnimationFrame(runDetection);
                return;
            }
            const resized = window.faceapi.resizeResults(detections, dims);

            const ctx = canvas.getContext("2d");
            if (ctx) {
                ctx.clearRect(0, 0, canvas.width, canvas.height);
                resized.forEach((det: any) => {
                    const { x, y, width, height } = det.box;
                    ctx.strokeStyle = "rgba(139,92,246,0.9)";
                    ctx.lineWidth = 3;
                    ctx.shadowColor = "#8b5cf6";
                    ctx.shadowBlur = 12;
                    ctx.strokeRect(x, y, width, height);
                    ctx.shadowBlur = 0;
                });
            }

            detectionLoopRef.current = requestAnimationFrame(runDetection);
        };

        detectionLoopRef.current = requestAnimationFrame(runDetection);
        return () => { if (detectionLoopRef.current) cancelAnimationFrame(detectionLoopRef.current); };
    }, [state]);

    // ─── Pick handler ────────────────────────────────────────────────────────
    const handlePick = useCallback(() => {
        const faces = detectedFacesRef.current;
        if (!faces || faces.length === 0) return;

        // Check refs BEFORE setState — resultCanvasRef is null until isResult=true!
        const video = videoRef.current;
        const overlayCanvas = overlayRef.current;
        if (!video || !overlayCanvas) return;

        if (detectionLoopRef.current) cancelAnimationFrame(detectionLoopRef.current);
        setState("selecting");

        const vW = video.videoWidth || 1280;
        const vH = video.videoHeight || 720;

        // Face boxes in VIDEO pixel space (source of truth)
        const videoBoxes: Array<{ x: number; y: number; width: number; height: number }> =
            window.faceapi.resizeResults(faces, { width: vW, height: vH })
                .map((d: any) => ({
                    x: d.box.x, y: d.box.y,
                    width: d.box.width, height: d.box.height,
                }));

        if (videoBoxes.length === 0) { setState("camera"); return; }

        // ── No-repeat selection ──
        // Build pool of indices not yet used in this cycle
        const available = videoBoxes
            .map((_: any, i: number) => i)
            .filter((i: number) => !usedIndicesRef.current.has(i));

        // If all faces already picked → reset cycle and use everyone again
        const isCycleReset = available.length === 0;
        if (isCycleReset) {
            usedIndicesRef.current.clear();
        }
        const pool = isCycleReset
            ? videoBoxes.map((_: any, i: number) => i)
            : available;

        const targetIdx = pool[Math.floor(Math.random() * pool.length)];
        // Mark this index as used
        usedIndicesRef.current.add(targetIdx);

        // Helper: map video coords → canvas display coords
        const toDisplay = (b: { x: number; y: number; width: number; height: number }) => {
            const dW = video.offsetWidth || overlayCanvas.width || vW;
            const dH = video.offsetHeight || overlayCanvas.height || vH;
            return {
                x: (b.x / vW) * dW,
                y: (b.y / vH) * dH,
                width: (b.width / vW) * dW,
                height: (b.height / vH) * dH,
            };
        };

        // Slot-machine animation
        let iter = 0;
        let currentHighlight = 0;
        const ctx = overlayCanvas.getContext("2d");

        const animate = () => {
            if (!ctx) { setState("result"); return; }
            iter++;

            const dW = video.offsetWidth || overlayCanvas.width || vW;
            const dH = video.offsetHeight || overlayCanvas.height || vH;
            // Sync canvas size if needed
            if (overlayCanvas.width !== dW) overlayCanvas.width = dW;
            if (overlayCanvas.height !== dH) overlayCanvas.height = dH;

            ctx.clearRect(0, 0, dW, dH);

            videoBoxes.forEach((_b: any, i: number) => {
                const db = toDisplay(_b);
                const isHighlighted = i === currentHighlight;
                ctx.strokeStyle = isHighlighted ? "#f59e0b" : "rgba(139,92,246,0.5)";
                ctx.lineWidth = isHighlighted ? 5 : 2;
                ctx.shadowColor = isHighlighted ? "#f59e0b" : "transparent";
                ctx.shadowBlur = isHighlighted ? 20 : 0;
                ctx.strokeRect(db.x, db.y, db.width, db.height);
                ctx.shadowBlur = 0;
            });

            const speedFactor = iter < animIterations - 5 ? 1 : 1 + (iter - (animIterations - 5)) * 0.4;
            const delay = animSpeed * speedFactor;

            if (iter < animIterations) {
                currentHighlight = iter > animIterations - (videoBoxes.length * 2)
                    ? targetIdx
                    : (currentHighlight + 1) % videoBoxes.length;
                animFrameRef.current = setTimeout(animate, delay) as unknown as number;
            } else {
                // ── Animation done: draw winner & capture result ──
                const wb = videoBoxes[targetIdx];
                const dWb = toDisplay(wb);
                ctx.clearRect(0, 0, dW, dH);
                ctx.strokeStyle = "#10b981";
                ctx.lineWidth = 6;
                ctx.shadowColor = "#10b981";
                ctx.shadowBlur = 25;
                ctx.strokeRect(dWb.x, dWb.y, dWb.width, dWb.height);
                ctx.shadowBlur = 0;

                // Create offscreen canvas — resultCanvasRef is null until isResult=true
                const margin = Math.round(Math.max(wb.width, wb.height) * 0.3 + 20);
                const sx = Math.max(0, Math.round(wb.x - margin));
                const sy = Math.max(0, Math.round(wb.y - margin));
                const sw = Math.min(vW - sx, Math.round(wb.width + margin * 2));
                const sh = Math.min(vH - sy, Math.round(wb.height + margin * 2));

                try {
                    const offscreen = document.createElement("canvas");
                    offscreen.width = Math.max(1, sw * zoomLevel);
                    offscreen.height = Math.max(1, sh * zoomLevel);
                    const rCtx = offscreen.getContext("2d");
                    if (rCtx && sw > 0 && sh > 0) {
                        rCtx.drawImage(video, sx, sy, sw, sh, 0, 0, offscreen.width, offscreen.height);
                        rCtx.strokeStyle = "#10b981";
                        rCtx.lineWidth = 8;
                        rCtx.strokeRect(4, 4, offscreen.width - 8, offscreen.height - 8);
                        setResultImageUrl(offscreen.toDataURL("image/jpeg", 0.95));
                    }
                } catch (e) {
                    console.warn("Result capture failed:", e);
                }

                // Check if this pick completed a full cycle (all faces used)
                const allUsed = videoBoxes.every((_: any, i: number) =>
                    usedIndicesRef.current.has(i)
                );
                setCycleComplete(allUsed);

                setState("result");
                setShowConfetti(true);
                setTimeout(() => setShowConfetti(false), 3000);


            }
        };

        animFrameRef.current = setTimeout(animate, animSpeed) as unknown as number;
    }, [animSpeed, animIterations, zoomLevel]);


    // ─── Restart ─────────────────────────────────────────────────────────────
    const handleRestart = useCallback(async () => {
        if (animFrameRef.current) clearTimeout(animFrameRef.current);
        setShowConfetti(false);
        setCycleComplete(false);
        if (!streamRef.current) {
            try {
                const stream = await navigator.mediaDevices.getUserMedia({
                    video: {
                        deviceId: selectedDeviceId ? { exact: selectedDeviceId } : undefined,
                        facingMode: selectedDeviceId ? undefined : "user"
                    },
                    audio: false
                });
                streamRef.current = stream;
                if (videoRef.current) { videoRef.current.srcObject = stream; await videoRef.current.play(); }
            } catch { }
        }
        setState("camera");
    }, [selectedDeviceId]);

    const handleSwitchCamera = useCallback(() => {
        if (videoDevices.length < 2) return;
        const currentIdx = videoDevices.findIndex(d => d.deviceId === selectedDeviceId);
        const nextIdx = (currentIdx + 1) % videoDevices.length;
        const nextDevice = videoDevices[nextIdx];

        stopAll();
        setSelectedDeviceId(nextDevice.deviceId);
        setState("loading");
    }, [videoDevices, selectedDeviceId]);

    // ─── Close ───────────────────────────────────────────────────────────────
    const handleClose = () => {
        stopAll();
        usedIndicesRef.current.clear(); // reset no-repeat on modal close
        setState("loading");
        onClose();
    };

    if (!isOpen) return null;

    const isSelecting = state === "selecting";
    const isResult = state === "result";
    const isCamera = state === "camera";
    const isLoading = state === "loading";

    return (
        <AnimatePresence>
            {isOpen && (
                <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4"
                    onClick={(e) => { if (e.target === e.currentTarget) handleClose(); }}
                >
                    <motion.div
                        initial={{ scale: 0.92, opacity: 0, y: 20 }}
                        animate={{ scale: 1, opacity: 1, y: 0 }}
                        exit={{ scale: 0.92, opacity: 0, y: 20 }}
                        transition={{ type: "spring", damping: 25, stiffness: 300 }}
                        className="relative w-full max-w-4xl bg-zinc-900/95 border border-violet-500/30 rounded-3xl shadow-2xl overflow-hidden"
                    >
                        {/* Glow effect */}
                        <div className="absolute inset-0 bg-gradient-to-br from-violet-900/20 to-indigo-900/20 pointer-events-none" />

                        {/* Confetti */}
                        {showConfetti && <ConfettiEffect />}

                        {/* Header */}
                        <div className="relative flex items-center justify-between px-6 py-4 border-b border-white/10">
                            <div className="flex items-center gap-3">
                                <div className="p-2 rounded-xl bg-violet-600/20 border border-violet-500/30">
                                    <Camera className="w-5 h-5 text-violet-400" />
                                </div>
                                <div>
                                    <h2 className="font-bold text-white text-lg">{t("face_picker_title")}</h2>
                                    <p className="text-xs text-white/40">
                                        {isLoading && t("face_picker_loading")}
                                        {isCamera && (faceCount > 0
                                            ? t("face_picker_faces_found", { count: faceCount })
                                            : t("face_picker_no_faces"))}
                                        {isSelecting && "⚡ Іздеуде..."}
                                        {isResult && "✅ Таңдалды!"}
                                    </p>
                                </div>
                            </div>
                            <button onClick={handleClose} className="p-2 rounded-xl hover:bg-white/10 text-white/60 hover:text-white transition-all">
                                <X className="w-5 h-5" />
                            </button>
                        </div>

                        {/* Body */}
                        <div className="relative p-6 space-y-4">
                            {/* Loading state */}
                            {isLoading && !error && (
                                <div className="flex flex-col items-center justify-center py-16 gap-4">
                                    <div className="relative">
                                        <div className="w-16 h-16 rounded-full border-4 border-violet-500/30 border-t-violet-500 animate-spin" />
                                        <Camera className="absolute inset-0 m-auto w-6 h-6 text-violet-400" />
                                    </div>
                                    <p className="text-white/60 text-sm">{t("face_picker_loading")}</p>
                                </div>
                            )}

                            {/* Error state */}
                            {error && (
                                <div className="flex flex-col items-center justify-center py-12 gap-3">
                                    <div className="p-4 rounded-2xl bg-red-500/10 border border-red-500/30">
                                        <Camera className="w-8 h-8 text-red-400" />
                                    </div>
                                    <p className="text-red-400 text-sm text-center">{error}</p>
                                    <button onClick={handleRestart} className="px-4 py-2 rounded-xl bg-red-500/20 text-red-300 hover:bg-red-500/30 transition text-sm">
                                        Қайталау
                                    </button>
                                </div>
                            )}

                            {/* Camera View */}
                            {!error && (isCamera || isSelecting) && (
                                <div className="relative rounded-2xl overflow-hidden bg-black shadow-inner" style={{ aspectRatio: "16/9" }}>
                                    <video
                                        ref={videoRef}
                                        autoPlay
                                        muted
                                        playsInline
                                        className="w-full h-full object-cover"
                                        style={{ transform: "scaleX(-1)" }}
                                    />

                                    <canvas
                                        ref={overlayRef}
                                        className="absolute inset-0 w-full h-full"
                                        style={{ transform: "scaleX(-1)" }}
                                    />
                                    {/* Face count badge */}
                                    {isCamera && (
                                        <div className={`absolute top-3 right-3 px-3 py-1 rounded-full text-xs font-bold backdrop-blur-sm border transition-all ${faceCount > 0
                                            ? "bg-violet-600/80 text-white border-violet-400/50"
                                            : "bg-zinc-800/80 text-zinc-400 border-zinc-600/50"
                                            }`}>
                                            {faceCount > 0 ? `${faceCount} 👤` : "0 👤"}
                                        </div>
                                    )}
                                    {/* Selecting overlay */}
                                    {isSelecting && (
                                        <div className="absolute inset-0 flex items-center justify-center">
                                            <div className="px-6 py-3 rounded-2xl bg-black/60 backdrop-blur border border-amber-400/50 text-amber-300 font-bold text-lg animate-pulse">
                                                ⚡ Іздеуде...
                                            </div>
                                        </div>
                                    )}
                                </div>
                            )}

                            {/* Result View */}
                            {isResult && (
                                <div className="flex flex-col items-center gap-4">
                                    <motion.div
                                        initial={{ scale: 0.5, opacity: 0 }}
                                        animate={{ scale: 1, opacity: 1 }}
                                        transition={{ type: "spring", damping: 15, stiffness: 300 }}
                                        className="relative"
                                    >
                                        <div className="p-1.5 rounded-2xl bg-gradient-to-br from-emerald-400 to-teal-500 shadow-lg shadow-emerald-500/30">
                                            {resultImageUrl
                                                ? <img
                                                    src={resultImageUrl}
                                                    alt="selected face"
                                                    className="rounded-xl block"
                                                    style={{ maxWidth: "100%", maxHeight: "400px", objectFit: "contain" }}
                                                />
                                                : <div className="w-48 h-48 rounded-xl bg-zinc-800 flex items-center justify-center text-zinc-500">?</div>
                                            }
                                        </div>
                                        <div className="absolute -top-3 -right-3 px-3 py-1 bg-emerald-500 rounded-full text-white text-xs font-bold shadow-lg shadow-emerald-500/40">
                                            ✓ Таңдалды!
                                        </div>
                                    </motion.div>

                                    {/* Cycle complete banner */}
                                    {cycleComplete && (
                                        <motion.div
                                            initial={{ opacity: 0, y: 8 }}
                                            animate={{ opacity: 1, y: 0 }}
                                            className="px-4 py-2 rounded-xl bg-amber-500/20 border border-amber-400/40 text-amber-300 text-sm font-semibold text-center"
                                        >
                                            {t("face_picker_cycle_complete")}
                                        </motion.div>

                                    )}
                                </div>
                            )}


                            {/* Controls */}
                            {!isLoading && !error && (
                                <div className="flex items-center gap-3">
                                    {/* Pick button */}
                                    {isCamera && (
                                        <button
                                            onClick={handlePick}
                                            disabled={faceCount === 0}
                                            className="flex-1 flex items-center justify-center gap-2 py-3 rounded-xl bg-gradient-to-r from-violet-600 to-indigo-600 text-white font-bold shadow-lg shadow-violet-500/30 hover:shadow-violet-500/50 hover:scale-[1.02] transition-all disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:scale-100"
                                        >
                                            <Shuffle className="w-5 h-5" />
                                            {t("face_picker_start")}
                                        </button>
                                    )}

                                    {/* Pick Again */}
                                    {isResult && (
                                        <>
                                            <button
                                                onClick={handleRestart}
                                                className="flex-1 flex items-center justify-center gap-2 py-3 rounded-xl bg-violet-600 text-white font-bold hover:bg-violet-700 transition"
                                            >
                                                <RotateCcw className="w-5 h-5" />
                                                {t("face_picker_again")}
                                            </button>
                                            <button
                                                onClick={handleClose}
                                                className="px-5 py-3 rounded-xl border border-white/10 text-white/70 hover:bg-white/5 transition text-sm"
                                            >
                                                Жабу
                                            </button>
                                        </>
                                    )}

                                    {/* Switch Camera button */}
                                    {isCamera && videoDevices.length > 1 && (
                                        <button
                                            onClick={handleSwitchCamera}
                                            className="p-3 rounded-xl border border-white/10 text-white/50 hover:text-white hover:bg-white/5 transition"
                                            title={t("face_picker_switch_camera")}
                                        >
                                            <RefreshCcw className="w-5 h-5" />
                                        </button>
                                    )}

                                    {/* Settings toggle */}
                                    {isCamera && (
                                        <button
                                            onClick={() => setShowSettings(v => !v)}
                                            className="p-3 rounded-xl border border-white/10 text-white/50 hover:text-white hover:bg-white/5 transition"
                                            title="Баптаулар"
                                        >
                                            <Settings2 className="w-5 h-5" />
                                        </button>
                                    )}
                                </div>
                            )}

                            {/* Settings Panel */}
                            <AnimatePresence>
                                {showSettings && isCamera && (
                                    <motion.div
                                        initial={{ opacity: 0, height: 0 }}
                                        animate={{ opacity: 1, height: "auto" }}
                                        exit={{ opacity: 0, height: 0 }}
                                        className="overflow-hidden"
                                    >
                                        <div className="pt-2 space-y-3 border-t border-white/10">
                                            <p className="text-xs text-white/40 font-medium uppercase tracking-widest">Баптаулар</p>

                                            <SettingSlider
                                                label="Анимация жылдамдығы"
                                                value={animSpeed}
                                                min={40} max={200} step={10}
                                                display={`${animSpeed}мс`}
                                                onChange={setAnimSpeed}
                                            />
                                            <SettingSlider
                                                label="Анимация итерациясы"
                                                value={animIterations}
                                                min={10} max={50} step={5}
                                                display={`${animIterations}`}
                                                onChange={setAnimIterations}
                                            />
                                            <SettingSlider
                                                label="Масштаб"
                                                value={zoomLevel}
                                                min={1.5} max={3} step={0.1}
                                                display={`${zoomLevel.toFixed(1)}x`}
                                                onChange={setZoomLevel}
                                            />

                                            <div className="flex items-center justify-between gap-3 pt-2">
                                                <span className="text-xs text-white/50">{t("face_picker_high_accuracy")}</span>
                                                <button
                                                    onClick={() => setHighAccuracy(!highAccuracy)}
                                                    className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none ${highAccuracy ? "bg-violet-600" : "bg-zinc-700"}`}
                                                >
                                                    <span
                                                        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${highAccuracy ? "translate-x-6" : "translate-x-1"}`}
                                                    />
                                                </button>
                                            </div>
                                        </div>
                                    </motion.div>
                                )}
                            </AnimatePresence>
                        </div>
                    </motion.div>
                </motion.div>
            )}
        </AnimatePresence>
    );
}

// ─── Helper Components ────────────────────────────────────────────────────────

function SettingSlider({ label, value, min, max, step, display, onChange }: {
    label: string; value: number; min: number; max: number; step: number;
    display: string; onChange: (v: number) => void;
}) {
    return (
        <div className="flex items-center gap-3">
            <span className="text-xs text-white/50 w-36 flex-shrink-0">{label}</span>
            <input
                type="range" min={min} max={max} step={step} value={value}
                onChange={e => onChange(Number(e.target.value))}
                className="flex-1 accent-violet-500"
            />
            <span className="text-xs text-violet-300 w-12 text-right font-mono">{display}</span>
        </div>
    );
}

function ConfettiEffect() {
    const colors = ["#8b5cf6", "#10b981", "#f59e0b", "#ef4444", "#3b82f6", "#ec4899"];
    return (
        <div className="absolute inset-0 pointer-events-none overflow-hidden z-10">
            {Array.from({ length: 40 }).map((_, i) => (
                <div
                    key={i}
                    style={{
                        position: "absolute",
                        left: `${Math.random() * 100}%`,
                        top: "-10px",
                        width: `${6 + Math.random() * 8}px`,
                        height: `${6 + Math.random() * 8}px`,
                        backgroundColor: colors[Math.floor(Math.random() * colors.length)],
                        borderRadius: Math.random() > 0.5 ? "50%" : "2px",
                        animation: `confettiFall ${1.5 + Math.random() * 2}s ease-in forwards`,
                        animationDelay: `${Math.random() * 0.8}s`,
                        transform: `rotate(${Math.random() * 360}deg)`,
                    }}
                />
            ))}
            <style>{`
                @keyframes confettiFall {
                    0% { transform: translateY(0) rotate(0deg); opacity: 1; }
                    100% { transform: translateY(600px) rotate(720deg); opacity: 0; }
                }
            `}</style>
        </div>
    );
}
