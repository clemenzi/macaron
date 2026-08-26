import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  output: 'static',
  integrations: [
    starlight({
      title: 'Macaron',
      description:
        'Turn your Mac into a securely accessible remote workstation.',
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/clemenzi/macaron',
        },
      ],
      editLink: {
        baseUrl: 'https://github.com/clemenzi/macaron/edit/main/',
      },
      locales: {
        root: {
          label: 'English',
          lang: 'en',
        },
        it: {
          label: 'Italiano',
          lang: 'it',
        },
      },
      sidebar: [
        {
          label: 'Getting started',
          translations: { it: 'Per iniziare' },
          items: [
            { slug: 'getting-started/installation' },
            { slug: 'getting-started/running' },
          ],
        },
        {
          label: 'Services',
          translations: { it: 'Servizi' },
          items: [
            { slug: 'services/installing' },
            { slug: 'services/official' },
            { slug: 'services/managing' },
          ],
        },
        {
          label: 'Service development',
          translations: { it: 'Sviluppo dei servizi' },
          items: [
            { slug: 'service-development/overview' },
            { slug: 'service-development/scripts' },
          ],
        },
        {
          label: 'Reference',
          translations: { it: 'Riferimenti' },
          items: [
            { slug: 'reference/configuration' },
            { slug: 'reference/cli' },
            { slug: 'reference/troubleshooting' },
          ],
        },
      ],
      customCss: ['./src/styles/custom.css'],
      lastUpdated: true,
    }),
  ],
});
