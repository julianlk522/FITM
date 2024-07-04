export function assign_scroll_observer({scroll_up_btn, target}: {scroll_up_btn: HTMLButtonElement, target: HTMLElement}) {
    if (!scroll_up_btn || !target) {
        return
    }
    const observer = new IntersectionObserver((entries) => {
        entries.forEach((entry) => {
            if (!entry.isIntersecting) {
                scroll_up_btn.style.display = 'flex'
            } else {
                scroll_up_btn.style.display = 'none'
            }
        })
    })
    observer.observe(target)
}

export function assign_scrollup_btn_click_handler(scroll_up_btn: HTMLButtonElement) {
    if (!scroll_up_btn) {
        return
    }

    scroll_up_btn.addEventListener('click', () => {
        window.scrollTo({ top: 0 })
    })
}