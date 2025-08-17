/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./web/**/*.templ"],
  theme: {
    container: {
      center: true,
      padding: '1rem'
    }
  },
  plugins: [
    require('@tailwindcss/typography'),
  ],
}

