export const formatPhoneInput = (value: string) => {
    // Remove all non-digits
    let digits = value.replace(/\D/g, '');

    // Convert leading 8 to 7
    if (digits.startsWith('8')) {
        digits = '7' + digits.slice(1);
    }

    // If it has digits but doesn't start with 7, prepend 7
    if (digits.length > 0 && digits[0] !== '7') {
        digits = '7' + digits;
    }

    // Limit to 11 digits
    digits = digits.slice(0, 11);

    if (digits.length === 0) return '';

    // Apply mask: +7 XXX XXX XX XX
    let res = '+7';
    if (digits.length > 1) {
        res += ' ' + digits.slice(1, 4);
    }
    if (digits.length > 4) {
        res += ' ' + digits.slice(4, 7);
    }
    if (digits.length > 7) {
        res += ' ' + digits.slice(7, 9);
    }
    if (digits.length > 9) {
        res += ' ' + digits.slice(9, 11);
    }

    return res;
};

export const formatIINInput = (value: string) => {
    // Remove all non-digits and limit to 12
    return value.replace(/\D/g, '').slice(0, 12);
};

export const validatePassword = (password: string) => {
    const hasUpper = /[A-Z]/.test(password);
    const hasLower = /[a-z]/.test(password);
    const hasNumber = /[0-9]/.test(password);
    const isLongEnough = password.length >= 8;

    return {
        isValid: hasUpper && hasLower && hasNumber && isLongEnough,
        hasUpper,
        hasLower,
        hasNumber,
        isLongEnough
    };
};

/**
 * For a combined "email or phone" field: format as a phone number only while the
 * value looks like a phone (digits, +, spaces, brackets, dashes). Anything with
 * letters or "@" is kept exactly as typed so e-mail addresses can be entered.
 */
export const formatEmailOrPhoneInput = (value: string) => {
    if (value === '' || /^[\d+\s()\-]+$/.test(value)) {
        return formatPhoneInput(value);
    }
    return value;
};
