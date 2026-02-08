/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'], // Adding a mono font for a tech/math vibe
      },
      colors: {
        background: '#09090b', // Zinc 950 - Deep Black
        surface: '#18181b',    // Zinc 900 - Dark Surface
        surfaceHighlight: '#27272a', // Zinc 800
        primary: {
          DEFAULT: '#fafafa', // Zinc 50 - White Text
          foreground: '#09090b',
        },
        secondary: {
          DEFAULT: '#27272a', // Zinc 800
          foreground: '#fafafa',
        },
        accent: {
          DEFAULT: '#ffffff', // White accent for high contrast minimalist look
          foreground: '#000000',
        },
        muted: {
          DEFAULT: '#71717a', // Zinc 500
          foreground: '#a1a1aa', // Zinc 400
        },
        border: '#27272a', // Zinc 800
      },
      borderRadius: {
        'xl': '0.75rem', // Slightly sharper corners for a tech look
        '2xl': '1rem',
        '3xl': '1.5rem',
      },
      boxShadow: {
        'glow': '0 0 20px -5px rgba(255, 255, 255, 0.1)',
        'subtle': '0 1px 2px 0 rgba(0, 0, 0, 0.5)',
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
      }
    }
  },
  plugins: []
};
