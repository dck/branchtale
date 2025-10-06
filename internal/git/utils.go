package git

import (
	"fmt"
	"os/user"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

func getSSHAuth() (*ssh.PublicKeys, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	keyPath := filepath.Join(usr.HomeDir, ".ssh", "id_rsa")
	auth, err := ssh.NewPublicKeysFromFile("git", keyPath, "")
	if err != nil {
		return nil, fmt.Errorf("failed to load ssh key: %w", err)
	}
	return auth, nil
}
