import { useEffect, useState } from 'preact/hooks'

interface Props {
	pfp: string
    login_name: string
}

export default function ProfilePic(props: Props) {
    const {login_name, pfp} = props
    const [url, set_url] = useState<string | undefined>(undefined)

    useEffect(() => {
        async function get_pfp() {
            
            const pfp_resp = await fetch(`http://127.0.0.1:8000/pic/${pfp}`, {
                headers: { 'Content-Type': 'image/png' },
            })
            if (pfp_resp.status > 399) {
                console.log(pfp_resp)
            }
            const pfp_blob = await pfp_resp.blob()
            const pfp_url = URL.createObjectURL(pfp_blob)
            set_url(pfp_url)
        }

        get_pfp()
    }, [pfp])
    

    return <img src={url} alt={login_name} width="150" />
}