import { effect, useSignal } from '@preact/signals'
import { useState } from 'preact/hooks'
import type { Period } from '../types'
import SearchCat from './SearchCats'
import './SearchFilters.css'
import SearchPeriod from './SearchPeriod'

interface Props {
	InitialCats: string[] | undefined
	InitialPeriod: Period
}

export default function SearchFilters(props: Props) {
	const [period, set_period] = useState<Period>(props.InitialPeriod)
	const [cats, set_cats] = useState<string[]>(props.InitialCats ?? [])

	// set search URL based on period and cats
	const base_URL = `/top`
	let search_URL = base_URL
	if (cats.length) {
		search_URL += `?cats=${cats.join(',')}`
	}
	if (period !== 'all') {
		if (cats.length) {
			search_URL += `&period=${period}`
		} else {
			search_URL += `?period=${period}`
		}
	}

	// pass added/deleted_cat signals to allow modifying cats state in SearchCat.tsx
	const added_cat = useSignal<string | undefined>(undefined)
	const deleted_cat = useSignal<string | undefined>(undefined)

	// pass changed_period to SearchPeriod.tsx to allow modifying period state in SearchFilters.tsx
	const changed_period = useSignal<Period>(props.InitialPeriod)

	// Check for added cat, set state accordingly
	effect(() => {
		if (added_cat.value?.length) {
			const new_cat = added_cat.value
			set_cats((c) => [...c, new_cat])
			added_cat.value = undefined
		} else if (deleted_cat.value) {
			set_cats((c) => c.filter((cat) => cat !== deleted_cat.value))
			deleted_cat.value = undefined
		} else if (changed_period.value) {
			set_period(changed_period.value)
		}
	})

	return (
		<section id='search-filters'>
			<form>
				<h2>Search Filters</h2>

				<SearchPeriod
					SelectedPeriod={period}
					SetPeriodSignal={changed_period}
				/>
				<SearchCat
					InitialCats={props.InitialCats ?? []}
					AddedSignal={added_cat}
					DeletedSignal={deleted_cat}
				/>

				<a id='search-from-filters' href={search_URL}>
					Search
				</a>
			</form>
		</section>
	)
}
