import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import { LanguageProvider } from './context/LanguageContext';
import { ErrorBoundary } from './components/ErrorBoundary';
import { Layout } from './components/Layout';
import { LoginPage } from './pages/LoginPage';
import { DashboardPage } from './pages/DashboardPage';
import { SessionPage } from './pages/SessionPage';
import { ReflectionPage } from './pages/ReflectionPage';
import { HivePage } from './pages/HivePage';
import { AnalyticsPage } from './pages/AnalyticsPage';
import { ProfilePage } from './pages/ProfilePage';

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated, authChecked } = useAuth();
  if (!authChecked) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background text-xs font-mono text-muted-foreground">
        INITIALIZING...
      </div>
    );
  }
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
};

// Placeholders removed

function App() {
  return (
    <LanguageProvider>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            
            <Route path="/" element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }>
              <Route index element={<DashboardPage />} />
              <Route path="hive" element={<HivePage />} />
              <Route path="analytics" element={<ErrorBoundary><AnalyticsPage /></ErrorBoundary>} />
              <Route path="profile" element={<ProfilePage />} />
              <Route path="session/:sessionId" element={<SessionPage />} />
              <Route path="session/:sessionId/reflection" element={<ReflectionPage />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </LanguageProvider>
  );
}

export default App;
