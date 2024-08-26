import netlify from '@astrojs/netlify'
import preact from '@astrojs/preact'
import { defineConfig } from 'astro/config'

// https://astro.build/config
export default defineConfig({
	output: 'server',
	integrations: [
		preact({
			devtools: true,
		}),
	],
	adapter: netlify({ edgeMiddlware: true }),
})
