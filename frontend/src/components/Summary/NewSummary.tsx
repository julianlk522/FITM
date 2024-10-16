import { useState } from 'preact/hooks'
import { SUMMARIES_ENDPOINT } from '../../constants'
import { is_error_response } from '../../types'
import fetch_with_handle_redirect from '../../util/fetch_with_handle_redirect'
import { save_path_then_redirect_to_login } from '../../util/login_redirect'
import './NewSummary.css'

interface Props {
	Token: string
	LinkID: string
}

export default function NewSummary(props: Props) {
	const { Token: token, LinkID: link_id } = props

	const [error, set_error] = useState<string | undefined>(undefined)

	const expected_submit_status = 201
	async function handle_submit(event: SubmitEvent, token: string) {
		event.preventDefault()

		if (!token) {
			save_path_then_redirect_to_login()
		}

		const form = event.target as HTMLFormElement
		const formData = new FormData(form)
		const summary = formData.get('summary')

		const new_summary_resp = await fetch_with_handle_redirect(
			SUMMARIES_ENDPOINT,
			{
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${token}`,
				},
				body: JSON.stringify({
					link_id: link_id,
					text: summary,
				}),
			}
		)
		if (!new_summary_resp.Response || new_summary_resp.RedirectTo) {
			return (window.location.href =
				new_summary_resp.RedirectTo ?? '/500')
		} else if (
			new_summary_resp.Response.status !== expected_submit_status
		) {
			const new_summary_data = await new_summary_resp.Response.json()
			if (is_error_response(new_summary_data)) {
				set_error(new_summary_data.error)
				return
			}
			return console.error('Whoops: ', new_summary_data)
		}

		return window.location.reload()
	}

	return (
		<form onSubmit={async (e) => await handle_submit(e, token)}>
			{error ? <p class='error'>{`Error: ${error}`}</p> : null}

			<label for='summary'>Add New Summary</label>
			<input type='text' id='summary' name='summary' />
			<button id='submit-new-summary' type='submit'>
				Submit
			</button>
		</form>
	)
}
