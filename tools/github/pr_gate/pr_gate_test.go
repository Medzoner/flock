package pr_gate

import (
	"testing"

	"github.com/google/go-github/v93/github"
)

func TestRunReopensOnlyGateClosedPullRequests(t *testing.T) {
	tests := []struct {
		name       string
		gateClosed bool
		wantState  string
	}{
		{name: "reopens gate-closed pull request", gateClosed: true, wantState: "open"},
		{name: "does not reopen manually closed pull request", gateClosed: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &fakeService{
				pullRequest: pullRequest("closed", "Fixes #73"),
				gateClosed:  tt.gateClosed,
				issue:       approvedIssue(),
			}
			result, err := Run(api, "flock-php", "flock", Event{Action: "edited", Number: 87, BodyEdited: true})
			if err != nil {
				t.Fatal(err)
			}
			if api.state != tt.wantState {
				t.Fatalf("state = %q, want %q", api.state, tt.wantState)
			}
			if tt.gateClosed && (result.ApprovedIssue == nil || result.ApprovedIssue.Number != 73) {
				t.Fatalf("approved issue = %+v, want flock#73", result.ApprovedIssue)
			}
		})
	}
}

func TestRunClosesUnapprovedPullRequest(t *testing.T) {
	api := &fakeService{pullRequest: pullRequest("open", "Fixes #73"), files: []*github.CommitFile{{Filename: github.String("internal/site.go")}}, issue: &github.Issue{State: github.String("open")}}
	if _, err := Run(api, "flock-php", "flock", Event{Action: "synchronize", Number: 87}); err != nil {
		t.Fatal(err)
	}
	if api.state != "closed" || !api.commented {
		t.Fatalf("state = %q, commented = %t", api.state, api.commented)
	}
}

type fakeService struct {
	pullRequest *github.PullRequest
	files       []*github.CommitFile
	issue       *github.Issue
	gateClosed  bool
	state       string
	commented   bool
}

func (f *fakeService) PullRequest(string, string, int) (*github.PullRequest, error) {
	return f.pullRequest, nil
}
func (f *fakeService) Permission(string, string, string) (string, error)       { return "read", nil }
func (f *fakeService) Files(string, string, int) ([]*github.CommitFile, error) { return f.files, nil }
func (f *fakeService) Issue(string, string, int) (*github.Issue, error)        { return f.issue, nil }
func (f *fakeService) GateClosed(string, string, int) (bool, error)            { return f.gateClosed, nil }
func (f *fakeService) Comment(string, string, int, string) error {
	f.commented = true
	return nil
}
func (f *fakeService) SetPullRequestState(_ string, _ string, _ int, state string) error {
	f.state = state
	return nil
}

func pullRequest(state, body string) *github.PullRequest {
	return &github.PullRequest{
		Number: github.Int(87),
		State:  github.String(state),
		Base:   &github.PullRequestBranch{Ref: github.String("main")},
		Body:   github.String(body),
		User:   &github.User{Login: github.String("contributor"), Type: github.String("User")},
	}
}

func approvedIssue() *github.Issue {
	return &github.Issue{State: github.String("open"), Labels: []*github.Label{{Name: openToAnyone}}}
}
