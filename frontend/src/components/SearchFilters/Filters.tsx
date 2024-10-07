import { useState } from 'preact/hooks'
import type { Period, SortMetric } from '../../types'
import SearchCats from './Cats'
import './Filters.css'
import SearchNSFW from './NSFW'
import SearchPeriod from './Period'
import SearchSortBy from './SortBy'

interface Props {
	Endpoint?: '/top' | '/more'
	InitialPeriod: Period
	InitialSortBy?: SortMetric
	InitialNSFW?: boolean
	InitialCats: string[]
}

export default function SearchFilters(props: Props) {
	let endpoint = props.Endpoint ?? '/top'

	const [period, set_period] = useState<Period>(props.InitialPeriod)
	const [cats, set_cats] = useState<string[]>(props.InitialCats)

	// only for links (/top)
	const [sort_by, set_sort_by] = useState<SortMetric>(
		props.InitialSortBy ?? 'rating'
	)
	const [nsfw, set_nsfw] = useState<boolean>(props.InitialNSFW ?? false)

	// set search URL based on period and cats
	const base_URL = endpoint ?? '/top'
	let search_URL = base_URL

	if (cats.length) {
		// encode reserved chars
		const encoded_cats = cats
			.map((cat) => encodeURIComponent(cat))
			.join(',')
		search_URL += `?cats=${encoded_cats}`
	}
	if (period !== 'all') {
		if (cats.length) {
			search_URL += `&period=${period}`
		} else {
			search_URL += `?period=${period}`
		}
	}

	// /top page endpoint only
	if (sort_by && sort_by !== 'rating') {
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

	return (
		<section id='search-filters'>
			<form>
				<h2>Search Filters</h2>

				<SearchPeriod Period={period} SetPeriod={set_period} />

				{/* only for links (/top) */}
				{endpoint === '/top' ? (
					<>
						<SearchSortBy
							SortBy={sort_by}
							SetSortBy={set_sort_by}
						/>
						<SearchNSFW NSFW={nsfw} SetNSFW={set_nsfw} />
					</>
				) : null}

				<SearchCats SelectedCats={cats} SetSelectedCats={set_cats} />

				<a id='search-from-filters' href={search_URL}>
					Search
				</a>
			</form>
		</section>
	)
}
