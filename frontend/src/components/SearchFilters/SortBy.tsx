import type { ChangeEvent } from 'preact/compat'
import type { Dispatch, StateUpdater } from 'preact/hooks'
import { SortMetrics, type SortMetric } from '../../types'

interface Props {
	SortBy: SortMetric
	SetSortBy: Dispatch<StateUpdater<SortMetric>>
}

export default function SearchSortBy(props: Props) {
	const { SortBy: sort_by, SetSortBy: set_sort_by } = props

	async function handle_set_sort_by(e: ChangeEvent<HTMLSelectElement>) {
		set_sort_by(e.currentTarget.value as SortMetric)
	}
	return (
		<div>
			<label id='search-sort-by' for='sort-by'>
				Sort By:
			</label>
			<select
				name='sort-by'
				id='sort-by'
				defaultValue={sort_by}
				value={sort_by}
				onChange={handle_set_sort_by}
			>
				{SortMetrics.map((met) => (
					<option value={met}>{met}</option>
				))}
			</select>
		</div>
	)
}
