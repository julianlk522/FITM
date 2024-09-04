import { effect, useSignal } from '@preact/signals'
import type { Dispatch, StateUpdater } from 'preact/hooks'
import './NewTag.css'
import TagCat from './TagCat'

interface Props {
	Cats: string[]
	SetCats: Dispatch<StateUpdater<string[]>>
	SetError: Dispatch<StateUpdater<string | undefined>>
}

export default function NewTag(props: Props) {
	const { Cats: cats, SetCats: set_cats, SetError: set_error } = props

	// Pass deleted_cat signal to children TagCategory.tsx
	// to allow removing their cat in this NewTag.tsx parent
	const deleted_cat = useSignal<string | undefined>(undefined)

	// Check for deleted cat and set cats accordingly
	effect(() => {
		if (deleted_cat.value) {
			set_cats((c) => c.filter((cat) => cat !== deleted_cat.value))
			deleted_cat.value = undefined
		}
	})

	function add_tag_cat(event: MouseEvent) {
		event.preventDefault()

		// @ts-ignore
		const form = event.target.form as HTMLFormElement
		if (!form) return set_error('Form not found')
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

		set_cats(
			[...cats, cat].sort((a, b) => {
				return a.localeCompare(b)
			})
		)
		set_error(undefined)

		const cat_field = document.getElementById('cat') as HTMLInputElement
		cat_field.value = ''
		return
	}

	return (
		<>
			<label for='cat'>Tag</label>
			<input type='text' id='cat' name='cat' />
			<button id='add-cat' onClick={(event) => add_tag_cat(event)}>
				Add Cat
			</button>

			<ol id='cat-list'>
				{cats.map((cat) => (
					<TagCat
						Cat={cat}
						EditActivated={true}
						Deleted={deleted_cat}
					/>
				))}
			</ol>
		</>
	)
}
