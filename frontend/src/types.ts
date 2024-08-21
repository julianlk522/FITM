// ERROR
type ErrorResponse = {
	status: string
	error: string
}

function is_error_response(obj: any): obj is ErrorResponse {
	return obj.error !== undefined
}

// USER
type Profile = {
	LoginName: string
	About: string
	PFP: string
	Created: string
}

// LINK
type Link = {
	ID: number
	URL: string
	SubmittedBy: string
	SubmitDate: string
	Cats: string | null
	ImgURL: string | undefined
	Summary: string | null
	SummaryCount: number
	TagCount: number
	LikeCount: number
	IsLiked: boolean | undefined
	IsTagged: boolean | undefined
	IsCopied: boolean | undefined
}

type PaginatedLinks = {
	Links: Link[]
	NextPage: number
}

const Periods = ['day', 'week', 'month', 'year', 'all'] as const
type Period = (typeof Periods)[number]

// TAG
type Tag = {
	ID: number
	Cats: string
	SubmittedBy: string
	LastUpdated: string
}

type TagRanking = Tag & { LifeSpanOverlap: number }

type TagPage = {
	Link: Link
	UserTag: Tag | undefined
	TagRankings: TagRanking[]
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
	Link: Link
	Summaries: Summary[]
}

// TREASURE MAP
type TmapLink = Link & { CatsFromUser: boolean | undefined }

type FilteredTreasureMap = {
	Submitted: TmapLink[]
	Copied: TmapLink[]
	Tagged: TmapLink[]
	Cats: CategoryCount[]
}

type TreasureMap = FilteredTreasureMap & { Profile: Profile }

const tmap_sections = ['Submitted', 'Copied', 'Tagged'] as const

export { Periods, is_error_response, tmap_sections }
export type {
	CategoryContributor,
	CategoryCount,
	ErrorResponse,
	FilteredTreasureMap,
	Link,
	PaginatedLinks,
	Period,
	Profile,
	Summary,
	SummaryPage,
	Tag,
	TagPage,
	TreasureMap,
}
