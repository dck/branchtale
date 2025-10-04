package pr

type State struct {
	CreatedBranch          bool
	BranchName             string
	CreatePullRequest      bool
	PullRequestTitle       string
	PullRequestDescription string
	PullRequestTags        []string
}

func Run(state *State) error {
	if state.CreatedBranch {
		// Logic for handling created branch
	}

	if state.CreatePullRequest {
		// Logic for creating pull request
	}

	return nil
}
