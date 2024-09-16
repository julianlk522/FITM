# flexible-internet-treasure-map

## Todos

In order of importance:
    1. output non-2xx responses to log file
    2. refactors / remove duplicate code
    3. ensure HTTP responses are accurate
    4. don't render shitty images

### Features

-add option in recommended cats handler for omitted cats
    -(that way it can keep giving new recommendations without repeats)
-look into not rendering images that dont successfully load
-surround related statements in transactions
    -(only new link as far as I can think of?)
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
-improve cat count lookup speed with fts5vocab table
    -(row type)

### Code Quality

-Refactors for simplicity / accuracy
    -repetition in tmap submitted/copied FromCats methods
    -Top Cats / Top Links / etc. components
    -move tmap cats json above links
-Purge code duplication
-Security
    -Look into input sequences that might produce problematic results
        -e.g., cats with "/" in them is not escaped in URL, might be read as different route path

## To-Maybe-Dos

### Features

-Look into broken auto og:image
    -e.g., coolers.co image should not have been added with invalid link
    -https://rss.com/blog/how-do-rss-feeds-work/
-use user summaries in tmap if they exist
-client:visible for tmap
-SQL prepared statements
    -more important if truly does help prevent injection... verify
-Redis caching
-Favorite tmaps?
-Show number of copies along with number of likes in frontend
-Better way to visualize how Global Cats are determined?
-Guidelines / heuristics for avoiding "marooned" tags
    -only proper nouns / abbreviations should be capitalized?
    -always use "and" instead of "&" unless part of proper noun
    -describe the containing group, not the object if one of many
        -e.g., channels not channel
-Optional summaries that can be edited if you submit / like enough links with a certain cat?
    -i.e., if you submit enough links with cat "FOSS" you get to add a wiki-like summary of "FOSS" that appears on the top page when it is applied alone
-Tmap period filter?
-Improve profile pic upload?
-Improve frontend A11y/semantic markup/looks
    -subtitle probably should not be h2
    -original favicon.ico
    -button titles
    -Link preview img srcset?
        probably not realistic
    -Tiny bit more space between like/copy buttons on mobile
    -maybe go through BrowserStack and see if anything is horrendous
-Audit CalculateGlobalCategories algo
-Implement RemoveProfilePic handler (securely) for when a user somehow ends up with a PFP that isn't found in the DB
    -should be basically impossible so not high priority

### Code Quality

-robots.txt
-VPS SSH key
-Ensure accurate / helpful http response codes
    -e.g., tag page for invalid link id returns 500 (should be 404)
    -replace "message":"deleted" with just 204
    -205 for successful logins/forms that require reload
    -500 for server fuckups
-Ensure consistent language:
    -get (request and retrieve things from an external source)
    -scan (copy rows from sql to structs)
    -extract (some data and carve out a different data type from it)
    -assign (take some data and a pointer and copy the data to the referenced var)
    -obtain (get, extract, assign)
    -resolve (take in a possibly incomplete form and translate to a correct form)
-Other lesser refactors and removal of duplicate code
    -sync.Once for db singleton?
    -duplicate handle_redirect() helpers on tag page / summary page
    -BuildTagPage helper to declutter GetTagPage handler
    -ScanTmapLinks tests
    -helpers for DB actions
        -(new link, new summary, update summary, etc.)
    -Fix SQL identifiers to use double quotes (?)
        -verify first
    -Move backend validation to /model unless using additional controller logic, e.g., JWT
    -repeat check for "." in cats
    -repeat redirect_to cookie logic using window.location.pathname
    -repeat delete modals
        -link, tag, tmap pfp
    -duplicate add_tag funcs (EditTag.tsx, SearchCats.tsx)
    -GetCatCountsFromTmapLinks probably possible in all sql
        -actually pretty clunky to achieve (break apart all global/user_cats_fts into words each time)
            -maybe consider revisiting if global/user_cats_fts vocab created for some reason later
        -also not that important since input is limited to user's tmap links, not entire links table. not going to be processing any more than a few hundred or thousand tags at absolute most (and not for a looong time). so perf differential is trivial
-Other tests
    -finish handlers
    -handler utils
        -TestExtractMetaDataFromGoogleAPIsResponse()
        -GetJWTFromLoginName: see if possible to verify JWT claims and AcceptableSkew
    -model utils


## Why?

Because there's a lot of good shit on the internet that's hard to be aware of and, to a lesser extent, hard to find even when you know about it.

Internet users deserve a portal that provides them an unbiased, direct view into the web's useful contents.

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
- LetsEncrypt / Certbot / NameCheap SSL
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