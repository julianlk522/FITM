import preact from '@astrojs/preact'
import { defineConfig } from 'astro/config'

import netlify from '@astrojs/netlify'

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
