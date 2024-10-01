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
				<button class='img-btn' onClick={handle_delete}>
					<img src='../../../x-lg.svg' height={20} width={20} />
				</button>
			) : addable && props.AddedSignal ? (
				<button class='img-btn plus-btn' onClick={handle_add}>
					<img src='../../../plus-lg.svg' height={20} width={20} />
				</button>
			) : null}
		</li>
	)
}
