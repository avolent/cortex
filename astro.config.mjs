import { defineConfig } from 'astro/config';

// https://astro.build/config
import mdx from "@astrojs/mdx";


export default defineConfig({
    site: 'https://avolent.io',
    base: '/cortex',
    integrations: [mdx()]
});