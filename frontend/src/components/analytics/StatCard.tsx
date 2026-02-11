import React from 'react';
import { Card } from '../ui/Card';

export const StatCard: React.FC<{
  icon: any;
  label: string;
  value: number;
  suffix?: string;
  hint?: string;
  className?: string;
}> = ({ icon: Icon, label, value, suffix, hint, className }) => (
  <Card className="flex flex-col items-start p-5 bg-surface border-border shadow-none transition-colors hover:bg-surfaceHighlight/50">
    <div className={`mb-3 flex h-8 w-8 items-center justify-center rounded bg-surfaceHighlight ${className}`}>
      <Icon size={16} className="text-primary" />
    </div>
    <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">{label}</span>
    <div className="mt-1 flex items-baseline gap-1">
      <span className="text-2xl font-bold tracking-tight text-primary font-mono">{value}</span>
      {suffix && <span className="text-xs text-muted-foreground font-mono">{suffix}</span>}
    </div>
    {hint && <p className="mt-2 text-[11px] leading-snug text-muted-foreground">{hint}</p>}
  </Card>
);
