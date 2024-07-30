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
type LinkData = {
	ID: number
	URL: string
	SubmittedBy: string
	SubmitDate: string
	Categories: string | null
	ImgURL: string | undefined
	Summary: string | null
	SummaryCount: number
	LikeCount: number
	IsLiked: boolean | undefined
	IsTagged: boolean | undefined
	IsCopied: boolean | undefined
}

type PaginatedLinks = {
	Links: LinkData[]
	NextPage: number
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
	EarliestTags: EarlyTag[]
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
type TmapLink = LinkData & { CategoriesFromUser: boolean | undefined }

type FilteredTreasureMap = {
	Submitted: TmapLink[]
	Copied: TmapLink[]
	Tagged: TmapLink[]
	Categories: CategoryCount[]
}

type TreasureMap = FilteredTreasureMap & { Profile: Profile }

const tmap_sections = ['Submitted', 'Copied', 'Tagged'] as const

export { Periods, is_error_response, tmap_sections }
export type { CategoryContributor, CategoryCount, ErrorResponse, FilteredTreasureMap, LinkData, PaginatedLinks, Period, Profile, Summary, SummaryPage, Tag, TagPage, TreasureMap }

