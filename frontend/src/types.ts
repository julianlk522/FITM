// ERROR
type ErrorResponse = {
	status: string
	error: string
}

function is_error_response(obj: any): obj is ErrorResponse {
    return obj.error !== undefined 
  }

// USER
type User = {
	LoginName: string
	About: string
	PFP: string
	Created: string
}

// LINK
type LinkData = {
	ID: number
	URL: string
	SubmittedBy: string
	SubmitDate: string
	Categories: string | null
	Summary: string | null
	SummaryCount: number
	LikeCount: number
	ImgURL: string | undefined
	IsLiked: boolean | undefined
	IsCopied: boolean | undefined
}

const Periods = ['day', 'week', 'month', 'year', 'all'] as const
type Period = typeof Periods[number]

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
	SubmitDate: string
	LikeCount: number
	IsLiked: boolean | undefined
}

type SummaryPage = {
	Link: LinkData
	Summaries: Summary[]
}

// TREASURE MAP
type TreasureMap = {
	Submitted: LinkData[]
	Tagged: LinkData[]
	Copied: LinkData[]
	Categories: CategoryCount[]
}

export { Periods, is_error_response }
export type { CategoryContributor, CategoryCount, ErrorResponse, LinkData, Period, Summary, SummaryPage, TreasureMap, User }

