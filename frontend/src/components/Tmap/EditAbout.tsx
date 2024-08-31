import { useState } from 'preact/hooks'
import { TMAP_ABOUT_ENDPOINT } from '../../constants'
import { is_error_response } from '../../types'
import './EditAbout.css'

interface Props {
	initial: string
	token: string | undefined
}

export default function EditAbout(props: Props) {
	const { initial, token } = props
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

		const resp = await fetch(TMAP_ABOUT_ENDPOINT, {
			method: 'PUT',
			headers: {
				'Content-Type': 'application/json',
				Authorization: `Bearer ${token}`,
			},
			body: JSON.stringify({ about }),
		})

		const data = await resp.json()

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
						title='Save changes to your Treasure Map About section'
						class='img-btn'
						type='submit'
						value='Submit'
					>
						<img
							src='../../../check2-circle.svg'
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
								<pre>
									<span>about:</span> {abbreviated}
								</pre>
							</summary>
							<pre>
								<span>about:</span> {initial}
							</pre>
						</details>
					) : (
						<pre>
							<span>about:</span> {initial}
						</pre>
					)}
					<button
						id='edit-about-btn'
						title='Edit About section'
						alt='Edit About'
						class='img-btn'
						onClick={() => {
							set_editing(true)
						}}
					>
						<img
							src='../../../edit_about.svg'
							height={20}
							width={20}
							alt='Toggle About section edit mode'
						/>
					</button>
				</>
			)}

			{error ? <p class='error'>{error}</p> : null}
		</div>
	)
}
