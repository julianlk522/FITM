import type { ResponseAndRedirect } from '../types'
export default async function fetch_with_handle_redirect(
	url: string,
	opts?: RequestInit
): Promise<ResponseAndRedirect> {
	try {
		const resp = await fetch(url, opts)
		switch (resp.status) {
			// unauthorized
			case 401:
				return { Response: undefined, RedirectTo: '/login' }
			// not found
			case 404:
				return { Response: undefined, RedirectTo: '/404' }
			// rate limited
			case 429:
				return { Response: undefined, RedirectTo: '/rate-limit' }
			// server error
			case 500:
				return { Response: undefined, RedirectTo: '/500' }
			// anything else: no redirect
			default:
				return { Response: resp, RedirectTo: undefined }
		}
	} catch {
		return { Response: undefined, RedirectTo: '/404' }
	}
}
