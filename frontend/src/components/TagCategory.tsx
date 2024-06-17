interface Props {
	Category: string
    EditActivated: boolean
    Deleted: Signal<string | undefined>
}

import type { Signal } from "@preact/signals"
import "./TagCategory.css"

export default function TagCategory(props: Props) {
    const { Category: category, EditActivated: edit_activated } = props
    async function handle_delete(e: MouseEvent) {
        e.preventDefault()

        props.Deleted.value = category
    }
    return (
        <li class='category'>
            <p>{props.Category}</p>
            {edit_activated 
                ? 
                    <button class='img-btn' onClick={handle_delete}>
                        <img src='../../../x-lg.svg' height={20} width={20} />
                    </button> 
                : null
            }
        </li>
    )
}