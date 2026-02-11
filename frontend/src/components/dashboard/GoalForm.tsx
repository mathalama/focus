import React, { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useLanguage } from '../../context/LanguageContext';
import { useI18n } from '../../lib/i18n';
import { api } from '../../api';
import { getSessionTemplates } from '../../lib/sessionTemplates';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Activity, Hash } from 'lucide-react';

export const GoalForm: React.FC = () => {
  const { language } = useLanguage();
  const { t } = useI18n();
  const navigate = useNavigate();
  const [topic, setTopic] = useState('');
  const [result, setResult] = useState('');
  const [minutes, setMinutes] = useState(25);
  const [tags, setTags] = useState('');
  const [manualLoading, setManualLoading] = useState(false);
  const [templateLoadingId, setTemplateLoadingId] = useState<string | null>(null);
  const [formError, setFormError] = useState<string | null>(null);
  const templates = useMemo(() => getSessionTemplates(language), [language]);

  const createAndStart = async (payload: {
    topic: string;
    desired_result: string;
    recommended_minutes: number;
    tags: string[];
  }) => {
    const { goal } = await api.goals.create(payload);
    const { session } = await api.sessions.start({
      goal_id: goal.id,
      recommended_minutes: goal.recommended_minutes ?? payload.recommended_minutes,
      is_strict: false,
    });
    navigate(`/session/${session.id}`);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setManualLoading(true);
    setFormError(null);
    try {
      const tagList = tags.split(',').map(t => t.trim()).filter(Boolean);
      await createAndStart({
        topic,
        desired_result: result,
        recommended_minutes: minutes,
        tags: tagList,
      });
    } catch (err) {
      console.error(err);
      setFormError(err instanceof Error ? err.message : t('goalForm.error'));
    } finally {
      setManualLoading(false);
    }
  };

  const handleTemplateStart = async (templateId: string) => {
    const template = templates.find((item) => item.id === templateId);
    if (!template) return;

    setTemplateLoadingId(template.id);
    setFormError(null);
    try {
      await createAndStart({
        topic: template.topic,
        desired_result: template.desiredResult,
        recommended_minutes: template.minutes,
        tags: template.tags,
      });
    } catch (err) {
      console.error(err);
      setFormError(err instanceof Error ? err.message : t('goalForm.error'));
    } finally {
      setTemplateLoadingId(null);
    }
  };

  const loading = manualLoading || templateLoadingId !== null;

  return (
    <Card className="bg-surface border-border shadow-none">
       <form onSubmit={handleSubmit} className="space-y-5">
         <div className="flex items-center gap-2 text-primary mb-2">
           <Activity size={18} />
           <span className="font-bold font-mono text-sm uppercase">{t('goalForm.newObjective')}</span>
         </div>
         
         <div className="grid gap-4 sm:grid-cols-2">
           <div className="space-y-1.5">
             <label className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">{t('goalForm.topic')}</label>
             <Input 
               placeholder={t('goalForm.placeholder.topic')}
               value={topic}
               onChange={e => setTopic(e.target.value)}
               className="font-mono text-xs bg-background border-border"
               required
             />
           </div>
           <div className="space-y-1.5">
             <label className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">{t('goalForm.duration')}</label>
             <div className="flex items-center gap-2">
               <Input 
                 type="number"
                 min={1}
                 max={480}
                 value={minutes}
                 onChange={e => {
                   const value = parseInt(e.target.value, 10);
                   if (!isNaN(value) && value >= 1) {
                     setMinutes(Math.min(value, 480));
                   }
                 }}
                 className="font-mono text-xs bg-background border-border"
                 required
               />
               <span className="text-xs text-muted-foreground font-mono whitespace-nowrap">min</span>
             </div>
           </div>
         </div>

         <div className="space-y-1.5">
           <label className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">{t('goalForm.desiredResult')}</label>
           <Input 
             placeholder={t('goalForm.placeholder.result')}
             value={result}
             onChange={e => setResult(e.target.value)}
             className="text-sm bg-background border-border"
             required
           />
         </div>

         <div className="space-y-1.5">
            <label className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">{t('goalForm.tags')}</label>
            <div className="relative">
              <Hash className="absolute left-3 top-3 h-3.5 w-3.5 text-muted-foreground" />
              <Input 
                className="pl-9 text-xs font-mono bg-background border-border"
                placeholder={t('goalForm.placeholder.tags')}
                value={tags}
                onChange={e => setTags(e.target.value)}
              />
            </div>
         </div>

         <div className="pt-2">
           <Button type="submit" disabled={loading} className="w-full bg-primary text-primary-foreground hover:bg-primary/90 font-mono text-xs uppercase tracking-wide">
             {loading ? t('goalForm.initializing') : t('goalForm.initialize')}
           </Button>
         </div>

         {formError && (
           <div className="rounded-lg border border-red-500/30 bg-red-950/20 px-4 py-2 text-xs font-mono text-red-400 animate-in fade-in">
             {formError}
           </div>
         )}

         <div className="space-y-2">
           <div className="flex items-center justify-between">
             <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">{t('goalForm.quickTemplates')}</span>
             <span className="text-[10px] font-mono text-muted-foreground">{t('goalForm.quickTemplatesHint')}</span>
           </div>
           <div className="grid gap-2">
             {templates.map((template) => (
               <button
                 key={template.id}
                 type="button"
                 onClick={() => void handleTemplateStart(template.id)}
                 disabled={loading}
                 className="flex items-center justify-between rounded-lg border border-border bg-background px-3 py-2 text-left transition-colors hover:bg-surfaceHighlight disabled:opacity-60"
               >
                 <div className="min-w-0">
                   <p className="truncate font-mono text-xs font-bold uppercase text-primary">{template.name}</p>
                   <p className="truncate text-[11px] text-muted-foreground">{template.topic}</p>
                 </div>
                 <div className="ml-3 text-right text-[10px] font-mono uppercase text-muted-foreground">
                   <p>{template.minutes}m</p>
                   <p>{templateLoadingId === template.id ? t('goalForm.initializing') : t('goalForm.templateStart')}</p>
                 </div>
               </button>
             ))}
           </div>
         </div>
       </form>
    </Card>
  );
};
