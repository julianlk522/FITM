import { useEffect, useState } from 'preact/hooks'
import { TMAP_PFP_ENDPOINT } from '../../constants'
import './ProfilePic.css'

interface Props {
	LoginName: string
	PFP?: string
	SignedInUser?: string
	Token?: string
}

export default function ProfilePic(props: Props) {
	const {
		LoginName: login_name,
		PFP: pfp,
		SignedInUser: signed_in_user,
		Token: token,
	} = props
	const [url, set_url] = useState<string | undefined>(undefined)
	const [error, set_error] = useState<string | undefined>(undefined)
	const [show_delete_modal, set_show_delete_modal] = useState<boolean>(false)

	const is_signed_in_user = login_name === signed_in_user

	useEffect(() => {
		async function get_pfp() {
			if (!pfp) {
				return
			}
			const pfp_resp = await fetch(TMAP_PFP_ENDPOINT + `/${pfp}`, {
				headers: { 'Content-Type': 'image/png' },
			})
			if (pfp_resp.status > 399) {
				console.error(pfp_resp)
				return
			}
			const pfp_blob = await pfp_resp.blob()
			const pfp_url = URL.createObjectURL(pfp_blob)
			set_url(pfp_url)
		}

		get_pfp()
	}, [pfp])

	async function handle_pic_change(e: Event) {
		if (!token) {
			return (window.location.href = '/login')
		}
		const target = e.target as HTMLInputElement
		if (!target.files || target.files.length === 0) {
			return
		}

		const new_pic = target.files[0]
		let formData = new FormData()
		formData.append('pic', new_pic)

		const new_pic_resp = await fetch(TMAP_PFP_ENDPOINT, {
			method: 'POST',
			headers: { Authorization: `Bearer ${token}` },
			body: formData,
		})
		if (new_pic_resp.status > 399) {
			const data = await new_pic_resp.json()
			set_error(data.error)
			return
		}

		const new_pic_data = await new_pic_resp.blob()
		set_url(URL.createObjectURL(new_pic_data))
	}

	async function handle_delete(e: MouseEvent) {
		e.preventDefault()
		if (!token) {
			return (window.location.href = '/login')
		}

		const delete_resp = await fetch(TMAP_PFP_ENDPOINT, {
			method: 'DELETE',
			headers: { Authorization: `Bearer ${token}` },
		})
		if (delete_resp.status > 399) {
			const data = await delete_resp.json()
			set_error(data.error)
			return
		}

		set_url(undefined)
		set_show_delete_modal(false)
	}

	return (
		<div id='profile-pic'>
			{error ? <p class='error'>{error}</p> : null}
			{url ? (
				<img
					src={url}
					id='pfp'
					alt={`${login_name}'s profile picture`}
					width='150'
				/>
			) : null}
			{is_signed_in_user ? (
				<form>
					<label id='new-pic-upload-label' for='new-pic-upload'>
						<img
							title='Upload new profile picture'
							src='../../../upload-pfp.svg'
							height={24}
							width={24}
							alt='Upload New Profile Picture'
						/>
					</label>
					<input
						id='new-pic-upload'
						type='file'
						accept={'image/*'}
						hidden
						onChange={handle_pic_change}
					/>
					{url ? (
						<button
							title='Delete profile picture'
							id='delete-pfp-btn'
							class='img-btn'
							onClick={(e) => {
								e.preventDefault()
								set_show_delete_modal(true)
							}}
						>
							<img
								src='../../../x-lg.svg'
								height={22}
								width={22}
							/>
						</button>
					) : null}

					{show_delete_modal ? (
						<>
							{/* delete modal */}
							<dialog id='delete-profile-pic-modal' open>
								<p>Delete profile pic?</p>
								<button onClick={handle_delete}>Yes</button>
								<button
									autofocus
									onClick={() => set_show_delete_modal(false)}
								>
									Cancel
								</button>
							</dialog>
						</>
					) : null}
				</form>
			) : null}
		</div>
	)
}
