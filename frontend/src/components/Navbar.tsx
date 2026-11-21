'use client';

import React, { useState, useEffect } from 'react';
import { Link } from '@/i18n/routing';
import { motion, AnimatePresence } from 'framer-motion';
import { LogIn, UserPlus, LayoutDashboard, LogOut, Menu, X, User } from 'lucide-react';
import { logout } from '@/lib/auth';
import { ThemeToggle } from '@/components/ThemeToggle';
import LanguageSwitcher from '@/components/LanguageSwitcher';
import { useTranslations } from 'next-intl';

export default function Navbar() {
    const tNav = useTranslations('nav');
    const tAuth = useTranslations('auth');
    const [isAuth, setIsAuth] = useState(false);
    const [role, setRole] = useState('');
    const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

    useEffect(() => {
        if (typeof window !== 'undefined') {
            const token = localStorage.getItem('access_token');
            const refreshToken = localStorage.getItem('refresh_token');
            const userRole = localStorage.getItem('user_role');
            if (token || refreshToken) {
                setIsAuth(true);
                setRole(userRole || '');
            }
        }
    }, []);

    const toggleMenu = () => setMobileMenuOpen(!mobileMenuOpen);
    const closeMenu = () => setMobileMenuOpen(false);

    // Links config
    const links = [
        { href: '#features', label: tNav('features') },
        { href: '#about', label: tNav('about') },
        { href: '#pricing', label: tNav('pricing') },
    ];

    return (
        <nav className="fixed top-0 left-0 right-0 z-50 flex justify-center p-4 md:p-6">
            <motion.div
                initial={{ y: -100, opacity: 0 }}
                animate={{ y: 0, opacity: 1 }}
                className="glass-card w-full max-w-6xl rounded-2xl relative z-50 backdrop-blur-xl"
            >
                <div className="flex items-center justify-between px-6 py-4">
                    {/* Logo */}
                    <Link href="/" className="text-2xl font-bold tracking-tighter text-gradient z-50 relative">
                        ZeinAI
                    </Link>

                    {/* Desktop Menu */}
                    <div className="hidden md:flex items-center gap-8">
                        {links.map(link => (
                            <Link key={link.href} href={link.href} className="text-sm font-medium hover:text-primary transition-colors">
                                {link.label}
                            </Link>
                        ))}
                    </div>

                    {/* Desktop Actions */}
                    <div className="hidden md:flex items-center gap-4">
                        <LanguageSwitcher />
                        <ThemeToggle />
                        {isAuth ? (
                            <>
                                <Link href={role === 'admin' ? '/admin' : (role === 'teacher' ? '/teacher' : '/student')}>
                                    <button className="flex items-center gap-2 px-5 py-2 text-sm font-medium bg-primary text-white hover:bg-primary-hover rounded-xl shadow-lg shadow-primary/20 transition-all active:scale-95">
                                        <LayoutDashboard className="w-4 h-4" />
                                        {tNav('dashboard')}
                                    </button>
                                </Link>
                                <button
                                    onClick={() => logout()}
                                    className="flex items-center gap-2 px-5 py-2 text-sm font-medium transition-all hover:bg-red-500/10 hover:text-red-500 rounded-xl active:scale-95"
                                >
                                    <LogOut className="w-4 h-4" />
                                    <span>{tAuth('logout')}</span>
                                </button>
                            </>
                        ) : (
                            <>
                                <Link href="/login">
                                    <button className="flex items-center gap-2 px-5 py-2 text-sm font-medium transition-all hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl active:scale-95">
                                        <LogIn className="w-4 h-4" />
                                        {tAuth('login_button')}
                                    </button>
                                </Link>
                                <Link href="/register">
                                    <button className="flex items-center gap-2 px-5 py-2 text-sm font-medium bg-primary text-white hover:bg-primary-hover rounded-xl shadow-lg shadow-primary/20 transition-all active:scale-95">
                                        <UserPlus className="w-4 h-4" />
                                        {tAuth('register_button')}
                                    </button>
                                </Link>
                            </>
                        )}
                    </div>

                    {/* Mobile Toggle (Visible on Mobile) */}
                    <div className="flex items-center gap-3 md:hidden">
                        <button
                            onClick={toggleMenu}
                            className="p-2 rounded-xl hover:bg-slate-100 dark:hover:bg-white/5 transition-colors active:scale-95"
                        >
                            {mobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
                        </button>
                    </div>
                </div>

                {/* Mobile Dropdown Menu */}
                <AnimatePresence>
                    {mobileMenuOpen && (
                        <motion.div
                            initial={{ height: 0, opacity: 0 }}
                            animate={{ height: 'auto', opacity: 1 }}
                            exit={{ height: 0, opacity: 0 }}
                            transition={{ duration: 0.3, ease: 'easeInOut' }}
                            className="overflow-hidden md:hidden border-t border-slate-200 dark:border-white/10"
                        >
                            <div className="p-6 space-y-4 flex flex-col">
                                {links.map(link => (
                                    <Link
                                        key={link.href}
                                        href={link.href}
                                        onClick={closeMenu}
                                        className="text-lg font-medium p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-colors text-center"
                                    >
                                        {link.label}
                                    </Link>
                                ))}

                                {/* Mobile Utilities */}
                                <div className="flex items-center justify-center gap-4 py-2">
                                    <LanguageSwitcher />
                                    <ThemeToggle />
                                </div>

                                <div className="h-px bg-slate-200 dark:bg-white/10 my-2" />

                                {isAuth ? (
                                    <div className="flex flex-col gap-3">
                                        <Link href={role === 'admin' ? '/admin' : (role === 'teacher' ? '/teacher' : '/student')} onClick={closeMenu}>
                                            <button className="w-full flex items-center justify-center gap-2 px-5 py-3 text-sm font-medium bg-primary text-white hover:bg-primary-hover rounded-xl shadow-lg shadow-primary/20">
                                                <LayoutDashboard className="w-4 h-4" />
                                                {tNav('dashboard')}
                                            </button>
                                        </Link>
                                        <button
                                            onClick={() => { logout(); closeMenu(); }}
                                            className="w-full flex items-center justify-center gap-2 px-5 py-3 text-sm font-medium transition-all bg-red-500/10 text-red-500 hover:bg-red-500/20 rounded-xl"
                                        >
                                            <LogOut className="w-4 h-4" />
                                            {tAuth('logout')}
                                        </button>
                                    </div>
                                ) : (
                                    <div className="flex flex-col gap-3">
                                        <Link href="/login" onClick={closeMenu}>
                                            <button className="w-full flex items-center justify-center gap-2 px-5 py-3 text-sm font-medium hover:bg-slate-100 dark:hover:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl">
                                                <LogIn className="w-4 h-4" />
                                                {tAuth('login_button')}
                                            </button>
                                        </Link>
                                        <Link href="/register" onClick={closeMenu}>
                                            <button className="w-full flex items-center justify-center gap-2 px-5 py-3 text-sm font-medium bg-primary text-white hover:bg-primary-hover rounded-xl shadow-lg shadow-primary/20">
                                                <UserPlus className="w-4 h-4" />
                                                {tAuth('register_button')}
                                            </button>
                                        </Link>
                                    </div>
                                )}
                            </div>
                        </motion.div>
                    )}
                </AnimatePresence>
            </motion.div>
        </nav>
    );
}
