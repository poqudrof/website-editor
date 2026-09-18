/** @type {import('tailwindcss').Config} */
// Les couleurs pointent vers les variables CSS définies dans le <style> de
// index.html (:root) : c'est là, et seulement là, qu'on change la palette.
module.exports = {
  content: ['./index.html'],
  theme: {
    extend: {
      colors: {
        primary: { DEFAULT: 'var(--primary)', dark: 'var(--primary-dark)' },
        secondary: { DEFAULT: 'var(--secondary)' },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['"JetBrains Mono"', 'monospace'],
      },
    },
  },
  // Ajoutées par le JS (menu mobile), donc absentes d'une partie du markup.
  safelist: ['hidden', 'flex'],
}
