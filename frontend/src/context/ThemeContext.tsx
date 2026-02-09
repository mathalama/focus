import React, { createContext, useContext, useEffect, useMemo, useState } from 'react';

export type AppTheme = 'midnight' | 'ivory' | 'forest' | 'ocean';

interface ThemeContextType {
  theme: AppTheme;
  setTheme: (theme: AppTheme) => void;
}

const THEME_STORAGE_KEY = 'app_theme';
const DEFAULT_THEME: AppTheme = 'midnight';

const isAppTheme = (value: string | null): value is AppTheme => {
  return value === 'midnight' || value === 'ivory' || value === 'forest' || value === 'ocean';
};

export const getInitialTheme = (): AppTheme => {
  const stored = localStorage.getItem(THEME_STORAGE_KEY);
  if (isAppTheme(stored)) {
    return stored;
  }
  return DEFAULT_THEME;
};

const applyThemeToDocument = (theme: AppTheme) => {
  document.documentElement.setAttribute('data-theme', theme);
};

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

export const ThemeProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [theme, setThemeState] = useState<AppTheme>(() => {
    const initialTheme = getInitialTheme();
    applyThemeToDocument(initialTheme);
    return initialTheme;
  });

  useEffect(() => {
    applyThemeToDocument(theme);
  }, [theme]);

  const setTheme = (nextTheme: AppTheme) => {
    localStorage.setItem(THEME_STORAGE_KEY, nextTheme);
    setThemeState(nextTheme);
  };

  const value = useMemo(() => ({ theme, setTheme }), [theme]);

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
};

export const useTheme = () => {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
};
