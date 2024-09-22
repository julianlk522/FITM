import type { Signal } from '@preact/signals'
import type { ChangeEvent } from 'preact/compat'
import { Periods, type Period } from '../types'

interface Props {
	SelectedPeriod: Period
	SetPeriodSignal: Signal<Period>
}

export default function SearchPeriod(props: Props) {
	const selected = props.SelectedPeriod

	async function handle_set_period(e: ChangeEvent<HTMLSelectElement>) {
		props.SetPeriodSignal.value = e.currentTarget.value as Period
	}
	return (
		<div>
			<label id='search-period' for='period'>
				Period:
			</label>
			<select
				name='period'
				id='period'
				defaultValue={selected}
				value={selected}
				onChange={handle_set_period}
			>
				{Periods.map((per) => (
					<option value={per}>{per}</option>
				))}
			</select>
		</div>
	)
}
