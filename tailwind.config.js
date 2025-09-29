module.exports = {
  darkMode: 'class',
  content: [
    './templates/**/*.templ',
    './static/**/*.js',
    './static/**/*.html',
  ],
  theme: {
    extend: {
      colors: {
        // Dark mode colors
        dark: {
          background: '#121212',
          card: '#1e1e1e',
          accent: '#2d2d2d',
          border: '#333',
        },
        primary: {
          DEFAULT: '#bb86fc',
          dark: '#a370e0',
        },
        danger: {
          DEFAULT: '#cf6679',
          dark: '#b55c6a',
        },
      },
      typography: {
        DEFAULT: {
          css: {
            maxWidth: '65ch',
            color: '#e0e0e0',
            a: {
              color: '#bb86fc',
              '&:hover': {
                color: '#a370e0',
              },
            },
          },
        },
      },
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
    require('@tailwindcss/forms'),
    require('@tailwindcss/aspect-ratio'),
    require('@tailwindcss/line-clamp'),
  ],
};
