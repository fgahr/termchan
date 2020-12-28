package output

// FIXME: Can use file embedding in Go 1.16
// Contains backticks so cannot use backtick notation.
const DefaultWelcome string = "{{ .FgGreen }}::::::::::::.,:::::: :::::::..   .        :     {{ .End }}\n" +
	"{{ .FgGreen }};;;;;;;;'''';;;;'''' ;;;;``;;;;  ;;,.    ;;;    {{ .End }}\n" +
	"{{ .FgGreen }}     [[      [[cccc   [[[,/[[['  [[[[, ,[[[[,   {{ .End }}\n" +
	"{{ .FgGreen }}     $$      $$\"\"\"\"   $$$$$$c    $$$$$$$$\"$$$   {{ .End }}\n" +
	"{{ .FgGreen }}     88,     888oo,__ 888b \"88bo,888 Y88\" 888o  {{ .End }}\n" +
	"{{ .FgGreen }}     MMM     \"\"\"\"YUMMMMMMM   \"W\" MMM  M'  \"MMM  {{ .End }}\n" +
	"{{ .FgBlue }}                                    .,-:::::   ::   .:   :::.   :::.    :::. {{ .End }}\n" +
	"{{ .FgBlue }}                                  ,;;;'````'  ,;;   ;;,  ;;`;;  `;;;;,  `;;; {{ .End }}\n" +
	"{{ .FgBlue }}                                  [[[        ,[[[,,,[[[ ,[[ '[[,  [[[[[. '[[ {{ .End }}\n" +
	"{{ .FgBlue }}                                  $$$        \"$$$\"\"\"$$$c$$$cc$$$c $$$ \"Y$c$$ {{ .End }}\n" +
	"{{ .FgBlue }}                                  `88bo,__,o, 888   \"88o888   888,888    Y88 {{ .End }}\n" +
	"{{ .FgBlue }}                                    \"YUMMMMMP\"MMM    YMMYMM   \"\"` MMM     YM {{ .End }}\n" +
	"Welcome!\n" +
	"{{ .Separator.Double }}\n" +
	"Boards\n" +
	"{{ range .Boards }}  {{ . | formatBoard }}\n{{ end }}" +
	"{{ .Separator.Single }}\n" +
	"How do I use it?\n" +
	"{{ .Separator.Double }}\n" +
	"{{ .FgGreen }}Viewing{{ .End }}\n" +
	"{{ .Separator.Single }}\n" +
	"{{ .FgGreen }}View{{ .End }} a board (e.g. /g/)\n" +
	"  curl -s '{{ .Hostname }}/g'\n" +
	"{{ .Separator.Single }}\n" +
	"{{ .FgGreen }}View{{ .End }} a board as HTML (e.g. /m/)\n" +
	"  curl -s '{{ .Hostname }}/m?format=html'\n" +
	"{{ .Separator.Single }}\n" +
	"{{ .FgGreen }}View{{ .End }} a thread (e.g. thread #23 on /v/)\n" +
	"  curl -s '{{ .Hostname}}/v/23'\n" +
	"{{ .Separator.Single }}\n" +
	"{{ .FgGreen }}View{{ .End }} as JSON\n" +
	"  curl -s '{{ .Hostname }}/d/69?format=json'\n" +
	"{{ .Separator.Double }}\n" +
	"{{ .FgBlue }}Posting{{ .End }}\n" +
	"{{ .FgBlue }}Post{{ .End }} a reply to a thread ({{ .FgBlue }}*{{ .End }})\n" +
	"  curl -s '{{ .Hostname }}/g/42' \\\n" +
	"      --data-urlencode \"format=json\" \\\n" +
	"      --data-urlencode \"name=ilovebsd\" \\\n" +
	"      --data-urlencode \"content=Have you considered OpenBSD?\"\\\n" +
	"{{ .Separator.Single }}\n" +
	"{{ .FgBlue }}Post{{ .End }} (i.e. create) a thread ({{ .FgBlue }}*{{ .End }})\n" +
	"  curl -s '{{ .Hostname }}/b' \\\n" +
	"      --data-urlencode \"name=m00t\" \\\n" +
	"      --data-urlencode \"topic=Candlejack\" \\\n" +
	"      --data-urlencode \"content=I'm not afraid of him, what's he gon-\\\n" +
	"{{ .Separator.Single }}\n" +
	"({{ .FgBlue }}*{{ .End }}) fields other than content are optional, thread/board has to exist.\n" +
	"{{ .Separator.Double }}\n" +
	"{{ .FgGreen }}HAVE{{ .End }} {{ .FgBlue }}FUN{{ .End }}!\n"

const DefaultPost = `[{{ .ID }}] {{ .Author }} wrote at {{ .Timestamp | timeANSIC }}

{{ .Content }}
`

const DefaultThread = `/{{ .Board.Name | highlight }}/{{ .ID }} {{ .Topic }}
{{ .Separator.Double }}
{{ with $sep := .Separator.Single }}{{ range .Posts }}
{{ . | formatPost }}
{{ $sep }}{{ end }}{{ end }}
{{ .NumReplies }} {{ with $n := len .Posts }}{{ if eq $n 2}}reply{{ else }}replies{{ end }}{{ end }}
`

const DefaultBoard = `/{{ .Name | highlight }}/ - {{ .Descr | highlight }}
{{ $dsep := .Separator.Double }}{{ $ssep := .Separator.Single }}{{ $dsep }}
{{ range .Threads }}
/{{ .Board.Name | highlight }}/{{ .ID }} {{ .Topic }} ({{ .NumReplies }} {{ if eq 1 .NumReplies }}reply{{ else }}replies{{ end }}) updated {{ .Active | timeANSIC }}
{{ $ssep }}
{{ .OP | formatPost }}
{{ $dsep }}
{{ end }}{{ $n := len .Threads }}{{ $n }} {{ if eq $n 1 }}thread{{ else }}threads{{ end }}
`

const DefaultError = `{{ .Status }} {{ .FgRed }}ERROR{{ .End }}: {{ .Error }}
`
