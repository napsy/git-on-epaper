# git-on-epaper

A gitlab webhook for push notifications on a project. The webhook serves a HTML that shows the last push on the project with the following information:

- commit title
- commit author
- added/deleted lines

Here's a screenshot of the page, shown on a e-paper device from [Visionect!](https://www.visionect.com):

[Screen shot](shot.png)

To compile, clone and run

```bash
go build ./...
```

The webhook listens on port ``8001`` where ``/`` is the webhook URL and ``/page`` the HTML template.
