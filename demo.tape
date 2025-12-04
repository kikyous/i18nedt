# Where should we write the GIF?
Output demo.gif
Output demo.mp4

Set FontSize 28
Set Width 1200
Set Height 800

Set Padding 20

Set Theme "iTerm2 Solarized Light"

Hide
Type 'git restore test-locales/'
Enter
Ctrl+l

Show

Type 'i18nedt test-locales/*.json -k home.start'

Sleep 1

Enter

Type '/en'
Enter

Type 'j^C'

Type@200ms "Begin"

Escape
Type '/zh'
Enter

Type 'j^C'

Type@200ms "初始"

Escape

Type@300ms ":wq"
Enter

Type 'git --no-pager diff'

Enter
Sleep 2
