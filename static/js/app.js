const themes = {
    default: {
        '--color-1': '#26251c',
        '--color-2': '#51493d',
        '--color-3': '#7c6c5e',
        '--color-4': '#e8dab2',
        '--color-5': '#cd4c3b',
        '--color-6': '#934041',
        '--color-7': '#593447',
        '--color-8': '#6d4e66',
    },
    ayu_mirage: {
        '--color-1': '#232736',
        '--color-2': '#101521',
        '--color-3': '#32384a',
        '--color-4': '#e8e8e8',
        '--color-5': '#cd4c3b',
        '--color-6': '#a63434',
        '--color-7': '#7c3b3b',
        '--color-8': '#6d4e66',
    }
};

function setTheme(themeName) {
    if (!themes[themeName]) {
        console.error(`Theme "${themeName}" not found`);
        return;
    }

    const theme = themes[themeName];
    const root = document.documentElement;

    for (const [key, value] of Object.entries(theme)) {
        root.style.setProperty(key, value);
    }

    document.cookie = `theme=${themeName}; path=/; max-age=31536000; SameSite=Lax`;
}

function getTheme() {
    const match = document.cookie.match(/theme=([^;]+)/);
    return match ? match[1] : 'default';
}

function applySavedTheme() {
    const savedTheme = getTheme();
    if (themes[savedTheme]) {
        setTheme(savedTheme);
    } else {
        setTheme('default');
    }
}

document.addEventListener('DOMContentLoaded', applySavedTheme);
