import type { ComponentChildren } from 'preact'
import type { Dispatch, StateUpdater } from 'preact/hooks'
import './Modal.css'
// import MaroonedLinksTip from '../components/Modal/MaroonedLinksTip.astro'

interface Props {
	children?: ComponentChildren
	Prompt?: string
	IsDeleteConfirmation?: boolean
	DeleteURL?: string
	HandleDelete?: (e: MouseEvent) => void | Promise<void | string>
	SetShowModal: Dispatch<StateUpdater<boolean>>
}

export default function Modal(props: Props) {
	const {
		children,
		Prompt: prompt,
		IsDeleteConfirmation: is_delete_confirmation,
		DeleteURL: delete_url,
		HandleDelete: handle_delete,
		SetShowModal: set_show_modal,
	} = props

	return (
		<dialog class='modal' open>
			{children ? (
				<>{children}</>
			) : prompt ? (
				<p>
					{prompt}
					{delete_url ? (
						<>
							{' '}
							<strong>{delete_url}</strong>?
						</>
					) : null}
				</p>
			) : null}
			{is_delete_confirmation ? (
				<>
					<button onClick={handle_delete}>Yes</button>
					<button autofocus onClick={() => set_show_modal(false)}>
						Cancel
					</button>
				</>
			) : (
				<button class='ok-btn' onClick={() => set_show_modal(false)}>
					Ok
				</button>
			)}
		</dialog>
	)
}
