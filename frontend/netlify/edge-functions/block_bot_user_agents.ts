// thanks to https://www.jeremiak.com/blog/block-bots-netlify-edge-functions/
import { Config, Context } from '@netlify/edge-functions'

const bot_user_agents = [
	'AdsBot-Google',
	'Amazonbot',
	'anthropic-ai',
	'Applebot',
	'AwarioRssBot',
	'AwarioSmartBot',
	'Bytespider',
	'CCBot',
	'ChatGPT',
	'ChatGPT-User',
	'Claude-Web',
	'ClaudeBot',
	'cohere-ai',
	'DataForSeoBot',
	'Diffbot',
	'Dotbot',
	'FacebookBot',
	'GoogleBot',
	'GoogleOther',
	'Google-Extended',
	'GPTBot',
	'ImagesiftBot',
	'magpie-crawler',
	'Meltwater',
	'omgili',
	'Omgilibot',
	'peer39_crawler',
	'PerplexityBot',
	'Seekr',
	'Semrushbot',
	'YouBot',
]

export default async (request: Request, context: Context) => {
	const ua_headers = request.headers.get('user-agent')?.toLowerCase()
	let is_bot = false

	bot_user_agents.forEach((ua) => {
		if (!ua_headers || ua_headers?.includes(ua.toLowerCase())) {
			is_bot = true
		}
	})

	const response = is_bot
		? new Response(null, { status: 401 })
		: await context.next()
	return response
}

export const config: Config = { path: '/*' }
