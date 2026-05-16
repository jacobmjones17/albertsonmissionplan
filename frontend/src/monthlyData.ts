/** Twelve monthly challenge goals (static copy for the SPA). */
export type Goal = {
  id: string
  title: string
  bullets: { who: string; text: string }[]
}

/** Month index 0–11 matches each goal (January → December). */
export const MONTHS_SHORT = [
  'Jan',
  'Feb',
  'Mar',
  'Apr',
  'May',
  'Jun',
  'Jul',
  'Aug',
  'Sep',
  'Oct',
  'Nov',
  'Dec',
] as const

export const MONTHS_LONG = [
  'January',
  'February',
  'March',
  'April',
  'May',
  'June',
  'July',
  'August',
  'September',
  'October',
  'November',
  'December',
] as const

export const WEEKDAYS_SHORT = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'] as const

export const MONTHLY_GOALS: Goal[] = [
  {
    id: 'g1',
    title: 'Encourage daily scripture study',
    bullets: [
      { who: 'Individuals', text: 'Read a chapter from the Book of Mormon each day.' },
      { who: 'Families', text: 'Hold a weekly family scripture study and discussion.' },
      { who: 'Ward', text: 'Share insights and favorite verses in a weekly ward email or social media group.' },
    ],
  },
  {
    id: 'g2',
    title: 'Promote service and kindness',
    bullets: [
      {
        who: 'Individuals',
        text: 'Perform one act of kindness each day (for example, compliment someone or help a neighbor).',
      },
      {
        who: 'Families',
        text: 'Plan and carry out a family service project (for example, clean the chapel or donate to those in need).',
      },
      { who: 'Ward', text: 'Organize a ward-wide service day (for example, a donation drive).' },
    ],
  },
  {
    id: 'g3',
    title: 'Foster connection with ancestors and temple work',
    bullets: [
      { who: 'Individuals', text: '' },
      { who: 'Families', text: 'Share family stories and work on a family tree together.' },
      { who: 'Ward', text: 'Host a family history workshop or temple trip.' },
    ],
  },
  {
    id: 'g4',
    title: 'Increase missionary efforts and outreach',
    bullets: [
      {
        who: 'Individuals',
        text: 'Share a gospel message or invitation to a church activity with a friend.',
      },
      { who: 'Families', text: 'Invite neighbors to family home evening or a church event.' },
      { who: 'Ward', text: 'Organize a "bring a friend" Sunday.' },
    ],
  },
  {
    id: 'g5',
    title: 'Encourage physical, mental, and spiritual well-being',
    bullets: [
      { who: 'Individuals', text: 'Set personal health goals (exercise, healthy eating, mindful rest).' },
      { who: 'Families', text: 'Plan and participate in a family fitness activity each week.' },
      { who: 'Ward', text: 'Host a health and wellness fair.' },
    ],
  },
  {
    id: 'g6',
    title: 'Strengthen faith and personal testimonies',
    bullets: [
      { who: 'Individuals', text: 'Write in a journal about personal faith experiences.' },
      { who: 'Families', text: 'Share testimonies during family home evening.' },
      { who: 'Ward', text: 'Hold a testimony meeting or fireside.' },
    ],
  },
  {
    id: 'g7',
    title: 'Cultivate an attitude of gratitude',
    bullets: [
      { who: 'Individuals', text: 'Keep a daily gratitude journal.' },
      { who: 'Families', text: 'Create a gratitude tree or jar and add to it each day.' },
      {
        who: 'Ward',
        text: 'Share expressions of gratitude in sacrament meeting or the ward group chat.',
      },
    ],
  },
  {
    id: 'g8',
    title: 'Improve and deepen prayer habits',
    bullets: [
      { who: 'Individuals', text: 'Commit to praying morning and night.' },
      { who: 'Families', text: 'Hold family prayer daily.' },
      { who: 'Ward', text: 'Organize a ward prayer chain.' },
    ],
  },
  {
    id: 'g9',
    title: 'Honor and keep the Sabbath day holy',
    bullets: [
      { who: 'Individuals', text: 'Study Sabbath-related scriptures and talks.' },
      { who: 'Families', text: 'Plan meaningful Sabbath activities.' },
      { who: 'Ward', text: 'Share Sabbath experiences in ward meetings.' },
    ],
  },
  {
    id: 'g10',
    title: 'Study and apply general conference',
    bullets: [
      { who: 'Individuals', text: 'Study talks from previous conferences.' },
      { who: 'Families', text: 'Discuss conference messages and set goals.' },
      { who: 'Ward', text: 'Host a conference recap and goal-setting session.' },
    ],
  },
  {
    id: 'g11',
    title: 'Serve others in meaningful ways',
    bullets: [
      { who: 'Individuals', text: 'Volunteer at local charities or in the community.' },
      { who: 'Families', text: 'Plan and carry out a family service project.' },
      { who: 'Ward', text: 'Organize a ward-wide service initiative.' },
    ],
  },
  {
    id: 'g12',
    title: 'Focus on the true meaning of Christmas',
    bullets: [
      { who: 'Individuals', text: 'Study the life of Christ and His teachings.' },
      { who: 'Families', text: 'Participate in "Light the World"–style activities.' },
      { who: 'Ward', text: 'Host a nativity event or Christmas program.' },
    ],
  },
]
