import { useState } from 'preact/hooks'
import { TAGS_ENDPOINT } from '../../constants'
import type { Tag } from '../../types'
import { is_error_response } from '../../types'
import { format_long_date } from '../../util/format_date'
import SearchCats from '../Search/Cats'
import './EditTag.css'
interface Props {
	LinkID: string
	OnlyTag: boolean
	UserTag?: Tag
	Token?: string
}

export default function EditTag(props: Props) {
	const {
		LinkID: link_id,
		OnlyTag: only_tag,
		UserTag: tag,
		Token: token,
	} = props
	const initial_cats = tag ? tag.Cats.split(',') : []

	const [cats, set_cats] = useState<string[]>(initial_cats)
	const [editing, set_editing] = useState(false)
	const [error, set_error] = useState<string | undefined>(undefined)
	const [show_delete_modal, set_show_delete_modal] = useState(false)

	async function confirm_changes() {
		if (!token) {
			document.cookie = `redirect_to=${window.location.pathname.replaceAll(
				'/',
				'%2F'
			)}; path=/login; max-age=14400; SameSite=strict; Secure`
			return (window.location.href = '/login')
		}

		let resp: Response

		// new tag
		if (!tag) {
			resp = await fetch(TAGS_ENDPOINT, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${token}`,
				},
				body: JSON.stringify({
					link_id: link_id,
					cats: cats.join(','),
				}),
			})

			// edit tag
		} else {
			resp = await fetch(TAGS_ENDPOINT, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${token}`,
				},
				body: JSON.stringify({
					tag_id: tag.ID,
					cats: cats.join(','),
				}),
			})
		}

		if (resp.status !== 200 && resp.status !== 201) {
			console.error(resp)
		}

		const edit_tag_data = await resp.json()

		if (is_error_response(edit_tag_data)) {
			set_error(edit_tag_data.error)
			return
		}

		// reload if successful
		window.location.reload()
	}

	async function handle_delete() {
		if (!tag) {
			return
		}

		if (!token) {
			document.cookie = `redirect_to=${window.location.pathname.replaceAll(
				'/',
				'%2F'
			)}; path=/login; max-age=14400; SameSite=strict; Secure`
			return (window.location.href = '/login')
		}

		const delete_resp = await fetch(TAGS_ENDPOINT, {
			method: 'DELETE',
			headers: {
				'Content-Type': 'application/json',
				Authorization: `Bearer ${token}`,
			},
			body: JSON.stringify({ tag_id: tag.ID }),
		})
		if (is_error_response(delete_resp) || delete_resp.status !== 204) {
			console.error(delete_resp)
			const err_msg = await delete_resp.json()
			set_error(err_msg.error)
			return
		}

		// reload if successful
		window.location.reload()
	}

	return (
		<form id='edit-tag' onSubmit={(e) => e.preventDefault()}>
			<div id='user-tags-header-bar'>
				<h2>Your Tag</h2>

				<button
					title={editing ? 'confirm tag changes' : 'edit tag'}
					onClick={() => {
						set_cats(cats.sort())

						if (
							editing &&
							(cats.length !== initial_cats.length ||
								cats.some((c, i) => c !== initial_cats[i]))
						) {
							confirm_changes()
						}
						set_editing((e) => !e)
					}}
					class='img-btn'
				>
					<img
						src={
							editing
								? '../../../check2-circle.svg'
								: '../../../bi-feather.svg'
						}
						height={20}
						width={20}
						alt={editing ? 'Confirm Edits' : 'Edit Tag'}
					/>
				</button>

				{editing && !only_tag ? (
					<button
						title='delete your tag'
						class='delete-tag-btn img-btn'
						onClick={() => set_show_delete_modal(true)}
					>
						<img src='../../../x-lg.svg' height={20} width={20} />
					</button>
				) : null}
			</div>

			{error ? <p class='error'>{error}</p> : null}

			{tag || editing ? (
				<SearchCats
					AbbreviatedCatsText
					SelectedCats={cats}
					SetSelectedCats={set_cats}
					Addable={editing}
					Removable={editing}
				/>
			) : null}

			{tag ? (
				<p>last updated: {format_long_date(tag.LastUpdated)}</p>
			) : editing ? null : (
				<p>(not tagged)</p>
			)}

			{show_delete_modal ? (
				<dialog class='delete-tag-modal' open>
					<p>Delete your tag?</p>
					<button onClick={handle_delete}>Yes</button>
					<button
						autofocus
						onClick={() => set_show_delete_modal(false)}
					>
						Cancel
					</button>
				</dialog>
			) : null}
		</form>
	)
}
