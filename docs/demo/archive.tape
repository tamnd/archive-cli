# Demo tape for archive. Rendered with ascii-gif (github.com/tamnd/ascii-gif),
# which supplies the window chrome and theme; this file is just the action.
#
#   ascii-gif render docs/demo/archive.tape -o docs/static/demo.gif
#
# archive must be on PATH inside the recording shell.

Hide
Type "export PS1='$ ' PATH=/Users/apple/bin:$PATH"
Enter
Type "clear"
Enter
Show

Sleep 500ms
Type "archive search 'collection:nasa mediatype:movies' -n 5 --fields identifier,title,downloads"
Sleep 800ms
Enter
Sleep 3s

Type "archive item nasa -o table"
Sleep 800ms
Enter
Sleep 2.5s

Type "archive wayback available https://google.com -o table"
Sleep 800ms
Enter
Sleep 2.5s
