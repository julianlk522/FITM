type CategoryCount = {
	Category: string
	Count: number
}

type CategoryContributor = {
	Category: string
	LoginName: string
	LinksSubmitted: number
}

type LinkData = {
	ID: number
	URL: string
	SubmittedBy: string
	SubmitDate: string
	Categories: string | null
	Summary: string | null
	LikeCount: number
}

type TreasureMap = {
	Links: LinkData[]
	Categories: CategoryCount[]
}

export type { CategoryContributor, CategoryCount, LinkData, TreasureMap }

