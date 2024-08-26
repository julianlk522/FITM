import { useState } from 'preact/hooks'
import { SUMMARIES_ENDPOINT } from '../../constants'
import { is_error_response } from '../../types'

interface Props {
	Token: string
	LinkID: string
}

export default function NewSummary(props: Props) {
	const { Token: token, LinkID: link_id } = props

	const [error, set_error] = useState<string | undefined>(undefined)

	async function handle_submit(event: SubmitEvent, token: string) {
		event.preventDefault()

		const form = event.target as HTMLFormElement
		const formData = new FormData(form)
		const summary = formData.get('summary')

		const new_summary_resp = await fetch(SUMMARIES_ENDPOINT, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				Authorization: `Bearer ${token}`,
			},
			body: JSON.stringify({
				link_id: link_id,
				text: summary,
			}),
		})
		if (new_summary_resp.statusText === 'Unauthorized') {
			document.cookie = `redirect_to=${window.location.pathname.replaceAll(
				'/',
				'%2F'
			)}; path=/login; max-age=21600; SameSite=strict; Secure`
			window.location.href = '/login'
		}
		let new_summary_data = await new_summary_resp.json()

		if (is_error_response(new_summary_data)) {
			set_error(new_summary_data.error)
			return
		} else {
			form.reset()
			window.location.reload()
		}

		return
	}

	return (
		<form onSubmit={async (e) => await handle_submit(e, token)}>
			{error ? <p class='error'>{`Error: ${error}`}</p> : null}

			<label for='summary'>Add New Summary</label>
			<input type='text' id='summary' name='summary' />
			<button type='submit'>Submit</button>
		</form>
	)
}
