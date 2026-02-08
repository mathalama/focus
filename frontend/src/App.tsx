import { FormEvent, useEffect, useMemo, useState } from 'react';
import {
  addInterruption,
  completeSession,
  createGoal,
  devLogin,
  getOverview,
  listGoals,
  pauseSession,
  resumeSession,
  saveReflection,
  startSession
} from './lib/api';
import type { AnalyticsOverview, FocusSession, Goal, User } from './types';

const USER_KEY = 'mathalama_focus_user';

type StoredUser = Pick<User, 'id' | 'email' | 'name'> & { token: string };

function formatClock(seconds: number): string {
  const safe = Math.max(seconds, 0);
  const mins = Math.floor(safe / 60)
    .toString()
    .padStart(2, '0');
  const secs = (safe % 60).toString().padStart(2, '0');
  return `${mins}:${secs}`;
}

const emptyOverview: AnalyticsOverview = {
  completed_goals: 0,
  calm_score: 0,
  focus_stability: 0,
  best_hour_of_day_utc: -1
};

export default function App() {
  const [user, setUser] = useState<StoredUser | null>(null);
  const [goals, setGoals] = useState<Goal[]>([]);
  const [overview, setOverview] = useState<AnalyticsOverview>(emptyOverview);
  const [activeSession, setActiveSession] = useState<FocusSession | null>(null);
  const [remainingSeconds, setRemainingSeconds] = useState(0);

  const [email, setEmail] = useState('learner@mathalama.dev');
  const [name, setName] = useState('Mathalama Learner');

  const [topic, setTopic] = useState('');
  const [desiredResult, setDesiredResult] = useState('');
  const [modeMinutes, setModeMinutes] = useState(25);

  const [whatLearned, setWhatLearned] = useState('');
  const [whatWasHard, setWhatWasHard] = useState('');
  const [nextAction, setNextAction] = useState('continue');

  const [isLoading, setIsLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    const persisted = localStorage.getItem(USER_KEY);
    if (!persisted) {
      return;
    }

    try {
      setUser(JSON.parse(persisted) as StoredUser);
    } catch {
      localStorage.removeItem(USER_KEY);
    }
  }, []);

  useEffect(() => {
    if (!user) {
      return;
    }

    void loadDashboard(user.token);
  }, [user]);

  useEffect(() => {
    if (!activeSession || activeSession.status !== 'active') {
      return;
    }

    const interval = window.setInterval(() => {
      setRemainingSeconds((prev) => {
        if (prev <= 1) {
          window.clearInterval(interval);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => window.clearInterval(interval);
  }, [activeSession?.id, activeSession?.status]);

  const calmLabel = useMemo(() => {
    if (overview.calm_score >= 80) {
      return 'Stable';
    }
    if (overview.calm_score >= 55) {
      return 'Building';
    }
    return 'Recovering';
  }, [overview.calm_score]);

  async function loadDashboard(token: string) {
    try {
      const [goalList, summary] = await Promise.all([listGoals(token), getOverview(token)]);
      setGoals(goalList);
      setOverview(summary);
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function handleLogin(event: FormEvent) {
    event.preventDefault();
    setError('');
    setMessage('');
    setIsLoading(true);

    try {
      const { user: account, token } = await devLogin(email.trim(), name.trim());
      const persistedUser = { id: account.id, email: account.email, name: account.name, token };
      localStorage.setItem(USER_KEY, JSON.stringify(persistedUser));
      setUser(persistedUser);
      setMessage('Dev login ready. You can start focus sessions.');
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  }

  async function handleCreateGoal(event: FormEvent) {
    event.preventDefault();
    if (!user) return;

    setError('');
    setMessage('');
    setIsLoading(true);

    try {
      const goal = await createGoal(user.token, {
        topic: topic.trim(),
        desired_result: desiredResult.trim(),
        recommended_minutes: modeMinutes
      });
      setGoals((prev) => [goal, ...prev]);
      setTopic('');
      setDesiredResult('');
      setMessage('Goal added. Start when you are ready.');
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  }

  async function handleStartSession(goal: Goal) {
    if (!user) return;
    setError('');
    setMessage('');
    setIsLoading(true);

    try {
      const session = await startSession(user.token, {
        goal_id: goal.id,
        recommended_minutes: modeMinutes
      });
      setActiveSession(session);
      setRemainingSeconds(session.recommended_minutes * 60);
      setWhatLearned('');
      setWhatWasHard('');
      setNextAction('continue');
      setMessage('Focus session started.');
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  }

  async function handlePauseResume() {
    if (!user || !activeSession) return;

    setError('');
    setMessage('');
    setIsLoading(true);

    try {
      if (activeSession.status === 'active') {
        const session = await pauseSession(user.token, activeSession.id);
        setActiveSession(session);
        setMessage('Session paused.');
      } else if (activeSession.status === 'paused') {
        const session = await resumeSession(user.token, activeSession.id);
        setActiveSession(session);
        setMessage('Session resumed.');
      }
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  }

  async function handleInterruption() {
    if (!user || !activeSession) return;

    setError('');
    setMessage('');
    setIsLoading(true);

    try {
      await addInterruption(user.token, activeSession.id, 'context switch');
      setMessage('Interruption logged without resetting progress.');
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  }

  async function handleComplete() {
    if (!user || !activeSession) return;

    setError('');
    setMessage('');
    setIsLoading(true);

    try {
      const session = await completeSession(user.token, activeSession.id);
      setActiveSession(session);
      setMessage('Session completed. Reflection is required.');
      await loadDashboard(user.token);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  }

  async function handleReflection(event: FormEvent) {
    event.preventDefault();
    if (!user || !activeSession) return;

    setError('');
    setMessage('');
    setIsLoading(true);

    try {
      await saveReflection(user.token, activeSession.id, {
        what_learned: whatLearned.trim(),
        what_was_hard: whatWasHard.trim(),
        next_action: nextAction.trim()
      });
      setActiveSession(null);
      setRemainingSeconds(0);
      await loadDashboard(user.token);
      setMessage('Reflection saved. Session closed calmly.');
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(false);
    }
  }

  function logout() {
    localStorage.removeItem(USER_KEY);
    setUser(null);
    setGoals([]);
    setOverview(emptyOverview);
    setActiveSession(null);
    setRemainingSeconds(0);
  }

  if (!user) {
    return (
      <main className="min-h-screen bg-paper px-4 py-10 text-ink">
        <div className="mx-auto max-w-xl rounded-3xl bg-white p-8 shadow-calm">
          <p className="text-sm uppercase tracking-[0.2em] text-sea">Mathalama Focus</p>
          <h1 className="mt-3 text-3xl font-semibold">Calm focus for learning</h1>
          <p className="mt-2 text-slate-600">Dev auth for MVP testing. Login once and start sessions.</p>

          <form className="mt-6 space-y-4" onSubmit={handleLogin}>
            <label className="block">
              <span className="mb-1 block text-sm">Email</span>
              <input
                className="w-full rounded-xl border border-slate-200 px-3 py-2"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </label>

            <label className="block">
              <span className="mb-1 block text-sm">Name</span>
              <input
                className="w-full rounded-xl border border-slate-200 px-3 py-2"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </label>

            <button
              disabled={isLoading}
              className="w-full rounded-xl bg-sea px-4 py-2 font-medium text-white disabled:opacity-60"
              type="submit"
            >
              {isLoading ? 'Signing in...' : 'Enter Focus'}
            </button>
          </form>

          {error && <p className="mt-4 text-sm text-red-700">{error}</p>}
          {message && <p className="mt-4 text-sm text-sea">{message}</p>}
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-paper px-4 py-8 text-ink">
      <div className="mx-auto grid max-w-6xl gap-5 lg:grid-cols-[1.2fr_1fr]">
        <section className="space-y-5">
          <header className="rounded-3xl bg-white p-6 shadow-calm">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm uppercase tracking-[0.2em] text-sea">Mathalama Focus</p>
                <h1 className="mt-2 text-2xl font-semibold">What will you work on now?</h1>
              </div>
              <button className="rounded-lg border px-3 py-2 text-sm" onClick={logout}>
                Logout
              </button>
            </div>
          </header>

          <section className="rounded-3xl bg-white p-6 shadow-calm">
            <h2 className="text-xl font-semibold">Create Goal</h2>
            <form className="mt-4 space-y-3" onSubmit={handleCreateGoal}>
              <input
                className="w-full rounded-xl border border-slate-200 px-3 py-2"
                placeholder="Topic or task"
                value={topic}
                onChange={(e) => setTopic(e.target.value)}
                required
              />

              <textarea
                className="w-full rounded-xl border border-slate-200 px-3 py-2"
                placeholder="Desired result"
                value={desiredResult}
                onChange={(e) => setDesiredResult(e.target.value)}
                rows={2}
                required
              />

              <div className="flex gap-2">
                {[15, 25, 40].map((minutes) => (
                  <button
                    key={minutes}
                    type="button"
                    onClick={() => setModeMinutes(minutes)}
                    className={`rounded-full px-3 py-1 text-sm ${
                      modeMinutes === minutes ? 'bg-sea text-white' : 'bg-slate-100 text-slate-700'
                    }`}
                  >
                    {minutes} min
                  </button>
                ))}
              </div>

              <button
                type="submit"
                disabled={isLoading}
                className="rounded-xl bg-sea px-4 py-2 font-medium text-white disabled:opacity-60"
              >
                Add Goal
              </button>
            </form>
          </section>

          <section className="rounded-3xl bg-white p-6 shadow-calm">
            <h2 className="text-xl font-semibold">Goals</h2>
            <div className="mt-4 space-y-3">
              {goals.length === 0 && <p className="text-sm text-slate-500">No goals yet.</p>}
              {goals.map((goal) => (
                <article key={goal.id} className="rounded-2xl border border-slate-200 p-4">
                  <h3 className="font-medium">{goal.topic}</h3>
                  <p className="mt-1 text-sm text-slate-600">{goal.desired_result}</p>
                  <button
                    type="button"
                    disabled={isLoading || Boolean(activeSession && activeSession.status !== 'completed')}
                    onClick={() => handleStartSession(goal)}
                    className="mt-3 rounded-lg bg-mint px-3 py-2 text-sm font-medium text-ink disabled:opacity-60"
                  >
                    Start Focus
                  </button>
                </article>
              ))}
            </div>
          </section>
        </section>

        <section className="space-y-5">
          <section className="rounded-3xl bg-white p-6 shadow-calm">
            <h2 className="text-xl font-semibold">Focus Session</h2>
            {!activeSession && <p className="mt-2 text-sm text-slate-500">Start a goal to begin.</p>}

            {activeSession && (
              <div className="mt-4">
                <p className="text-sm text-slate-500">Status: {activeSession.status}</p>
                <p className="mt-2 text-5xl font-semibold tracking-tight text-sea">{formatClock(remainingSeconds)}</p>
                <p className="mt-2 text-sm text-slate-500">Pauses: {activeSession.pause_count}</p>

                {activeSession.status !== 'completed' && (
                  <div className="mt-4 flex flex-wrap gap-2">
                    <button
                      type="button"
                      onClick={handlePauseResume}
                      className="rounded-lg border px-3 py-2 text-sm"
                    >
                      {activeSession.status === 'active' ? 'Pause' : 'Resume'}
                    </button>
                    <button
                      type="button"
                      onClick={handleInterruption}
                      className="rounded-lg border px-3 py-2 text-sm"
                    >
                      Log interruption
                    </button>
                    <button
                      type="button"
                      onClick={handleComplete}
                      className="rounded-lg bg-sea px-3 py-2 text-sm text-white"
                    >
                      Complete session
                    </button>
                  </div>
                )}
              </div>
            )}
          </section>

          <section className="rounded-3xl bg-white p-6 shadow-calm">
            <h2 className="text-xl font-semibold">Reflection</h2>
            {!activeSession || activeSession.status !== 'completed' ? (
              <p className="mt-2 text-sm text-slate-500">Complete a session to unlock reflection.</p>
            ) : (
              <form className="mt-4 space-y-3" onSubmit={handleReflection}>
                <textarea
                  className="w-full rounded-xl border border-slate-200 px-3 py-2"
                  placeholder="What did you understand?"
                  value={whatLearned}
                  onChange={(e) => setWhatLearned(e.target.value)}
                  rows={2}
                  required
                />
                <textarea
                  className="w-full rounded-xl border border-slate-200 px-3 py-2"
                  placeholder="What was difficult?"
                  value={whatWasHard}
                  onChange={(e) => setWhatWasHard(e.target.value)}
                  rows={2}
                  required
                />
                <select
                  className="w-full rounded-xl border border-slate-200 px-3 py-2"
                  value={nextAction}
                  onChange={(e) => setNextAction(e.target.value)}
                >
                  <option value="continue">Continue</option>
                  <option value="repeat">Repeat</option>
                  <option value="postpone">Postpone</option>
                </select>

                <button type="submit" className="rounded-xl bg-sea px-4 py-2 font-medium text-white">
                  Save Reflection
                </button>
              </form>
            )}
          </section>

          <section className="rounded-3xl bg-white p-6 shadow-calm">
            <h2 className="text-xl font-semibold">Calm Analytics</h2>
            <div className="mt-3 grid grid-cols-2 gap-3 text-sm">
              <div className="rounded-xl bg-slate-50 p-3">
                <p className="text-slate-500">Completed goals</p>
                <p className="text-xl font-semibold">{overview.completed_goals}</p>
              </div>
              <div className="rounded-xl bg-slate-50 p-3">
                <p className="text-slate-500">Calm score</p>
                <p className="text-xl font-semibold">{overview.calm_score}%</p>
              </div>
              <div className="rounded-xl bg-slate-50 p-3">
                <p className="text-slate-500">Focus stability</p>
                <p className="text-xl font-semibold">{overview.focus_stability}%</p>
              </div>
              <div className="rounded-xl bg-slate-50 p-3">
                <p className="text-slate-500">Best hour (UTC)</p>
                <p className="text-xl font-semibold">
                  {overview.best_hour_of_day_utc >= 0 ? `${overview.best_hour_of_day_utc}:00` : 'N/A'}
                </p>
              </div>
            </div>
            <p className="mt-3 text-sm text-slate-600">Current calm level: {calmLabel}</p>
          </section>

          {error && <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{error}</p>}
          {message && <p className="rounded-xl bg-mint/30 p-3 text-sm text-ink">{message}</p>}
        </section>
      </div>
    </main>
  );
}
