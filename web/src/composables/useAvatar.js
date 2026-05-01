/**
 * Avatar utility composable.
 * Provides functions to generate text-based avatars matching daisyUI's avatar placeholder pattern:
 * - getInitials(name): extracts initials (e.g., "Inkedus" → "I", "Hi Go" → "HG")
 * - getNameColor(name): returns a deterministic DaisyUI semantic color pair based on the name
 *
 * Usage with daisyUI:
 *   <div class="avatar placeholder">
 *     <div :class="['w-8 rounded-full', getNameColor(name).bg, getNameColor(name).text]">
 *       <span class="text-xs">{{ getInitials(name) }}</span>
 *     </div>
 *   </div>
 */

// DaisyUI semantic color pairs: [background, text-color]
const COLOR_PAIRS = [
  ['bg-primary', 'text-primary-content'],
  ['bg-secondary', 'text-secondary-content'],
  ['bg-accent', 'text-accent-content'],
  ['bg-neutral', 'text-neutral-content'],
  ['bg-info', 'text-info-content'],
  ['bg-success', 'text-success-content'],
  ['bg-warning', 'text-warning-content'],
  ['bg-error', 'text-error-content'],
]

/**
 * Simple string hash function (djb2) for deterministic color selection.
 * Returns a positive integer hash for the given string.
 */
function hashString(str) {
  let hash = 5381
  for (let i = 0; i < str.length; i++) {
    hash = ((hash << 5) + hash) + str.charCodeAt(i)
    hash = hash & hash // Convert to 32-bit integer
  }
  return Math.abs(hash)
}

/**
 * Returns a deterministic pair of DaisyUI semantic color classes for a given name.
 * The same name always returns the same colors.
 *
 * @param {string} name - The user's name
 * @returns {{ bg: string, text: string }} Classes for background and text color
 *
 * @example
 * getNameColor('Inkedus') // { bg: 'bg-warning', text: 'text-warning-content' }
 */
export function getNameColor(name) {
  if (!name) {
    return { bg: 'bg-neutral', text: 'text-neutral-content' }
  }
  const hash = hashString(name.toLowerCase())
  const pair = COLOR_PAIRS[hash % COLOR_PAIRS.length]
  return {
    bg: pair[0],
    text: pair[1],
  }
}

/**
 * Extracts initials from a name.
 * For single-word names, returns the first letter.
 * For multi-word names, returns the first letter of the first word + first letter of the last word.
 *
 * @param {string} name - The user's name
 * @returns {string} Uppercase initials (1-2 characters)
 *
 * @example
 * getInitials('Inkedus')   // 'I'
 * getInitials('Hi Go')     // 'HG'
 * getInitials('John Doe')  // 'JD'
 */
export function getInitials(name) {
  if (!name) return '?'
  const words = name.trim().split(/\s+/)
  if (words.length === 1) {
    return words[0][0].toUpperCase()
  }
  return (words[0][0] + words[words.length - 1][0]).toUpperCase()
}
