'use client';

import * as React from 'react';
import { useLocale } from 'next-intl';
import { usePathname, useRouter } from '@/i18n/routing';
import { motion, AnimatePresence } from 'framer-motion';
import { Globe, Check } from 'lucide-react';


const languages = [
    { code: 'en', label: 'English', short: 'EN' },
    { code: 'ru', label: 'Русский', short: 'RU' },
    { code: 'kk', label: 'Қазақша', short: 'KK' }
];

export default function LanguageSwitcher() {
    const locale = useLocale();
    const router = useRouter();
    const pathname = usePathname();
    const [isOpen, setIsOpen] = React.useState(false);
    const dropdownRef = React.useRef<HTMLDivElement>(null);

    const toggleDropdown = () => setIsOpen(!isOpen);

    const switchLanguage = (nextLocale: string) => {
        if (typeof window !== 'undefined') {
            // 1. Store in localStorage as fallback
            localStorage.setItem('NEXT_LOCALE', nextLocale);

            // 2. Set the cookie with proper flags for modern browsers (including mobile)
            const expires = new Date();
            expires.setFullYear(expires.getFullYear() + 1);

            // next-intl expects NEXT_LOCALE cookie
            const secure = window.location.protocol === 'https:' ? '; Secure' : '';
            document.cookie = `NEXT_LOCALE=${nextLocale}; expires=${expires.toUTCString()}; path=/; SameSite=Lax${secure}`;

            // 3. Force a reload to let the server-side middleware pick up the new locale
            window.location.reload();
        }
    };

    // Close dropdown when clicking outside
    React.useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const currentLang = languages.find(l => l.code === locale) || languages[0];

    return (
        <div className="relative" ref={dropdownRef}>
            <button
                onClick={toggleDropdown}
                className="flex items-center gap-2 px-3 py-2 text-sm font-medium transition-all rounded-xl hover:bg-slate-100 dark:hover:bg-white/5 border border-transparent hover:border-slate-200 dark:hover:border-white/10 active:scale-95"
            >
                <Globe className="w-4 h-4 text-foreground/80" />
                <span>{currentLang.short}</span>
            </button>

            <AnimatePresence>
                {isOpen && (
                    <motion.div
                        initial={{ opacity: 0, y: 10, scale: 0.95 }}
                        animate={{ opacity: 1, y: 0, scale: 1 }}
                        exit={{ opacity: 0, y: 10, scale: 0.95 }}
                        transition={{ duration: 0.2 }}
                        className="absolute right-0 mt-2 w-40 bg-white/80 dark:bg-slate-900/80 backdrop-blur-md rounded-xl border border-slate-200 dark:border-white/10 overflow-hidden z-50 shadow-xl"
                    >
                        <div className="py-1">
                            {languages.map((lang) => (
                                <button
                                    key={lang.code}
                                    onClick={() => switchLanguage(lang.code)}
                                    className={`w-full text-left px-4 py-2.5 text-sm flex items-center justify-between transition-colors ${locale === lang.code
                                        ? 'bg-primary/10 text-primary font-medium'
                                        : 'hover:bg-slate-100 dark:hover:bg-white/5 text-slate-600 dark:text-slate-300'
                                        }`}
                                >
                                    <span>{lang.label}</span>
                                    {locale === lang.code && (
                                        <Check className="w-3.5 h-3.5" />
                                    )}
                                </button>
                            ))}
                        </div>
                    </motion.div>
                )}
            </AnimatePresence>
        </div>
    );
}
