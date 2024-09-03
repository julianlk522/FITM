import type { ResponseAndRedirect } from '../types'
export default async function fetch_with_handle_redirect(
	url: string,
	opts?: RequestInit
): Promise<ResponseAndRedirect> {
	try {
		const resp = await fetch(url, opts)
		switch (resp.status) {
			case 401:
				return { Response: undefined, RedirectTo: '/login' }
			case 404:
				return { Response: undefined, RedirectTo: '/404' }
			case 429:
				return { Response: undefined, RedirectTo: '/rate-limit' }
			case 500:
				return { Response: undefined, RedirectTo: '/500' }
			default:
				return { Response: resp, RedirectTo: undefined }
		}
	} catch {
		return { Response: undefined, RedirectTo: '/404' }
	}
}
