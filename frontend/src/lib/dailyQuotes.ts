import { AppLanguage } from '../context/LanguageContext';

export interface DailyQuote {
  text: string;
  author: string;
}

const quotesByLanguage: Record<AppLanguage, DailyQuote[]> = {
  en: [
    { text: 'Knowing yourself is the beginning of all wisdom.', author: 'Aristotle' },
    { text: 'The future depends on what you do today.', author: 'Mahatma Gandhi' },
    { text: 'It always seems impossible until it is done.', author: 'Nelson Mandela' },
    { text: 'The secret of getting ahead is getting started.', author: 'Mark Twain' },
    { text: 'Everyone thinks of changing the world, but no one thinks of changing himself.', author: 'Leo Tolstoy' },
    { text: 'He who opens a school door, closes a prison.', author: 'Victor Hugo' },
    { text: 'The man who moves a mountain begins by carrying away small stones.', author: 'Confucius' },
    { text: 'Nothing in life is to be feared, it is only to be understood.', author: 'Marie Curie' },
    { text: 'A room without books is like a body without a soul.', author: 'Cicero' },
    { text: 'Love all humanity as your brother.', author: 'Abai Qunanbaiuly' },
    { text: 'Find your place in the world, like a brick in a wall.', author: 'Abai Qunanbaiuly' },
    { text: 'Education without upbringing is the enemy of humanity.', author: 'Al-Farabi' },
    { text: 'Come, children, let us learn.', author: 'Ybyrai Altynsarin' },
    { text: 'A nation without discipline cannot stand.', author: 'Bauyrzhan Momyshuly' },
  ],
  ru: [
    { text: 'Познание себя — начало всякой мудрости.', author: 'Аристотель' },
    { text: 'Будущее зависит от того, что вы делаете сегодня.', author: 'Махатма Ганди' },
    { text: 'Пока не сделано, всё кажется невозможным.', author: 'Нельсон Мандела' },
    { text: 'Секрет движения вперёд — начать.', author: 'Марк Твен' },
    { text: 'Все хотят изменить мир, но никто не хочет измениться сам.', author: 'Лев Толстой' },
    { text: 'Тот, кто открывает дверь школы, закрывает дверь тюрьмы.', author: 'Виктор Гюго' },
    { text: 'Тот, кто сдвигает гору, начинает с малого камня.', author: 'Конфуций' },
    { text: 'В жизни нет ничего, чего нужно бояться, есть лишь то, что нужно понять.', author: 'Мария Кюри' },
    { text: 'Дом без книг — как тело без души.', author: 'Цицерон' },
    { text: 'Люби всё человечество, как брата своего.', author: 'Абай Кунанбаев' },
    { text: 'И ты стань кирпичом в мироздании, найди свою нишу и будь уложен.', author: 'Абай Кунанбаев' },
    { text: 'Воспитание без знаний бесплодно, знания без воспитания опасны.', author: 'Аль-Фараби' },
    { text: 'Давайте, дети, учиться.', author: 'Ыбырай Алтынсарин' },
    { text: 'Народ без дисциплины не устоит.', author: 'Бауыржан Момышулы' },
  ],
  kk: [
    { text: 'Адамзаттың бәрін сүй, бауырым деп.', author: 'Абай Құнанбайұлы' },
    { text: 'Ғылым таппай мақтанба.', author: 'Абай Құнанбайұлы' },
    { text: 'Сен де бір кірпіш дүниеге, кетігін тап та, бар, қалан.', author: 'Абай Құнанбайұлы' },
    { text: 'Тәрбиесіз берілген білім — адамзаттың қас жауы.', author: 'Әл-Фараби' },
    { text: 'Кел, балалар, оқылық.', author: 'Ыбырай Алтынсарин' },
    { text: 'Тәртіпке бас иген құл болмайды, тәртіпсіз ел болмайды.', author: 'Бауыржан Момышұлы' },
    { text: 'Өзіңді тану — даналықтың басы.', author: 'Аристотель' },
    { text: 'Болашақ бүгін не істейтініңе байланысты.', author: 'Махатма Ганди' },
    { text: 'Бітпейінше, бәрі мүмкін еместей көрінеді.', author: 'Нельсон Мандела' },
    { text: 'Алға шығудың құпиясы — бастау.', author: 'Марк Твен' },
    { text: 'Әлемді өзгерткісі келмейтін адам жоқ, өзін өзгерткісі келетін адам аз.', author: 'Лев Толстой' },
    { text: 'Мектеп есігін ашқан адам түрме есігін жабады.', author: 'Виктор Гюго' },
    { text: 'Тауды жылжытқан адам ұсақ тастан бастайды.', author: 'Конфуций' },
    { text: 'Кітапсыз үй — жансыз тәнмен тең.', author: 'Цицерон' },
  ],
};

const localDateKey = (date: Date) => {
  const y = date.getFullYear();
  const m = `${date.getMonth() + 1}`.padStart(2, '0');
  const d = `${date.getDate()}`.padStart(2, '0');
  return `${y}-${m}-${d}`;
};

const hashString = (value: string) => {
  let hash = 0;
  for (let i = 0; i < value.length; i += 1) {
    hash = (hash * 31 + value.charCodeAt(i)) | 0;
  }
  return Math.abs(hash);
};

export const getDailyQuote = (language: AppLanguage, date: Date): DailyQuote => {
  const pool = quotesByLanguage[language];
  const key = localDateKey(date);
  const idx = hashString(key) % pool.length;
  return pool[idx];
};

export const msUntilNextLocalDay = (now: Date) => {
  const next = new Date(now);
  next.setHours(24, 0, 0, 0);
  return Math.max(1000, next.getTime() - now.getTime());
};
