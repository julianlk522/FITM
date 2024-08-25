interface Props {
	Cat: string
	EditActivated: boolean
	Deleted: Signal<string | undefined> | undefined
}

import type { Signal } from '@preact/signals'
import './TagCat.css'

export default function TagCat(props: Props) {
	const { Cat: cat, EditActivated: edit_activated } = props
	async function handle_delete(e: MouseEvent) {
		e.preventDefault()

		if (!props.Deleted) return
		props.Deleted.value = cat
	}
	return (
		<li class='cat'>
			<p>{props.Cat}</p>
			{edit_activated ? (
				<button class='img-btn' onClick={handle_delete}>
					<img src='../../../x-lg.svg' height={20} width={20} />
				</button>
			) : null}
		</li>
	)
}
