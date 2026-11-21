'use client';

import * as React from 'react';
import { Moon, Sun } from 'lucide-react';
import { useTheme } from 'next-themes';
import { motion } from 'framer-motion';

export function ThemeToggle() {
    const { theme, setTheme, resolvedTheme } = useTheme();
    const [mounted, setMounted] = React.useState(false);

    React.useEffect(() => {
        setMounted(true);
    }, []);

    if (!mounted) {
        return <div className="w-10 h-10" />; // Placeholder to avoid hydration mismatch
    }

    const isDark = resolvedTheme === 'dark';

    return (
        <button
            onClick={() => setTheme(isDark ? 'light' : 'dark')}
            className={`relative w-14 h-8 rounded-full border p-1 transition-colors flex items-center ${isDark ? 'bg-white/10 border-white/20' : 'bg-slate-200 border-slate-300'}`}
            aria-label="Toggle theme"
        >
            <motion.div
                initial={false}
                animate={{
                    x: isDark ? 24 : 0,
                    backgroundColor: isDark ? '#8b5cf6' : '#f59e0b' // purple for dark, amber/orange for sun
                }}
                className="w-6 h-6 rounded-full flex items-center justify-center shadow-md relative"
            >
                <div className="absolute inset-0 flex items-center justify-center">
                    <motion.div
                        initial={false}
                        animate={{ opacity: isDark ? 1 : 0, scale: isDark ? 1 : 0.5 }}
                        className="absolute"
                    >
                        <Moon className="w-4 h-4 text-white" />
                    </motion.div>
                    <motion.div
                        initial={false}
                        animate={{ opacity: isDark ? 0 : 1, scale: isDark ? 0 : 1 }}
                        className="absolute"
                    >
                        <Sun className="w-4 h-4 text-white" />
                    </motion.div>
                </div>
            </motion.div>
        </button>
    );
}
