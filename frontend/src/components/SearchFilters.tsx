import { effect, useSignal } from '@preact/signals'
import { useState } from 'preact/hooks'
import type { Period, SortMetric } from '../types'
import SearchCat from './SearchCats'
import './SearchFilters.css'
import SearchPeriod from './SearchPeriod'
import SearchSortBy from './SearchSortBy'

interface Props {
	InitialPeriod: Period
	InitialSortBy: SortMetric
	InitialCats: string[] | undefined
}

export default function SearchFilters(props: Props) {
	const [period, set_period] = useState<Period>(props.InitialPeriod)
	const [sort_by, set_sort_by] = useState<SortMetric>(props.InitialSortBy)
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

	if (sort_by !== 'rating') {
		if (cats.length || period !== 'all') {
			search_URL += `&sort_by=${sort_by}`
		} else {
			search_URL += `?sort_by=${sort_by}`
		}
	}

	// pass changed_period to SearchPeriod.tsx to allow modifying period state in SearchFilters.tsx
	const changed_period = useSignal<Period>(props.InitialPeriod)

	// pass changed_sort_by to SearchSortBy.tsx to allow modifying sort_by state in SearchFilters.tsx
	const changed_sort_by = useSignal<SortMetric>(props.InitialSortBy)

	// pass added/deleted_cat signals to allow modifying cats state in SearchCat.tsx
	const added_cat = useSignal<string | undefined>(undefined)
	const deleted_cat = useSignal<string | undefined>(undefined)

	// Check for update period / sort_by / cats, set state accordingly
	effect(() => {
		if (changed_period.value) {
			set_period(changed_period.value)
		}
		if (changed_sort_by.value) {
			set_sort_by(changed_sort_by.value)
		}

		if (added_cat.value?.length) {
			const new_cat = added_cat.value
			set_cats((c) => [...c, new_cat])
			added_cat.value = undefined
		} else if (deleted_cat.value) {
			set_cats((c) => c.filter((cat) => cat !== deleted_cat.value))
			deleted_cat.value = undefined
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

				<SearchSortBy
					SelectedSortBy={sort_by}
					SetSortBySignal={changed_sort_by}
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
