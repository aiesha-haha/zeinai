import type { FieldDef } from './StudioTool';

/** Content language select shared by the AI tools */
export function languageField(t: (k: string) => string): FieldDef {
    return {
        name: 'language', label: t('language'), type: 'select', defaultValue: 'kaz',
        options: [
            { value: 'kaz', label: 'Қазақша' },
            { value: 'rus', label: 'Русский' },
            { value: 'eng', label: 'English' },
        ],
    };
}

export function gradeOptions(suffix: string) {
    return Array.from({ length: 11 }, (_, i) => ({ value: String(i + 1), label: `${i + 1} ${suffix}` }));
}
