import { useEffect, useState } from 'preact/hooks'

interface Props {
    LoginName: string
	PFP: string | undefined
    SignedInUser: string | undefined
    Token: string | undefined
}

export default function ProfilePic(props: Props) {
    const {LoginName: login_name, PFP: pfp, SignedInUser: signed_in_user, Token: token} = props
    const [url, set_url] = useState<string | undefined>(undefined)

    const is_signed_in_user = login_name === signed_in_user

    useEffect(() => {
        async function get_pfp() {
            if (!pfp) {
                return
            }
            const pfp_resp = await fetch(`http://127.0.0.1:8000/pic/${pfp}`, {
                headers: { 'Content-Type': 'image/png' },
            })
            if (pfp_resp.status > 399) {
                console.error(pfp_resp)
                return
            }
            const pfp_blob = await pfp_resp.blob()
            const pfp_url = URL.createObjectURL(pfp_blob)
            set_url(pfp_url)
        }

        get_pfp()
    }, [pfp])

    async function handle_pic_change(e: Event) {
        if (!token) {
            return (window.location.href = '/login')
        }
        const target = e.target as HTMLInputElement
        if (!target.files) {
            return
        }

        const new_pic = target.files[0]
        let formData = new FormData()
        formData.append('pic', new_pic)
        
        const new_pic_resp = await fetch(`http://127.0.0.1:8000/pic`, {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${token}` },
            body: formData
        })
        if (new_pic_resp.status > 399) {
            console.error(new_pic_resp)
            return
        }

        const new_pic_data = await new_pic_resp.blob()
        set_url(URL.createObjectURL(new_pic_data))
    }
    

    return url
        // user has profile pic
        ? 
            is_signed_in_user 
                ?
                    <>
                        <img src={url} alt="profile pic" width='150' />
                        <div>
                            <label for="new_pic_upload">Upload New: </label>
                            <input id='new_pic_upload' type="file" accept={"image/*"} onChange={handle_pic_change} />
                        </div>
                    </>
                :
                    <img src={url} alt="profile pic" width='150' />
        // no profile pic
        : 
            is_signed_in_user
                ? 
                <div>
                    <label for="new_pic_upload">Upload New: </label>
                    <input id='new_pic_upload' type="file" accept={"image/*"} onChange={handle_pic_change} />
                </div>
                : null
}