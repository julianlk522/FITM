import type { Signal } from '@preact/signals'
import type { ChangeEvent } from 'preact/compat'
import { SortMetrics, type SortMetric } from '../../types'

interface Props {
	SelectedSortBy: SortMetric
	SetSortBySignal: Signal<SortMetric>
}

export default function SearchSortBy(props: Props) {
	const selected = props.SelectedSortBy

	async function handle_set_sort_by(e: ChangeEvent<HTMLSelectElement>) {
		props.SetSortBySignal.value = e.currentTarget.value as SortMetric
	}
	return (
		<div>
			<label id='search-sort-by' for='sort-by'>
				Sort By:
			</label>
			<select
				name='sort-by'
				id='sort-by'
				defaultValue={selected}
				value={selected}
				onChange={handle_set_sort_by}
			>
				{SortMetrics.map((met) => (
					<option value={met}>{met}</option>
				))}
			</select>
		</div>
	)
}
