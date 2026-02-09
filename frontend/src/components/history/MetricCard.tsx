import React from 'react';
import { Card } from '../ui/Card';

export const MetricCard: React.FC<{
  icon: any;
  label: string;
  value: number;
  suffix?: string;
}> = ({ icon: Icon, label, value, suffix }) => (
  <Card className="border-border bg-surface p-5 shadow-none">
    <div className="mb-2 flex h-8 w-8 items-center justify-center rounded bg-surfaceHighlight">
      <Icon size={15} className="text-primary" />
    </div>
    <p className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground">{label}</p>
    <p className="mt-1 text-2xl font-mono font-bold text-primary">
      {value}
      {suffix ? <span className="ml-1 text-xs text-muted-foreground">{suffix}</span> : null}
    </p>
  </Card>
);
