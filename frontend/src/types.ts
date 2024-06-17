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

// TAG
type Tag = {
	ID: number
	Categories: string
	SubmittedBy: string
	LastUpdated: string
}

type EarlyTag = Tag & {LifeSpanOverlap: number}

type TagPage = {
	Link: LinkData
	UserTag: Tag | undefined
	TopTags: EarlyTag[]
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
	LastUpdated: string
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

const tmap_sections = ['Submitted', 'Tagged', 'Copied'] as const

export { Periods, is_error_response, tmap_sections }
export type { CategoryContributor, CategoryCount, ErrorResponse, LinkData, Period, Summary, SummaryPage, Tag, TagPage, TreasureMap, User }

