import { effect, useSignal } from "@preact/signals";
import type { Dispatch, StateUpdater } from "preact/hooks";
import './NewTag.css';
import TagCategory from './TagCategory';

interface Props {
    Categories: string[]
    SetCategories: Dispatch<StateUpdater<string[]>>
    SetError: Dispatch<StateUpdater<string | undefined>>
}

export default function NewTag(props: Props) {
    const { Categories: categories, SetCategories: set_categories, SetError: set_error } = props

    // Pass deleted_cat signal to children TagCategory.tsx
    // to allow removing their category in this NewTag.tsx parent
    const deleted_cat = useSignal<string | undefined>(undefined)
    
    // Check for deleted category and set categories accordingly
    effect(() => {
        if (deleted_cat.value) {
            set_categories((c) => c.filter(cat => cat !== deleted_cat.value))
            deleted_cat.value = undefined
        }
    })
    
    function add_category(event: MouseEvent) {
        event.preventDefault()

        // @ts-ignore
        const form = event.target.form as HTMLFormElement
        if (!form) return set_error('Form not found')
        const formData = new FormData(form)
        const category = formData.get('category')?.toString()

        if (!category) {
            set_error('Missing category')
            return
        }

        if (categories.includes(category)) {
            set_error('Category already added')
            return
        }

        set_categories([...categories, category].sort())
        set_error(undefined)

        const cat_field = document.getElementById("category") as HTMLInputElement
        cat_field.value = ""
        return
    }

    return (
        <>
            <label for='category'>Tag Category(ies)</label>
            <input type='text' id='category' name='category' />
            <button onClick={(event) => add_category(event)}>Add Category</button>

            <ol id='categories_grid'>
                {categories.map((cat) => (
                    <TagCategory
                    Category={cat}
                    EditActivated={true}
                    Deleted={deleted_cat}
                />
                ))}
            </ol>
        </>
    )
    
}