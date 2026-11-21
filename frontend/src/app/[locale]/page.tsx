'use client';

import React from 'react';
import { motion } from 'framer-motion';
import { Bot, Sparkles, GraduationCap, Users, ArrowRight, LayoutDashboard, MessageCircle, Instagram } from 'lucide-react';
import Navbar from '@/components/Navbar';
import Link from 'next/link';
import { useTranslations } from 'next-intl';

export default function LandingPage() {
  const t = useTranslations('landing');
  const tNav = useTranslations('nav');
  const tFeatures = useTranslations('features');
  const [isAuth, setIsAuth] = React.useState(false);
  const [role, setRole] = React.useState('');

  React.useEffect(() => {
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

  return (
    <main className="relative">
      <Navbar />

      {/* Hero Section */}
      <section className="relative pt-32 pb-20 overflow-hidden min-h-screen flex items-center">
        <div className="container mx-auto px-6 relative z-10">
          {/* Centered Content */}
          <div className="flex flex-col items-center text-center max-w-4xl mx-auto">
            <motion.div
              initial={{ opacity: 0, scale: 0.5 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.5 }}
              className="px-4 py-1.5 mb-8 rounded-full glass-card border-primary/20 flex items-center gap-2 text-sm font-medium"
            >
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-primary"></span>
              </span>
              <span className="text-primary italic">{t('redefining')}</span>
            </motion.div>

            <motion.h1
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
              className="text-5xl md:text-7xl lg:text-8xl font-bold tracking-tight mb-8 font-outfit"
            >
              {t('hero_subtitle_1')} <span className="text-gradient font-black">AI-Powered</span> {t('hero_subtitle_2')}
            </motion.h1>

            <motion.p
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.3 }}
              className="text-lg md:text-xl text-foreground/60 mb-10 leading-relaxed max-w-2xl"
            >
              {t('hero_desc')}
            </motion.p>

            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.4 }}
              className="flex flex-wrap items-center justify-center gap-6"
            >
              {isAuth ? (
                <Link href={role === 'admin' ? '/admin' : (role === 'teacher' ? '/teacher' : '/student')}>
                  <button className="px-10 py-5 bg-primary text-white rounded-2xl font-bold text-xl hover:bg-primary-hover transition-all flex items-center gap-3 shadow-2xl shadow-primary/40 active:scale-95 group">
                    <LayoutDashboard className="w-6 h-6 group-hover:rotate-12 transition-transform" />
                    {tNav('dashboard')}
                    <ArrowRight className="w-6 h-6 group-hover:translate-x-1 transition-transform" />
                  </button>
                </Link>
              ) : (
                <>
                  <Link href="/register">
                    <button className="px-10 py-5 bg-primary text-white rounded-2xl font-bold text-xl hover:bg-primary-hover transition-all flex items-center gap-3 shadow-2xl shadow-primary/40 active:scale-95 group">
                      {t('get_started')}
                      <ArrowRight className="w-6 h-6 group-hover:translate-x-1 transition-transform" />
                    </button>
                  </Link>
                  <Link href="/login">
                    <button className="px-10 py-5 glass-card rounded-2xl font-bold text-xl hover:bg-white/10 transition-all active:scale-95 border-white/10">
                      {t('watch_demo')}
                    </button>
                  </Link>
                </>
              )}
            </motion.div>
          </div>
        </div>

        {/* Abstract Background Shapes */}
        <div className="hidden md:block absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full h-full -z-10 opacity-30">
          <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-primary blur-[120px] rounded-full animate-pulse" />
          <div className="absolute bottom-1/4 right-1/4 w-[500px] h-[500px] bg-secondary blur-[160px] rounded-full delay-1000 animate-pulse" />
        </div>
      </section>

      {/* Features Preview */}
      <section id="features" className="py-24 container mx-auto px-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {[
            {
              icon: <Bot className="w-10 h-10 text-primary" />,
              title: tFeatures('ai_assistant_title'),
              desc: tFeatures('ai_assistant_desc')
            },
            {
              icon: <GraduationCap className="w-10 h-10 text-secondary" />,
              title: tFeatures('smart_content_title'),
              desc: tFeatures('smart_content_desc')
            },
            {
              icon: <Users className="w-10 h-10 text-accent" />,
              title: tFeatures('teacher_dashboard_title'),
              desc: tFeatures('teacher_dashboard_desc')
            }
          ].map((feature, i) => (
            <motion.div
              key={i}
              whileHover={{ y: -10 }}
              className="glass-card p-10 rounded-3xl"
            >
              <div className="mb-6">{feature.icon}</div>
              <h3 className="text-2xl font-bold mb-4 font-outfit">{feature.title}</h3>
              <p className="text-foreground/60 leading-relaxed">
                {feature.desc}
              </p>
            </motion.div>
          ))}
        </div>
      </section>

      {/* Contact Section */}
      <section className="py-20 container mx-auto px-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="glass-card rounded-3xl p-12 text-center max-w-2xl mx-auto"
        >
          <h2 className="text-3xl md:text-4xl font-bold font-outfit mb-4">Байланыс / Контакты</h2>
          <p className="text-foreground/60 mb-10 text-lg">Сұрақтарыңыз болса, хабарласыңыз!</p>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <a
              href="https://wa.me/77088710495"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-3 px-8 py-4 bg-[#25D366] text-white rounded-2xl font-bold text-lg hover:bg-[#1ebe5d] transition-all shadow-lg shadow-[#25D366]/30 active:scale-95 w-full sm:w-auto justify-center"
            >
              <MessageCircle className="w-6 h-6" />
              WhatsApp
            </a>
            <a
              href="https://www.instagram.com/aibek.orazbakhov/"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-3 px-8 py-4 rounded-2xl font-bold text-lg text-white transition-all active:scale-95 w-full sm:w-auto justify-center"
              style={{ background: 'linear-gradient(135deg, #f09433, #e6683c, #dc2743, #cc2366, #bc1888)', boxShadow: '0 10px 30px rgba(220,39,67,0.3)' }}
            >
              <Instagram className="w-6 h-6" />
              Instagram
            </a>
          </div>
        </motion.div>
      </section>
    </main>
  );
}
