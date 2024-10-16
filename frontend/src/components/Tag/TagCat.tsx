interface Props {
	Cat: string
	Count?: number
	Addable?: boolean
	Removable?: boolean
	AddedSignal?: Signal<string | undefined>
	DeletedSignal?: Signal<string | undefined>
}

import type { Signal } from '@preact/signals'
import './TagCat.css'

export default function TagCat(props: Props) {
	const { Cat: cat, Addable: addable, Removable: removable } = props

	async function handle_add(e: MouseEvent) {
		e.preventDefault()

		if (!props.AddedSignal) return
		props.AddedSignal.value = cat
	}

	async function handle_delete(e: MouseEvent) {
		e.preventDefault()

		if (!props.DeletedSignal) return
		props.DeletedSignal.value = cat
	}
	return (
		<li
			title={addable ? `Add cat '${cat}'` : ''}
			class={`cat${addable ? ' addable' : ''}`}
		>
			<p>
				{props.Cat}
				{props.Count ? ` (${props.Count})` : ''}
			</p>
			{removable && props.DeletedSignal ? (
				<button
					title='Remove cat'
					class='img-btn'
					onClick={handle_delete}
				>
					<img src='../../../delete.svg' height={20} width={20} />
				</button>
			) : addable && props.AddedSignal ? (
				<button
					title='Add cat'
					class='img-btn plus-btn'
					onClick={handle_add}
				>
					<img src='../../../add.svg' height={20} width={20} />
				</button>
			) : null}
		</li>
	)
}
