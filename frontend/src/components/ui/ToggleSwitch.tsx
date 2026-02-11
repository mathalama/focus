import React from 'react';
import { Check, X } from 'lucide-react';

interface ToggleSwitchProps {
  enabled: boolean;
  onChange: () => void;
  label?: string;
}

/**
 * Accessible toggle switch with proper contrast on all themes
 */
export const ToggleSwitch: React.FC<ToggleSwitchProps> = ({ enabled, onChange, label }) => {
  return (
    <div className="flex items-center gap-3">
      <button
        onClick={onChange}
        className={`relative inline-flex h-8 w-14 flex-shrink-0 items-center rounded-full transition-all duration-200 border-2 ${
          enabled
            ? 'border-accent bg-accent'
            : 'border-border bg-surfaceHighlight'
        }`}
        role="switch"
        aria-checked={enabled}
        aria-label={label}
      >
        {/* Animated circle background */}
        <span
          className={`inline-flex h-6 w-6 transform items-center justify-center rounded-full transition-transform duration-200 ${
            enabled
              ? 'translate-x-7 bg-accent-foreground text-accent'
              : 'translate-x-1 bg-muted-foreground/50 text-surface'
          }`}
        >
          {/* Icon inside circle */}
          {enabled ? (
            <Check size={14} strokeWidth={3} />
          ) : (
            <X size={14} strokeWidth={3} />
          )}
        </span>
      </button>
      {label && (
        <span className={`text-sm font-medium ${enabled ? 'text-primary' : 'text-muted-foreground'}`}>
          {label}
        </span>
      )}
    </div>
  );
};
