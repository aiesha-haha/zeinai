'use client';

import { useState, useEffect, useCallback } from 'react';
import api from '@/lib/api';
import { Notification, NotificationsResponse } from '@/types/notification';

const LIMIT = 10;

export const useNotifications = () => {
    const [notifications, setNotifications] = useState<Notification[]>([]);
    const [unreadCount, setUnreadCount] = useState(0);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const [totalCount, setTotalCount] = useState(0);
    const [page, setPage] = useState(1);
    const [unreadOnly, setUnreadOnly] = useState(false);

    const fetchNotifications = useCallback(async (isSilent: boolean = false) => {
        if (typeof window === 'undefined') return;

        const hasToken = !!localStorage.getItem('access_token');
        if (!hasToken) return;

        if (!isSilent) setIsLoading(true);
        try {
            const offset = (page - 1) * LIMIT;
            const response = await api.get<any>(`/notifications?limit=${LIMIT}&offset=${offset}${unreadOnly ? '&unread=true' : ''}`);
            const data = response.data.data;
            setNotifications(data.notifications || []);
            setUnreadCount(data.unread_count || 0);
            setTotalCount(data.total_count || 0);
            setError(null);
        } catch (err: any) {
            if (err.response?.status !== 401) {
                setError(err.message || 'Failed to fetch notifications');
            }
        } finally {
            if (!isSilent) setIsLoading(false);
        }
    }, [page, unreadOnly]);

    const handlePageChange = useCallback((newPage: number) => {
        setPage(newPage);
    }, []);

    const handleFilterChange = useCallback((unread: boolean) => {
        setUnreadOnly(unread);
        setPage(1);
    }, []);

    const markAsRead = async (id: string) => {
        try {
            await api.put(`/notifications/${id}/read`);
            setNotifications((prev) =>
                prev.map((n) => (n.id === id ? { ...n, is_read: true } : n))
            );
            setUnreadCount((prev) => Math.max(0, prev - 1));
        } catch (err: any) {
            console.error('Failed to mark notification as read:', err);
        }
    };

    const markAllAsRead = async () => {
        try {
            await api.put('/notifications/read-all');
            setNotifications((prev) => prev.map((n) => ({ ...n, is_read: true })));
            setUnreadCount(0);
        } catch (err: any) {
            console.error('Failed to mark all notifications as read:', err);
        }
    };

    const deleteNotification = async (id: string) => {
        try {
            await api.delete(`/notifications/${id}`);
            setNotifications((prev) => prev.filter((n) => n.id !== id));
            const wasUnread = notifications.find(n => n.id === id)?.is_read === false;
            if (wasUnread) {
                setUnreadCount((prev) => Math.max(0, prev - 1));
            }
        } catch (err: any) {
            console.error('Failed to delete notification:', err);
        }
    };

    // Effect for manual changes (page, filter)
    useEffect(() => {
        fetchNotifications();
    }, [fetchNotifications]);

    // Separate effect for polling (silent)
    useEffect(() => {
        const interval = setInterval(() => {
            fetchNotifications(true);
        }, 30000);
        return () => clearInterval(interval);
    }, [fetchNotifications]);

    return {
        notifications,
        unreadCount,
        totalCount,
        page,
        unreadOnly,
        pageCount: Math.ceil(totalCount / LIMIT),
        setPage: handlePageChange,
        setUnreadOnly: handleFilterChange,
        isLoading,
        error,
        refresh: fetchNotifications,
        markAsRead,
        markAllAsRead,
        deleteNotification,
    };
};
