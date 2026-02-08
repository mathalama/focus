import React from 'react';
import { getInitialLanguage } from '../context/LanguageContext';
import { translate } from '../lib/i18n';

interface Props {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error('[ErrorBoundary]', error, info.componentStack);
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) return this.props.fallback;
      const language = getInitialLanguage();
      return (
        <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 text-center">
          <p className="text-sm font-mono uppercase text-red-500">{translate(language, 'error.render')}</p>
          <pre className="max-w-lg overflow-auto rounded border border-border bg-surface p-4 text-xs text-muted-foreground">
            {this.state.error?.message}
          </pre>
          <button
            onClick={() => this.setState({ hasError: false, error: null })}
            className="rounded bg-surfaceHighlight px-4 py-2 text-xs font-mono uppercase text-primary hover:bg-accent hover:text-accent-foreground transition-colors"
          >
            {translate(language, 'error.retry')}
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}
