
// LINK
type LinkData = {
	ID: number
	URL: string
	SubmittedBy: string
	SubmitDate: string
	Categories: string | null
	Summary: string | null
	LikeCount: number
}

// CATEGORY
type CategoryCount = {
	Category: string
	Count: number
}

type CategoryContributor = {
	Category: string
	LoginName: string
	LinksSubmitted: number
}

// SUMMARY
type Summary = {
	ID: number
	Text: string
	SubmittedBy: string
	LikeCount: number
}

type SummaryPage = {
	Link: LinkData
	Summaries: Summary[]
}

// TREASURE MAP
type TreasureMap = {
	Links: LinkData[]
	Categories: CategoryCount[]
}

export type { CategoryContributor, CategoryCount, LinkData, Summary, SummaryPage, TreasureMap }

