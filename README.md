# flexible-internet-treasure-map

## Todos

In order of importance:
    1. find some way to cache the stupid github.com/lestrrat-go/httpcc download
    2. anti-spam/naughty stuff
    3. refactors

nice to do:
- some preventative actions in place to prevent spamming
    - probably some way to detect porn/gore and add NSFW tag
        - (and prevent it from being removed)
    - way to report links as NSFW

- sync rpi test data with updated (with NSFW tags)
- change monkey / bradley names
- update marvel char name tags to be uppercase
- replace cookies expiring in 21600 secs (6 hrs) with 14400 (4 hrs)
- check if any remaining fetches should be wrapped in redirect util
- replace frontend (and backend, if any) magic numbers

### Features

-Pagination
    -User Treasure Map
        -Submitted / Copied / Tagged links
        -Cats
        -Subcats
    -Global Cats
    -Global Subcats

### Code Quality

-Purge duplication
-Simplicity / accuracy
    -Top Cats / Top Links / etc. components
    -move tmap cats json above links
    -move contributors queries from query/link.go to query/contributors.go
        -will also allow sharing WHERE_NO_NSFW_CATS between TopLinks / tmap queries
-Readability
    -ErrServerFail => Err500 etc.
-Security
    -Look into input sequences that might produce problematic results
    -prevent whacky chars from usernames/passwords
        -/, ;, ", ', etc.
    -fuzz test

## To-Maybe-Dos

### Features

-favorite / followed tmaps?
-Show number of copies along with number of likes in frontend
-Improve frontend A11y/semantic markup/looks
    -button titles
    -subtitle probably should not be h2
    -original favicon.ico
    -Tiny bit more space between like/copy buttons on mobile
    -maybe go through BrowserStack and see if anything is horrendous
    -Link preview img srcset?
        probably not realistic
-Tmap period filter?
-Audit CalculateGlobalCategories algo?
-Improve profile pic upload?
    -cropping, more file formats, etc.
-Redis caching

### Code Quality

-find way to cache github.com/lestrrat-go/httpcc in GHA workflow
-Look into broken auto og:image
    -e.g., coolers.co image should not have been added with invalid link
    -https://rss.com/blog/how-do-rss-feeds-work/
-SQL prepared statements
    -more important if truly does help prevent injection... verify
-Ensure accurate / helpful http response codes
    -start by making sure all ErrInvalidRequests are actually that
    -e.g., tag page for invalid link id returns 500 (should be 404)
    -replace "message":"deleted" with just 204
    -205 for successful logins/forms that require reload
    -500 for server fuckups
-Other lesser refactors and removal of duplicate code
    -Fix SQL identifiers to use "" and string literals to use ''
    -duplicate SearchParams dropdown components
        -merge Period / SortBy into same component with unique options set as props
    -duplicate add_tag funcs (EditTag.tsx, SearchCats.tsx)
    -duplicate handle_redirect() helpers on tag page / summary page
    -duplicate redirect_to cookie logic using window.location.pathname
    -duplicate delete modals
        -link, tag, tmap pfp
    -BuildTagPage helper to declutter GetTagPage handler
    -helpers for DB actions
        -(new link, new summary, update summary, etc.)
    -os.LookupEnv?
        -not sure if makes any practical difference
    -Ensure backend validation is all in /model unless using additional controller logic, e.g., JWT
    -sync.Once for db singleton?
    -GetCatCountsFromTmapLinks probably possible in all sql
        -actually pretty clunky to achieve (break apart all global/user_cats_fts into words each time)
            -maybe consider revisiting if global/user_cats_fts vocab created for some reason later
        -also not that important since input is limited to user's tmap links, not entire links table. not going to be processing any more than a few hundred or thousand tags at absolute most (and not for a looong time). so perf differential is trivial
-Other tests
    -handler utils
        -TestExtractMetaDataFromGoogleAPIsResponse()
        -GetJWTFromLoginName: see if possible to verify JWT claims and AcceptableSkew / expiration
        -ScanTmapLinks
        -Increment/DecrementSpellfixRanksForCats
    -finish handlers
    -middleware
        -test err responses are logged to $FITM_ERR_LOG_FILE?
    -model utils
-VPS SSH key
-improve cat count lookup speed with fts5vocab table
    -(row type)
-Ensure consistent language:
    -get (request and retrieve things from an external source)
    -scan (copy rows from sql to structs)
    -extract (some data and carve out a different data type from it)
    -assign (take some data and a pointer and copy the data to the referenced var)
    -obtain (get, extract, assign)
    -resolve (take in a possibly incomplete form and translate to a correct form)
    -verify (instead of check, ensure, etc.)
-replace spellfix transactions with triggers
    -(that way can make changes over CLI without worrying about unsync)
    -too complicated for now ... workaround might be just resetting on cron job or something though that requires downtime...

## Why is it different and better than existing alternatives?

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
    - testing features
- Debugging Docker install
    - edit etc/apt/sources.list.d/docker.list to add specific Ubuntu codename to retrieve correct release package
    - repeated "dial unix /var/run/docker.sock: connect: connection refused" cryptic errors, tried editing group permissions, starting/stopping docker daemon etc. but nothing working
    - finally got it by authenticating DockerHub acct. via CLI (docker login)
- Linode
    - Getting connections other than SSH (http/https) to open despite firewall config explicitly allowing them
    - other firewall problems:
        - allow only implicit outbound SSH so git pull possible (and explicit inbound for ad hoc connections from trusted machine)
- LetsEncrypt / Certbot / NameCheap SSL
    - Certbot CLI
- tmux
    - detach from / reattach to SSH session to safely exit terminal and leave running
- YouTube Data API
    - Register Google API key
- Bash scripting
    - modified https://stackoverflow.com/a/76544267 for FITM package rename
    - sourcing .bashrc from /etc/profile on if exists and readable
    - backup_now.sh on cronjob
    - update script: process redirection, debugging double echos
- SQLite optimization
    - FTS5 virtual table
    - Spellfix1 extension
        - cross-compile errors, getting correct headers/.dll files and passing correct flags to x86_64-w64-mingw32-gcc
            - .so compiled by gcc in WSL not compatible with go:alpine Docker image architecture, recompile on test runner and save path in env
        - debugging debounce / re-render errors in React
        - tweaking the spellfix rankings system to improve result relevance (still WIP)
- CI/CD
    - GH actions workflow
        - Raspbian Buster firmware outdated (no Node.js 20 support needed for test runner)
            - flash memory card to update to Raspbian Lite Bookworm
            - no networking, configure manually with nmcli
        - didn't want private test data stored on GH: store on test runner local filesystem and pass path as GH Actions secret through workflow .yml file to Docker container where test suite runs
        - GH deploy key (SSH)
- Rate limiting
    - limiting from backend only not possible while using Netlify
        - Netlify CDN using numerous edge servers with hidden IPs: no way to whitelist appropriate server IPs
    - limiting from frontend only not practical / useful
        - Netlify only offers to Enterprise clients, but even if it were available it would not prevent abuse via direct communication to server
    - compromise by adding multiple IP-based rate limits to be shared among frontend and all users making requests from client, plus app-wide hard limit
        - 1min timeframe for ordinary usage limits,
        - 1sec timeframe for quick abuse resolution
- Bot traffic
    - add Netlify edge func to identify and block bot user agents
- Design challenges
    - global cats calc system
