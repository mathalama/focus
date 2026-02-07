/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        paper: '#F5F3EF',
        ink: '#1E293B',
        mint: '#83C5BE',
        sea: '#006D77',
        clay: '#B8A58B'
      },
      boxShadow: {
        calm: '0 12px 30px rgba(0, 77, 87, 0.15)'
      }
    }
  },
  plugins: []
};
