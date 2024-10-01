// ERROR
type ErrorResponse = {
	status: string
	error: string
}

function is_error_response(obj: any): obj is ErrorResponse {
	return obj.error !== undefined
}

// FETCH
const Redirects = ['/404', '/500', '/login', '/rate-limit'] as const
type RedirectTo = (typeof Redirects)[number]

type ResponseAndRedirect = {
	Response: Response | undefined
	RedirectTo: RedirectTo | undefined
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
	Summary: string | null
	SummaryCount: number
	TagCount: number
	ImgURL?: string
	LikeCount: number
	IsLiked?: boolean
	IsCopied?: boolean
}

type PaginatedLinks = {
	Links: Link[]
	NextPage: number
}

const Periods = ['day', 'week', 'month', 'year', 'all'] as const
type Period = (typeof Periods)[number]

const SortMetrics = ['rating', 'newest'] as const
type SortMetric = (typeof SortMetrics)[number]

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
	UserTag?: Tag
	TagRankings: TagRanking[]
}

// CATEGORY
type CatCount = {
	Category: string
	Count: number
}

type Contributor = {
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
	IsLiked?: boolean
}

type SummaryPage = {
	Link: Link
	Summaries: Summary[]
}

// TREASURE MAP
type TmapLink = Link & { CatsFromUser?: boolean }

type FilteredTreasureMap = {
	Submitted: TmapLink[]
	Copied: TmapLink[]
	Tagged: TmapLink[]
	Cats: CatCount[]
}

type TreasureMap = FilteredTreasureMap & { Profile: Profile }

const tmap_sections = ['Submitted', 'Copied', 'Tagged'] as const

export { Periods, SortMetrics, is_error_response, tmap_sections }
export type {
	CatCount,
	Contributor,
	ErrorResponse,
	FilteredTreasureMap,
	Link,
	PaginatedLinks,
	Period,
	Profile,
	ResponseAndRedirect,
	SortMetric,
	Summary,
	SummaryPage,
	Tag,
	TagPage,
	TreasureMap,
}
