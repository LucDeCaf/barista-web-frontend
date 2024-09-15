/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ['templates/*.html'],
    theme: {
        extend: {
            colors: {
                current: 'currentColor',
                white: '#ffffff',
                black: '#000000',
                primary: '#E9C379',
                light: '#FFF5DD',
                dark: '#483722',
            },
            fontFamily: {
                sans: ['Laila', "sans-serif"],
            },
        },
        fontFamily: {
            cursive: ['"Kaushan Script"', 'cursive'],
        },
    },
    plugins: [],
};
