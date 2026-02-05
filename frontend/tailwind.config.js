/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: ['class', '[data-theme="dark"]'],
  theme: {
    extend: {
      colors: {
        base: 'rgb(var(--ds-bg) / <alpha-value>)',
        surface: 'rgb(var(--ds-surface) / <alpha-value>)',
        card: 'rgb(var(--ds-card) / <alpha-value>)',
        border: 'rgb(var(--ds-border) / <alpha-value>)',
        ink: 'rgb(var(--ds-text) / <alpha-value>)',
        muted: 'rgb(var(--ds-muted) / <alpha-value>)',
        subtle: 'rgb(var(--ds-subtle) / <alpha-value>)',
        primary: 'rgb(var(--ds-primary) / <alpha-value>)',
        'primary-strong': 'rgb(var(--ds-primary-strong) / <alpha-value>)',
        success: 'rgb(var(--ds-success) / <alpha-value>)',
        warning: 'rgb(var(--ds-warning) / <alpha-value>)',
        danger: 'rgb(var(--ds-danger) / <alpha-value>)',
        info: 'rgb(var(--ds-info) / <alpha-value>)',
      },
      boxShadow: {
        card: '0 1px 2px rgb(var(--ds-shadow) / 0.12), 0 10px 28px rgb(var(--ds-shadow) / 0.18)',
        'card-soft': '0 1px 2px rgb(var(--ds-shadow) / 0.08), 0 6px 18px rgb(var(--ds-shadow) / 0.12)',
        focus: '0 0 0 3px rgb(var(--ds-primary) / 0.3)',
      },
      borderRadius: {
        xl: '16px',
        '2xl': '20px',
        '3xl': '26px',
      },
      fontFamily: {
        sans: ['Inter', 'SF Pro Text', 'SF Pro Display', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'SFMono-Regular', 'Menlo', 'monospace'],
      },
    },
  },
  plugins: [],
}
