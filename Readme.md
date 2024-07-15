# got

A very basic git client written in golang. This project isn't meant to be a complete git client but a project for me to learn go. I will only implement some very common used git commands and flags. My longterm goal is that this repo can be tracked by both the git client and the got binary. Probably use them interchangebly.

Current Progress: working on `git commit`, but before that I'm working on writing tests.

## Project Structure

`/cmd` folder contains implemetation for each commands I've implemented. Currently, I've added support for `git init`, `git add`, `git status`. Along with them, I've also added `git ls-files`, `git hash-object`, `git cat-file` as they were used to debug and test some internal working of git.

`/internals` contains all the internal stuff, `/internal/git` contains git internal implementation like `object`, `head`, `index` file parsing. File names should be self-explanatory.

I have also documented my process on Twitter in [this thread](https://x.com/quacky_batak/status/1799424455586017747) (also check quote tweets)

## Build and Run

I have very simple `Makefile` just run `make got` and it should build the executable in `/build` folder.

## Found a bug / Have suggestions

Great! please open a PR or an issue and I will definately look at it.
