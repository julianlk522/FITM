import { useState } from 'preact/hooks'
import { is_error_response } from '../../types'
import './EditAbout.css'

interface Props {
	initial: string
	token: string | undefined
}

export default function EditAbout(props: Props) {
	const { initial, token } = props

	const [editing, set_editing] = useState<boolean>(false)
	const [error, set_error] = useState<string | undefined>(undefined)

	async function update_about(event: SubmitEvent) {
		event.preventDefault()

		if (!token) {
			return (window.location.href = '/login')
		}

		const form = event.target as HTMLFormElement
		const formData = new FormData(form)
		const about = formData.get('about')?.toString()

		if (about === initial) {
			set_error('No changes made')
			return
		}

		const resp = await fetch('http://127.0.0.1:8000/users/about', {
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
		<div id='edit-about'>
			{editing ? (
				<form onSubmit={(event) => update_about(event)}>
					<label for='about'>About</label>
					<textarea name='about' cols={50} rows={1}>
						{initial}
					</textarea>
					<button class='img-btn' type='submit' value='Submit'>
						<img
							src='../../../check2-circle.svg'
							height={24}
							width={24}
							alt='Save Changes'
						/>
					</button>
				</form>
			) : initial ? (
				<figcaption>about: {initial}</figcaption>
			) : null}

			<button
				onClick={() => {
					set_error(undefined)
					set_editing((prev) => !prev)
				}}
				class='img-btn'
			>
				<img
					src='../../../edit_about.svg'
					height={24}
					width={24}
					alt='Toggle About Section edit mode'
				/>
			</button>

			{error ? <p class='error'>{error}</p> : null}
		</div>
	)
}
