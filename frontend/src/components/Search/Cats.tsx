import { effect, useSignal } from '@preact/signals'
import {
	useCallback,
	useEffect,
	useRef,
	useState,
	type Dispatch,
	type StateUpdater,
} from 'preact/hooks'
import { CATS_ENDPOINT } from '../../constants'
import * as types from '../../types'
import { type CatCount } from '../../types'
import TagCat from '../Tag/TagCat'
import './Cats.css'

interface Props {
	AbbreviatedCatsText?: boolean
	Addable?: boolean
	Removable?: boolean
	SelectedCats: string[]
	SetSelectedCats: Dispatch<StateUpdater<string[]>>
	SubmittedLinks?: types.Link[]
}

export default function SearchCats(props: Props) {
	const {
		AbbreviatedCatsText: abbreviated_cats_text,
		Removable: removable,
		SelectedCats: selected_cats,
		SetSelectedCats: set_selected_cats,
	} = props
	const addable = props.Addable ?? true

	const [snippet, set_snippet] = useState<string>('')
	const [recommended_cats, set_recommended_cats] = useState<
		CatCount[] | undefined
	>(undefined)
	const [error, set_error] = useState<string | undefined>(undefined)

	const MIN_SNIPPET_CHARS = 2
	const search_snippet_recommendations = useCallback(async () => {
		if (!snippet || snippet.length < MIN_SNIPPET_CHARS) {
			set_recommended_cats(undefined)
			return
		}

		let spellfix_matches_url = CATS_ENDPOINT + `/${snippet}`
		if (selected_cats.length) {
			spellfix_matches_url += `?omitted=${selected_cats.join(',')}`
		}

		try {
			const spellfix_matches_resp = await fetch(spellfix_matches_url)
			if (!spellfix_matches_resp.ok) {
				const msg: types.ErrorResponse =
					await spellfix_matches_resp.json()
				set_error(msg.error)
				throw new Error('API request failed')
			}

			const spellfix_matches: CatCount[] =
				await spellfix_matches_resp.json()
			set_recommended_cats(spellfix_matches)
		} catch (error) {
			set_recommended_cats([])
			set_error(error instanceof Error ? error.message : String(error))
		}
	}, [snippet])

	const timeout_ref = useRef<number | null>(null)
	const DEBOUNCE_INTERVAL = 500
	useEffect(() => {
		// refresh debounce interval if searching
		if (snippet?.length >= MIN_SNIPPET_CHARS) {
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

	const handle_enter = (event: KeyboardEvent) => {
		if (event.key === 'Enter') {
			event.preventDefault()
			add_cat(event)
		}
	}

	function add_cat(event: Event) {
		event.preventDefault()
		if (!snippet) {
			set_error('Input is empty')
			return
		}

		let new_cat = snippet
		// correct 'nsfw' to 'NSFW'
		if (snippet === 'nsfw') {
			new_cat = 'NSFW'
		}

		if (selected_cats.includes(new_cat)) {
			set_error('Already added')
			return
		}

		set_selected_cats((prev) =>
			[...prev, new_cat].sort((a, b) => a.localeCompare(b))
		)

		set_error(undefined)
		set_recommended_cats((prev) =>
			prev?.filter((cat) => cat.Category !== new_cat)
		)
	}

	// Pass added_cat / deleted_cat signals to children TagCat.tsx
	// to allow adding recommended cats / removing selected cats here
	const added_cat = useSignal<string | undefined>(undefined)
	const deleted_cat = useSignal<string | undefined>(undefined)

	// Listen for add / delete cat signals from TagCat
	effect(() => {
		if (added_cat.value?.length) {
			const new_cat = added_cat.value
			set_selected_cats((c) =>
				[...c, new_cat].sort((a, b) => {
					return a.localeCompare(b)
				})
			)
			set_recommended_cats((c) =>
				c?.filter((cat) => cat.Category !== new_cat)
			)
			added_cat.value = undefined

			set_error(undefined)
		} else if (deleted_cat.value) {
			const to_delete = deleted_cat.value
			set_selected_cats((c) => c.filter((cat) => cat !== to_delete))
			deleted_cat.value = undefined

			// prevent weird case where deleting a hidden recommended cat causes it to suddenly appear
			set_recommended_cats((c) =>
				c?.filter((cat) => cat.Category !== to_delete)
			)

			set_error(undefined)
		}
	})

	useEffect(() => {
		if (props.SubmittedLinks && props.SubmittedLinks.length) {
			set_snippet('')
		}
	}, [props.SubmittedLinks])

	return (
		<div id='search-cats-container'>
			{addable ? (
				<>
					<label id='search-cats' for='cats'>
						{abbreviated_cats_text ? 'Cats:' : 'Tag Cats:'}
					</label>
					<input
						type='text'
						name='cats'
						id='cats'
						onInput={(event) =>
							set_snippet(
								(event.target as HTMLInputElement).value
							)
						}
						onKeyPress={handle_enter}
						value={snippet}
					/>
					<input
						id='add-cat-filter'
						title='Add cat'
						type='submit'
						value='Add'
						onClick={add_cat}
					/>
				</>
			) : null}

			{selected_cats.length ? (
				<ol id='cat-list'>
					{selected_cats.map((cat) => (
						<TagCat
							key={cat}
							Cat={cat}
							Count={undefined}
							Removable={removable ?? true}
							Addable={false}
							AddedSignal={undefined}
							DeletedSignal={deleted_cat}
						/>
					))}
				</ol>
			) : null}

			{recommended_cats?.length ? (
				<ul id='recommendations-list'>
					{recommended_cats
						.filter((rc) => !selected_cats.includes(rc.Category))
						.map((cat) => (
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
		</div>
	)
}
