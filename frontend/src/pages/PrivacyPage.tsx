import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '../components/ui/Button';
import { ArrowLeft } from 'lucide-react';

export const PrivacyPage: React.FC = () => {
  const navigate = useNavigate();

  return (
    <div className="flex min-h-screen flex-col bg-background p-6 font-sans text-primary">
      <div className="mx-auto w-full max-w-2xl">
        <header className="mb-8 flex items-center justify-between border-b border-border pb-4">
          <div className="flex items-center gap-4">
            <button 
              onClick={() => navigate(-1)}
              className="flex h-8 w-8 items-center justify-center rounded-lg border border-border bg-surface text-muted-foreground transition-colors hover:text-primary"
            >
              <ArrowLeft size={18} />
            </button>
            <h1 className="font-mono text-xl font-bold uppercase tracking-tight">Privacy Policy</h1>
          </div>
        </header>

        <div className="space-y-6 text-sm leading-relaxed text-muted-foreground">
          <section>
            <h2 className="mb-2 font-mono text-xs font-bold uppercase tracking-wider text-primary">1. Data Collection</h2>
            <p>
              Mathalama Focus collects minimal data necessary to provide its focus-tracking services. This includes your email address for authentication and your focus session history to provide analytics and productivity insights.
            </p>
          </section>

          <section>
            <h2 className="mb-2 font-mono text-xs font-bold uppercase tracking-wider text-primary">2. Data Usage</h2>
            <p>
              Your data is used solely to enhance your personal productivity experience. We do not sell, trade, or otherwise transfer your personally identifiable information to outside parties.
            </p>
          </section>

          <section>
            <h2 className="mb-2 font-mono text-xs font-bold uppercase tracking-wider text-primary">3. Data Security</h2>
            <p>
              We implement a variety of security measures to maintain the safety of your personal information. Your focus sessions and personal settings are stored securely in our database.
            </p>
          </section>

          <section>
            <h2 className="mb-2 font-mono text-xs font-bold uppercase tracking-wider text-primary">4. Cookies</h2>
            <p>
              We use local storage and essential cookies to remember your login session and application preferences (such as language and theme settings).
            </p>
          </section>

          <section>
            <h2 className="mb-2 font-mono text-xs font-bold uppercase tracking-wider text-primary">5. Consent</h2>
            <p>
              By using Mathalama Focus, you consent to our privacy policy and the collection of the aforementioned data.
            </p>
          </section>

          <div className="pt-8 text-center">
            <Button variant="outline" size="sm" onClick={() => navigate(-1)}>
              Go Back
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};
