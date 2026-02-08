import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { api } from '../lib/api';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { Activity, ArrowRight } from 'lucide-react';

export const LoginPage: React.FC = () => {
  const [email, setEmail] = useState('');
  const [name, setName] = useState('');
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const { user, token } = await api.auth.devLogin(email, name);
      login(user, token);
      navigate('/');
    } catch (error) {
      console.error('Login failed', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4 font-sans text-primary">
      <Card className="w-full max-w-sm border-border bg-surface p-8 shadow-none">
        <div className="mb-8 flex flex-col items-start">
          <div className="mb-6 flex h-10 w-10 items-center justify-center rounded bg-accent text-accent-foreground">
            <Activity size={20} />
          </div>
          <h1 className="text-xl font-bold tracking-tight font-mono uppercase">MathalamaFocus</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Enter the zone.
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div className="space-y-1.5">
            <label htmlFor="email" className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              Email
            </label>
            <Input
              id="email"
              type="email"
              placeholder="user@mathalama.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="bg-background border-border text-primary placeholder:text-muted/20"
              required
              autoFocus
            />
          </div>

          <div className="space-y-1.5">
            <label htmlFor="name" className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              Name
            </label>
            <Input
              id="name"
              type="text"
              placeholder="Your name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="bg-background border-border text-primary placeholder:text-muted/20"
              required
            />
          </div>

          <Button 
            type="submit" 
            className="w-full justify-between bg-accent text-accent-foreground hover:bg-accent/90 mt-2" 
            size="md" 
            disabled={loading}
          >
            <span>{loading ? 'Authenticating...' : 'Initialize Session'}</span>
            {!loading && <ArrowRight size={16} />}
          </Button>
        </form>
      </Card>
    </div>
  );
};