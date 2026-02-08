import { AppLanguage } from '../context/LanguageContext';
import { translate } from './i18n';

export interface SessionTemplate {
  id: string;
  name: string;
  topic: string;
  desiredResult: string;
  tags: string[];
  minutes: number;
}

const templateDefinitions: Array<{ id: string; minutes: number }> = [
  { id: 'deepWork', minutes: 50 },
  { id: 'study', minutes: 40 },
  { id: 'code', minutes: 45 },
  { id: 'review', minutes: 25 },
];

export const getSessionTemplates = (language: AppLanguage): SessionTemplate[] => {
  return templateDefinitions.map((template) => ({
    id: template.id,
    minutes: template.minutes,
    name: translate(language, `templates.${template.id}.name`),
    topic: translate(language, `templates.${template.id}.topic`),
    desiredResult: translate(language, `templates.${template.id}.result`),
    tags: translate(language, `templates.${template.id}.tags`)
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean),
  }));
};

