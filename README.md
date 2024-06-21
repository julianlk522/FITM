# open-internet-treasure-map

## Todos:

-Change "Add Tag" when you already have a tag for that link
-Fix popular tags links so they check global cats
-Allow http in AddLink handler
-Update submit new link page to use TagCategory component and form a la EditTag component
-Treasure Map skip to each main section (sidebar)
-Improve profile pic upload
    -Crop to square
    -Shrink to 100x100
-Pagination
    -Home
    -Top Links 
    -User Treasure Map
-Separate tag categories into distinct rows in Tags table
    -(Simplifies add/delete and maybe global category calculations, but might not be necessary at this point?)
-Improve frontend look/semantic markup
    -A11y, responsive layouts for phones / tablets
    -Link articles
    -Proper color scheme
-Tests
-Remove repeat code wherever possible
-Some way to visualize how Global Cats are determined

## Why?

Because there's a lot of good shit on the internet that's hard to be aware of and, to a lesser extent, hard to find even when you know about it.

Search engines help somewhat, but they are biased in favor of results who optimize endlessly for SEO features (which are pretty irrelevant with regard to content quality, and gaming SEO buries a lot of barebones but high-quality stuff) or who pay to be promoted. These conditions funnel engagement mainly to big businesses rather than new and varied creators, and especially to businesses who specialize in farming the engagement at the expense of the internet visitor. Internet users deserve a portal that provides them an unbiased, direct view into the web's useful contents. 

## What would be better than the status quo?

A network of links that are selected and organized by end users who closely relate to the individuals seeking those resources.

e.g., if I wanted to find all of the internet resources that Tim Ferriss recommends, or that the public at large recommends about the topic of surfing, they will all be in one place for quick and easy discovery.

Links are tagged with various categories to best organize and produce relevant recommendations.

e.g., Vim Adventures could be tagged with 'programming', 'learning games', 'typing games', 'keyboard ergonometry', etc.

Users can like listed links to boost them, so in theory the most univerally appreciated resources are easiest to find.

## What will each user have / be able to do?

Have:
-Tree of user's submitted links + copied links
-User profile/summary screen with avatar, stats

Can:
-Create an account (user/pass or OAUTH/provider)
-Log in (same options as above)
-Change profile settings

-Add (and tag) new links
-Like existing links
-Copy existing links to user's own tree
-Like link summary
-Submit alternative link summary
-Browse global link tree
-Browse other user trees

## How is the global tree ("treasure map"?) derived from numerous varying, individual trees?

Tags are aggregated across all users in the form of weighted scores. Weights are calculated based on 1) commonness of the topic and 2) amount of information provided by each user.

e.g., user A who has a scant individual tree and who contributes/edits a link related to a popular topic (News - Political) will not greatly influence the likelihood of their chosen tags appearing on links in the global tree. OTOH, user B who contributes/edits heavily and submits tags on niche topics (Programming - OpenGL - Learning Resources) will have a significant influence on the likelihood of those tags being set in the global tree.

Since a tag's weight is partially dependent on time since creation, tags need to be globally re-evaluated at a regular interval (maybe every morning).

## Stack:

Astro, TS
Go
Netlify (frontend)
MySQL - ~$15/m for 1GB RAM, 1vCPU
[some VPS - Akamai? DO?] (backend) - ~$24/m for 4GB RAM / 2 CPUs, ~$1/m for 10GB block storage

## Why is this service different and better than existing alternatives?

-LinkTree: only about social links for a particular person (no concept of global tree / treasure map)

-PinBoard: ugly, unintuitive to navigate, non-hierarchical (no guarantee of link quality), unclear what links actually are (inkscape.org reads as, "Draw Freely" with the destination address hidden), founder is AWOL

-LinkHut: weird layout (one long index at a time), duplicate links, giant unwieldy block of searchable tags (fix by using weight system and limiting tags that are publicly visible), no tree structure for tags for most-specific sorting

-Shaarli: self-hosted / personal but also shareable links? what? also ugly and not easily accessible (web-based)

-Del.icio.us: doesn't exist anymore, non-hierarchical. note: really like that they have a preview img for front page links

-StumbleUpon: doesn't exist anymore, hard to browse many things at once

-CloudHiker (redirected from stumbled.to): good but hard to browse many things at once, too few categories / not specific enough

-Digg: doesn't exist anymore, too few categories / not specific enough, no personal trees / profiles only global (I think?), "Digg faced problems due to so-called "power users" who would manipulate the article recommendation features to only support one another's postings"

-Are.na: some good ideas but requires signup to do anything and paid plan to do much, not explicitly web content (just too big a can of worms at that point), requires too much upfront learning/adapting (e.g. blocks and channels methodology), nested channels impose arbitrary complexity and are too confusing to navigate, kind of dull and scary looking

## Anecdotes to keep in mind:

"what made these sites awesome for me was the backfill of amazing content that other people had already cataloged. I didn’t share much, I just enjoyed other people’s content."

"There was a time I would only search in del.icio.us instead of Google because the content quality was much better. So if you go this way, please don't fill it with content from botfarms posting to reddit."

"To me, Search is the number 1 need."
