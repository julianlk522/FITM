import type { ComponentChildren } from 'preact'
import { useState } from 'preact/hooks'
import Modal from './Modal'
import './ToolTip.css'

interface Props {
	Prompt: string
	children?: ComponentChildren
}

export default function ToolTip(props: Props) {
	const [show_modal, set_show_modal] = useState(false)

	return (
		<>
			<button
				class='tip'
				title='View more information'
				onClick={() => set_show_modal(!show_modal)}
			>
				{props.Prompt}
			</button>
			{show_modal ? (
				<Modal SetShowModal={set_show_modal}>{props.children}</Modal>
			) : null}
		</>
	)
}
