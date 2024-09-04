import { effect, useSignal } from '@preact/signals'
import { useState } from 'preact/hooks'
import { TAGS_ENDPOINT } from '../../constants'
import type { Tag } from '../../types'
import { is_error_response } from '../../types'
import { format_long_date } from '../../util/format_date'
import './EditTag.css'
import TagCat from './TagCat'
interface Props {
	LinkID: string
	Token: string | undefined
	UserTag: Tag | undefined
}

export default function EditTag(props: Props) {
	const { LinkID: link_id, Token: token, UserTag: tag } = props
	const initial_cats = tag ? tag.Cats.split(',') : []

	const [cats, set_cats] = useState<string[]>(initial_cats)

	// Pass deleted_cat signal to children TagCategory.tsx
	// to allow removing their cat in this EditTag.tsx parent
	const deleted_cat = useSignal<string | undefined>(undefined)

	// Check for deleted cat and set cats accordingly
	effect(() => {
		if (deleted_cat.value) {
			set_cats((c) => c.filter((cat) => cat !== deleted_cat.value))
			deleted_cat.value = undefined
		}
	})

	const [editing, set_editing] = useState(false)
	const [error, set_error] = useState<string | undefined>(undefined)

	function add_cat(event: SubmitEvent) {
		event.preventDefault()

		const form = event.target as HTMLFormElement
		const formData = new FormData(form)
		const cat = formData.get('cat')?.toString()

		if (!cat) {
			set_error('Input is empty')
			return
		}

		if (cats.includes(cat)) {
			set_error('Cat already added')
			return
		}

		set_cats([...cats, cat])
		set_error(undefined)

		const cat_field = document.getElementById('cat') as HTMLInputElement
		cat_field.value = ''
		return
	}

	async function confirm_changes() {
		if (!token) {
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
		} else {
			window.location.reload()
		}
	}

	return (
		<section id='edit-tag'>
			<div id='user_tags_title_bar'>
				<h2>Your Tag</h2>

				<button
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
			</div>

			{error ? <p class='error'>{`Error: ${error}`}</p> : null}

			{tag || (editing && cats.length) ? (
				<ol id='cat-list'>
					{cats.map((cat) => (
						<TagCat
							Cat={cat}
							EditActivated={editing}
							Deleted={deleted_cat}
						/>
					))}
				</ol>
			) : null}

			{editing ? (
				<form id='edit_tag_form' onSubmit={(event) => add_cat(event)}>
					<label for='cat'>Cat</label>
					<input type='text' id='cat' name='cat' />
					<input type='Submit' value='Add Cat' />
				</form>
			) : null}

			{tag ? (
				<p>(last updated: {format_long_date(tag.LastUpdated)})</p>
			) : editing ? null : (
				<p>(not tagged)</p>
			)}
		</section>
	)
}
