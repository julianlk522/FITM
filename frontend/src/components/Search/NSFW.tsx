import type { ChangeEvent } from 'preact/compat'
import type { Dispatch, StateUpdater } from 'preact/hooks'

interface Props {
	NSFW: boolean
	SetNSFW: Dispatch<StateUpdater<boolean>>
}

export default function SearchNSFW(props: Props) {
	const { NSFW: nsfw, SetNSFW: set_nsfw } = props

	return (
		<div>
			<label id='search-nsfw' for='nsfw'>
				NSFW
			</label>
			<input
				type='checkbox'
				id='nsfw'
				name='nsfw'
				checked={nsfw}
				onChange={(e: ChangeEvent<HTMLInputElement>) =>
					set_nsfw((e.target as HTMLInputElement).checked)
				}
			/>
		</div>
	)
}
