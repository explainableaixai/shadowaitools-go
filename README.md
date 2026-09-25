# shadowaitools-go

A Go library, and a pattern for a small command-line tool, that reads a network log and reports the AI applications in it. Point it at a DNS query log, a proxy access log or a firewall CSV. It pulls out the hostnames, checks each one, and gives you back the ones that belong to AI products, with category and data-use terms attached.

It is the Go counterpart of the hosted [shadow AI discovery tools](https://www.shadowaitools.com/how-it-works.php), which work in a browser. Use this module when you want the same result inside your own tooling, on a schedule, or on logs that must not leave your servers.

```bash
go get github.com/explainableaixai/shadowaitools-go
```

## Two calls

```go
c := shadowaitools.New(os.Getenv("AQ_API_KEY"))

one, err := c.Check(ctx, "gemini.google.com")
all, err := c.Scan(ctx, "/var/log/resolver/queries.log")
```

`Check` looks up a single hostname. `Scan` does the whole file and returns `[]Result`, one entry per unique hostname.

## What Scan extracts

`Scan` reads the file into memory and splits it on commas, semicolons, spaces, tabs and line breaks. For every token it:

- adds `https://` if there is no scheme, then parses it as a URL
- keeps the lower-cased host if parsing worked, or the raw token otherwise
- drops tokens without a dot and tokens already seen

The result is a de-duplicated set of candidate hostnames, each looked up once. Because the split ignores column structure, the same code handles BIND query logs, Squid access logs, Pi-hole exports and most vendor CSVs.

Anything with a dot counts as a candidate, and that includes IP addresses and decimal numbers. They cost a lookup each and come back as "not an AI tool". On a wide CSV, extract the host column first with `cut`, `awk` or a few lines of Go, and scan that.

## A usable command in forty lines

```go
package main

import (
	"context"
	"encoding/csv"
	"flag"
	"log"
	"os"

	"github.com/explainableaixai/shadowaitools-go"
)

func main() {
	in := flag.String("in", "", "log file to scan")
	flag.Parse()

	c := shadowaitools.New(os.Getenv("AQ_API_KEY"))
	rows, err := c.Scan(context.Background(), *in)
	if err != nil {
		log.Fatal(err)
	}

	w := csv.NewWriter(os.Stdout)
	w.Write([]string{"domain", "category", "trains_on_data"})
	for _, r := range rows {
		if hit, _ := r["blocked"].(bool); !hit {
			continue
		}
		cat, _ := r["primary_category"].(string)
		tr, _ := r["trains_on_data"].(string)
		dom, _ := r["domain"].(string)
		w.Write([]string{dom, cat, tr})
	}
	w.Flush()
}
```

Build it once with `go build -o ai-inventory`, then run `./ai-inventory -in queries.log > inventory.csv`. The CSV opens in any spreadsheet and is a sensible first artefact for an AI governance meeting.

## Result fields worth keeping

For an AI tool, a result carries `blocked: true`, `primary_category`, `ai_type`, a `categories` list, and four data-use fields: `trains_on_data`, `opt_out_available`, `enterprise_no_training` and `api_no_training`. `terms_checked` dates the review. A value of `unstated` means the vendor terms do not address the question, which most compliance teams treat as a risk rather than a pass.

## Privacy by construction

Only hostnames leave the machine. Timestamps, client IPs, usernames and every other column stay where they are, because the library never sends anything except the host in a lookup. That makes it practical to run against logs covered by internal data rules. For the report itself, join the findings back to your log locally to see which users or subnets reached each tool.

## Speed and quotas

Lookups run one at a time. That is intentional: it keeps load on the service steady and makes quota use predictable, one call per unique host. A day of DNS logs from a mid-sized office usually has a few thousand unique hosts and finishes in a few minutes.

If you scan daily, keep a local record of hosts already checked and scan only new ones. Most of any network's traffic goes to the same hosts every day, so the second run is much shorter.

## When something fails

`Scan` stops at the first failed lookup and returns that error with no partial results. For long files where you would rather skip failures, write the loop yourself:

```go
for _, h := range hosts {
	r, err := c.Check(ctx, h)
	var apiErr *shadowaitools.APIError
	if errors.As(err, &apiErr) && apiErr.Status == 429 {
		time.Sleep(30 * time.Second)
		continue
	}
	if err != nil {
		log.Printf("skip %s: %v", h, err)
		continue
	}
	results = append(results, r)
}
```

An `*APIError` carries `Status` and `Body`. Statuses 401 and 403 mean the key is wrong or the month's quota is used, and retrying will not help. Missing files, bad JSON and cancelled contexts come back as ordinary errors.

## Scheduling it

A weekly systemd timer or cron job running the command above, writing to a dated file, gives you a history. Comparing two weeks shows new tools as they arrive, and those are usually the ones worth a conversation.

## Frameworks this supports

An AI inventory is an early step in most governance frameworks. The EU AI Act expects deployers to know which AI systems they use. ISO/IEC 42001 asks for an inventory as part of an AI management system. The NIST AI RMF places it under the Map function. A dated CSV from this tool gives an auditor something concrete to review.

## The rest of the stack

- Lookups use the same register that drives [AI data leakage prevention](https://www.aitoolsblocklist.com/prevent-data-leakage-ai.php) policies.
- Hosts that turn out not to be AI can be labelled with [check domain categorization](https://www.urlcategorizationdatabase.com/check-domain.php) lookups.
- If you run AI agents yourself, put an [AI agent allow list as a guardrail](https://www.aiagentallowlist.com/agent-guardrails.php) in front of their browsing tool.

Prefer a scripting language? [shadowaitools for Python](https://pypi.org/project/shadowaitools/) fits notebooks, [the npm build](https://www.npmjs.com/package/shadowaitools) fits Node tooling, and [the Dart package](https://pub.dev/packages/shadowaitools) runs on desktop and servers.

## License

MIT
