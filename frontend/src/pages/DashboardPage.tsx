import React, { useEffect, useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { api, Goal, Insight } from '../lib/api';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { useNavigate } from 'react-router-dom';
import { Activity, Clock, ArrowRight, Hash, Sparkles, AlertTriangle, Lightbulb } from 'lucide-react';
import { motion } from 'framer-motion';

export const DashboardPage: React.FC = () => {
  const { user } = useAuth();
  const [goals, setGoals] = useState<Goal[]>([]);
  const [insight, setInsight] = useState<Insight | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchData = async () => {
    try {
      const [goalsData, insightData] = await Promise.all([
        api.goals.list(),
        api.analytics.insights()
      ]);
      setGoals(goalsData.goals || []);
      setInsight(insightData.insight);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  return (
    <div className="space-y-8 font-sans">
      <header className="flex items-center justify-between border-b border-border pb-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight font-mono uppercase text-primary">Dashboard</h1>
          <p className="text-sm text-muted-foreground font-mono">
            // STATUS: ONLINE | USER: {user?.name.toUpperCase()}
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={fetchData} className="hidden sm:flex">
          Refresh Data
        </Button>
      </header>

      {insight && (
        <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }}>
          <Card className="bg-surfaceHighlight/30 border-accent/20 border-l-4 border-l-accent p-4 relative overflow-hidden">
            <div className="flex items-start gap-4 relative z-10">
              <div className="p-2 rounded bg-accent/10 text-accent">
                {insight.type === 'warning' ? <AlertTriangle size={18} /> : 
                 insight.type === 'tip' ? <Lightbulb size={18} /> : 
                 <Sparkles size={18} />}
              </div>
              <div>
                <h3 className="text-xs font-bold uppercase tracking-wider text-accent mb-1 font-mono">AI Insight: {insight.title}</h3>
                <p className="text-sm text-primary leading-relaxed">{insight.content}</p>
              </div>
            </div>
            <div className="absolute top-0 right-0 p-2 opacity-5">
               <Sparkles size={64} />
            </div>
          </Card>
        </motion.div>
      )}

      <div className="grid gap-8 lg:grid-cols-[1fr,1.5fr]">
        <div className="space-y-6">
           <CreateGoalForm onGoalCreated={fetchData} />
        </div>

        <div className="space-y-4">
           <div className="flex items-center justify-between">
             <h2 className="text-sm font-medium tracking-wider text-muted-foreground uppercase">Recent Objectives</h2>
             <span className="text-xs font-mono text-muted-foreground">{goals.length} ACTIVE</span>
           </div>
           
           {loading ? (
             <div className="py-12 text-center text-xs font-mono text-muted-foreground animate-pulse">
               LOADING OBJECTIVES...
             </div>
           ) : goals.length === 0 ? (
             <div className="rounded-xl border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
               No active objectives found. Initialize new protocol.
             </div>
           ) : (
             <div className="grid gap-3">
               {goals.map((goal) => (
                 <GoalCard key={goal.id} goal={goal} />
               ))}
             </div>
           )}
        </div>
      </div>
    </div>
  );
};

const CreateGoalForm: React.FC<{ onGoalCreated: () => void }> = ({ onGoalCreated }) => {
  const [topic, setTopic] = useState('');
  const [result, setResult] = useState('');
  const [minutes, setMinutes] = useState(25);
  const [tags, setTags] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const tagList = tags.split(',').map(t => t.trim()).filter(Boolean);
      await api.goals.create({
        topic,
        desired_result: result,
        recommended_minutes: minutes,
        tags: tagList
      });
      setTopic('');
      setResult('');
      setTags('');
      onGoalCreated();
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card className="bg-surface border-border shadow-none">
       <form onSubmit={handleSubmit} className="space-y-5">
         <div className="flex items-center gap-2 text-primary mb-2">
           <Activity size={18} />
           <span className="font-bold font-mono text-sm uppercase">New Objective</span>
         </div>
         
         <div className="grid gap-4 sm:grid-cols-2">
           <div className="space-y-1.5">
             <label className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">Topic</label>
             <Input 
               placeholder="SYS.BACKEND.API" 
               value={topic}
               onChange={e => setTopic(e.target.value)}
               className="font-mono text-xs bg-background border-border"
               required
             />
           </div>
           <div className="space-y-1.5">
             <label className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">Duration (m)</label>
             <Input 
               type="number"
               min={5}
               max={180}
               value={minutes}
               onChange={e => setMinutes(parseInt(e.target.value))}
               className="font-mono text-xs bg-background border-border"
               required
             />
           </div>
         </div>

         <div className="space-y-1.5">
           <label className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">Desired Result</label>
           <Input 
             placeholder="Implement authentication..." 
             value={result}
             onChange={e => setResult(e.target.value)}
             className="text-sm bg-background border-border"
             required
           />
         </div>

         <div className="space-y-1.5">
            <label className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">Tags (CSV)</label>
            <div className="relative">
              <Hash className="absolute left-3 top-3 h-3.5 w-3.5 text-muted-foreground" />
              <Input 
                className="pl-9 text-xs font-mono bg-background border-border"
                placeholder="dev, urgent"
                value={tags}
                onChange={e => setTags(e.target.value)}
              />
            </div>
         </div>

         <div className="pt-2">
           <Button type="submit" disabled={loading} className="w-full bg-primary text-primary-foreground hover:bg-primary/90 font-mono text-xs uppercase tracking-wide">
             {loading ? 'Initializing...' : 'Initialize Objective'}
           </Button>
         </div>
       </form>
    </Card>
  );
};

const GoalCard: React.FC<{ goal: Goal }> = ({ goal }) => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);

  const handleStart = async () => {
    setLoading(true);
    try {
      const { session } = await api.sessions.start({
        goal_id: goal.id,
        recommended_minutes: goal.recommended_minutes,
        is_strict: false
      });
      navigate(`/session/${session.id}`);
    } catch (err) {
      console.error(err);
      setLoading(false);
    }
  };

  return (
    <motion.div
      whileHover={{ x: 4 }}
      transition={{ duration: 0.2 }}
    >
      <div 
        onClick={handleStart}
        className="group relative flex cursor-pointer flex-col gap-2 rounded-lg border border-border bg-surface p-4 transition-colors hover:border-accent/50 hover:bg-surfaceHighlight"
      >
        <div className="flex items-start justify-between">
            <div className="flex flex-col gap-1">
                <h3 className="font-mono text-xs font-bold uppercase tracking-wide text-primary group-hover:text-accent transition-colors">
                    {goal.topic}
                </h3>
                <span className="text-[10px] text-muted-foreground font-mono">
                    ID: {goal.id.substring(0, 8)}
                </span>
            </div>
            <div className="flex items-center gap-1 rounded bg-background px-1.5 py-0.5 text-[10px] font-mono font-medium text-muted-foreground border border-border">
              <Clock size={10} />
              <span>{goal.recommended_minutes}m</span>
            </div>
        </div>
          
        <p className="text-sm text-muted-foreground line-clamp-1 group-hover:text-primary transition-colors">
            {goal.desired_result}
        </p>

        {goal.tags && goal.tags.length > 0 && (
            <div className="flex flex-wrap gap-1 mt-1">
              {goal.tags.map(tag => (
                <span key={tag} className="text-[9px] px-1 py-0 rounded bg-secondary text-muted-foreground font-mono uppercase">
                  #{tag}
                </span>
              ))}
            </div>
        )}
      </div>
    </motion.div>
  );
};