export default function AboutText(props: { About: string }) {
	return (
		<pre>
			about:
			{props.About ? '\n\n' + props.About : ''}
		</pre>
	)
}
