

I want to:
- give Claude a to-do-list; it can be a markdown file, with structured waves and individual items.
- give Claude access to a private GitHub repo; it must pull the code in
- Claude runs in the background, working on the codebase. It is not executing on my local machine, so I can lose internet / turn off my machine.
- I must review the changes Claude is making. What is helpful is (1) a diff-view of the changes so far (like filesystem view), and (2) a stream of Claude's output / thoughts.
- I must be able to talk to Claude; ask it questions, and tell it to do things, not included in the original PR. That way I can correct it mid-flight if needed.
- Claude must perform the desired changes, and then open a PR

Two interfaces for this:
- The terminal on my computer
- My phone (presumably on a website)

---

Possible solutions:

VM:
- Spin up a VM (Digital Ocean) ( will it already have docker installed?)
- Pull in a docker container, which has: (1) Ubuntu, (2) Claude Code, (3) git, (4) our custom API-server / client.
- Start the container, providing (1) github credentials, (2) claude code credentials.
- The container receives messages via API; it git-clones the repo, gets the TO_DO.md, and then prompts Claude to start working on it.
- The container streams Claude-output (and maybe filesystem state?) to the API.
- 
