package copyright

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"sourcegraph.com/sourcegraph/go-diff/diff"

	"github.com/mattmoor/knobots/pkg/botinfo"
	"github.com/mattmoor/knobots/pkg/comment"
	"github.com/mattmoor/knobots/pkg/handler"
	"github.com/mattmoor/knobots/pkg/reviewrequest"
	"github.com/mattmoor/knobots/pkg/reviewresult"
	"github.com/mattmoor/knobots/pkg/visitor"
)

var (
	re = regexp.MustCompile("Copyright \\d{4} [a-zA-Z0-9]+")
)

type copyright struct{}

var _ handler.Interface = (*copyright)(nil)

func New() handler.Interface {
	return &copyright{}
}

func (*copyright) GetType() interface{} {
	return &reviewrequest.Response{}
}

func updateCopyrightYear(orig string) string {
	if !re.MatchString(orig) {
		return orig
	}

	return string(re.ReplaceAllFunc([]byte(orig), func(in []byte) []byte {
		before := string(in)
		return []byte(fmt.Sprintf("Copyright %d%s",
			time.Now().Year(), before[len("Copyright 2018"):]))
	}))
}

func (*copyright) Handle(x interface{}) (handler.Response, error) {
	rrr := x.(*reviewrequest.Response)

	var comments []*github.DraftReviewComment
	err := visitor.Hunks(rrr.Owner, rrr.Repository, rrr.PullRequest,
		func(path string, hs []*diff.Hunk) (visitor.VisitControl, error) {
			// TODO(mattmoor): Base this on .gitattributes (we should build a library).
			if strings.HasPrefix(path, "vendor/") {
				return visitor.Continue, nil
			}
			// Each hunk header @@ takes a line.
			// For subsequent hunks, this is covered by the trailing `\n`
			// in each hunk, but the first needs to start at offset 1.
			offset := 1
			for _, hunk := range hs {
				lines := strings.Split(string(hunk.Body), "\n")
				for _, line := range lines {
					// Increase our offset for each line we see.
					if strings.HasPrefix(line, "+") {
						orig := line[1:]
						updated := updateCopyrightYear(orig)
						if updated != orig {
							position := offset // Copy it because of &.
							comments = append(comments, &github.DraftReviewComment{
								Path:     &path,
								Position: &position,
								Body:     comment.WithSuggestion(updated),
							})
						}
					}
					// Increase our offset for each line we see.
					offset++
				}
			}

			return visitor.Continue, nil
		})
	if err != nil {
		return nil, err
	}

	return &reviewresult.Payload{
		Name:        botinfo.GetName(),
		Description: `Check for incorrect year in Copyright headers.`,
		Owner:       rrr.Owner,
		Repository:  rrr.Repository,
		PullRequest: rrr.PullRequest,
		SHA:         rrr.SHA,
		Comments:    comments,
	}, nil
}