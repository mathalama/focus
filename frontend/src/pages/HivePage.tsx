import React, { useEffect, useState } from 'react';
import { api, LeaderboardEntry } from '../lib/api';
import { Card } from '../components/ui/Card';
import { useAuth } from '../context/AuthContext';
import { cn } from '../components/ui/Button';
import { Trophy } from 'lucide-react';

export const HivePage: React.FC = () => {
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const { user } = useAuth();

  useEffect(() => {
    api.gamification.leaderboard()
      .then(data => setEntries(data.leaderboard || []))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="text-xs font-mono text-muted-foreground">LOADING DATA...</div>;

  return (
    <div className="space-y-6">
      <header className="border-b border-border pb-4">
        <h1 className="text-2xl font-bold tracking-tight font-mono uppercase text-primary">Leaderboard</h1>
        <p className="text-sm text-muted-foreground font-mono mt-1">
          // TOP PERFORMERS
        </p>
      </header>

      <Card className="overflow-hidden border-border bg-surface p-0 shadow-none">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm font-mono">
            <thead className="bg-surfaceHighlight text-xs uppercase text-muted-foreground">
              <tr>
                <th className="px-6 py-3 font-medium tracking-wider">Rank</th>
                <th className="px-6 py-3 font-medium tracking-wider">User</th>
                <th className="px-6 py-3 font-medium tracking-wider text-right">Points</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {entries.map((entry) => (
                <tr 
                  key={entry.user_id} 
                  className={cn(
                    'transition-colors hover:bg-surfaceHighlight/50',
                    entry.user_id === user?.id && 'bg-surfaceHighlight/20'
                  )}
                >
                  <td className="px-6 py-4">
                    <div className={cn(
                      "flex h-6 w-6 items-center justify-center rounded text-xs font-bold",
                      entry.rank === 1 ? "text-yellow-500" :
                      entry.rank === 2 ? "text-gray-400" :
                      entry.rank === 3 ? "text-orange-500" :
                      "text-muted-foreground"
                    )}>
                      {entry.rank <= 3 && <Trophy size={12} className="mr-1" />}
                      #{entry.rank}
                    </div>
                  </td>
                  <td className="px-6 py-4 font-medium text-primary">
                    {entry.name}
                    {entry.user_id === user?.id && <span className="ml-2 text-[10px] text-accent uppercase tracking-wider">[YOU]</span>}
                  </td>
                  <td className="px-6 py-4 text-right font-bold text-accent">
                    {entry.total_nectar_earned.toLocaleString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
};