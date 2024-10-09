import { useState } from 'preact/hooks'
import { TMAP_ABOUT_ENDPOINT } from '../../constants'
import { is_error_response } from '../../types'
import fetch_with_handle_redirect from '../../util/fetch_with_handle_redirect'
import AboutText from './AboutText'
import './EditAbout.css'

interface Props {
	Initial: string
	Token?: string
}

export default function EditAbout(props: Props) {
	const { Initial: initial, Token: token } = props
	const abbreviated =
		initial.length > 200 ? `${initial.slice(0, 200)}...` : undefined

	const [editing, set_editing] = useState<boolean>(false)
	const [error, set_error] = useState<string | undefined>(undefined)

	function handle_finished_editing(event: SubmitEvent) {
		event.preventDefault()

		const form = event.target as HTMLFormElement
		const formData = new FormData(form)
		let about = formData.get('about')?.toString()

		if (about === initial || (!about && !initial)) {
			set_editing(false)
			set_error(undefined)
			return
		} else if (!about) {
			about = ''
		}

		update_about(about)
	}

	async function update_about(about: string) {
		if (!token) {
			return (window.location.href = '/login')
		}

		const edit_about_resp = await fetch_with_handle_redirect(
			TMAP_ABOUT_ENDPOINT,
			{
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${token}`,
				},
				body: JSON.stringify({ about }),
			}
		)
		if (!edit_about_resp.Response || edit_about_resp.RedirectTo) {
			return (window.location.href = edit_about_resp.RedirectTo ?? '/500')
		}

		const data = await edit_about_resp.Response.json()

		if (is_error_response(data)) {
			set_error(data.error)
		} else {
			window.location.reload()
		}
	}

	return (
		<div id='profile-about'>
			{editing ? (
				<form onSubmit={(event) => handle_finished_editing(event)}>
					<label for='about'>about: </label>
					<textarea name='about' cols={100} rows={8}>
						{initial}
					</textarea>
					<button
						id='confirm-changes'
						title='Save changes to your Treasure Map about section'
						class='img-btn'
						type='submit'
						value='Submit'
					>
						<img
							src='../../../confirm.svg'
							height={24}
							width={24}
							alt='Save Changes'
						/>
					</button>
				</form>
			) : (
				<>
					{abbreviated ? (
						<details>
							<summary>
								<AboutText About={abbreviated} />
							</summary>
							<AboutText About={initial} />
						</details>
					) : (
						<AboutText About={initial} />
					)}
					<button
						id='edit-about-btn'
						title='Edit about'
						alt='Edit about'
						class='img-btn'
						onClick={() => {
							set_editing(true)
						}}
					>
						<img
							src='../../../edit.svg'
							height={20}
							width={20}
							alt='Toggle about section editing'
						/>
					</button>
				</>
			)}

			{error ? <p class='error'>{error}</p> : null}
		</div>
	)
}
