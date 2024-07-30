# open-internet-treasure-map

## Todos:

### Features:

-Prevent tag last_updated from being changed if no cats were edited
-Merge GetProfile handler logic into GetTreasureMap handler
    -(since it is only used on tmap page alongside link data)
-Middleware to alphabetize tag categories automatically
    -e.g., http://localhost:4321/cat/shit,nerd => http://localhost:4321/cat/nerd,shit
    -/top/{period}/{cat}, /cat/{cat}
-Use traditional "?" URL params instead of routes
    -one single link results page spanning all of /cat/{categories}, /top/{period}, /top/{period}/{categories} depending on URL params passed
    -/cat, /top, etc. get confusing
-Pagination
    -User Treasure Map
    -Fix top tag cats so they are specific to page being shown?
-Improve frontend look/semantic markup
    -A11y, responsive layouts for phones / tablets
    -Proper color scheme
-Rethink CalculateGlobalCategories algo
    -currently makes it impossible, unless submitting first tag, to affect global cats unless extremely new link and fast tag submission...

### Code Quality:

-Enforce consistent names
    -e.g., Global Categories vs. Global Tag categories
    -Popular Categories vs. Top Categories 
-Rebrand subcategories as category overlaps since that is a bit more accurate
-Remove repeat code wherever possible
-Follow patterns shown in https://github.com/jonnylangefeld/go-api
-Tests
    -https://github.com/ory/dockertest
-Update JWT to use actual secret

## To Maybe-Dos:

-Show number of copies along with number of likes in frontend
-Edit category filters directly on top links by period/category(ies) page
    -Add or remove multiple at a time, so e.g., scanning for 3 cats does not take 3 page loads
-Search for existing tag cats while adding/editing
    -Fuzzysort?
-Way to prevent many tags from flooding global tag
    -might not happen actually? would require many different cats which is not super likely i would not imagine
-Some way to visualize how Global Cats are determined
-Improve profile pic upload
    -more file formats?
    -maybe optimize min/max aspect ratios
-Separate tag categories into distinct rows in Tags table
    -(Simplifies add/delete and maybe global category calculations, but might not be necessary at this point?)
    -would help optimize GetTopTagCategories / GetTopTagCategoriesByPeriod handlers since queries could all be done in sql (as of now requires splitting global_cats field in Go)
-Better logging?
    (Zap)




## Why?

Because there's a lot of good shit on the internet that's hard to be aware of and, to a lesser extent, hard to find even when you know about it.

Internet users deserve a portal that provides them an unbiased, direct view into the web's useful contents. 

## What would be better than the status quo?

A network of links that are selected and organized by end users who closely relate to the individuals seeking those resources.

e.g., if I wanted to find all of the internet resources that Tim Ferriss recommends, or that the public at large recommends about the topic of surfing, they will all be in one place for quick and easy discovery.

Links are tagged with various categories to best organize and produce relevant recommendations.

e.g., Vim Adventures could be tagged with 'programming', 'learning games', 'typing games', 'keyboard ergonometry', etc.

Users can like listed links to boost them, so in theory the most univerally appreciated resources are easiest to find.

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
