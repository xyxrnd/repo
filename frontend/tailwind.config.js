/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{svelte,js,ts,jsx,tsx}",
    ],
    darkMode: "class",
    theme: {
        extend: {
            colors: {
                "primary": "#1173d4",
                "background-light": "#f6f7f8",
                "background-dark": "#101922",
                "surface-dark": "#192633",
                "surface-highlight": "#233648",
            },
            fontFamily: {
                "display": ["Lexend", "sans-serif"],
                "body": ["Lexend", "sans-serif"],
            },
            borderRadius: {
                "DEFAULT": "0.25rem",
                "lg": "0.5rem",
                "xl": "0.75rem",
                "full": "9999px",
            },
        },
    },
    plugins: [
        require("@tailwindcss/forms"),
        require("@tailwindcss/container-queries"),
    ],
};
