import type { ChangeEvent } from 'preact/compat'
import type { Dispatch, StateUpdater } from 'preact/hooks'
import { Periods, type Period } from '../../types'

interface Props {
	Period: Period
	SetPeriod: Dispatch<StateUpdater<Period>>
}

export default function SearchPeriod(props: Props) {
	const { Period: period, SetPeriod: set_period } = props

	async function handle_set_period(e: ChangeEvent<HTMLSelectElement>) {
		set_period(e.currentTarget.value as Period)
	}
	return (
		<div>
			<label id='search-period' for='period'>
				Period:
			</label>
			<select
				name='period'
				id='period'
				defaultValue={period}
				value={period}
				onChange={handle_set_period}
			>
				{Periods.map((per) => (
					<option value={per}>{per}</option>
				))}
			</select>
		</div>
	)
}
