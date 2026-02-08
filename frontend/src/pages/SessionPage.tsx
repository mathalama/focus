import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api, FocusSession } from '../lib/api';
import { Button } from '../components/ui/Button';
import { Play, Pause, CheckCircle, AlertOctagon, Coffee, BatteryCharging } from 'lucide-react';
import { motion } from 'framer-motion';
import { addMinutes, differenceInSeconds, parseISO } from 'date-fns';

export const SessionPage: React.FC = () => {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  const [session, setSession] = useState<FocusSession | null>(null);
  const [timeLeft, setTimeLeft] = useState(0); 
  const [breakTime, setBreakTime] = useState(0); // Local break timer in seconds
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  
  // Focus Timer
  useEffect(() => {
    if (!session || session.status !== 'active') return;

    const interval = setInterval(() => {
      const startDate = parseISO(session.started_at);
      const endDate = addMinutes(startDate, session.recommended_minutes);
      const diff = differenceInSeconds(endDate, new Date());
      
      if (diff <= 0) {
        setTimeLeft(0);
        clearInterval(interval);
      } else {
        setTimeLeft(diff);
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [session]);

  // Break Timer
  useEffect(() => {
    if (breakTime <= 0) return;
    const interval = setInterval(() => {
      setBreakTime((prev) => Math.max(0, prev - 1));
    }, 1000);
    return () => clearInterval(interval);
  }, [breakTime]);

  const fetchSession = async () => {
    if (!sessionId) return;
    try {
      const data = await api.sessions.get(sessionId);
      setSession(data.session);
      
      if (data.session.status === 'active') {
         const startDate = parseISO(data.session.started_at);
         const endDate = addMinutes(startDate, data.session.recommended_minutes);
         const diff = differenceInSeconds(endDate, new Date());
         setTimeLeft(Math.max(0, diff));
      } else {
         setTimeLeft(data.session.recommended_minutes * 60);
      }
    } catch (err) {
      console.error(err);
      navigate('/');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSession();
  }, [sessionId]);

  const handleAction = async (action: 'pause' | 'resume' | 'abandon' | 'complete' | 'reset') => {
    if (!sessionId) return;
    setActionLoading(true);
    try {
      let res;
      switch (action) {
        case 'pause': res = await api.sessions.pause(sessionId); break;
        case 'resume': 
          res = await api.sessions.resume(sessionId); 
          setBreakTime(0); // Clear break timer on resume
          break;
        case 'reset': 
          res = await api.sessions.reset(sessionId); 
          setBreakTime(0);
          break;
        case 'abandon': res = await api.sessions.abandon(sessionId); break;
        case 'complete': res = await api.sessions.complete(sessionId); break;
      }
      
      if (action === 'abandon') {
        navigate('/');
        return;
      }
      if (action === 'complete') {
        // Increment local session count for smart breaks
        const currentCount = parseInt(localStorage.getItem('completedSessionsCount') || '0', 10);
        localStorage.setItem('completedSessionsCount', (currentCount + 1).toString());
        
        navigate(`/session/${sessionId}/reflection`);
        return;
      }

      setSession(res.session);
    } catch (err) {
      console.error(err);
    } finally {
      setActionLoading(false);
    }
  };

  const startBreak = (minutes: number) => {
    setBreakTime(minutes * 60);
  };

  if (loading) return <div className="flex h-screen items-center justify-center font-mono text-xs">INITIALIZING...</div>;
  if (!session) return <div className="flex h-screen items-center justify-center font-mono text-xs text-red-500">SESSION_NOT_FOUND</div>;

  // Smart Break Logic
  const sessionsCompleted = parseInt(localStorage.getItem('completedSessionsCount') || '0', 10);
  const isLongBreakDue = sessionsCompleted > 0 && sessionsCompleted % 3 === 0;

  // Display logic
  const displayTime = breakTime > 0 ? breakTime : timeLeft;
  const minutes = Math.floor(displayTime / 60);
  const seconds = displayTime % 60;
  
  // Progress Ring Logic
  const R = 120;
  const C = 2 * Math.PI * R;
  let totalSeconds = session.recommended_minutes * 60;
  if (breakTime > 0) totalSeconds = breakTime; 
  
  const progress = totalSeconds > 0 ? displayTime / totalSeconds : 0;
  const dashOffset = C * (1 - progress);
  const isBreak = breakTime > 0;

  return (
    <div className="flex min-h-[70vh] flex-col items-center justify-center font-sans text-primary">
      <motion.div 
        initial={{ scale: 0.95, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        className="w-full max-w-md text-center"
      >
        <div className="relative mx-auto mb-10 flex h-64 w-64 items-center justify-center">
          {/* Static Ring */}
          <circle
              cx="128"
              cy="128"
              r="120"
              fill="none"
              stroke="currentColor"
              strokeWidth="1"
              className="text-surfaceHighlight"
            />

          {/* Progress Ring */}
          <svg className="absolute inset-0 h-full w-full -rotate-90">
             {/* Background track */}
            <circle
              cx="128"
              cy="128"
              r={R}
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              className="text-surfaceHighlight"
            />
            {/* Active progress */}
            <circle
              cx="128"
              cy="128"
              r={R}
              fill="none"
              stroke="currentColor"
              strokeWidth="4"
              strokeDasharray={C}
              strokeDashoffset={dashOffset} // This visual might be jumpy for breaks without tracking total break time, but acceptable for MVP
              strokeLinecap="round"
              className={`transition-all duration-1000 ease-linear shadow-glow ${isBreak ? 'text-green-500' : 'text-accent'}`}
            />
          </svg>

          <div className="relative z-10 flex flex-col items-center">
             <span className={`text-6xl font-bold tracking-tighter tabular-nums font-mono ${isBreak ? 'text-green-500' : 'text-primary'}`}>
               {minutes.toString().padStart(2, '0')}:{seconds.toString().padStart(2, '0')}
             </span>
             <span className="mt-4 flex items-center gap-2 rounded-full border border-border bg-surface px-3 py-1 text-[10px] font-mono font-bold uppercase tracking-widest text-muted-foreground">
               <span className={`h-1.5 w-1.5 rounded-full ${isBreak ? 'bg-green-500 animate-pulse' : session.status === 'active' ? 'bg-accent animate-pulse' : 'bg-yellow-500'}`} />
               {isBreak ? 'RECHARGING' : session.status}
             </span>
          </div>
        </div>

        <div className="mb-8 space-y-1">
           <h2 className="text-sm font-bold uppercase tracking-widest text-muted-foreground">Current Protocol</h2>
           <p className="text-lg font-medium">{isBreak ? 'BREAK IN PROGRESS' : `${session.recommended_minutes} MIN FOCUS SESSION`}</p>
        </div>

        <div className="flex flex-col items-center gap-4">
           {/* Primary Controls */}
           <div className="flex items-center justify-center gap-6">
             {session.status === 'active' ? (
               <>
                 <Button 
                   variant="secondary" 
                   size="icon" 
                   onClick={() => handleAction('reset')}
                   disabled={actionLoading}
                   className="h-10 w-10 rounded-full border border-border bg-surface hover:bg-surfaceHighlight"
                   title="Reset Timer"
                 >
                   <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/></svg>
                 </Button>

                 {!session.is_strict && (
                   <Button 
                     variant="secondary" 
                     size="icon" 
                     onClick={() => handleAction('pause')}
                     disabled={actionLoading}
                     className="h-14 w-14 rounded-full border border-border bg-surface hover:bg-surfaceHighlight"
                   >
                     <Pause size={20} />
                   </Button>
                 )}
                 <Button 
                   variant="primary" 
                   size="lg"
                   className="h-14 gap-3 rounded-full px-8 font-mono text-xs uppercase tracking-wider bg-accent text-accent-foreground hover:bg-white/90"
                   onClick={() => handleAction('complete')}
                   disabled={actionLoading}
                 >
                   <CheckCircle size={18} />
                   Complete
                 </Button>
               </>
             ) : (
               <Button 
                  variant="primary" 
                  size="icon" 
                  onClick={() => handleAction('resume')}
                  disabled={actionLoading}
                  className="h-14 w-14 rounded-full bg-accent text-accent-foreground hover:bg-white/90"
                  title="Resume Focus"
               >
                  <Play size={20} fill="currentColor" />
               </Button>
             )}

             <Button 
                variant="ghost" 
                size="icon"
                className="h-10 w-10 rounded-full text-muted-foreground hover:bg-red-950/30 hover:text-red-500"
                onClick={() => handleAction('abandon')}
                disabled={actionLoading}
                title="Abandon Protocol"
             >
                <AlertOctagon size={16} />
             </Button>
           </div>

           {/* Break Controls (Only visible when paused and not strict) */}
           {session.status === 'paused' && !isBreak && (
             <div className="mt-4 grid grid-cols-2 gap-3 w-full max-w-xs animate-in fade-in slide-in-from-bottom-4">
               <Button 
                 variant={!isLongBreakDue ? "primary" : "outline"} 
                 size="sm" 
                 onClick={() => startBreak(5)} 
                 className="font-mono text-xs gap-2"
               >
                 <Coffee size={14} /> Short (5m)
               </Button>
               <Button 
                 variant={isLongBreakDue ? "primary" : "outline"} 
                 size="sm" 
                 onClick={() => startBreak(15)} 
                 className="font-mono text-xs gap-2"
               >
                 <BatteryCharging size={14} /> 
                 {isLongBreakDue ? "Suggested (15m)" : "Long (15m)"}
               </Button>
             </div>
           )}
           
           {isBreak && (
             <div className="mt-2 animate-in fade-in">
               <p className="text-xs text-muted-foreground font-mono mb-2">Break active. Click Resume to end early.</p>
             </div>
           )}
        </div>
      </motion.div>
    </div>
  );
};
