import { effect, useSignal } from '@preact/signals'
import { useState } from 'preact/hooks'
import type { Period, SortMetric } from '../../types'
import SearchCats from './Cats'
import './Filters.css'
import SearchNSFW from './NSFW'
import SearchPeriod from './Period'
import SearchSortBy from './SortBy'

interface Props {
	InitialPeriod: Period
	InitialSortBy: SortMetric
	InitialCats: string[]
	InitialNSFW: boolean
}

export default function SearchFilters(props: Props) {
	const [period, set_period] = useState<Period>(props.InitialPeriod)
	const [sort_by, set_sort_by] = useState<SortMetric>(props.InitialSortBy)
	const [cats, set_cats] = useState<string[]>(props.InitialCats)
	const [nsfw, set_nsfw] = useState<boolean>(props.InitialNSFW)

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
	if (nsfw) {
		if (cats.length || period !== 'all' || sort_by !== 'rating') {
			search_URL += `&nsfw=true`
		} else {
			search_URL += `?nsfw=true`
		}
	}

	// Check for update period / sort_by, set state accordingly
	const changed_period = useSignal<Period>(props.InitialPeriod)
	const changed_sort_by = useSignal<SortMetric>(props.InitialSortBy)
	effect(() => {
		if (changed_period.value) {
			set_period(changed_period.value)
		}
		if (changed_sort_by.value) {
			set_sort_by(changed_sort_by.value)
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

				<SearchCats SelectedCats={cats} SetSelectedCats={set_cats} />

				<SearchNSFW NSFW={nsfw} SetNSFW={set_nsfw} />

				<a id='search-from-filters' href={search_URL}>
					Search
				</a>
			</form>
		</section>
	)
}
