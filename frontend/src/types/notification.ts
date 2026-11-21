export type NotificationType =
    | 'subscription_expiring'
    | 'subscription_expired'
    | 'token_limit_warning'
    | 'token_limit_reached'
    | 'student_limit_warning'
    | 'monthly_reset'
    | 'system'
    | 'material_ready'
    | 'whiteboard_started'
    | 'whiteboard_summary_ready';

export interface Notification {
    id: string;
    user_id: string;
    sender_id?: string;
    sender_name?: string;
    title: string;
    message: string;
    type: NotificationType;
    link?: string;
    is_read: boolean;
    created_at: string;
    read_at?: string;
}

export interface RegisterPushTokenRequest {
    token: string;
    platform: 'web' | 'ios' | 'android';
    device_name?: string;
}

export interface NotificationsResponse {
    notifications: Notification[];
    unread_count: number;
    total_count: number;
    limit: number;
    offset: number;
}

export interface SendNotificationRequest {
    student_id: string;
    title: string;
    message: string;
}

export interface SendClassNotificationRequest {
    class_id: number;
    title: string;
    message: string;
}

export interface SendAllStudentsNotificationRequest {
    title: string;
    message: string;
}
