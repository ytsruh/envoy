package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	gitDirName    = ".git"
	gitConfigFile = "config"
	remoteName    = "remote"
	originName    = "origin"
	urlName       = "url"
	sectionHeader = "["
	sectionFooter = "]"
	gitHubPrefix  = "github.com:"
	gitLabPrefix  = "gitlab.com:"
	gitSuffix     = ".git"
)

type GitInfo struct {
	HasGit    bool
	RemoteURL string
	RepoName  string
	Owner     string
	IsGitHub  bool
	IsGitLab  bool
}

func DetectGitInfo() (*GitInfo, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	gitDir, err := findGitDir(currentDir)
	if err != nil {
		return &GitInfo{HasGit: false}, nil
	}

	gitConfigPath := filepath.Join(gitDir, gitConfigFile)
	remoteURL, err := parseGitConfig(gitConfigPath)
	if err != nil {
		return &GitInfo{HasGit: true}, nil
	}

	return parseRemoteURL(remoteURL), nil
}

func findGitDir(startDir string) (string, error) {
	currentDir := startDir

	for {
		gitPath := filepath.Join(currentDir, gitDirName)
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return gitPath, nil
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return "", fmt.Errorf("git repository not found")
		}
		currentDir = parentDir
	}
}

func parseGitConfig(configPath string) (string, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to open git config: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentRemote string
	var remoteURL string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, sectionHeader) && strings.Contains(line, sectionFooter) {
			header := line[1 : len(line)-1]
			parts := strings.Fields(header)

			if len(parts) == 2 && parts[0] == remoteName {
				currentRemote = parts[1]
				remoteURL = ""
			} else {
				currentRemote = ""
			}
			continue
		}

		if currentRemote == originName && strings.HasPrefix(line, urlName) {
			urlParts := strings.SplitN(line, "=", 2)
			if len(urlParts) == 2 {
				remoteURL = strings.TrimSpace(urlParts[1])
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading git config: %w", err)
	}

	if remoteURL == "" {
		return "", fmt.Errorf("no origin URL found in git config")
	}

	return remoteURL, nil
}

func parseRemoteURL(remoteURL string) *GitInfo {
	info := &GitInfo{
		HasGit:    true,
		RemoteURL: remoteURL,
	}

	cleanURL := strings.TrimSuffix(remoteURL, gitSuffix)

	if strings.Contains(cleanURL, gitHubPrefix) {
		info.IsGitHub = true
		parts := strings.Split(cleanURL, gitHubPrefix)
		if len(parts) == 2 {
			repoParts := strings.Split(parts[1], "/")
			if len(repoParts) >= 2 {
				info.Owner = repoParts[0]
				info.RepoName = repoParts[1]
			}
		}
	} else if strings.Contains(cleanURL, gitLabPrefix) {
		info.IsGitLab = true
		parts := strings.Split(cleanURL, gitLabPrefix)
		if len(parts) == 2 {
			repoParts := strings.Split(parts[1], "/")
			if len(repoParts) >= 2 {
				info.Owner = repoParts[0]
				info.RepoName = repoParts[1]
			}
		}
	}

	return info
}

func GetGitRepoString() (string, error) {
	gitInfo, err := DetectGitInfo()
	if err != nil {
		return "", err
	}

	if !gitInfo.HasGit {
		return "", nil
	}

	if gitInfo.Owner != "" && gitInfo.RepoName != "" {
		return fmt.Sprintf("%s/%s", gitInfo.Owner, gitInfo.RepoName), nil
	}

	return "", fmt.Errorf("could not determine repo name from git config")
}

func GetFullGitRepoURL() (string, error) {
	gitInfo, err := DetectGitInfo()
	if err != nil {
		return "", err
	}

	if !gitInfo.HasGit {
		return "", nil
	}

	return gitInfo.RemoteURL, nil
}
