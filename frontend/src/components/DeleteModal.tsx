import type { Dispatch, StateUpdater } from 'preact/hooks'
import './DeleteModal.css'

interface Props {
	Prompt: string
	DeleteURL?: string
	HandleDelete: (
		e: MouseEvent
	) => void | Promise<void | '/login' | '/404' | '/500' | '/rate-limit'>
	SetShowDeleteModal: Dispatch<StateUpdater<boolean>>
}

export default function DeleteModal(props: Props) {
	const {
		Prompt: prompt,
		DeleteURL: delete_url,
		HandleDelete: handle_delete,
		SetShowDeleteModal: set_show_delete_modal,
	} = props

	return (
		<dialog class='delete-modal' open>
			<p>
				{prompt}
				{delete_url ? (
					<>
						{' '}
						<strong>{delete_url}</strong>?
					</>
				) : null}
			</p>
			<button onClick={handle_delete}>Yes</button>
			<button autofocus onClick={() => set_show_delete_modal(false)}>
				Cancel
			</button>
		</dialog>
	)
}
