type CategoryCount = {
	Category: string
	Count: number
}

type CategoryContributor = {
	Category: string
	LoginName: string
	LinksSubmitted: number
}

type Link = {
	ID: number
	URL: string
	SubmittedBy: string
	SubmitDate: string
	Categories: string
	LikeCount: number
}

export type { CategoryContributor, CategoryCount, Link }

