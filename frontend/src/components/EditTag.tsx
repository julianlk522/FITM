import { effect, useSignal } from '@preact/signals'
import { useState } from 'preact/hooks'
import type { Tag } from '../types'
import { is_error_response } from '../types'
import format_date from '../util/format_date'
import './EditTag.css'
import TagCategory from './TagCategory'
interface Props {
	LinkID: string
	Token: string | undefined
	UserTag: Tag | undefined
}

export default function EditTag(props: Props) {
	const { LinkID: link_id, Token: token, UserTag: tag } = props
	const initial_cats = tag ? tag.Categories.split(',') : []

	const [categories, set_categories] = useState<string[]>(initial_cats)

	// Pass deleted_cat signal to children TagCategory.tsx
	// to allow removing their category in this EditTag.tsx parent
	const deleted_cat = useSignal<string | undefined>(undefined)

	// Check for deleted category and set categories accordingly
	effect(() => {
		if (deleted_cat.value) {
			set_categories((c) => c.filter((cat) => cat !== deleted_cat.value))
			deleted_cat.value = undefined
		}
	})

	const [editing, set_editing] = useState(false)
	const [error, set_error] = useState<string | undefined>(undefined)

	function add_category(event: SubmitEvent) {
		event.preventDefault()
		2
		const form = event.target as HTMLFormElement
		const formData = new FormData(form)
		const category = formData.get('category')?.toString()

		if (!category) {
			set_error('Missing category')
			return
		}

		if (categories.includes(category)) {
			set_error('Category already added')
			return
		}

		set_categories([...categories, category])
		set_error(undefined)

		const cat_field = document.getElementById(
			'category'
		) as HTMLInputElement
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
			resp = await fetch('http://127.0.0.1:8000/tags', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${token}`,
				},
				body: JSON.stringify({
					link_id: link_id,
					categories: categories.join(','),
				}),
			})

			// edit tag
		} else {
			resp = await fetch('http://127.0.0.1:8000/tags', {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${token}`,
				},
				body: JSON.stringify({
					tag_id: tag.ID,
					categories: categories.join(','),
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
				<h2>Your Tag For This Link</h2>

				<button
					onClick={() => {
						set_categories(categories.sort())

						// update if changes detected, else skip
						if (
							editing &&
							(categories.length !== initial_cats.length ||
								categories.some(
									(c, i) => c !== initial_cats[i]
								))
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

			<ol id='cat_list'>
				{categories.map((cat) => (
					<TagCategory
						Category={cat}
						EditActivated={editing}
						Deleted={deleted_cat}
					/>
				))}
			</ol>

			{editing ? (
				<form
					id='edit_tag_form'
					onSubmit={(event) => add_category(event)}
				>
					<label for='category'>Category</label>
					<input type='text' id='category' name='category' />
					<input type='Submit' value='Add Category' />
				</form>
			) : null}

			{tag ? (
				<p>(Last Updated: {format_date(tag.LastUpdated)})</p>
			) : editing ? null : (
				<p>(Not tagged.)</p>
			)}
		</section>
	)
}
