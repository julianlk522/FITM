import type { APIContext } from 'astro'
import { sequence } from 'astro:middleware'
import type { VerifyErrors } from 'jsonwebtoken'
import jwt from 'jsonwebtoken'
import { API_URL } from './constants'

export const onRequest = sequence(handle_jwt_auth, handle_redirect_action)

async function handle_jwt_auth(
	context: APIContext,
	next: () => Promise<Response>
) {
	const req_token = context.cookies.get('token')?.value
	const req_user = context.cookies.get('user')?.value

	// authenticate token cookie if found
	if (req_token) {
		try {
			jwt.verify(
				req_token,
				import.meta.env.VITE_FITM_JWT_SECRET,
				function <JwtPayload>(
					err: VerifyErrors | null,
					decoded: JwtPayload | undefined
				) {
					// delete cookies and redirect to login if invalid
					// @ts-ignore
					if (err || !decoded || !decoded.login_name) {
						context.cookies.delete('token')
						context.cookies.delete('user')

						return Response.redirect(
							new URL('/login', context.request.url),
							302
						)

						// set user cookie if verified
					} else {
						// @ts-ignore
						context.cookies.set('user', decoded.login_name, {
							path: '/',
							maxAge: 14400,
							sameSite: 'strict',
							secure: true,
						})
					}
				}
			)
		} catch (err) {
			// TODO: maybe add saved logging
			console.log('jwt errors: ', err)
			return Response.redirect(
				new URL('/login', context.request.url),
				302
			)
		}

		// reset and redirect to login if user cookie found but not token cookie
	} else if (req_user) {
		context.cookies.delete('user')
		return Response.redirect(new URL('/login', context.request.url))
	}

	return next()
}

async function handle_redirect_action(
	context: APIContext,
	next: () => Promise<Response>
) {
	// "like summary 78", "copy summary 89", etc.
	const redirect_action = context.cookies.get('redirect_action')?.value

	// continue normally if no redirect action cookie found
	// or if user aborted redirect process from login page
	// (in the 2nd case, there should be a redirect_action cookie but no token
	// cookie and the request URL should not be "/login", so token check below
	// will apply)
	if (
		!redirect_action ||
		context.request.url === 'https://fitm.online/login' ||
		context.request.url === 'https://www.fitm.online/login'
		// || context.request.url === 'http://localhost:4321/login'
		// || context.request.url === 'http://127.0.0.1:4321/login'
	) {
		return next()
	}

	// remove cookie if unauthenticated
	const token = context.cookies.get('token')?.value
	if (!token) {
		context.cookies.delete('redirect_action')
		return next()
	}

	// "like", "copy"
	const action = redirect_action.split(' ')[0]

	// "link", "summary"
	const item = redirect_action.split(' ')[1]
	let api_section
	if (item === 'link') {
		api_section = 'links'
	} else if (item === 'summary') {
		api_section = 'summaries'
	}

	const item_id = redirect_action.split(' ')[2]
	const redirect_action_url = `${API_URL}/${api_section}/${item_id}/${action}`

	const resp = await fetch(redirect_action_url, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
			Authorization: `Bearer ${token}`,
		},
	})

	if (resp.status !== 200) {
		// TODO: maybe add saved logging
		console.error('redirect action failed')
	} else {
		// cleanup cookie if successful
		context.cookies.delete('redirect_action', {
			path: context.url.pathname,
		})
	}

	return next()
}
