# open-internet-treasure-map

## Todos

### Features

-Pagination
    -User Treasure Map
        -Submitted / Copied / Tagged links
        -Cats
        -Subcats
    -Global Cats
    -Global Subcats
-NSFW tags:
    -automatically correct 'nsfw' to 'NSFW'
    -Tests
    -Restrict from tmap/top unless specifically chosen in filter
-Look into input sequences that might produce problematic results
    -e.g., cats with "/" in them is not escaped in URL, might be read as different route path
-Cat search on index/top.astro
-Fix times off by 4 hrs... again
    -12h time might be screwing it up (just got 1:28pm for 5:28am)
        -btw lol
-Quick link to existing link if attempting to submit duplicate
-Home page pagination

### Code Quality

-Refactors for simplicity / accuracy
    -Move backend validation to /model unless using additional controller logic, e.g., JWT
    -GetTmapCatCounts probably possible in all sql
    -remove superfluous is_tagged
    -Make Top Cats / Top Links / etc. into components
-Purge code duplication
    -Layout.astro, Feedback.astro styles
        -NoFooter.astro for Feedback, BaseLayout extends NoFooter and is used elsewhere
    -handler_test / util_test TestMain()s
-CI/CD
    -cronjob to backup db every day or so

## To-Maybe-Dos

-robots.txt
-Guidelines / heuristics for avoiding "marooned" tags
    -only proper nouns / abbreviations should be capitalized?
-Optional topical summaries that can be provided if you submit / like enough links with a certain cat?
-Tmap period filter
-Better way to visualize how Global Cats are determined
-Show number of copies along with number of likes in frontend
-Edit category filters directly on top links by period/category(ies) page
    -Add or remove multiple at a time, so e.g., scanning for 3 cats does not take 3 page loads
-Search for existing tag cats while adding/editing
    -Fuzzysort?
    -For nearly identical cats with slight differences, maybe have a prompt on load that says like "would you like to reload and include these results too?"
-Properly backup DB
-Rethink CalculateGlobalCategories algo
-Favorite tmaps
    -add favorites col to users table
    -'Add to Favorites' button on other user's tmap
    -'Favorites' link on tmap
    -{user}'s favorite tmaps page
-SQL prepared statements
-Better logging?
    (Zap)
-Other lesser refactors and removal of duplicate code
    -shared TestClient?
    -helpers for DB actions
        -(new link, new summary, update summary, etc.)
    -ScanTmapLinks tests
    -BuildTagPage helper to declutter GetTagPage handler
    -duplicate handle_redirect() helpers on tag page / summary page
    -Fix SQL identifiers to use double quotes (?)
        -verify first
-Other tests
    -finish handlers
    -handler utils
        -TestExtractMetaDataFromGoogleAPIsResponse()
        -GetJWTFromLoginName: see if possible to verify JWT claims and AcceptableSkew
    -model utils
-Look into broken auto og:image
    -e.g., coolers.co image should not have been added with invalid link
-Improve profile pic upload?
-Improve frontend A11y/semantic markup/looks
    -Link preview img srcset
    -Tiny bit more space between like/copy buttons on mobile
    -maybe go through BrowserStack and see if anything is horrendous
    -add proper favicon.ico
-Redis caching

## Why?

Because there's a lot of good shit on the internet that's hard to be aware of and, to a lesser extent, hard to find even when you know about it.

Internet users deserve a portal that provides them an unbiased, direct view into the web's useful contents.

## What would be better than the status quo?

A network of links that are selected and organized by end users who closely relate to the individuals seeking those resources.

e.g., if I wanted to find all of the internet resources that Tim Ferriss recommends, or that the public at large recommends about the topic of surfing, they will all be in one place for quick and easy discovery.

Links are tagged with various categories to best organize and produce relevant recommendations.

e.g., Vim Adventures could be tagged with 'programming', 'learning games', 'typing games', 'keyboard ergonometry', etc.

Users can like listed links to boost them, so in theory the most univerally appreciated resources are easiest to find.

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

## Anecdotes to keep in mind

"what made these sites awesome for me was the backfill of amazing content that other people had already cataloged. I didn’t share much, I just enjoyed other people’s content."

"There was a time I would only search in del.icio.us instead of Google because the content quality was much better. So if you go this way, please don't fill it with content from botfarms posting to reddit."

"To me, Search is the number 1 need."

## Challenges

- Learning Go
    - interfaces
    - pointers
- Debugging Docker install
    - edit etc/apt/sources.list.d/docker.list to add specific Ubuntu codename to retrieve correct release package
    - repeated "dial unix /var/run/docker.sock: connect: connection refused" cryptic errors, tried editing group permissions, starting/stopping docker daemon etc. but nothing working
    - finally got it by authenticating DockerHub acct. via CLI (docker login)
- Linode
    - Getting connections other than SSH (http/https) to open despite firewall config explicitly allowing them
- LetsEncrypt / Certbot / NameCheap SSL
- tmux
    - detach from / reattach to SSH session to safely exit terminal and leave running
- YouTube Data API
    - Register Google API key
- Bash scripting
    - modified https://stackoverflow.com/a/76544267 for FITM
    - sourcing .bashrc from /etc/profile on if exists and readable