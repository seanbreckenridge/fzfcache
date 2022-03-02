## fzfcache

A small cache for unique lines of text, used to speedup the load time of expensive [fzf](https://github.com/junegunn/fzf) invocations

Given any shell command as positional arguments:

- If that command has been run in the past (determined by hashing the command itself), immediately prints the cached value -- this means you can immediately select something with `fzf`
- Prints any lines from that shell command to STDOUT, which haven't already been printed from the cachefile, removing any duplicates
- Once the shell command exits, saves the output of the shell command to a file in `~/.cache/fzfcache/`

This keeps a history of one command, so its possible that lines from the previous result are included in the current `fzf` buffer/cachefile. So, if exact results are very important every time this is run, this probably isn't for you.

As an example:

<img src="https://raw.githubusercontent.com/seanbreckenridge/fzfcache/master/.github/demo.gif">

As some other examples of me using this:

- [cache food items](https://github.com/seanbreckenridge/ttally#shell-scripts) (in [`cz`](https://github.com/seanbreckenridge/ttally/blob/master/bin/cz))
- jump to directories I use often in [`tttjump`](https://sean.fish/d/tttjump?dark)
- pick a [config file to edit](https://github.com/seanbreckenridge/dotfiles/blob/2c579f32e6c3a5d42736816e4d38e0a409a847e4/.config/shortcuts.toml#L5-L21)
- pick a [config file to send to someone](https://sean.fish/d/give?dark)
- search my [github stars](https://sean.fish/d/mystarsfzf?dark)

### Install

Using `go install` to put it on your `$GOBIN`:

`go install github.com/seanbreckenridge/fzfcache@latest`
