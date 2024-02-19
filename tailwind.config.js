/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./{assets,views}/**/*.{html,js}"],
  theme: {
    extend: {},
  },
  plugins: [require("tailwind-fontawesome")],
  safelist: [
    {
      pattern: /icon-(person-running)/,
    },
    {
      pattern: /text-(green|rose)-500/,
    },
  ],
};
