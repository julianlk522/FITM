import { format_long_date } from '../../util/format_date'

interface Props {
	LastUpdated: string
}

export default function CSRLastUpdated(props: Props) {
	return <p>last updated: {format_long_date(props.LastUpdated)}</p>
}
