# CLI tool for Pivotal Tracker

A simple tool to interact with Pivotal Tracker. The primary
focus is on ease of use for daily usage. This tool is not meant to
replace the web interface.

## Installation

```
go get github.com/mckean/ptcli
```

There are two env variables that should be set:

```
PIVOTAL_TOKEN
```

and

```
PIVOTAL_PROJECT
```

However they can be passed/overriden with `-t` (token) and `-p` (project).

## Usage

```
ptcli story My new feature
```

This will add a new feature story to the icebox.

```
ptcli story -b This must be fixed asap -i -l api,urgent
```

This will add a new bug to the end of the top of the backlog with labels `api`
and `urgent`.  
Since it's a bug with no estimate it will be at the bottom of the current
iteration.

```
ptcli story -c refactor the signup component -l frontend
```

This will add a new chore to the icebox.

```
ptcli story User should be able to create stories from cli -i -e 8
```

This will add a new feature story at the top of the backlog with an estimate
of 8 points.
