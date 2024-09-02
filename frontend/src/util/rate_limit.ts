export default async function fetch_with_handle_rate_limit(
	url: string,
	opts?: RequestInit
): Promise<Response | undefined> {
	try {
		const resp = await fetch(url, opts)
		if (resp.status === 429) {
			return undefined
		}
		return resp
	} catch {
		return undefined
	}
}
