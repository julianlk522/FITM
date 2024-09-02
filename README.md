# open-internet-treasure-map

## Todos

### Features

-Cat search on index/top.astro
    -Add or remove multiple at a time, so e.g., scanning for 3 cats does not take 3 page loads
    -For nearly identical cats with slight differences, maybe have a prompt on load that says like "would you like to reload and include these results too?"
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
-look into not rendering images that dont successfully load
-be sure to fix custom fetch directing to /rate-limit when it shouldnt
-add about scrollup
-make sure new link submissions show correct time
    (they dont now.. but then it's correct elsewhere? weird)
-YT channel links

### Code Quality

-VPS SSH key
-Refactors for simplicity / accuracy
    -Move backend validation to /model unless using additional controller logic, e.g., JWT
    -GetTmapCatCounts probably possible in all sql
-Purge code duplication
    -Astro / Preact components:
        -Top Cats / Top Links / etc. lists
        -search filters (top, more)
    -handler_test / util_test TestMain()s
-CI/CD
    -cronjob to backup db every day or so
-Security
    -Look into input sequences that might produce problematic results
        -e.g., cats with "/" in them is not escaped in URL, might be read as different route path
    -refactor fetch_with_handle_rate_limit() to redirect to /404 in the catch block
        -maybe have it return an object with props Response (Response or undefined) and RedirectTo ("/404". "/rate-limit", or "")

## To-Maybe-Dos

### Features

-SQL prepared statements
    -more important if truly does help prevent injection... verify
-Redis caching
-Favorite tmaps?
-Show number of copies along with number of likes in frontend
-Better way to visualize how Global Cats are determined?
-Optional summaries that can be edited if you submit / like enough links with a certain cat?
    -i.e., if you submit enough links with cat "FOSS" you get to add a wiki-like summary of "FOSS" that appears on the top page when it is applied alone
-Guidelines / heuristics for avoiding "marooned" tags
    -only proper nouns / abbreviations should be capitalized?
-Tmap period filter?
-Improve profile pic upload?
-Improve frontend A11y/semantic markup/looks
    -proper favicon.ico
    -Link preview img srcset
    -Tiny bit more space between like/copy buttons on mobile
    -maybe go through BrowserStack and see if anything is horrendous
-Rethink CalculateGlobalCategories algo

### Code Quality

-Better logging?
    (Zap)
-Other lesser refactors and removal of duplicate code
    -duplicate handle_redirect() helpers on tag page / summary page
    -BuildTagPage helper to declutter GetTagPage handler
    -ScanTmapLinks tests
    -helpers for DB actions
        -(new link, new summary, update summary, etc.)
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
    -https://rss.com/blog/how-do-rss-feeds-work/
-robots.txt

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
