// Test types stored in tests.test_type on the backend.
// "game_*" and "tug_of_war" are played by students as games; the rest are tests.
export const TEST_TYPES = ['standard', 'pisa', 'unt', 'olympiad', 'game_quiz', 'game_true_false', 'tug_of_war'] as const;
export type TestType = (typeof TEST_TYPES)[number];

export const GAME_TYPES: TestType[] = ['game_quiz', 'game_true_false', 'tug_of_war'];

export function isGameType(t?: string | null): boolean {
    return !!t && (GAME_TYPES as string[]).includes(t);
}

/** Translation key (namespace "test_types") for a type */
export function testTypeKey(t?: string | null): string {
    return (TEST_TYPES as readonly string[]).includes(t || '') ? (t as string) : 'standard';
}
