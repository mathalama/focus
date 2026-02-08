import React from 'react';
import { cn } from './Button';

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  hover?: boolean;
}

export const Card = React.forwardRef<HTMLDivElement, CardProps>(
  ({ className, hover = false, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(
          'rounded-2xl bg-surface border border-border/50 p-6 shadow-soft transition-all duration-300',
          hover && 'hover:-translate-y-1 hover:shadow-float cursor-pointer',
          className
        )}
        {...props}
      />
    );
  }
);
