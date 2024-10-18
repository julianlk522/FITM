import { useState } from 'preact/hooks'
import { TAGS_ENDPOINT } from '../../constants'
import type { Tag } from '../../types'
import { is_error_response } from '../../types'
import fetch_with_handle_redirect from '../../util/fetch_with_handle_redirect'
import { format_long_date } from '../../util/format_date'
import Modal from '../Modal/Modal'
import SearchCats from '../SearchFilters/Cats'
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

		const request_method = tag ? 'PUT' : 'POST'
		const cats_str = cats.join(',')
		const payload = tag
			? { tag_id: tag.ID, cats: cats_str }
			: { link_id: link_id, cats: cats_str }
		const resp = await fetch_with_handle_redirect(TAGS_ENDPOINT, {
			method: request_method,
			headers: {
				'Content-Type': 'application/json',
				Authorization: `Bearer ${token}`,
			},
			body: JSON.stringify(payload),
		})
		if (!resp.Response || resp.RedirectTo) {
			return (window.location.href = resp.RedirectTo ?? '/500')
		}

		const expected_status = tag ? 200 : 201
		if (resp.Response.status !== expected_status) {
			console.error('WHOOPS: ', resp)

			const tag_data = await resp.Response.json()
			if (is_error_response(tag_data)) {
				return set_error(tag_data.error)
			}
		}

		// success: reload
		return window.location.reload()
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

		const delete_resp = await fetch_with_handle_redirect(TAGS_ENDPOINT, {
			method: 'DELETE',
			headers: {
				'Content-Type': 'application/json',
				Authorization: `Bearer ${token}`,
			},
			body: JSON.stringify({ tag_id: tag.ID }),
		})
		if (!delete_resp.Response || delete_resp.RedirectTo) {
			return (window.location.href = delete_resp.RedirectTo ?? '/500')
		}

		const expected_status = 204
		if (delete_resp.Response.status !== expected_status) {
			console.error('WHOOPS: ', delete_resp)

			const delete_tag_data = await delete_resp.Response.json()
			if (is_error_response(delete_tag_data)) {
				return set_error(delete_tag_data.error)
			}
		}

		// success: reload
		return window.location.reload()
	}

	return (
		<form id='edit-tag' onSubmit={(e) => e.preventDefault()}>
			<div id='user-tags-header-bar'>
				<h2>Your Tag</h2>

				<button
					title={editing ? 'Save tag changes' : 'Edit tag'}
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
								? '../../../confirm.svg'
								: '../../../edit.svg'
						}
						height={20}
						width={20}
						alt={editing ? 'Save Edits' : 'Edit Tag'}
					/>
				</button>

				{editing && !only_tag ? (
					<button
						title='Delete tag'
						alt='Delete Tag'
						class='delete-tag-btn img-btn'
						onClick={() => set_show_delete_modal(true)}
					>
						<img src='../../../delete.svg' height={20} width={20} />
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
				<Modal
					Prompt={'Delete your tag?'}
					IsDeleteConfirmation
					HandleDelete={handle_delete}
					SetShowModal={set_show_delete_modal}
				/>
			) : null}
		</form>
	)
}
