/**
 * Leader password rules, matching the Go backend (backend/internal/auth/password.go): any
 * non-empty password up to 72 bytes (bcrypt's limit); leading/trailing spaces are trimmed.
 * Shown under every "new password" field so registration and reset give the same guidance.
 */
export const PASSWORD_MAX_LENGTH = 72

export const PASSWORD_GUIDANCE =
  'Use any password up to 72 characters. There are no other length or character rules, but spaces at the start or end are ignored.'
