import { effect, useSignal } from '@preact/signals'
import { useCallback, useEffect, useRef, useState } from 'preact/hooks'
import { CATS_ENDPOINT } from '../constants'
import * as types from '../types'
import { type CatCount } from '../types'
import './CatSearch.css'
import TagCat from './Tag/TagCat'

export default function CatSearch() {
	const [snippet, set_snippet] = useState<string>('')
	const [populated_cats, set_populated_cats] = useState<string[]>([])
	const [recommended_cats, set_recommended_cats] = useState<
		CatCount[] | undefined
	>(undefined)
	const [error, set_error] = useState<string | undefined>(undefined)

	const timeout_ref = useRef<number | null>(null)
	const DEBOUNCE_INTERVAL = 500

	const search_snippet_recommendations = useCallback(async () => {
		if (!snippet || snippet.length < 2) {
			set_recommended_cats(undefined)
			return
		}

		try {
			const spellfix_matches_resp = await fetch(
				CATS_ENDPOINT + `/${snippet}`
			)
			if (!spellfix_matches_resp.ok) {
				const msg: types.ErrorResponse =
					await spellfix_matches_resp.json()
				set_error(msg.error)
				throw new Error('API request failed')
			}

			const spellfix_matches = await spellfix_matches_resp.json()
			set_recommended_cats(spellfix_matches)
			console.log('spellfix_matches', spellfix_matches)
		} catch (error) {
			set_recommended_cats([])
			set_error(error instanceof Error ? error.message : String(error))
		}
	}, [snippet])

	useEffect(() => {
		// refresh debounce interval if searching
		if (snippet?.length >= 2) {
			timeout_ref.current = window.setTimeout(() => {
				search_snippet_recommendations()
			}, DEBOUNCE_INTERVAL)

			// or clear recommendations if empty input
		} else {
			set_recommended_cats(undefined)
		}

		// cleanup: clear any old debounce interval
		return () => {
			if (timeout_ref.current) {
				window.clearTimeout(timeout_ref.current)
			}
		}
	}, [search_snippet_recommendations])

	function handle_input(event: Event) {
		const value = (event.target as HTMLInputElement).value
		set_snippet(value)
	}

	function add_cat(event: Event) {
		event.preventDefault()

		if (!snippet) {
			set_error('Input is empty')
			return
		}

		if (populated_cats.includes(snippet)) {
			set_error('Cat already added')
			return
		}

		set_populated_cats((prev) =>
			[...prev, snippet].sort((a, b) => a.localeCompare(b))
		)
		set_error(undefined)
		set_recommended_cats((prev) =>
			prev?.filter((cat) => cat.Category !== snippet)
		)
	}

	// Pass added_cat / deleted_cat signals to children TagCategory.tsx
	// to allow adding / removing them here
	const added_cat = useSignal<string | undefined>(undefined)
	const deleted_cat = useSignal<string | undefined>(undefined)

	// Check for added and deleted <TagCat />s accordingly
	// (cats can be added by clicking one of the recommendations)
	effect(() => {
		if (added_cat.value?.length) {
			const new_cat = added_cat.value
			set_populated_cats((c) =>
				[...c, new_cat].sort((a, b) => {
					return a.localeCompare(b)
				})
			)
			set_recommended_cats((c) =>
				c?.filter((cat) => cat.Category !== new_cat)
			)
			added_cat.value = undefined
		} else if (deleted_cat.value) {
			set_populated_cats((c) =>
				c.filter((cat) => cat !== deleted_cat.value)
			)
			deleted_cat.value = undefined
		}
	})

	return (
		<>
			<form id='cat-search-form' onSubmit={add_cat}>
				<label id='cat-label' for='cat'>
					Cats:
				</label>
				<br />
				<input
					type='text'
					name='cat-search'
					id='cat-search'
					onInput={handle_input}
					value={snippet}
				/>
				<input type='submit' value='Add Cat' />
				{populated_cats?.length ? (
					<a href={`/top?cats=${populated_cats.join(',')}`}>search</a>
				) : null}
			</form>

			{populated_cats.length ? (
				<ol id='cat-list'>
					{populated_cats.map((cat) => (
						<TagCat
							key={cat}
							Cat={cat}
							Count={undefined}
							Removable={true}
							Addable={false}
							AddedSignal={undefined}
							DeletedSignal={deleted_cat}
						/>
					))}
				</ol>
			) : null}

			{/* cat / rank list */}
			{recommended_cats?.length ? (
				<ul id='recommendations-list'>
					{recommended_cats.map((cat) => (
						<TagCat
							key={cat}
							Cat={cat.Category}
							Count={cat.Count}
							Removable={false}
							Addable={true}
							AddedSignal={added_cat}
							DeletedSignal={undefined}
						/>
					))}
				</ul>
			) : null}

			{error ? <p class='error'>{error}</p> : null}
		</>
	)
}
