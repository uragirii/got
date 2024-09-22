# got

got is a lightweight Git client written in Go. This project serves as a learning experience for me to explore the Go programming language. While not a comprehensive Git replacement, it implements some commonly used commands and flags.

**Goals**

- Implement essential Git commands.
- Learn and improve Go development practices.
- Track this project's development using both Git and got (interchangebly).

**Current Status**

Currently working on handling errors gracefully. The next focus will be on implementing `git checkout and `git branch`.

## Supported Commands

Note: Not all flags are currently supported. Basic functionality is implemented. Refer to the corresponding files in the `/cmd` folder for details.

- `git init`: Initializes a new Git repository. (Does nothing if already existing)
- `git status`: Displays the working directory status, including staged, tracked, and untracked files.
- `git add`: Starts tracking a file, adding it to the staging area (index).
- `git commit`: Commits staged changes. The output may differ slightly from the standard git command.

**Internal Commands**

- `git cat-file`: Pretty prints or uncompresses a Git object.
- `git hash-objects`: Calculates the SHA hash of an object, supporting multiple files.
- `git ls-files`: Lists all files tracked in the index, in their tracked order.

## Project Structure

- `/cmd`: Contains individual files for each supported command.
- `/internals`: Houses packages used by got:
  - `color`: Simple color printing functionality
  - `test_utils`: Suprise suprise! Contains test utilites
  - `git`: Core Git functionality. Further subfolders represent internal Git operations.
- `/testdata`: Contains raw files copied from this project's own .git folder for testing purposes.

I have also documented my process on Twitter in [this thread](https://x.com/quacky_batak/status/1799424455586017747) (also check quote tweets)

## Build and Run

I have very simple `Makefile` just run `make got` and it should build the executable in `/build` folder.

## Found a bug / Have suggestions

Great! please open a PR or an issue and I will definately look at it.
