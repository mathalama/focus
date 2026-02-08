import React from 'react';
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { LayoutDashboard, BarChart3, LogOut, Trophy, Activity } from 'lucide-react';
import { cn } from './ui/Button';

export const Layout: React.FC = () => {
  const { user, logout } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();

  const navItems = [
    { path: '/', label: 'Focus', icon: LayoutDashboard },
    { path: '/hive', label: 'Leaderboard', icon: Trophy }, // Renamed from "The Hive"
    { path: '/analytics', label: 'Analytics', icon: BarChart3 },
  ];

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-background text-primary selection:bg-accent/20 font-sans">
      {/* Top Navigation */}
      <header className="fixed top-0 left-0 right-0 z-50 h-14 border-b border-border bg-background/90 backdrop-blur-md">
        <div className="mx-auto flex h-full max-w-6xl items-center justify-between px-4 sm:px-6">
          <div className="flex items-center gap-8">
            <Link to="/" className="flex items-center gap-2 group">
              <div className="flex h-6 w-6 items-center justify-center rounded bg-accent text-accent-foreground">
                <Activity size={16} />
              </div>
              <span className="text-sm font-bold tracking-wide uppercase font-mono group-hover:text-muted-foreground transition-colors">MathalamaFocus</span>
            </Link>

            <nav className="hidden md:flex items-center gap-1">
              {navItems.map((item) => {
                const Icon = item.icon;
                const isActive = location.pathname === item.path;
                return (
                  <Link
                    key={item.path}
                    to={item.path}
                    className={cn(
                      'flex items-center gap-2 rounded-md px-3 py-1.5 text-xs font-medium transition-all duration-200',
                      isActive
                        ? 'bg-surfaceHighlight text-primary'
                        : 'text-muted-foreground hover:bg-surfaceHighlight/50 hover:text-primary'
                    )}
                  >
                    <Icon size={14} />
                    {item.label}
                  </Link>
                );
              })}
            </nav>
          </div>

          <div className="flex items-center gap-4">
            {user && (
              <div className="flex items-center gap-2 px-3 py-1 text-xs font-mono text-muted-foreground">
                <span>{user.name.split(' ')[0]}</span>
                <span className="text-border">|</span>
                <span className="text-accent">{user.nectar_balance} pts</span>
              </div>
            )}
            
            <button
              onClick={handleLogout}
              className="text-muted-foreground hover:text-primary transition-colors"
              title="Logout"
            >
              <LogOut size={16} />
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="mx-auto max-w-4xl px-4 pt-24 pb-12 sm:px-6">
        <Outlet />
      </main>

      {/* Mobile Nav (Bottom) */}
      <nav className="fixed bottom-0 left-0 right-0 z-50 flex h-14 items-center justify-around border-t border-border bg-background px-4 md:hidden">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = location.pathname === item.path;
          return (
            <Link
              key={item.path}
              to={item.path}
              className={cn(
                'flex flex-col items-center justify-center gap-1 p-2 transition-colors',
                isActive
                  ? 'text-accent'
                  : 'text-muted-foreground hover:text-primary'
              )}
            >
              <Icon size={18} />
            </Link>
          );
        })}
      </nav>
    </div>
  );
};