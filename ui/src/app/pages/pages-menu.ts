import { NbMenuItem } from '@nebular/theme';

export const MENU_ITEMS: NbMenuItem[] = [
  {
    title: 'MainFlux',
    icon: 'nb-e-commerce',
    link: '/pages/dashboard',
    home: true,
  },
  {
    title: 'MANAGEMENT',
    group: true,
  },
  {
    title: 'Devices',
    icon: 'nb-tables',
    children: [
      {
        title: 'Things',
        link: '/pages/tables/things',
      },
      {
        title: 'Channels',
        link: '/pages/tables/channels',
      },
      {
        title: 'Control',
        link: '/pages/tables/smart-table',
      },
    ],
  },
];
