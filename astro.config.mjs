import { defineConfig } from 'astro/config';
import mdx from '@astrojs/mdx';
import sitemap from '@astrojs/sitemap';

export default defineConfig({
  site: 'https://avolent.github.io/cortex',
  integrations: [mdx(), sitemap()],
});
