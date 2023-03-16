import { defineConfig } from 'astro/config';
import mdx from "@astrojs/mdx";

export default defineConfig({
    site: 'https://avolent.io',
    base: '/cortex',
    integrations: [mdx()],
    metadata: {
        lastUpdated: new Date().toISOString()
    }
});