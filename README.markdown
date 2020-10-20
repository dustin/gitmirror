# What's This?

I've got a few git repos and a few places I automatically clone git
repos and wanted to make sure these things stay up-to-date.  Some
repos are on github, some are on other machines around the internet.
They all look the same here.

gitmirror updates git repositories on webhook triggers.  This can be
anything from a simple invocation of `curl` from a post-commit hook to
github's post-receive hooks.

# How Do I Use This?

## Setting Up a Repo Path

First, you'll need stuff for it to do.  Let's say we wanted to set up
a repository mirror location in `/tmp/gitmirrors` and our first mirror
will be of my awesome `gitmirror` project.

First, install the software:

    go get github.com/dustin/gitmirror

Now, create a location for your mirrors and (as an example), check out
the gitmirror source into it:

    mkdir /tmp/gitmirrors
    cd /tmp/gitmirrors
    git clone --mirror git://github.com/dustin/gitmirror

(note, don't actually use `/tmp/` as your permanent mirror path)

### Note for Github Usage

If you're planning to use gitmirror with github, it will automatically
create the mirrors for you on first contact, so you just need to make
sure the default directory is there.

## Getting gitmirror Running

gitmirror is a standalone web server written in [go][golang].  It's
been tested on several platforms, but I mostly use it on Linux and
MacOS X.

Once you have your `gitmirror` binary built (Which happens
automatically with the `go get` command above), you run it like this:

    /path/to/gitmirror -git=/path/to/git -dir=/tmp/gitmirrors

## Trying it Out

Now, you can use [curl][curl] to play around and do repo syncs:

    curl http://localhost:8124/gitmirror.git

The above does a background sync and responds immediately with an http
202 (you can add `-D -` if you don't believe me).  If you want a
foreground sync, you can do the following:

    curl 'http://localhost:8124/gitmirror.git?bg=false'

Now you'll either get an http 200 or 500 depending on whether it was
successful along with the contents of stdout and stderr so you can see
what happened.

## Productionalizing

I've got a sample [launchd][launchd] `.plist` file in the `support`
directory because I happen to be running it on a mac.

See my blog post on [running processes][startup] for more detail on
actually running stuff.  I'm sure you can figure out the rest.

The machine I'm running this on doesn't have a web server, so I'm
actually doing a path translating proxy through nginx to get it here.
My nginx config looks not entirely unlike this:

    upstream gitmirror {
        server 10.10.3.21:8124;
    }
    server {
        [...];
        location /gm/ { proxy_pass http://gitmirror/; }
    }

Now I've got a URL available on the internet I can use to set up
github post-receive-hooks to update as well as git `post-commit` hooks
for the more private/weird stuff.

### But I Don't Have a Public Address

See [wwcp][wwcp] for a way to run entirely behind a firewall.

## Hooks

gitmirror will run `post-fetch` hooks for you if you have them
available.  One or both of the following will be executed (in this
order):

* `$gitmirrordir/current_repo.git/hooks/post-fetch`
* `$gitmirrordir/bin/post-fetch`

The first is the repository specific hook, allowing you to do stuff
like CI integration or doc builds or something.

The second is a single global hook that will run for every repo
allowing you to have a common behavior across all updates (e.g. you
might want to `touch 'git-daemon-export-ok'` or post something to
twitter or chain a different hook or something.

### Batches of Hooks

If you have a ton of hooks to set up, check out the
[setuphooks][setuphooks] command.  It works great for hundreds of
repos with simple patterns to express how you want them to map to your
mirror.

[golang]: http://golang.org/
[launchd]: http://developer.apple.com/macosx/launchd.html
[curl]: http://curl.haxx.se/
[startup]: http://dustin.github.com/2010/02/28/running-processes.html
[setuphooks]: gitmirror/tree/master/setuphooks
[wwcp]: //github.com/dustin/wwcp
